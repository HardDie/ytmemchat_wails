// Package alerts matches chat text against commands.yaml and emits overlay
// media clips. It does not start HTTP and does not import Wails. Serving
// files under /obs/media/ belongs in obs, wired from app.go using MediaPath.
package alerts
