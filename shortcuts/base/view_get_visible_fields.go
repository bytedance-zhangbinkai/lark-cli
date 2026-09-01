// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package base

import (
	"context"

	"github.com/larksuite/cli/shortcuts/common"
)

var BaseViewGetVisibleFields = common.Shortcut{
	Service:     "base",
	Command:     "+view-get-visible-fields",
	Description: "Get view visible fields configuration",
	Risk:        "read",
	Scopes:      []string{"base:view:read"},
	AuthTypes:   authTypes(),
	Flags:       []common.Flag{baseTokenFlag(true), tableRefFlag(true), viewRefFlag(true)},
	Tips: []string{
		"Supported view types: grid, kanban, gallery, calendar, gantt, and form; query form is not supported.",
		"For a form, the result is the current visible questions in question order. Use +form-questions-list to obtain stable question IDs before reordering.",
	},
	DryRun: dryRunViewGetVisibleFields,
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		return executeViewGetProperty(runtime, "visible_fields", "visible_fields")
	},
}
