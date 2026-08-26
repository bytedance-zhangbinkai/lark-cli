// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package base

import (
	"strings"
	"testing"

	"github.com/larksuite/cli/internal/httpmock"
)

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
	t.Run("rejects ai analysis table names string", func(t *testing.T) {
		factory, stdout, _ := newExecuteFactory(t)
		err := runShortcut(t, BaseWorkflowCreate, []string{"+workflow-create", "--base-token", "app_x", "--json", `{"steps":[{"id":"step_ai","type":"AIAnalysisAction","data":{"analysis_table_names":"订单表","identity_type":"maker"}}]}`}, factory, stdout)
		assertInvalidArgumentValidation(t, err, "--json", []string{"--json"}, "steps[0].data.analysis_table_names")
	})
	t.Run("rejects ai analysis table names item type", func(t *testing.T) {
		factory, stdout, _ := newExecuteFactory(t)
		err := runShortcut(t, BaseWorkflowCreate, []string{"+workflow-create", "--base-token", "app_x", "--json", `{"steps":[{"id":"step_ai","type":"AIAnalysisAction","data":{"analysis_table_names":["订单表",1],"identity_type":"maker"}}]}`}, factory, stdout)
		assertInvalidArgumentValidation(t, err, "--json", []string{"--json"}, "steps[0].data.analysis_table_names[1]")
	})
	t.Run("rejects ai analysis identity type enum", func(t *testing.T) {
		factory, stdout, _ := newExecuteFactory(t)
		err := runShortcut(t, BaseWorkflowCreate, []string{"+workflow-create", "--base-token", "app_x", "--json", `{"steps":[{"id":"step_ai","type":"AIAnalysisAction","data":{"analysis_table_names":["订单表"],"identity_type":"unknownIdentity"}}]}`}, factory, stdout)
		assertInvalidArgumentValidation(t, err, "--json", []string{"--json"}, "maker, triggerPersonal")
		if !strings.Contains(err.Error(), "steps[0].data.identity_type") {
			t.Fatalf("err=%v, want field path", err)
		}
	})
	t.Run("accepts ai analysis maker", func(t *testing.T) {
		factory, stdout, reg := newExecuteFactory(t)
		reg.Register(&httpmock.Stub{
			Method: "POST",
			URL:    "/open-apis/base/v3/bases/app_x/workflows",
			Body: map[string]interface{}{
				"code": 0,
				"data": map[string]interface{}{"workflow_id": "wkf_new"},
			},
		})
		err := runShortcut(t, BaseWorkflowCreate, []string{"+workflow-create", "--base-token", "app_x", "--json", `{"steps":[{"id":"step_ai","type":"AIAnalysisAction","data":{"analysis_table_names":["订单表"],"identity_type":"maker"}}]}`}, factory, stdout)
		if err != nil {
			t.Fatalf("err=%v", err)
		}
	})
	t.Run("accepts ai analysis triggerPersonal", func(t *testing.T) {
		factory, stdout, reg := newExecuteFactory(t)
		reg.Register(&httpmock.Stub{
			Method: "POST",
			URL:    "/open-apis/base/v3/bases/app_x/workflows",
			Body: map[string]interface{}{
				"code": 0,
				"data": map[string]interface{}{"workflow_id": "wkf_new"},
			},
		})
		err := runShortcut(t, BaseWorkflowCreate, []string{"+workflow-create", "--base-token", "app_x", "--json", `{"steps":[{"id":"step_ai","type":"AIAnalysisAction","data":{"analysis_table_names":["订单表"],"identity_type":"triggerPersonal"}}]}`}, factory, stdout)
		if err != nil {
			t.Fatalf("err=%v", err)
		}
	})
}

func TestBaseWorkflowExecuteUpdateValidateAIAnalysisData(t *testing.T) {
	t.Run("rejects table names string", func(t *testing.T) {
		factory, stdout, _ := newExecuteFactory(t)
		err := runShortcut(t, BaseWorkflowUpdate, []string{"+workflow-update", "--base-token", "app_x", "--workflow-id", "wkf_1", "--json", `{"steps":[{"id":"step_ai","type":"AIAnalysisAction","data":{"analysis_table_names":"订单表","identity_type":"maker"}}]}`}, factory, stdout)
		assertInvalidArgumentValidation(t, err, "--json", []string{"--json"}, "steps[0].data.analysis_table_names")
	})
	t.Run("rejects identity type enum", func(t *testing.T) {
		factory, stdout, _ := newExecuteFactory(t)
		err := runShortcut(t, BaseWorkflowUpdate, []string{"+workflow-update", "--base-token", "app_x", "--workflow-id", "wkf_1", "--json", `{"steps":[{"id":"step_ai","type":"AIAnalysisAction","data":{"analysis_table_names":["订单表"],"identity_type":"unknownIdentity"}}]}`}, factory, stdout)
		assertInvalidArgumentValidation(t, err, "--json", []string{"--json"}, "maker, triggerPersonal")
		if !strings.Contains(err.Error(), "steps[0].data.identity_type") {
			t.Fatalf("err=%v, want field path", err)
		}
	})
}

func TestBaseWorkflowExecuteListReturnsEmptyItemsArray(t *testing.T) {
	factory, stdout, reg := newExecuteFactory(t)
	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/open-apis/base/v3/bases/app_x/workflows/list",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"items": nil, "total": 0},
		},
	})
	if err := runShortcut(t, BaseWorkflowList, []string{"+workflow-list", "--base-token", "app_x"}, factory, stdout); err != nil {
		t.Fatalf("err=%v", err)
	}
	data := decodeBaseEnvelope(t, stdout)
	items, ok := data["items"].([]interface{})
	if !ok {
		t.Fatalf("items=%#v, want []", data["items"])
	}
	if len(items) != 0 || data["total"] != float64(0) {
		t.Fatalf("data=%#v, want empty items and total=0", data)
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
