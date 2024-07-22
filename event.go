package hyprland

import (
	"fmt"
	"net"
	"strings"
)

const SEPARATOR = ">>"

func MustEventClient() *EventClient {
	return must1(NewEventClient(mustSocket(".socket2.sock")))
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

// Low-level receive event method, should be avoided unless there is no
// alternative.
func (c *EventClient) Receive() ([]ReceivedData, error) {
	buf := make([]byte, BUF_SIZE)
	n, err := c.conn.Read(buf)
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

		split := strings.Split(event, SEPARATOR)
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
