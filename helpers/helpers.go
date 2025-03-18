package helpers

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

var ErrorEmptyHis = errors.New("HYPRLAND_INSTANCE_SIGNATURE is empty")

// Returns a Hyprland socket path.
func GetSocket(socket Socket) (string, error) {
	his := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if his == "" {
		return "", fmt.Errorf("%w, are you using Hyprland?", ErrorEmptyHis)
	}

	// https://github.com/hyprwm/Hyprland/blob/83a5395eaa99fecef777827fff1de486c06b6180/hyprctl/main.cpp#L53-L62
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")

	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("error while getting the current user: %w", err)
	}

	if runtimeDir == "" {
		user := u.Uid
		runtimeDir = filepath.Join("/run/user", user)
	}

	return filepath.Join(runtimeDir, "hypr", his, string(socket)), nil
}
