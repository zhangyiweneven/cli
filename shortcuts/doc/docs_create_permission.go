// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package doc

import (
	"fmt"
	"strings"

	"github.com/larksuite/cli/internal/validate"
	"github.com/larksuite/cli/shortcuts/common"
)

const (
	docsCreatePermissionGranted = "granted"
	docsCreatePermissionSkipped = "skipped"
	docsCreatePermissionFailed  = "failed"
)

type docsPermissionTarget struct {
	Token    string
	Type     string
	PermType string
}

func augmentDocsCreateResult(runtime *common.RuntimeContext, result map[string]interface{}) {
	if !runtime.IsBot() {
		return
	}
	result["permission_grant"] = autoGrantCurrentUserDocPermission(runtime, result)
}

func autoGrantCurrentUserDocPermission(runtime *common.RuntimeContext, result map[string]interface{}) map[string]interface{} {
	userOpenID := strings.TrimSpace(runtime.UserOpenId())
	if userOpenID == "" {
		return buildDocsCreatePermissionGrantResult(
			docsCreatePermissionSkipped,
			"",
			"",
			"Document was created with bot identity, but no current CLI user open_id is configured, so user permission was not granted. You can retry later or continue using bot identity.",
		)
	}

	target := selectDocsPermissionTarget(result)
	if target.Token == "" || target.Type == "" {
		return buildDocsCreatePermissionGrantResult(
			docsCreatePermissionSkipped,
			userOpenID,
			"",
			"Document was created, but no permission target was returned (missing doc_id/doc_url), so user permission was not granted. You can retry later or continue using bot identity.",
		)
	}

	body := map[string]interface{}{
		"member_type": "openid",
		"member_id":   userOpenID,
		"perm":        "full_access",
		"type":        "user",
	}
	if target.PermType != "" {
		body["perm_type"] = target.PermType
	}

	_, err := runtime.CallAPI(
		"POST",
		fmt.Sprintf("/open-apis/drive/v1/permissions/%s/members", validate.EncodePathSegment(target.Token)),
		map[string]interface{}{
			"type":              target.Type,
			"need_notification": false,
		},
		body,
	)
	if err != nil {
		return buildDocsCreatePermissionGrantResult(
			docsCreatePermissionFailed,
			userOpenID,
			target.Type,
			fmt.Sprintf("Document was created, but granting current user full_access failed: %s. You can retry later or continue using bot identity.", compactDocsCreateError(err)),
		)
	}

	return buildDocsCreatePermissionGrantResult(
		docsCreatePermissionGranted,
		userOpenID,
		target.Type,
		fmt.Sprintf("Granted the current CLI user full_access on the new %s.", docsPermissionTargetLabel(target.Type)),
	)
}

func buildDocsCreatePermissionGrantResult(status, userOpenID, resourceType, message string) map[string]interface{} {
	result := map[string]interface{}{
		"status":  status,
		"perm":    "full_access",
		"message": message,
	}
	if userOpenID != "" {
		result["user_open_id"] = userOpenID
		result["member_type"] = "openid"
	}
	if resourceType != "" {
		result["resource_type"] = resourceType
	}
	return result
}

func selectDocsPermissionTarget(result map[string]interface{}) docsPermissionTarget {
	if ref, ok := parseDocsPermissionTargetFromURL(common.GetString(result, "doc_url")); ok {
		return ref
	}

	docID := strings.TrimSpace(common.GetString(result, "doc_id"))
	if docID != "" {
		return docsPermissionTarget{Token: docID, Type: "docx"}
	}
	return docsPermissionTarget{}
}

func parseDocsPermissionTargetFromURL(docURL string) (docsPermissionTarget, bool) {
	if strings.TrimSpace(docURL) == "" {
		return docsPermissionTarget{}, false
	}

	ref, err := parseDocumentRef(docURL)
	if err != nil {
		return docsPermissionTarget{}, false
	}

	switch ref.Kind {
	case "wiki":
		return docsPermissionTarget{Token: ref.Token, Type: "wiki", PermType: "container"}, true
	case "doc", "docx":
		return docsPermissionTarget{Token: ref.Token, Type: ref.Kind}, true
	default:
		return docsPermissionTarget{}, false
	}
}

func docsPermissionTargetLabel(resourceType string) string {
	switch resourceType {
	case "wiki":
		return "wiki node"
	case "doc":
		return "document"
	case "docx":
		return "docx document"
	default:
		return "document"
	}
}

func compactDocsCreateError(err error) string {
	if err == nil {
		return ""
	}
	return strings.Join(strings.Fields(err.Error()), " ")
}
