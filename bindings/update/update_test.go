package update

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	intupdate "github.com/HardDie/ytmemchat_wails/internal/update"
)

type stub struct{}

func (stub) DialogContext() context.Context { return context.Background() }

func TestCheck_needsHTTP(t *testing.T) {
	u := New(stub{})
	u.c.API = "http://127.0.0.1:1"
	u.c.HTTP = &http.Client{Timeout: 200 * time.Millisecond}
	_, err := u.Check()
	if err == nil {
		t.Fatal("want error")
	}
}

func TestDownload_beforeCheck(t *testing.T) {
	u := New(stub{})
	if err := u.Download(); !errors.Is(err, intupdate.ErrNotReady) {
		t.Fatalf("err = %v", err)
	}
}

func TestApply_beforeDownload(t *testing.T) {
	u := New(stub{})
	if err := u.ApplyAndQuit(); !errors.Is(err, intupdate.ErrNoDownload) {
		t.Fatalf("err = %v", err)
	}
}
