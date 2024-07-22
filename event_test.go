package hyprland

import (
	"fmt"
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
	if ec == nil {
		t.Skip("HYPRLAND_INSTANCE_SIGNATURE not set, skipping test")
	}
	msg, err := ec.Receive()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(msg)
}
