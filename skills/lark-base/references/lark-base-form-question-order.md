# Base Form Question Visibility and Ordering

Use `+view-get-visible-fields` and `+view-set-visible-fields` to read and atomically update which Form questions are visible and their order. The Form ID is a View ID at this API boundary, so pass it through `--view-id`; do not use `+form-questions-update`, `pre_question_id`, `index`, or per-question PATCH requests for visibility or ordering.

## Contract

For standard Grid, Kanban, Gallery, Calendar, and Gantt views, `visible_fields` controls both visibility and order, and the API may place the primary field first.

For a Form, `visible_fields` is the complete final ordered list of questions that should be visible:

- A currently visible question omitted from the array becomes hidden.
- An existing hidden Form question included in the array becomes visible again.
- Questions in the array are ordered exactly as requested; hidden questions keep their internal relative order.
- The primary field is not inserted automatically. It is shown only when explicitly included.
- An empty array hides every question.
- A no-op target is safe and idempotent.
- Duplicate questions and table fields that are not already Form members are rejected.
- Query Form is not supported.
- A visible question's `visible_rule` may reference only visible questions before it. Hiding its dependency or moving the dependent question before its dependency is rejected without a write.

The update changes only visibility and order. It preserves question title, description, required state, option display mode, `visible_rule`, and other presentation configuration.

Use stable question IDs from `+form-questions-list` when writing. The GET `visible_fields` response preserves its field-name output contract, so names may be ambiguous or may change. Save a question's stable ID before hiding it because the current read interfaces may not expose a complete hidden-member list.

## Safe visibility and ordering workflow

1. Fresh-read the Form questions. Save each visible question's ID, current order, title, description, required state, option display mode, `visible_rule`, and other returned configuration. Preserve any previously known hidden-member IDs that may need to be restored.
2. Fresh-read `visible_fields` for the same Form and reconcile the returned names with the stable IDs.
3. Build the complete final visible target. Include each question that should remain or become visible exactly once, in final order; omit every question that should be hidden.
4. Check every surviving `visible_rule`: its referenced questions must remain visible and occur before the dependent question.
5. Dry-run the setter and inspect the exact PUT URL and body. The CLI must not add the primary field, fetch state implicitly, or rewrite the target array.
6. Execute the identical target array. Visibility changes and ordering are submitted in one server changeset.
7. Fresh-read both `+view-get-visible-fields` and `+form-questions-list`. Treat the write as persisted only when they reflect the target projection and retained questions preserve their configuration.

```bash
# 1. Read stable IDs and preserve complete question configuration.
lark-cli base +form-questions-list \
  --base-token <base_token> \
  --table-id <table_id> \
  --form-id <form_id> \
  --as user

# 2. Read the current visible-question projection.
lark-cli base +view-get-visible-fields \
  --base-token <base_token> \
  --table-id <table_id> \
  --view-id <form_id> \
  --as user

# 3. Preview the complete final target. fld_b is omitted and becomes hidden;
# fld_c may be an existing hidden Form member that becomes visible again.
lark-cli base +view-set-visible-fields \
  --base-token <base_token> \
  --table-id <table_id> \
  --view-id <form_id> \
  --json '{"visible_fields":["fld_c","fld_a"]}' \
  --dry-run \
  --as user

# 4. Apply the exact same target after the preview is correct.
lark-cli base +view-set-visible-fields \
  --base-token <base_token> \
  --table-id <table_id> \
  --view-id <form_id> \
  --json '{"visible_fields":["fld_c","fld_a"]}' \
  --as user

# 5. Fresh-read the target projection; repeat step 1 to verify question config.
lark-cli base +view-get-visible-fields \
  --base-token <base_token> \
  --table-id <table_id> \
  --view-id <form_id> \
  --as user
```

## Failure handling

If the server rejects the request, report the exact typed error, hint, and `log_id`, then fresh-read both endpoints to prove that no write occurred. Resolve the specific cause:

- Duplicate reference: remove the duplicate and retry with the intended complete target.
- Ambiguous field name: use the stable question ID.
- Field not in Form: add it explicitly with `+form-questions-create` using `use_existing_field:true`, then reconstruct the final target. Do not use `visible_fields` as an implicit question-create operation.
- Field access denied: use another accessible field or obtain access; do not classify it as a non-Form field.
- `visible_rule` dependency violation: keep the dependency visible and before the dependent question, or update the rule separately with explicit user intent.
- Operation limit: report that the requested atomic reorder exceeds the limit. Do not silently split one target into partial writes.
- Unsupported Query Form: stop; this workflow does not apply.

## Related commands

- `+form-questions-create` adds new or existing table fields as Form questions.
- `+form-questions-delete --keep-field` hides a question while preserving its backing field and record data.
- `+form-questions-update` changes complete question presentation configuration and is not a visibility or ordering API.
