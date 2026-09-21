// Package config loads and saves desktop settings as JSON under the OS
// user config directory. It does not read .env files and does not panic
// on a missing file (unlike the console app).
//
// Default path: os.UserConfigDir()/ytmemchat/config.json (mode 0600).
// The YouTube API key is stored in the OS keychain when that vault works.
// Bindings on the Configuration pane should call [Store.Load] / [Store.Save]; this package
// does not import Wails.
package config
