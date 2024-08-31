package helpers

import (
	"os"
	"os/user"
	"path/filepath"

	"github.com/thiagokokada/hyprland-go/internal/assert"
)

// Returns a Hyprland socket or panics.
func MustSocket(socket string) string {
	his := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if his == "" {
		panic("HYPRLAND_INSTANCE_SIGNATURE is empty, are you using Hyprland?")
	}

	// https://github.com/hyprwm/Hyprland/blob/83a5395eaa99fecef777827fff1de486c06b6180/hyprctl/main.cpp#L53-L62
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		user := assert.Must1(user.Current()).Uid
		runtimeDir = filepath.Join("/run/user", user)
	}
	return filepath.Join(runtimeDir, "hypr", his, socket)
}
