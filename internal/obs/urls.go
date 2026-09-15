package obs

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// ChatSourceURL is the OBS Browser Source URL for chat (host 127.0.0.1).
func ChatSourceURL(addr string) string {
	return httpBase(addr) + PathChat
}

// OverlaySourceURL is the OBS Browser Source URL for alerts and TTS.
func OverlaySourceURL(addr string) string {
	return httpBase(addr) + PathOverlay
}

func httpBase(addr string) string {
	addr = strings.TrimSpace(addr)
	host, port, err := net.SplitHostPort(strings.TrimPrefix(addr, "http://"))
	if err != nil {
		// "8080" with no host
		if p, perr := strconv.Atoi(strings.TrimPrefix(addr, ":")); perr == nil && p > 0 {
			return fmt.Sprintf("http://127.0.0.1:%d", p)
		}
		return "http://127.0.0.1"
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return "http://" + net.JoinHostPort(host, port)
}
