package core

import (
	"encoding/json"
	"testing"
)

func TestMultiAppConfig_StrictMode_JSON(t *testing.T) {
	// StrictMode=false should be omitted (omitempty)
	m := &MultiAppConfig{
		Apps: []AppConfig{{AppId: "a", AppSecret: PlainSecret("s"), Brand: BrandFeishu, Users: []AppUser{}}},
	}
	data, _ := json.Marshal(m)
	if string(data) != `{"apps":[{"appId":"a","appSecret":"s","brand":"feishu","users":[]}]}` {
		t.Errorf("StrictMode=false should be omitted, got: %s", data)
	}

	// StrictMode=true should be present
	m.StrictMode = true
	data, _ = json.Marshal(m)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)
	if parsed["strictMode"] != true {
		t.Errorf("StrictMode=true should be present, got: %s", data)
	}
}

func TestAppConfig_StrictMode_JSON(t *testing.T) {
	// StrictMode nil should be omitted
	app := &AppConfig{AppId: "a", AppSecret: PlainSecret("s"), Brand: BrandFeishu, Users: []AppUser{}}
	data, _ := json.Marshal(app)
	var parsed map[string]interface{}
	json.Unmarshal(data, &parsed)
	if _, ok := parsed["strictMode"]; ok {
		t.Errorf("nil StrictMode should be omitted, got: %s", data)
	}

	// StrictMode = pointer to true
	v := true
	app.StrictMode = &v
	data, _ = json.Marshal(app)
	json.Unmarshal(data, &parsed)
	if parsed["strictMode"] != true {
		t.Errorf("StrictMode=true should be present, got: %s", data)
	}

	// StrictMode = pointer to false (should be present, not omitted)
	vf := false
	app.StrictMode = &vf
	data, _ = json.Marshal(app)
	json.Unmarshal(data, &parsed)
	// Note: omitempty with *bool omits nil but keeps false — verify
	if val, ok := parsed["strictMode"]; !ok || val != false {
		t.Errorf("StrictMode=false (explicit) should be present, got: %s", data)
	}
}
