package sidebar

import "strings"

// buildVersion is the git tag or short commit stamped at link time
// (`-X github.com/HardDie/ytmemchat_wails/bindings/sidebar.buildVersion=…`).
// Unset builds show "dev".
var buildVersion = "dev"

func versionString() string {
	v := strings.TrimSpace(buildVersion)
	if v == "" {
		return "dev"
	}
	return v
}

// AppVersion is the git tag or commit the binary was built from.
func (s *Sidebar) AppVersion() string {
	return versionString()
}
