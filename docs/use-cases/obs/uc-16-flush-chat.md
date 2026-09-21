# UC-16: Flush OBS chat from the Test pane

**Module:** `internal/obs`  
**Status:** Ported  
**Actors:** Operator (Test pane)  
**Goal:** Clear every on-screen chat line without stopping YouTube or overlay  
**Preconditions:** OBS HTTP is listening. Chat page is open.  

## Main scenario (happy path)

1. Operator clicks **Flush chat**.
2. Binding calls `PublishChat` with `type` `chat_flush`.
3. Chat page removes all rows. Overlay and TTS are unchanged.

## Alternative scenarios and errors

* **OBS HTTP down:** `ErrOBSNotListening`.
* **No chat clients:** publish still succeeds.

## Postconditions

* Chat lines after flush still appear. Start is not required.
