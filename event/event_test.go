package event

import (
	"testing"
)

var ec *EventClient

var (
	h = &FakeEventHandler{}
	c = &FakeEventClient{}
)

type FakeEventClient struct {
	EventClient
}

type FakeEventHandler struct {
	DummyEvHandler
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
			Data: "jetbrains-goland,hyprland-ipc-ipc – ipc.go",
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

func TestSubscribe(t *testing.T) {
	err := c.SubscribeWithoutLoop(h, GetAllEvents()...)
	if err != nil {
		t.Error(err)
	}
}

func (c *FakeEventClient) SubscribeWithoutLoop(ev EventHandler, events ...EventType) error {
	msg, err := c.Receive()
	if err != nil {
		return err
	}

	for _, data := range msg {
		processEvent(ev, data, events)
	}

	return nil
}
