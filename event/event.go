package event

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/thiagokokada/hyprland-go/helpers"
	"github.com/thiagokokada/hyprland-go/internal/assert"
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
// will not panic on error, use [NewClient] instead.
func MustClient() *EventClient {
	return assert.Must1(NewClient(
		assert.Must1(helpers.GetSocket(helpers.EventSocket))),
	)
}

// Initiate a new event client.
// Receive as parameters a socket that is generally localised in
// '$XDG_RUNTIME_DIR/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket2.sock'.
// The ctx ([context.Context]) parameter is passed to the underlying socket to
// allow cancellations and timeouts for the Hyprland event socket.
func NewClient(socket string) (*EventClient, error) {
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to socket: %w", err)
	}
	return &EventClient{conn: conn}, err
}

// Close the underlying connection.
func (c *EventClient) Close() error {
	err := c.conn.Close()
	if err != nil {
		return fmt.Errorf("error while closing socket: %w", err)
	}
	return err
}

// Low-level receive event method, should be avoided unless there is no
// alternative.
func (c *EventClient) Receive(ctx context.Context) ([]ReceivedData, error) {
	buf := make([]byte, bufSize)

	n, err := readWithContext(ctx, c.conn, buf)
	if err != nil {
		return nil, fmt.Errorf("error while reading from socket: %w", err)
	}

	buf = buf[:n]

	var recv []ReceivedData
	raw := strings.Split(string(buf), "\n")
	for _, event := range raw {
		if event == "" {
			continue
		}

		split := strings.Split(event, sep)
		if len(split) < 2 || split[0] == "" || split[1] == "" || split[1] == "," {
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
func (c *EventClient) Subscribe(ctx context.Context, ev EventHandler, events ...EventType) error {
	for {
		// Process an event
		if err := receiveAndProcessEvent(ctx, c, ev, events...); err != nil {
			return fmt.Errorf("event processing: %w", err)
		}
	}
}

func readWithContext(ctx context.Context, conn net.Conn, buf []byte) (n int, err error) {
	done := make(chan struct{})

	// Start a goroutine to perform the read
	go func() {
		n, err = conn.Read(buf)
		close(done)
	}()

	select {
	case <-done:
		return n, err
	case <-ctx.Done():
		// Set a short deadline to unblock the Read()
		err = conn.SetReadDeadline(time.Now())
		if err != nil {
			return 0, err
		}
		// Reset read deadline
		defer func() {
			if e := conn.SetReadDeadline(time.Time{}); e != nil {
				err = errors.Join(err, e)
			}
		}()
		// Make sure that the goroutine is done to avoid leaks
		<-done
		return 0, errors.Join(err, ctx.Err())
	}
}

func receiveAndProcessEvent(ctx context.Context, c eventClient, ev EventHandler, events ...EventType) error {
	msg, err := c.Receive(ctx)
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
				// e.g. nvim,nvim event/event.go
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
