package k3s

import (
	"fmt"
)

func k3sArch(nodeArch string) (string, error) {
	switch nodeArch {
	case "amd64", "x86_64":
		return "amd64", nil
	case "arm64", "aarch64":
		return "arm64", nil
	case "arm", "armv7", "armv7l", "armhf":
		return "armhf", nil
	default:
		return "", fmt.Errorf("unsupported k3s arch %q", nodeArch)
	}
}
