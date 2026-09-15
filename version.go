package main

import "strings"

// buildVersion is the git tag or short commit stamped at link time
// (`-X main.buildVersion=…`). Unset builds show "dev".
var buildVersion = "dev"

func versionString() string {
	v := strings.TrimSpace(buildVersion)
	if v == "" {
		return "dev"
	}
	return v
}

// AppVersion is the git tag or commit the binary was built from.
func (a *App) AppVersion() string {
	return versionString()
}
