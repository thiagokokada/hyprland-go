package helpers

import (
	"os/user"
	"testing"

	"github.com/thiagokokada/hyprland-go/internal/assert"
)

func TestGetSocketWithXdg(t *testing.T) {
	t.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "foo")
	t.Setenv("XDG_RUNTIME_DIR", "/xdg")

	socket, err := GetSocket(RequestSocket)
	assert.NoError(t, err)
	assert.Equal(t, socket, "/xdg/hypr/foo/.socket.sock")

	socket, err = GetSocket(EventSocket)
	assert.NoError(t, err)
	assert.Equal(t, socket, "/xdg/hypr/foo/.socket2.sock")
}

func TestGetSocketWithoutXdg(t *testing.T) {
	t.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "bar")
	t.Setenv("XDG_RUNTIME_DIR", "")

	socket, err := GetSocket(RequestSocket)
	assert.NoError(t, err)
	assert.Equal(t, socket, "/run/user/"+getUid(t)+"/hypr/bar/.socket.sock")

	socket, err = GetSocket(EventSocket)
	assert.NoError(t, err)
	assert.Equal(t, socket, "/run/user/"+getUid(t)+"/hypr/bar/.socket2.sock")
}

func getUid(t *testing.T) string {
	t.Helper()
	u, err := user.Current()
	assert.NoError(t, err)
	return u.Uid
}
