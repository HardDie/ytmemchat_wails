# UC-13: Edit commands.yaml from the window

**Module:** `internal/alerts` (`LoadFile` / `SaveFile`) and Wails `GetAlertCommands` / `SaveAlertCommands`  
**Status:** Ported  
**Actors:** Operator in the Commands pane  
**Goal:** Add and edit alert commands, writing volume and scale only when set  
**Preconditions:** A `commands.yaml` path is saved in Configuration  

## Main scenario (happy path)

1. The window calls `GetAlertCommands`. A missing file yields an empty list (path is still returned).
2. The operator adds a command with `name` and `file`. Volume and scale may be left empty.
3. `SaveAlertCommands` writes YAML. Nil volume and scale keys are omitted. Duplicate names (case-insensitive) are rejected. Volume and scale, when set, must be non-negative with at most two decimal digits.
4. The matcher is reloaded if OBS HTTP is running (YouTube Start is not required).

## Alternative scenarios and errors

* **1a. Empty path:** error telling the operator to set the path in Configuration.
* **3a. Empty name or file:** save rejected; the previous file is left unchanged.
* **3b. Volume or scale has more than two decimal digits, is negative, or is not a number:** save rejected.
* **3c. Invalid YAML on load:** Get returns the decode error.

## Postconditions

* On success the file contains only the fields the operator set. Runtime matching still defaults missing volume/scale to 1.
