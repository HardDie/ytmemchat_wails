# 28. Commands pane blocks duplicate names before save

* **Status:** Accepted
* **Date:** 2026-09-25
* **Authors:** @oleg

---

## Context

1. Command names must be unique ignoring case.
   1. Chat matching already ignores case.
   2. `validateCommands` rejects a duplicate on Save YAML.
2. The Commands pane lets the operator type the duplicate first.
3. The failure shows up only after Save, as a flash error.
4. A new row and an edited row can both collide with another name.

## Considered options

1. **Keep the save error only** — no extra UI; the operator sees the problem after Save.
2. **Flag the Name fields and block Save** — same rule as YAML validation; the bad rows are visible while typing.
3. **Auto-rename the new row** — Save always works; the operator does not pick the final name.

## Decision

Use option 2.

1. Compare trimmed names with `toLowerCase`.
   1. Same key as `strings.ToLower` in `validateCommands`.
2. An empty name is not a duplicate.
3. Every row that shares a name is marked.
   1. New row and existing row both count.
4. The Name input uses a red border, including while focused.
5. The field shows "This command already exists".
6. **Save YAML** stays disabled until every name is unique.
7. The click handler returns without saving while a duplicate remains.
8. YAML validation on save stays in place.

## Consequences

### Positive

* The operator sees the clash before Save.
* The pane and the YAML file use one uniqueness rule.

### Negative and risks

* Two names that differ only by case cannot both be saved.
* The button stays disabled until the extra name is changed or the row is deleted.

### Neutral

* Blank volume and scale are unchanged.
* A missing name or file still fails on Save, as before.
