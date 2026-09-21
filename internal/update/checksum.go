package update

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"strings"
)

func checksumFor(sums, assetName string) (string, error) {
	sc := bufio.NewScanner(strings.NewReader(sums))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		sum, name := fields[0], fields[len(fields)-1]
		name = strings.TrimPrefix(name, "*")
		if name == assetName || strings.HasSuffix(assetName, name) || strings.HasSuffix(name, assetName) {
			if _, err := hex.DecodeString(sum); err != nil || len(sum) != 64 {
				return "", fmt.Errorf("update: bad checksum line")
			}
			return strings.ToLower(sum), nil
		}
	}
	if err := sc.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("update: %s not in SHA256SUMS.txt", assetName)
}
