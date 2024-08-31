package event

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"github.com/thiagokokada/hyprland-go/internal/assert"
	"github.com/thiagokokada/hyprland-go/internal/helpers"
)

const (
	bufSize = 8192
	sep     = ">>"
)

// Initiate a new client or panic.
// This should be the preferred method for user scripts, since it will
// automatically find the proper socket to connect and use the
// HYPRLAND_INSTANCE_SIGNATURE for the current user.
// If you need to connect to arbitrary user instances or need a method that
// will not panic on error, use [NewEventClient] instead.
func MustEventClient() *EventClient {
	return assert.Must1(NewEventClient(helpers.MustSocket(".socket2.sock")))
}

// Initiate a new event client.
// Receive as parameters a socket that is generally localised in
// '$XDG_RUNTIME_DIR/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket2.sock'.
func NewEventClient(socket string) (*EventClient, error) {
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to socket: %w", err)
	}
	return &EventClient{conn: conn}, nil
}

// Close the underlying connection.
func (c *EventClient) Close() error {
	return c.conn.Close()
}

// Low-level receive event method, should be avoided unless there is no
// alternative.
func (c *EventClient) Receive() ([]ReceivedData, error) {
	buf := make([]byte, bufSize)
	reader := bufio.NewReader(c.conn)
	n, err := reader.Read(buf)
	if err != nil {
		return nil, err
	}

	buf = buf[:n]

	var recv []ReceivedData
	raw := strings.Split(string(buf), "\n")
	for _, event := range raw {
		if event == "" {
			continue
		}

		split := strings.Split(event, sep)
		if split[0] == "" || split[1] == "" || split[1] == "," {
			continue
		}

		recv = append(recv, ReceivedData{
			Type: EventType(split[0]),
			Data: RawData(split[1]),
		})
	}

	return recv, nil
}

// Subscribe to events.
// You need to pass an implementation of [EventHandler] interface for each of
// the events you want to handle and all event types you want to handle.
func Subscribe(c *EventClient, ev EventHandler, events ...EventType) error {
	for {
		subscribeOnce(c, ev, events...)
	}
}

func subscribeOnce(c eventClient, ev EventHandler, events ...EventType) error {
	msg, err := c.Receive()
	if err != nil {
		return err
	}

	for _, data := range msg {
		processEvent(ev, data, events)
	}

	return nil
}

func processEvent(ev EventHandler, msg ReceivedData, events []EventType) {
	for _, event := range events {
		raw := strings.Split(string(msg.Data), ",")
		if msg.Type == event {
			switch event {
			case EventWorkspace:
				// e.g. "1" (workspace number)
				ev.Workspace(WorkspaceName(raw[0]))
			case EventFocusedMonitor:
				// idk
				ev.FocusedMonitor(FocusedMonitor{
					MonitorName:   MonitorName(raw[0]),
					WorkspaceName: WorkspaceName(raw[1]),
				})
			case EventActiveWindow:
				// e.g. jetbrains-goland,hyprland-ipc-ipc – main.go
				ev.ActiveWindow(ActiveWindow{
					Name:  raw[0],
					Title: raw[1],
				})
			case EventFullscreen:
				// e.g. "true" or "false"
				ev.Fullscreen(raw[0] == "1")
			case EventMonitorRemoved:
				// e.g. idk
				ev.MonitorRemoved(MonitorName(raw[0]))
			case EventMonitorAdded:
				// e.g. idk
				ev.MonitorAdded(MonitorName(raw[0]))
			case EventCreateWorkspace:
				// e.g. "1" (workspace number)
				ev.CreateWorkspace(WorkspaceName(raw[0]))
			case EventDestroyWorkspace:
				// e.g. "1" (workspace number)
				ev.DestroyWorkspace(WorkspaceName(raw[0]))
			case EventMoveWorkspace:
				// e.g. idk
				ev.MoveWorkspace(MoveWorkspace{
					WorkspaceName: WorkspaceName(raw[0]),
					MonitorName:   MonitorName(raw[1]),
				})
			case EventActiveLayout:
				// e.g. AT Translated Set 2 keyboard,Russian
				ev.ActiveLayout(ActiveLayout{
					Type: raw[0],
					Name: raw[1],
				})
			case EventOpenWindow:
				// e.g. 80864f60,1,Alacritty,Alacritty
				ev.OpenWindow(OpenWindow{
					Address:       raw[0],
					WorkspaceName: WorkspaceName(raw[1]),
					Class:         raw[2],
					Title:         raw[3],
				})
			case EventCloseWindow:
				// e.g. 5
				ev.CloseWindow(CloseWindow{
					Address: raw[0],
				})
			case EventMoveWindow:
				// e.g. 5
				ev.MoveWindow(MoveWindow{
					Address:       raw[0],
					WorkspaceName: WorkspaceName(raw[1]),
				})
			case EventOpenLayer:
				// e.g. wofi
				ev.OpenLayer(OpenLayer(raw[0]))
			case EventCloseLayer:
				// e.g. wofi
				ev.CloseLayer(CloseLayer(raw[0]))
			case EventSubMap:
				// e.g. idk
				ev.SubMap(SubMap(raw[0]))
			case EventScreencast:
				ev.Screencast(Screencast{
					Sharing: raw[0] == "1",
					Owner:   raw[1],
				})
			}
		}
	}
}
