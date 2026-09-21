// Package sidebar is the Wails binding for the window sidebar.
package sidebar

// Sidebar exposes chrome that is not tied to a pane.
type Sidebar struct{}

// New returns the sidebar binding.
func New() *Sidebar {
	return &Sidebar{}
}
