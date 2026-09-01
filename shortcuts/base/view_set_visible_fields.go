// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package base

import (
	"context"

	"github.com/larksuite/cli/shortcuts/common"
)

var BaseViewSetVisibleFields = common.Shortcut{
	Service:     "base",
	Command:     "+view-set-visible-fields",
	Description: "Set view visible fields",
	Risk:        "write",
	Scopes:      []string{"base:view:write_only"},
	AuthTypes:   authTypes(),
	Flags: []common.Flag{
		baseTokenFlag(true),
		tableRefFlag(true),
		viewRefFlag(true),
		{Name: "json", Desc: `visible fields JSON object, e.g. {"visible_fields":["Name","Status"]}`, Required: true},
	},
	Tips: []string{
		"Supported view types: grid, kanban, gallery, calendar, gantt, and form; query form is not supported.",
		"Use a JSON object, not a bare array.",
		"For a standard view, visible_fields controls both visibility and order; include every field that should remain visible, and the API may force the primary field to the first position.",
		"For a form, visible_fields must contain every currently visible question exactly once and only reorders that same set; it does not add, remove, show, or hide questions, and it does not force the primary field first.",
		"Before reordering a form, use +form-questions-list to obtain stable question IDs and preserve question configuration. Keep every visible_rule dependency before the question that references it, then perform a fresh readback after the write.",
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		return validateViewJSONObject(runtime)
	},
	DryRun: dryRunViewSetVisibleFields,
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		return executeViewSetVisibleFields(runtime)
	},
}
