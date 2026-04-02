// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package auth

const (
	PathDeviceAuthorization     = "/oauth/v1/device_authorization"
	PathAppRegistration         = "/oauth/v1/app/registration"
	PathOAuthTokenV2            = "/open-apis/authen/v2/oauth/token"
	PathUserInfoV1              = "/open-apis/authen/v1/user_info"
	PathApplicationInfoV6Prefix = "/open-apis/application/v6/applications/"
)

func ApplicationInfoPath(appId string) string {
	return PathApplicationInfoV6Prefix + appId
}
