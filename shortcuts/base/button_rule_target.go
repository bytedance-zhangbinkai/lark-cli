// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package base

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

const (
	buttonTargetWorkflow   = "workflow"
	buttonTargetOpenRecord = "open_record"
	buttonTargetOpenLink   = "open_link"
	buttonTargetOpenForm   = "open_form"
)

type buttonTargetValueInfo struct {
	ValueType string `json:"value_type"`
	Value     string `json:"value"`
}

type buttonRuleTarget struct {
	Type    string                  `json:"type"`
	ID      string                  `json:"id,omitempty"`
	TableID string                  `json:"table_id,omitempty"`
	FormID  string                  `json:"form_id,omitempty"`
	Link    []buttonTargetValueInfo `json:"link,omitempty"`

	presentFields map[string]struct{}
}

func (t *buttonRuleTarget) UnmarshalJSON(data []byte) error {
	type targetAlias buttonRuleTarget
	var decoded targetAlias
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	*t = buttonRuleTarget(decoded)
	t.presentFields = make(map[string]struct{}, len(fields))
	for field := range fields {
		t.presentFields[field] = struct{}{}
	}
	return nil
}

func parseButtonRuleTarget(raw string) (*buttonRuleTarget, error) {
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.DisallowUnknownFields()

	var target buttonRuleTarget
	if err := decoder.Decode(&target); err != nil {
		return nil, baseFlagErrorf("--target-json must be a strict ButtonTarget JSON object: %v", err)
	}
	if err := ensureButtonTargetJSONEnd(decoder); err != nil {
		return nil, err
	}
	if err := validateButtonRuleTarget(&target); err != nil {
		return nil, err
	}
	return &target, nil
}

func ensureButtonTargetJSONEnd(decoder *json.Decoder) error {
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return baseFlagErrorf("--target-json must contain exactly one JSON object")
		}
		return baseFlagErrorf("--target-json contains trailing invalid JSON: %v", err)
	}
	return nil
}

func validateButtonRuleTarget(target *buttonRuleTarget) error {
	target.Type = strings.TrimSpace(target.Type)
	target.ID = strings.TrimSpace(target.ID)
	target.TableID = strings.TrimSpace(target.TableID)
	target.FormID = strings.TrimSpace(target.FormID)
	allowedFields := map[string]map[string]bool{
		buttonTargetWorkflow:   {"type": true, "id": true},
		buttonTargetOpenRecord: {"type": true},
		buttonTargetOpenForm:   {"type": true, "table_id": true, "form_id": true},
		buttonTargetOpenLink:   {"type": true, "link": true},
	}
	if allowed, ok := allowedFields[target.Type]; ok {
		for field := range target.presentFields {
			if !allowed[field] {
				return baseFlagErrorf("--target-json %s does not support field %s", target.Type, field)
			}
		}
	}

	switch target.Type {
	case buttonTargetWorkflow:
		if target.ID == "" || !strings.HasPrefix(target.ID, "wkf") {
			return baseFlagErrorf("--target-json workflow.id must be a public wkf workflow ID")
		}
		if target.TableID != "" || target.FormID != "" || target.Link != nil {
			return baseFlagErrorf("--target-json workflow accepts only type and id")
		}
	case buttonTargetOpenRecord:
		if target.ID != "" || target.TableID != "" || target.FormID != "" || target.Link != nil {
			return baseFlagErrorf("--target-json open_record accepts only type")
		}
	case buttonTargetOpenForm:
		if target.TableID == "" || target.FormID == "" {
			return baseFlagErrorf("--target-json open_form requires non-empty table_id and form_id")
		}
		if !strings.HasPrefix(target.TableID, "tbl") || !strings.HasPrefix(target.FormID, "vew") {
			return baseFlagErrorf("--target-json open_form requires stable tbl... table_id and vew... form_id values")
		}
		if target.ID != "" || target.Link != nil {
			return baseFlagErrorf("--target-json open_form accepts only type, table_id, and form_id")
		}
	case buttonTargetOpenLink:
		if target.ID != "" || target.TableID != "" || target.FormID != "" {
			return baseFlagErrorf("--target-json open_link accepts only type and link")
		}
		if len(target.Link) == 0 {
			return baseFlagErrorf("--target-json open_link.link must contain at least one item")
		}
		for index := range target.Link {
			item := &target.Link[index]
			item.ValueType = strings.TrimSpace(item.ValueType)
			if item.Value == "" {
				return baseFlagErrorf("--target-json open_link.link[%d].value must not be empty", index)
			}
			switch item.ValueType {
			case "text":
			case "ref":
				if err := validateButtonTargetRef(item.Value); err != nil {
					return err
				}
			default:
				return baseFlagErrorf("--target-json open_link.link[%d].value_type must be text or ref", index)
			}
		}
	default:
		return baseFlagErrorf("--target-json type must be one of workflow, open_record, open_link, or open_form")
	}
	return nil
}

func validateButtonTargetRef(ref string) error {
	const prefix = "$.button."
	if !strings.HasPrefix(ref, prefix) {
		return baseFlagErrorf("--target-json open_link ref %q must use the $.button namespace", ref)
	}
	path := strings.TrimPrefix(ref, prefix)
	switch path {
	case "triggerTime", "recordId", "recordLink", "recordCreatedBy", "recordCreatedTime",
		"recordModifiedBy", "recordModifiedTime", "baseLink":
		return nil
	}
	if strings.HasPrefix(path, "tableLink.") {
		tableID := strings.TrimPrefix(path, "tableLink.")
		if strings.HasPrefix(tableID, "tbl") && !strings.Contains(tableID, ".") {
			return nil
		}
	}
	if strings.HasPrefix(path, "viewLink.") {
		parts := strings.Split(strings.TrimPrefix(path, "viewLink."), ".")
		if len(parts) == 2 && strings.HasPrefix(parts[0], "tbl") && strings.HasPrefix(parts[1], "vew") {
			return nil
		}
	}
	if strings.HasPrefix(path, "fld") && !strings.Contains(path, ".") {
		return nil
	}
	return baseFlagErrorf("--target-json open_link ref %q is not a supported $.button path", ref)
}
