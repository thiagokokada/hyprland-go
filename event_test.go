package hyprland

import (
	"os"
	"testing"
)

var ec *EventClient

func init() {
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		ec = MustEventClient()
	}
}

func TestReceive(t *testing.T) {
	t.Skip("temporary disabled")

	msg, err := ec.Receive()
	if err != nil {
		t.Error(err)
	}
	t.Log(msg)
}
