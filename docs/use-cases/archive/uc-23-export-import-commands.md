# UC-23: Export and import a command collection

**Module:** `pkg/archive`  
**Status:** Implemented  
**Actors:** Operator (Commands pane)  
**Goal:** Share saved commands and the media files they use, then restore that zip as a media folder  
**Preconditions:** A media folder is saved. Export also needs a saved `commands.yaml` on disk.  

`pkg/archive` writes one named document plus selected files. It does not know commands. The Commands binding passes `commands.yaml` and the used files.

## Main scenario (happy path)

1. Operator clicks **Export** and chooses a zip path.
2. The zip root is `commands.yaml`, even when the saved path is custom.
3. Each used media file keeps its path relative to the media folder. Unused files are omitted.
4. Operator clicks **Import**, chooses that zip, then a folder.
5. Files unpack into that folder. The media folder becomes that path. A custom commands path is cleared.

## Alternative scenarios and errors

* **Export cancel:** disk is unchanged.
* **Import cancel:** config is unchanged. Cancelling the folder dialog does not unpack.
* **Unsaved editor rows:** Export reads the file on disk.
* **Missing commands file:** Export asks to save first.
* **Missing media file, or a path outside the media folder:** Export fails. No zip is left behind.
* **A command file named `commands.yaml`:** Export rejects that name.
* **Zip has no root `commands.yaml`:** Import rejects the archive.

## Postconditions

* A successful import uses `<media folder>/commands.yaml`.
* YouTube Start is not required.
