// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package base

import (
	"fmt"
	"strings"
)

const workflowAIClassificationStepType = "AIClassificationBranch"

func validateWorkflowAIClassificationBranches(body map[string]interface{}) error {
	stepsRaw, ok := body["steps"]
	if !ok || stepsRaw == nil {
		return nil
	}
	steps, ok := stepsRaw.([]interface{})
	if !ok {
		return nil
	}

	stepIDs := make(map[string]int, len(steps))
	for i, raw := range steps {
		step, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		id, _ := step["id"].(string)
		if id != "" {
			stepIDs[id] = i
		}
	}

	for i, raw := range steps {
		step, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		stepType, _ := step["type"].(string)
		if stepType != workflowAIClassificationStepType {
			continue
		}
		if err := validateWorkflowAIClassificationStep(step, i, stepIDs); err != nil {
			return err
		}
	}
	return nil
}

func validateWorkflowAIClassificationStep(step map[string]interface{}, index int, stepIDs map[string]int) error {
	path := fmt.Sprintf("--json steps[%d]", index)
	stepID, _ := step["id"].(string)
	if strings.TrimSpace(stepID) == "" {
		return baseValidationErrorf("%s.id must be a non-empty string for AIClassificationBranch", path)
	}
	currentIndex, ok := stepIDs[stepID]
	if !ok {
		currentIndex = index
	}

	data, ok := step["data"].(map[string]interface{})
	if !ok || data == nil {
		return baseValidationErrorf("%s.data must be an object for AIClassificationBranch", path)
	}
	if err := validateAIClassificationMode(data, path); err != nil {
		return err
	}
	if err := validateAIClassificationPrompt(data, path, stepID, currentIndex, stepIDs); err != nil {
		return err
	}
	if err := validateAIClassificationBranches(data, path, stepIDs); err != nil {
		return err
	}
	if err := validateAIClassificationDefaultBranch(data, path, stepIDs); err != nil {
		return err
	}
	if err := validateAIClassificationTextRefList(data, "classifyPrompt", "classify_prompt", path, false, false, stepID, currentIndex, stepIDs); err != nil {
		return err
	}
	return validateAIClassificationChildLinks(step, path, stepIDs)
}

func validateAIClassificationMode(data map[string]interface{}, path string) error {
	raw, ok := data["mode"]
	if !ok || raw == nil {
		return nil
	}
	mode, ok := raw.(string)
	if !ok {
		return baseValidationErrorf("%s.data.mode must be a string when set", path)
	}
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "exclusive", "parallel":
		return nil
	default:
		return baseValidationErrorf("%s.data.mode must be Exclusive or Parallel", path)
	}
}

func validateAIClassificationPrompt(data map[string]interface{}, path string, stepID string, currentIndex int, stepIDs map[string]int) error {
	if _, ok := data["prompt"]; !ok {
		return baseValidationErrorf("%s.data.prompt is required for AIClassificationBranch", path)
	}
	return validateAIClassificationTextRefList(data, "prompt", "prompt", path, true, true, stepID, currentIndex, stepIDs)
}

func validateAIClassificationBranches(data map[string]interface{}, path string, stepIDs map[string]int) error {
	raw, branchKey, ok := firstExisting(data, "childBranchList", "child_branch_list")
	if !ok {
		return baseValidationErrorf("%s.data.childBranchList is required for AIClassificationBranch", path)
	}
	branches, ok := raw.([]interface{})
	if !ok {
		return baseValidationErrorf("%s.data.%s must be an array", path, branchKey)
	}
	if len(branches) < 2 {
		return baseValidationErrorf("%s.data.%s must contain at least 2 classifications", path, branchKey)
	}

	seenNames := map[string]int{}
	seenEntries := map[string]string{}
	for i, rawBranch := range branches {
		branchPath := fmt.Sprintf("%s.data.%s[%d]", path, branchKey, i)
		branch, ok := rawBranch.(map[string]interface{})
		if !ok {
			return baseValidationErrorf("%s must be an object", branchPath)
		}
		name, err := textRefPlainText(branch["name"])
		if err != nil {
			return baseValidationErrorf("%s.name must be plain text", branchPath)
		}
		name = strings.TrimSpace(name)
		switch {
		case name == "":
			return baseValidationErrorf("%s.name must not be blank", branchPath)
		case strings.ContainsAny(name, "\r\n"):
			return baseValidationErrorf("%s.name must not contain newlines", branchPath)
		case name == "其他":
			return baseValidationErrorf("%s.name must not use reserved name %q", branchPath, "其他")
		}
		if prev, exists := seenNames[name]; exists {
			return baseValidationErrorf("%s.name duplicates %s.data.%s[%d].name", branchPath, path, branchKey, prev)
		}
		seenNames[name] = i

		if descRaw, ok := branch["description"]; ok && descRaw != nil {
			if _, err := textRefPlainText(descRaw); err != nil {
				return baseValidationErrorf("%s.description must be plain text", branchPath)
			}
		}
		if entry, ok := firstString(branch, "entryChildStepId", "entry_child_step_id"); ok {
			entry = strings.TrimSpace(entry)
			if entry == "" {
				return baseValidationErrorf("%s.entryChildStepId must not be blank when set", branchPath)
			}
			if _, exists := stepIDs[entry]; !exists {
				return baseValidationErrorf("%s.entryChildStepId references unknown step id %q", branchPath, entry)
			}
			if prevName, exists := seenEntries[entry]; exists {
				return baseValidationErrorf("%s.entryChildStepId %q is already used by classification %q", branchPath, entry, prevName)
			}
			seenEntries[entry] = name
		}
	}
	return nil
}

func validateAIClassificationDefaultBranch(data map[string]interface{}, path string, stepIDs map[string]int) error {
	raw, branchKey, ok := firstExisting(data, "defaultBranchInfo", "default_branch_info")
	if ok && raw != nil {
		info, ok := raw.(map[string]interface{})
		if !ok {
			return baseValidationErrorf("%s.data.%s must be an object when set", path, branchKey)
		}
		if mode, ok := firstString(info, "mode"); ok {
			switch strings.ToLower(strings.TrimSpace(mode)) {
			case "execute", "fail":
			default:
				return baseValidationErrorf("%s.data.%s.mode must be Execute or Fail", path, branchKey)
			}
		}
		if entry, ok := firstString(info, "entryStepId", "entry_step_id"); ok {
			entry = strings.TrimSpace(entry)
			if entry != "" {
				if _, exists := stepIDs[entry]; !exists {
					return baseValidationErrorf("%s.data.%s.entryStepId references unknown step id %q", path, branchKey, entry)
				}
			}
		}
	}

	if action, ok := firstString(data, "no_match_action", "noMatchAction"); ok {
		switch strings.ToLower(strings.TrimSpace(action)) {
		case "classifytoother", "classify_to_other", "fail":
		default:
			return baseValidationErrorf("%s.data.no_match_action must be classifyToOther or fail", path)
		}
	}
	return nil
}

func validateAIClassificationChildLinks(step map[string]interface{}, path string, stepIDs map[string]int) error {
	children, ok := step["children"].(map[string]interface{})
	if !ok || children == nil {
		return baseValidationErrorf("%s.children.links is required for AIClassificationBranch", path)
	}
	linksRaw, ok := children["links"]
	if !ok {
		return baseValidationErrorf("%s.children.links is required for AIClassificationBranch", path)
	}
	links, ok := linksRaw.([]interface{})
	if !ok {
		return baseValidationErrorf("%s.children.links must be an array", path)
	}
	seenTargets := map[string]int{}
	for i, raw := range links {
		linkPath := fmt.Sprintf("%s.children.links[%d]", path, i)
		link, ok := raw.(map[string]interface{})
		if !ok {
			return baseValidationErrorf("%s must be an object", linkPath)
		}
		kind, _ := link["kind"].(string)
		if strings.TrimSpace(kind) != "case" {
			return baseValidationErrorf("%s.kind must be case for AIClassificationBranch", linkPath)
		}
		to, _ := link["to"].(string)
		to = strings.TrimSpace(to)
		if to == "" {
			return baseValidationErrorf("%s.to must not be blank", linkPath)
		}
		if _, exists := stepIDs[to]; !exists {
			return baseValidationErrorf("%s.to references unknown step id %q", linkPath, to)
		}
		if prev, exists := seenTargets[to]; exists {
			return baseValidationErrorf("%s.to duplicates %s.children.links[%d].to", linkPath, path, prev)
		}
		seenTargets[to] = i
	}
	return nil
}

func validateAIClassificationTextRefList(data map[string]interface{}, camelKey string, snakeKey string, path string, required bool, allowRef bool, stepID string, currentIndex int, stepIDs map[string]int) error {
	raw, key, ok := firstExisting(data, camelKey, snakeKey)
	if !ok || raw == nil {
		if required {
			return baseValidationErrorf("%s.data.%s is required for AIClassificationBranch", path, camelKey)
		}
		return nil
	}
	items, ok := raw.([]interface{})
	if !ok {
		return baseValidationErrorf("%s.data.%s must be a TextRefItem array", path, key)
	}

	hasContent := false
	for i, rawItem := range items {
		itemPath := fmt.Sprintf("%s.data.%s[%d]", path, key, i)
		item, ok := rawItem.(map[string]interface{})
		if !ok {
			return baseValidationErrorf("%s must be an object", itemPath)
		}
		valueType, _ := firstString(item, "value_type", "valueType", "type")
		valueType = strings.ToLower(strings.TrimSpace(valueType))
		value, _ := item["value"].(string)
		if valueType == "" {
			return baseValidationErrorf("%s.value_type must be text or ref", itemPath)
		}
		switch valueType {
		case "text":
			if strings.TrimSpace(value) != "" {
				hasContent = true
			}
		case "ref":
			if !allowRef {
				return baseValidationErrorf("%s.value_type must be text; ref is not supported here", itemPath)
			}
			if strings.TrimSpace(value) == "" {
				return baseValidationErrorf("%s.value must not be blank for ref", itemPath)
			}
			refStep, ok := workflowRefStepID(value)
			if !ok {
				return baseValidationErrorf("%s.value must be a workflow ref path starting with $.step_id", itemPath)
			}
			refIndex, exists := stepIDs[refStep]
			if !exists {
				return baseValidationErrorf("%s.value references unknown step id %q", itemPath, refStep)
			}
			if refStep == stepID || refIndex >= currentIndex {
				return baseValidationErrorf("%s.value must reference a previous step, got %q", itemPath, refStep)
			}
			hasContent = true
		default:
			return baseValidationErrorf("%s.value_type must be text or ref", itemPath)
		}
	}
	if required && !hasContent {
		return baseValidationErrorf("%s.data.%s must contain non-empty text or a valid ref", path, key)
	}
	return nil
}

func workflowRefStepID(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, "$.") {
		return "", false
	}
	value = strings.TrimPrefix(value, "$.")
	if value == "" {
		return "", false
	}
	if idx := strings.Index(value, "."); idx >= 0 {
		value = value[:idx]
	}
	return value, value != ""
}

func textRefPlainText(raw interface{}) (string, error) {
	switch value := raw.(type) {
	case string:
		return value, nil
	case []interface{}:
		var b strings.Builder
		for _, itemRaw := range value {
			item, ok := itemRaw.(map[string]interface{})
			if !ok {
				return "", fmt.Errorf("item must be an object")
			}
			valueType, _ := firstString(item, "value_type", "valueType", "type")
			if strings.ToLower(strings.TrimSpace(valueType)) != "text" {
				return "", fmt.Errorf("item must be text")
			}
			text, ok := item["value"].(string)
			if !ok {
				return "", fmt.Errorf("value must be string")
			}
			b.WriteString(text)
		}
		return b.String(), nil
	default:
		return "", fmt.Errorf("value must be string or TextRefItem array")
	}
}

func firstExisting(data map[string]interface{}, keys ...string) (interface{}, string, bool) {
	for _, key := range keys {
		if value, ok := data[key]; ok {
			return value, key, true
		}
	}
	return nil, "", false
}

func firstString(data map[string]interface{}, keys ...string) (string, bool) {
	for _, key := range keys {
		if value, ok := data[key].(string); ok {
			return value, true
		}
	}
	return "", false
}
