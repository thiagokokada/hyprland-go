package hyprland

import (
	"fmt"
	"strings"
	"testing"
)

var client = MustClient()

func TestMakeRequest(t *testing.T) {
	// missing command
	_, err := makeRequest("", nil)
	if err == nil {
		t.Error("should have been an error")
	}

	// test params
	tests := []struct {
		command  string
		params   []string
		expected string
	}{
		{"command", nil, "command"},
		{"command", []string{"param0"}, "command param0"},
		{"command", []string{"param0", "param1"}, "[[BATCH]]command param0;command param1;"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v-%v", tt.command, tt.params), func(t *testing.T) {
			request, err := makeRequest(tt.command, tt.params)
			if err != nil {
				t.Error(err)
			}
			if string(request) != tt.expected {
				t.Errorf("got: %s, want: %s", request, tt.expected)
			}
		})
	}
}

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

func TestDispatch(t *testing.T) {
	response, err := client.Dispatch("exec kitty")
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
