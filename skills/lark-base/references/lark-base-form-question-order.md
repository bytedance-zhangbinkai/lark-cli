# Base Form Question Ordering

Use `+view-get-visible-fields` and `+view-set-visible-fields` to read and reorder Form questions. The Form ID is a View ID at this API boundary, so pass it through `--view-id`; do not use `+form-questions-update` or invent `pre_question_id`, `index`, or per-question move fields for ordering.

## Contract

For standard Grid, Kanban, Gallery, Calendar, and Gantt views, `visible_fields` controls both visibility and order, and the API may place the primary field first.

For a Form, `visible_fields` is the complete ordered set of questions that are currently visible in that Form:

- The request must contain every currently visible question exactly once.
- Extra table fields, hidden questions, missing questions, and duplicate questions are rejected.
- A successful request only changes order. It does not add, remove, show, or hide questions, change question configuration, or force the primary field first.
- A no-op order is safe and idempotent.
- Query Form is not supported.
- A `visible_rule` may only reference questions before the question carrying the rule. A target order that reverses such a dependency is rejected without a write.

Use stable question IDs from `+form-questions-list` when writing. The GET `visible_fields` response preserves its existing field-name output contract, so names may be ambiguous or may change.

## Safe reorder workflow

1. Fresh-read the full question list. Save each question's ID, current order, visibility, title, description, required state, option display mode, `visible_rule`, and any other returned configuration as the before snapshot.
2. Fresh-read `visible_fields` for the same Form. Reconcile it with the question list and construct a target ID array containing the same currently visible questions exactly once.
3. Check every existing `visible_rule`: each referenced question must remain before the question that carries the rule.
4. Dry-run the setter and inspect the exact PUT URL and body. The CLI must not add the primary field, fetch state implicitly, or rewrite the target array.
5. Execute the identical complete target array.
6. Fresh-read both `+view-get-visible-fields` and `+form-questions-list`. Treat the write as persisted only when both reflect the target order and all question IDs, visibility values, and configuration match the before snapshot.

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

# 3. Preview the complete target order.
lark-cli base +view-set-visible-fields \
  --base-token <base_token> \
  --table-id <table_id> \
  --view-id <form_id> \
  --json '{"visible_fields":["fld_a","fld_c","fld_b"]}' \
  --dry-run \
  --as user

# 4. Apply the same request after the preview is correct.
lark-cli base +view-set-visible-fields \
  --base-token <base_token> \
  --table-id <table_id> \
  --view-id <form_id> \
  --json '{"visible_fields":["fld_a","fld_c","fld_b"]}' \
  --as user

# 5. Fresh-read both views of the state; repeat step 1 as well.
lark-cli base +view-get-visible-fields \
  --base-token <base_token> \
  --table-id <table_id> \
  --view-id <form_id> \
  --as user
```

## Failure handling

If the server rejects the request, do not retry with a smaller or partial field set. Report the exact typed error, hint, and `log_id`, then fresh-read both endpoints to prove that no write occurred. Resolve the specific cause:

- Missing or duplicate question: rebuild the complete array from fresh reads.
- Extra, hidden, missing, or ambiguous field reference: use the stable question IDs from `+form-questions-list`.
- `visible_rule` dependency inversion: keep the dependency before the dependent question, or update the rule separately with explicit user intent.
- Operation limit: reduce the number of effective moves only if the intended final complete order can still be expressed in one request; do not split one reorder into partial writes.
- Unsupported Query Form: stop; this workflow does not apply.

## Related commands

- `+form-questions-create` adds questions.
- `+form-questions-delete --keep-field` removes a question from the Form while preserving the backing field and record data.
- `+form-questions-update` changes complete question configuration and is not an ordering API.
