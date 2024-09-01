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

	u, err := user.Current()
	assert.NoError(t, err)

	socket, err := GetSocket(RequestSocket)
	assert.NoError(t, err)
	assert.Equal(t, socket, "/run/user/"+u.Uid+"/hypr/bar/.socket.sock")

	socket, err = GetSocket(EventSocket)
	assert.NoError(t, err)
	assert.Equal(t, socket, "/run/user/"+u.Uid+"/hypr/bar/.socket2.sock")
}

func TestGetSocketError(t *testing.T) {
	t.Setenv("HYPRLAND_INSTANCE_SIGNATURE", "")

	_, err := GetSocket(RequestSocket)
	assert.Error(t, err)
}
