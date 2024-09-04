package event

import (
	"bufio"
	"context"
	"errors"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"

	"github.com/thiagokokada/hyprland-go"
	"github.com/thiagokokada/hyprland-go/internal/assert"
)

const socketPath = "/tmp/bench_unix_socket.sock"

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

	c := MustClient()
	defer c.Close()
	data, err := c.Receive(context.Background())

	// We must capture the event
	assert.NoError(t, err)
	assert.True(t, len(data) >= 0)
	for _, d := range data {
		assert.NotEqual(t, string(d.Data), "")
		assert.NotEqual(t, string(d.Type), "")
	}
}

func TestSubscribe(t *testing.T) {
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") == "" {
		t.Skip("HYPRLAND_INSTANCE_SIGNATURE not set, skipping test")
	}
	c := MustClient()
	defer c.Close()

	// Make sure that we can exit a Subscribe loop by cancelling the
	// context
	ctx, cancel := context.WithTimeout(
		context.Background(),
		100*time.Millisecond,
	)

	err := c.Subscribe(ctx, &DefaultEventHandler{}, AllEvents...)
	cancel()

	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))

	// Make sure that we can call Subscribe again it can still be used,
	// e.g.: the conn read deadline is not set otherwise it will exit
	// immediatelly
	ctx, cancel = context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err = c.Subscribe(ctx, &DefaultEventHandler{}, AllEvents...)
	elapsed := time.Since(start)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
	assert.True(t, elapsed >= 100*time.Millisecond)
}

func TestProcessEvent(t *testing.T) {
	h := &FakeEventHandler{t: t}
	c := &FakeEventClient{}
	err := receiveAndProcessEvent(context.Background(), c, h, AllEvents...)
	assert.NoError(t, err)
}

func (f *FakeEventClient) Receive(context.Context) ([]ReceivedData, error) {
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

func BenchmarkReceive(b *testing.B) {
	go RandomStringServer()

	// Make sure the socket exist
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		if _, err := os.Stat(socketPath); err != nil {
			break
		}
	}

	c := assert.Must1(NewClient(socketPath))
	defer c.Close()

	ctx := context.Background()

	// Reset setup time
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Receive(ctx)
	}
}

// This function needs to be as fast as possible, otherwise this is the
// bottleneck
// https://stackoverflow.com/a/31832326
const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ\n"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandomBytes(n int) []byte {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return b
}

func RandomStringServer() {
	// Remove the previous socket file if it exists
	if err := os.RemoveAll(socketPath); err != nil {
		panic(err)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		writer := bufio.NewWriter(conn)

		go func(c net.Conn) {
			defer c.Close()

			for {
				prefix := []byte(">>>")
				randomData := RandomBytes(16)
				message := append(prefix, randomData...)

				// Send the message to the client
				_, err := writer.Write(message)
				if err != nil {
					return
				}
			}
		}(conn)
	}
}
