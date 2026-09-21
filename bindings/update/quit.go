//go:build !nomain

package update

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func quit(ctx context.Context) {
	if ctx == nil {
		return
	}
	runtime.Quit(ctx)
}
