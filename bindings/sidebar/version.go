package sidebar

import "github.com/HardDie/ytmemchat_wails/pkg/version"

// AppVersion is the git tag or commit the binary was built from.
func (s *Sidebar) AppVersion() string {
	return version.String()
}
