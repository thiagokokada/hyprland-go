package hyprland

import (
	"strings"
	"testing"
)

var client = MustClient()

func TestRequest(t *testing.T) {
	response, err := client.Request([]byte("dispatch exec kitty"))
	if err != nil {
		t.Error(err)
	}
	if len(response) == 0 {
		t.Error("empty response")
	}
	trimmedResponse := strings.TrimSpace(string(response))
	if trimmedResponse != "ok" {
		t.Errorf("non-ok response: %s\n", trimmedResponse)
	}
}
