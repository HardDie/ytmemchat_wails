package update

import (
	"fmt"
	"strings"
)

// AssetHint is the substring that identifies this OS archive in a release name.
func AssetHint(goos, goarch string) (string, error) {
	switch goos {
	case "darwin":
		return "darwin-universal.zip", nil
	case "windows":
		if goarch != "amd64" {
			return "", fmt.Errorf("%w: windows/%s", ErrNoAsset, goarch)
		}
		return "windows-amd64.zip", nil
	case "linux":
		switch goarch {
		case "amd64", "arm64":
			return "linux-" + goarch + ".tar.gz", nil
		default:
			return "", fmt.Errorf("%w: linux/%s", ErrNoAsset, goarch)
		}
	default:
		return "", fmt.Errorf("%w: %s", ErrNoAsset, goos)
	}
}

func pickAsset(assets []ghAsset, goos, goarch string) (ghAsset, error) {
	hint, err := AssetHint(goos, goarch)
	if err != nil {
		return ghAsset{}, err
	}
	for _, a := range assets {
		if strings.Contains(a.Name, hint) {
			return a, nil
		}
	}
	return ghAsset{}, fmt.Errorf("%w (%s)", ErrNoAsset, hint)
}

func pickSums(assets []ghAsset) (ghAsset, error) {
	for _, a := range assets {
		if a.Name == "SHA256SUMS.txt" || strings.HasSuffix(a.Name, "SHA256SUMS.txt") {
			return a, nil
		}
	}
	return ghAsset{}, fmt.Errorf("update: SHA256SUMS.txt missing")
}
