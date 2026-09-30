package sidebar

import (
	"testing"

	"github.com/HardDie/ytmemchat_wails/pkg/version"
)

func TestAppVersion_defaultDev(t *testing.T) {
	if New().AppVersion() != "dev" {
		t.Fatalf("version = %q", New().AppVersion())
	}
}

func TestAppVersion_usesStamp(t *testing.T) {
	old := version.Build
	t.Cleanup(func() { version.Build = old })
	version.Build = "v1.2.3"
	if New().AppVersion() != "v1.2.3" {
		t.Fatalf("version = %q", New().AppVersion())
	}
}
