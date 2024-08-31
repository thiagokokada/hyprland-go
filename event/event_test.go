package event

import (
	"os"
	"testing"
	"time"

	"github.com/thiagokokada/hyprland-go"
	"github.com/thiagokokada/hyprland-go/internal/assert"
)

type FakeEventClient struct {
	EventClient
}

type FakeEventHandler struct {
	t *testing.T
	EventHandler
}

func TestReceive(t *testing.T) {
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") == "" {
		t.Skip("HYPRLAND_INSTANCE_SIGNATURE not set, skipping test")
	}

	// Generate an event
	go func() {
		c := hyprland.MustClient()
		time.Sleep(100 * time.Millisecond)
		c.Dispatch("exec kitty sh -c 'echo Testing hyprland-go events && sleep 1'")
	}()

	// We must capture this event
	c := MustEventClient()
	data, err := c.Receive()

	assert.NoError(t, err)
	assert.True(t, len(data) >= 0)
	for _, d := range data {
		assert.NotEqual(t, string(d.Data), "")
		assert.NotEqual(t, string(d.Type), "")
	}
}

func TestSubscribe(t *testing.T) {
	h := &FakeEventHandler{t: t}
	c := &FakeEventClient{}
	err := subscribeOnce(c, h, AllEvents...)
	assert.NoError(t, err)
}

func (f *FakeEventClient) Receive() ([]ReceivedData, error) {
	return []ReceivedData{
		{
			Type: EventWorkspace,
			Data: "1",
		},
		{
			Type: EventFocusedMonitor,
			Data: "1,1",
			// TODO I only have one monitor, so I didn't check this
		},
		{
			Type: EventActiveWindow,
			Data: "nvim,nvim event/event_test.go",
		},
		{
			Type: EventFullscreen,
			Data: "1",
		},
		{
			Type: EventMonitorRemoved,
			Data: "1",
			// TODO I only have one monitor, so I didn't check this
		},
		{
			Type: EventMonitorAdded,
			Data: "1",
			// TODO I only have one monitor, so I didn't check this
		},
		{
			Type: EventCreateWorkspace,
			Data: "1",
		},
		{
			Type: EventDestroyWorkspace,
			Data: "1",
		},

		{
			Type: EventMoveWorkspace,
			Data: "1,1",
			// TODO I only have one monitor, so I didn't check this
		},
		{
			Type: EventActiveLayout,
			Data: "AT Translated Set 2 keyboard,Russian",
		},
		{
			Type: EventOpenWindow,
			Data: "80e62df0,2,jetbrains-goland,win430",
		},
		{
			Type: EventCloseWindow,
			Data: "80e62df0",
		},
		{
			Type: EventMoveWindow,
			Data: "80e62df0,1",
		},
		{
			Type: EventOpenLayer,
			Data: "wofi",
		},
		{
			Type: EventCloseLayer,
			Data: "wofi",
		},
		{
			Type: EventSubMap,
			Data: "1",
			// idk
		},
		{
			Type: EventScreencast,
			Data: "1,0",
		},
	}, nil
}

func (h *FakeEventHandler) Workspace(w WorkspaceName) {
	assert.Equal(h.t, w, "1")
}

func (h *FakeEventHandler) FocusedMonitor(m FocusedMonitor) {
	assert.Equal(h.t, m.WorkspaceName, "1")
	assert.Equal(h.t, m.MonitorName, "1")
}

func (h *FakeEventHandler) ActiveWindow(w ActiveWindow) {
	assert.Equal(h.t, w.Name, "nvim")
	assert.Equal(h.t, w.Title, "nvim event/event_test.go")
}

func (h *FakeEventHandler) Fullscreen(f Fullscreen) {
	assert.Equal(h.t, f, true)
}

func (h *FakeEventHandler) MonitorRemoved(m MonitorName) {
	assert.Equal(h.t, m, "1")
}

func (h *FakeEventHandler) MonitorAdded(m MonitorName) {
	assert.Equal(h.t, m, "1")
}

func (h *FakeEventHandler) CreateWorkspace(w WorkspaceName) {
	assert.Equal(h.t, w, "1")
}

func (h *FakeEventHandler) DestroyWorkspace(w WorkspaceName) {
	assert.Equal(h.t, w, "1")
}

func (h *FakeEventHandler) MoveWorkspace(w MoveWorkspace) {
	assert.Equal(h.t, w.WorkspaceName, "1")
	assert.Equal(h.t, w.MonitorName, "1")
}

func (h *FakeEventHandler) ActiveLayout(l ActiveLayout) {
	assert.Equal(h.t, l.Name, "Russian")
	assert.Equal(h.t, l.Type, "AT Translated Set 2 keyboard")
}

func (h *FakeEventHandler) OpenWindow(o OpenWindow) {
	assert.Equal(h.t, o.Address, "80e62df0")
	assert.Equal(h.t, o.Class, "jetbrains-goland")
	assert.Equal(h.t, o.Title, "win430")
	assert.Equal(h.t, o.WorkspaceName, "2")
}

func (h *FakeEventHandler) CloseWindow(c CloseWindow) {
	assert.Equal(h.t, c.Address, "80e62df0")
}

func (h *FakeEventHandler) MoveWindow(m MoveWindow) {
	assert.Equal(h.t, m.Address, "80e62df0")
	assert.Equal(h.t, m.WorkspaceName, "1")
}

func (h *FakeEventHandler) OpenLayer(l OpenLayer) {
	assert.Equal(h.t, l, "wofi")
}

func (h *FakeEventHandler) CloseLayer(l CloseLayer) {
	assert.Equal(h.t, l, "wofi")
}

func (h *FakeEventHandler) SubMap(s SubMap) {
	assert.Equal(h.t, s, "1")
}

func (h *FakeEventHandler) Screencast(s Screencast) {
	assert.Equal(h.t, s.Owner, "0")
	assert.Equal(h.t, s.Sharing, true)
}
