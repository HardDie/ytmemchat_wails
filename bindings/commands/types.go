package commands

// AlertCommand is one row in commands.yaml for the editor.
type AlertCommand struct {
	// Name is the token word (for example "jump" in @jump).
	Name string `json:"name"`
	// File is a media filename inside the alerts media folder.
	File string `json:"file"`
	// Volume is overlay gain 0–1. Nil omits the YAML key (runtime default 1).
	Volume *float64 `json:"volume,omitempty"`
	// Scale is visual size. Nil omits the YAML key (runtime default 1).
	Scale *float64 `json:"scale,omitempty"`
}

// AlertCommandsFile is the commands.yaml editor payload.
type AlertCommandsFile struct {
	// Path is the saved commands.yaml location.
	Path string `json:"path"`
	// Commands is the list to show and write.
	Commands []AlertCommand `json:"commands"`
}
