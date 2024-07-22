package hyprland

import "net"

type RawData string

type EventType string

type ReceivedData struct {
	Type EventType
	Data RawData
}

// EventClient is the event struct from hyprland-go.
// Experimental: WIP
type EventClient struct {
	conn net.Conn
}
