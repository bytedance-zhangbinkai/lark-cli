// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package base

import (
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/larksuite/cli/errs"
	"github.com/larksuite/cli/internal/httpmock"
	"github.com/larksuite/cli/shortcuts/common"
)

func TestBaseButtonRuleScopesDoNotRequireWorkflowAccess(t *testing.T) {
	tests := []struct {
		name       string
		shortcut   common.Shortcut
		wantScopes []string
	}{
		{name: "bind", shortcut: BaseButtonRuleBind, wantScopes: []string{"base:field:read", "base:field:update"}},
		{name: "get", shortcut: BaseButtonRuleGet, wantScopes: []string{"base:field:read"}},
		{name: "unbind", shortcut: BaseButtonRuleUnbind, wantScopes: []string{"base:field:read", "base:field:update"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !slices.Equal(tt.shortcut.Scopes, tt.wantScopes) {
				t.Fatalf("Scopes=%v want=%v", tt.shortcut.Scopes, tt.wantScopes)
			}
		})
	}
}

func TestBaseWorkflowExecuteGet(t *testing.T) {
	factory, stdout, reg := newExecuteFactory(t)
	reg.Register(&httpmock.Stub{
		Method: "GET",
		URL:    "/open-apis/base/v3/bases/app_x/workflows/wkf_1",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"workflow_id": "wkf_1", "title": "My Workflow"},
		},
	})
	if err := runShortcut(t, BaseWorkflowGet, []string{"+workflow-get", "--base-token", "app_x", "--workflow-id", "wkf_1"}, factory, stdout); err != nil {
		t.Fatalf("err=%v", err)
	}
	if got := stdout.String(); !strings.Contains(got, `"wkf_1"`) || !strings.Contains(got, `"My Workflow"`) {
		t.Fatalf("stdout=%s", got)
	}
}

func TestBaseWorkflowExecuteGetWithUserIDType(t *testing.T) {
	factory, stdout, reg := newExecuteFactory(t)
	reg.Register(&httpmock.Stub{
		Method: "GET",
		URL:    "user_id_type=open_id",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"workflow_id": "wkf_1", "creator": map[string]interface{}{"open_id": "ou_abc"}},
		},
	})
	if err := runShortcut(t, BaseWorkflowGet, []string{"+workflow-get", "--base-token", "app_x", "--workflow-id", "wkf_1", "--user-id-type", "open_id"}, factory, stdout); err != nil {
		t.Fatalf("err=%v", err)
	}
	if got := stdout.String(); !strings.Contains(got, `"ou_abc"`) {
		t.Fatalf("stdout=%s", got)
	}
}

func TestBaseWorkflowExecuteGetValidate(t *testing.T) {
	t.Run("missing base-token", func(t *testing.T) {
		factory, stdout, _ := newExecuteFactory(t)
		err := runShortcut(t, BaseWorkflowGet, []string{"+workflow-get", "--workflow-id", "wkf_1"}, factory, stdout)
		if err == nil || !strings.Contains(err.Error(), "base-token") {
			t.Fatalf("err=%v", err)
		}
	})
	t.Run("missing workflow-id", func(t *testing.T) {
		factory, stdout, _ := newExecuteFactory(t)
		err := runShortcut(t, BaseWorkflowGet, []string{"+workflow-get", "--base-token", "app_x"}, factory, stdout)
		if err == nil || !strings.Contains(err.Error(), "workflow-id") {
			t.Fatalf("err=%v", err)
		}
	})
}

func TestBaseWorkflowExecuteCreate(t *testing.T) {
	factory, stdout, reg := newExecuteFactory(t)
	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/open-apis/base/v3/bases/app_x/workflows",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"workflow_id": "wkf_new", "title": "My Workflow"},
		},
	})
	if err := runShortcut(t, BaseWorkflowCreate, []string{"+workflow-create", "--base-token", "app_x", "--json", `{"title":"My Workflow","steps":[]}`}, factory, stdout); err != nil {
		t.Fatalf("err=%v", err)
	}
	if got := stdout.String(); !strings.Contains(got, `"wkf_new"`) {
		t.Fatalf("stdout=%s", got)
	}
}

func TestBaseWorkflowExecuteCreatePreservesAIClassificationBranch(t *testing.T) {
	factory, stdout, reg := newExecuteFactory(t)
	stub := &httpmock.Stub{
		Method: "POST",
		URL:    "/open-apis/base/v3/bases/app_x/workflows",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"workflow_id": "wkf_ai", "title": "Feedback classify"},
		},
	}
	reg.Register(stub)

	body := `{
		"title": "Feedback classify",
		"steps": [
			{"id": "step_trigger", "type": "AddRecordTrigger", "next": "step_classify", "data": {"table_name": "Feedback"}},
			{
				"id": "step_classify",
				"type": "AIClassificationBranch",
				"next": null,
				"children": {"links": [
					{"kind": "case", "to": "step_bug", "label": "branch_1", "desc": "Bug"},
					{"kind": "case", "to": "step_feature", "label": "branch_2", "desc": "Feature"},
					{"kind": "case", "to": "step_other", "label": "other", "desc": "Other"}
				]},
				"data": {
					"mode": "Exclusive",
					"prompt": [
						{"value_type": "text", "value": "Classify feedback: "},
						{"value_type": "ref", "value": "$.step_trigger.fldFeedback"}
					],
					"childBranchList": [
						{"name": [{"value_type": "text", "value": "Bug"}], "description": [{"value_type": "text", "value": "Broken behavior"}], "entryChildStepId": "step_bug"},
						{"name": [{"value_type": "text", "value": "Feature"}], "description": [{"value_type": "text", "value": "New capability"}], "entryChildStepId": "step_feature"}
					],
					"defaultBranchInfo": {"mode": "Execute", "entryStepId": "step_other"},
					"classifyPrompt": [{"value_type": "text", "value": "Use Other when unsure."}],
					"future_server_field": {"keep": true}
				}
			},
			{"id": "step_bug", "type": "SetRecordAction", "next": null, "data": {"unknown": true}},
			{"id": "step_feature", "type": "SetRecordAction", "next": null, "data": {}},
			{"id": "step_other", "type": "LarkMessageAction", "next": null, "data": {}}
		]
	}`
	if err := runShortcut(t, BaseWorkflowCreate, []string{"+workflow-create", "--base-token", "app_x", "--json", body}, factory, stdout); err != nil {
		t.Fatalf("err=%v", err)
	}
	if got := string(stub.CapturedBody); !strings.Contains(got, `"type":"AIClassificationBranch"`) || !strings.Contains(got, `"future_server_field":{"keep":true}`) {
		t.Fatalf("AI classification payload was not forwarded verbatim enough: %s", got)
	}
}

func TestBaseWorkflowExecuteCreateValidate(t *testing.T) {
	t.Run("missing base-token", func(t *testing.T) {
		factory, stdout, _ := newExecuteFactory(t)
		err := runShortcut(t, BaseWorkflowCreate, []string{"+workflow-create", "--json", `{"title":"x"}`}, factory, stdout)
		if err == nil || !strings.Contains(err.Error(), "base-token") {
			t.Fatalf("err=%v", err)
		}
	})
	t.Run("invalid json", func(t *testing.T) {
		factory, stdout, _ := newExecuteFactory(t)
		err := runShortcut(t, BaseWorkflowCreate, []string{"+workflow-create", "--base-token", "app_x", "--json", `not-json`}, factory, stdout)
		if err == nil {
			t.Fatalf("expected error for invalid json")
		}
	})
}

func TestBaseWorkflowExecuteValidateAIClassificationBranch(t *testing.T) {
	base := func(data string, children string) string {
		return `{
			"title": "Feedback classify",
			"steps": [
				{"id": "step_trigger", "type": "AddRecordTrigger", "next": "step_classify", "data": {}},
				{"id": "step_classify", "type": "AIClassificationBranch", "children": ` + children + `, "data": ` + data + `},
				{"id": "step_bug", "type": "SetRecordAction", "next": null, "data": {}},
				{"id": "step_feature", "type": "SetRecordAction", "next": null, "data": {}}
			]
		}`
	}
	validChildren := `{"links":[{"kind":"case","to":"step_bug"},{"kind":"case","to":"step_feature"}]}`
	validData := `{
		"mode": "Exclusive",
		"prompt": [{"value_type": "text", "value": "Classify: "}, {"value_type": "ref", "value": "$.step_trigger.fldFeedback"}],
		"child_branch_list": [
			{"name": "Bug", "description": "Broken behavior", "entry_child_step_id": "step_bug"},
			{"name": "Feature", "description": "New capability", "entry_child_step_id": "step_feature"}
		],
		"no_match_action": "fail",
		"classify_prompt": [{"value_type": "text", "value": "Use the closest category."}]
	}`

	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "blank prompt",
			body: base(strings.Replace(validData, `"prompt": [{"value_type": "text", "value": "Classify: "}, {"value_type": "ref", "value": "$.step_trigger.fldFeedback"}]`, `"prompt": [{"value_type": "text", "value": "   "}]`, 1), validChildren),
			want: "data.prompt must contain non-empty text or a valid ref",
		},
		{
			name: "duplicate classification name",
			body: base(strings.Replace(validData, `"Feature"`, `"Bug"`, 1), validChildren),
			want: "duplicates",
		},
		{
			name: "unknown child entry",
			body: base(strings.Replace(validData, `"step_feature"`, `"step_missing"`, 1), validChildren),
			want: "references unknown step id",
		},
		{
			name: "classify prompt ref",
			body: base(strings.Replace(validData, `"classify_prompt": [{"value_type": "text", "value": "Use the closest category."}]`, `"classify_prompt": [{"value_type": "ref", "value": "$.step_trigger.fldFeedback"}]`, 1), validChildren),
			want: "ref is not supported here",
		},
		{
			name: "non case link",
			body: base(validData, `{"links":[{"kind":"if_true","to":"step_bug"}]}`),
			want: "kind must be case",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory, stdout, _ := newExecuteFactory(t)
			err := runShortcut(t, BaseWorkflowCreate, []string{"+workflow-create", "--base-token", "app_x", "--json", tt.body}, factory, stdout)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("err=%v want substring %q", err, tt.want)
			}
			var validationErr *errs.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("err type=%T want *errs.ValidationError", err)
			}
		})
	}
}

func TestBaseWorkflowExecuteUpdateValidatesAIClassificationBranch(t *testing.T) {
	factory, stdout, _ := newExecuteFactory(t)
	body := `{
		"title": "Feedback classify",
		"steps": [
			{"id": "step_trigger", "type": "AddRecordTrigger", "next": "step_classify", "data": {}},
			{
				"id": "step_classify",
				"type": "AIClassificationBranch",
				"children": {"links":[{"kind":"case","to":"step_bug"},{"kind":"case","to":"step_feature"}]},
				"data": {
					"mode": "NotAMode",
					"prompt": [{"value_type": "text", "value": "Classify"}],
					"childBranchList": [
						{"name": "Bug", "entryChildStepId": "step_bug"},
						{"name": "Feature", "entryChildStepId": "step_feature"}
					]
				}
			},
			{"id": "step_bug", "type": "SetRecordAction", "next": null, "data": {}},
			{"id": "step_feature", "type": "SetRecordAction", "next": null, "data": {}}
		]
	}`
	err := runShortcut(t, BaseWorkflowUpdate, []string{"+workflow-update", "--base-token", "app_x", "--workflow-id", "wkf_1", "--json", body}, factory, stdout)
	if err == nil || !strings.Contains(err.Error(), "data.mode must be Exclusive or Parallel") {
		t.Fatalf("err=%v", err)
	}
}

func TestBaseWorkflowExecuteDisable(t *testing.T) {
	factory, stdout, reg := newExecuteFactory(t)
	reg.Register(&httpmock.Stub{
		Method: "PATCH",
		URL:    "/open-apis/base/v3/bases/app_x/workflows/wkf_1/disable",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"workflow_id": "wkf_1", "status": "disabled"},
		},
	})
	if err := runShortcut(t, BaseWorkflowDisable, []string{"+workflow-disable", "--base-token", "app_x", "--workflow-id", "wkf_1"}, factory, stdout); err != nil {
		t.Fatalf("err=%v", err)
	}
	if got := stdout.String(); !strings.Contains(got, `"disabled"`) {
		t.Fatalf("stdout=%s", got)
	}
}

func TestBaseWorkflowExecuteDisableValidate(t *testing.T) {
	t.Run("missing base-token", func(t *testing.T) {
		factory, stdout, _ := newExecuteFactory(t)
		err := runShortcut(t, BaseWorkflowDisable, []string{"+workflow-disable", "--workflow-id", "wkf_1"}, factory, stdout)
		if err == nil || !strings.Contains(err.Error(), "base-token") {
			t.Fatalf("err=%v", err)
		}
	})
	t.Run("missing workflow-id", func(t *testing.T) {
		factory, stdout, _ := newExecuteFactory(t)
		err := runShortcut(t, BaseWorkflowDisable, []string{"+workflow-disable", "--base-token", "app_x"}, factory, stdout)
		if err == nil || !strings.Contains(err.Error(), "workflow-id") {
			t.Fatalf("err=%v", err)
		}
	})
}

func TestBaseButtonRuleExecuteResolvesFieldReference(t *testing.T) {
	tests := []struct {
		name              string
		shortcut          common.Shortcut
		args              []string
		fieldRef          string
		canonicalFieldID  string
		fieldIdentityKey  string
		buttonRuleMethod  string
		wantWorkflowID    string
		wantWorkflowField bool
	}{
		{
			name: "bind by name", shortcut: BaseButtonRuleBind,
			args:     []string{"+button-rule-bind", "--base-token", "app_x", "--table-id", "tbl_1", "--field-id", "按钮", "--workflow-id", "wkf_1"},
			fieldRef: "按钮", canonicalFieldID: "fld_bind", fieldIdentityKey: "id",
			buttonRuleMethod: "PUT", wantWorkflowID: "wkf_1", wantWorkflowField: true,
		},
		{
			name: "get by name", shortcut: BaseButtonRuleGet,
			args:     []string{"+button-rule-get", "--base-token", "app_x", "--table-id", "tbl_1", "--field-id", "按钮"},
			fieldRef: "按钮", canonicalFieldID: "fld_get", fieldIdentityKey: "id",
			buttonRuleMethod: "GET",
		},
		{
			name: "unbind by name with field_id compatibility", shortcut: BaseButtonRuleUnbind,
			args:     []string{"+button-rule-unbind", "--base-token", "app_x", "--table-id", "tbl_1", "--field-id", "按钮"},
			fieldRef: "按钮", canonicalFieldID: "fld_unbind", fieldIdentityKey: "field_id",
			buttonRuleMethod: "PUT", wantWorkflowID: "", wantWorkflowField: true,
		},
		{
			name: "ID input is still resolved", shortcut: BaseButtonRuleGet,
			args:     []string{"+button-rule-get", "--base-token", "app_x", "--table-id", "tbl_1", "--field-id", "fld_input"},
			fieldRef: "fld_input", canonicalFieldID: "fld_canonical", fieldIdentityKey: "id",
			buttonRuleMethod: "GET",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factory, stdout, reg := newExecuteFactory(t)
			callOrder := 0
			reg.Register(&httpmock.Stub{
				Method: "GET",
				URL:    baseV3Path("bases", "app_x", "tables", "tbl_1", "fields", tt.fieldRef),
				Body: map[string]interface{}{
					"code": 0,
					"data": map[string]interface{}{tt.fieldIdentityKey: tt.canonicalFieldID, "name": tt.fieldRef},
				},
				OnMatch: func(_ *http.Request) {
					if callOrder != 0 {
						t.Fatalf("field resolution call order=%d want=0", callOrder)
					}
					callOrder++
				},
			})
			buttonRuleStub := &httpmock.Stub{
				Method: tt.buttonRuleMethod,
				URL:    baseV3Path("bases", "app_x", "tables", "tbl_1", "fields", tt.canonicalFieldID, "button_rule"),
				Body: map[string]interface{}{
					"code": 0,
					"data": map[string]interface{}{"table_id": "tbl_1", "field_id": tt.canonicalFieldID, "workflow_id": tt.wantWorkflowID, "bound": tt.wantWorkflowID != ""},
				},
				OnMatch: func(_ *http.Request) {
					if callOrder != 1 {
						t.Fatalf("ButtonRule call order=%d want=1", callOrder)
					}
					callOrder++
				},
			}
			reg.Register(buttonRuleStub)

			if err := runShortcut(t, tt.shortcut, tt.args, factory, stdout); err != nil {
				t.Fatalf("err=%v", err)
			}
			if callOrder != 2 {
				t.Fatalf("call order count=%d want=2", callOrder)
			}
			if got := stdout.String(); !strings.Contains(got, `"field_id": "`+tt.canonicalFieldID+`"`) {
				t.Fatalf("stdout=%s", got)
			}
			if tt.wantWorkflowField {
				var body map[string]interface{}
				if err := json.Unmarshal(buttonRuleStub.CapturedBody, &body); err != nil {
					t.Fatalf("decode ButtonRule body: %v", err)
				}
				if got, ok := body["workflow_id"].(string); !ok || got != tt.wantWorkflowID {
					t.Fatalf("workflow_id=%#v want=%q body=%s", body["workflow_id"], tt.wantWorkflowID, buttonRuleStub.CapturedBody)
				}
			}
		})
	}
}

func TestBaseButtonRuleFieldResolutionFailureStopsBeforeButtonRule(t *testing.T) {
	factory, stdout, reg := newExecuteFactory(t)
	buttonRuleCalls := 0
	reg.Register(&httpmock.Stub{
		Method: "GET",
		URL:    baseV3Path("bases", "app_x", "tables", "tbl_1", "fields", "missing"),
		Body: map[string]interface{}{
			"code": 1254045,
			"msg":  "field not found",
			"data": map[string]interface{}{"error": map[string]interface{}{"logid": "log_field_resolution"}},
		},
	})
	reg.Register(&httpmock.Stub{
		Method: "PUT", URL: "/button_rule", Optional: true,
		OnMatch: func(_ *http.Request) { buttonRuleCalls++ },
	})

	err := runShortcut(t, BaseButtonRuleBind, []string{"+button-rule-bind", "--base-token", "app_x", "--table-id", "tbl_1", "--field-id", "missing", "--workflow-id", "wkf_1"}, factory, stdout)
	problem, ok := errs.ProblemOf(err)
	if !ok || problem.Category != errs.CategoryAPI || problem.Code != 1254045 || problem.LogID != "log_field_resolution" {
		t.Fatalf("expected preserved typed field resolution error, got %T %#v", err, problem)
	}
	if buttonRuleCalls != 0 {
		t.Fatalf("ButtonRule calls=%d want=0", buttonRuleCalls)
	}
}

func TestBaseButtonRuleFieldResolutionRejectsMissingCanonicalID(t *testing.T) {
	factory, stdout, reg := newExecuteFactory(t)
	buttonRuleCalls := 0
	reg.Register(&httpmock.Stub{
		Method: "GET",
		URL:    baseV3Path("bases", "app_x", "tables", "tbl_1", "fields", "按钮"),
		Body:   map[string]interface{}{"code": 0, "data": map[string]interface{}{"name": "按钮"}},
	})
	reg.Register(&httpmock.Stub{
		Method: "GET", URL: "/button_rule", Optional: true,
		OnMatch: func(_ *http.Request) { buttonRuleCalls++ },
	})

	err := runShortcut(t, BaseButtonRuleGet, []string{"+button-rule-get", "--base-token", "app_x", "--table-id", "tbl_1", "--field-id", "按钮"}, factory, stdout)
	problem, ok := errs.ProblemOf(err)
	if !ok || problem.Category != errs.CategoryInternal || problem.Subtype != errs.SubtypeInvalidResponse {
		t.Fatalf("expected typed invalid-response error, got %T %#v", err, problem)
	}
	if buttonRuleCalls != 0 {
		t.Fatalf("ButtonRule calls=%d want=0", buttonRuleCalls)
	}
}

func TestBaseButtonRuleAPIFailurePreservesTypedCause(t *testing.T) {
	factory, stdout, reg := newExecuteFactory(t)
	cause := errors.New("button rule transport failed")
	reg.Register(&httpmock.Stub{
		Method: "GET",
		URL:    baseV3Path("bases", "app_x", "tables", "tbl_1", "fields", "按钮"),
		Body:   map[string]interface{}{"code": 0, "data": map[string]interface{}{"id": "fld_1"}},
	})
	reg.Register(&httpmock.Stub{
		Method: "GET",
		URL:    baseV3Path("bases", "app_x", "tables", "tbl_1", "fields", "fld_1", "button_rule"),
		Error:  cause,
	})

	err := runShortcut(t, BaseButtonRuleGet, []string{"+button-rule-get", "--base-token", "app_x", "--table-id", "tbl_1", "--field-id", "按钮"}, factory, stdout)
	problem, ok := errs.ProblemOf(err)
	if !ok || problem.Category != errs.CategoryNetwork || !errors.Is(err, cause) {
		t.Fatalf("expected typed network error preserving cause, got %T %#v", err, problem)
	}
}

func TestBaseButtonRuleValidateRejectsInternalWorkflowID(t *testing.T) {
	factory, stdout, _ := newExecuteFactory(t)
	err := runShortcut(t, BaseButtonRuleBind, []string{"+button-rule-bind", "--base-token", "app_x", "--table-id", "tbl_1", "--field-id", "fld_1", "--workflow-id", "123456"}, factory, stdout)
	if err == nil || !strings.Contains(err.Error(), "public wkf workflow ID") {
		t.Fatalf("err=%v", err)
	}
}
