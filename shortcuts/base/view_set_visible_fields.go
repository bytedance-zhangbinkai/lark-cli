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
		"For a form, visible_fields is the complete final ordered list of visible questions: omitted visible questions are hidden, included hidden Form questions are shown, and the primary field is not added automatically. An empty list hides every question.",
		"A table field that is not already a Form question is rejected; add it first with +form-questions-create and use_existing_field:true. Use +form-questions-list to obtain stable question IDs before changing visibility or order.",
		"Keep every visible_rule dependency visible and before the question that references it. After the write, fresh-read both visible_fields and Form questions to verify persisted visibility, order, and preserved question configuration.",
	},
	Validate: func(ctx context.Context, runtime *common.RuntimeContext) error {
		return validateViewJSONObject(runtime)
	},
	DryRun: dryRunViewSetVisibleFields,
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		return executeViewSetVisibleFields(runtime)
	},
}
