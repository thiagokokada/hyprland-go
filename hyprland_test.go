package hyprland

import (
	"fmt"
	"os"
	"reflect"
	"testing"
)

var client *IPCClient

func init() {
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		client = MustClient()
	}
}

func genParams(param string, nParams int) (params []string) {
	for i := 0; i < nParams; i++ {
		params = append(params, param)
	}
	return params
}

func TestMakeRequest(t *testing.T) {
	// missing command
	_, err := prepareRequests("", nil)
	if err == nil {
		t.Error("should have been an error")
	}

	// test params
	tests := []struct {
		command  string
		params   []string
		expected []string
	}{
		{"command", nil, []string{"command"}},
		{"command", []string{"param0"}, []string{"command param0"}},
		{"command", []string{"param0", "param1"}, []string{"[[BATCH]]command param0;command param1;"}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("tests_%v-%v", tt.command, tt.params), func(t *testing.T) {
			requests, err := prepareRequests(tt.command, tt.params)
			if err != nil {
				t.Error(err)
			}
			for i, e := range tt.expected {
				if string(requests[i]) != e {
					t.Errorf("got: %s, want: %s", requests[i], e)
				}
			}
		})
	}

	// test massive amount of parameters
	massTests := []struct {
		command  string
		params   []string
		expected int
	}{
		{"command", genParams("param", 5), 1},
		{"command", genParams("param", 15), 1},
		{"command", genParams("param", 30), 1},
		{"command", genParams("param", 60), 2},
		{"command", genParams("param", 90), 3},
		{"command", genParams("param", 100), 4},
	}
	for _, tt := range massTests {
		t.Run(fmt.Sprintf("mass_tests_%v-%d", tt.command, len(tt.params)), func(t *testing.T) {
			requests, err := prepareRequests(tt.command, tt.params)
			if err != nil {
				t.Error(err)
			}
			if len(requests) != tt.expected {
				t.Errorf("got: %d, want: %d", len(requests), tt.expected)
			}
		})
	}
}

func TestRequest(t *testing.T) {
	if client == nil {
		t.Skip("HYPRLAND_INSTANCE_SIGNATURE not set, skipping test")
	}

	response, err := client.Request([]byte("dispatch exec"))
	if err != nil {
		t.Error(err)
	}
	if len(response) == 0 {
		t.Error("empty response")
	}
}

func TestDispatch(t *testing.T) {
	if client == nil {
		t.Skip("HYPRLAND_INSTANCE_SIGNATURE not set, skipping test")
	}

	err := client.Dispatch("exec kitty")
	if err != nil {
		t.Error(err)
	}
}

func TestDispatchMassive(t *testing.T) {
	if client == nil {
		t.Skip("HYPRLAND_INSTANCE_SIGNATURE not set, skipping test")
	}
	if testing.Short() {
		t.Skip("skipping slow test")
	}

	err := client.Dispatch(genParams("exec kitty", 40)...)
	if err != nil {
		t.Error(err)
	}
}

func TestReload(t *testing.T) {
	if client == nil {
		t.Skip("HYPRLAND_INSTANCE_SIGNATURE not set, skipping test")
	}

	err := client.Reload()
	if err != nil {
		t.Error(err)
	}
}

func TestActiveWorkspace(t *testing.T) {
	if client == nil {
		t.Skip("HYPRLAND_INSTANCE_SIGNATURE not set, skipping test")
	}

	got, err := client.ActiveWorkspace()
	if err != nil {
		t.Error(err)
	}
	if reflect.DeepEqual(got, Workspace{}) {
		t.Error("got empty struct")
	}
}

func TestActiveWindow(t *testing.T) {
	if client == nil {
		t.Skip("HYPRLAND_INSTANCE_SIGNATURE not set, skipping test")
	}

	got, err := client.ActiveWindow()
	if err != nil {
		t.Error(err)
	}
	if reflect.DeepEqual(got, Window{}) {
		t.Error("got empty struct")
	}
}

func TestClient(t *testing.T) {
	if client == nil {
		t.Skip("HYPRLAND_INSTANCE_SIGNATURE not set, skipping test")
	}

	got, err := client.Clients()
	if err != nil {
		t.Error(err)
	}
	if len(got) == 0 {
		t.Error("got empty response")
	}
}

func TestGetOption(t *testing.T) {
	if client == nil {
		t.Skip("HYPRLAND_INSTANCE_SIGNATURE not set, skipping test")
	}

	tests := []struct{ option string }{
		{"general:border_size"},
		{"gestures:workspace_swipe"},
		{"misc:vrr"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("mass_tests_%v", tt.option), func(t *testing.T) {
			got, err := client.GetOption(tt.option)
			if err != nil {
				t.Error(err)
			}
			if reflect.DeepEqual(got, Option{}) {
				t.Error("got empty struct")
			}
		})
	}
}

func TestKill(t *testing.T) {
	if client == nil {
		t.Skip("HYPRLAND_INSTANCE_SIGNATURE not set, skipping test")
	}

	err := client.Kill()
	if err != nil {
		t.Error(err)
	}
}

func TestVersion(t *testing.T) {
	if client == nil {
		t.Skip("HYPRLAND_INSTANCE_SIGNATURE not set, skipping test")
	}

	got, err := client.Version()
	if err != nil {
		t.Error(err)
	}
	if reflect.DeepEqual(got, Version{}) {
		t.Error("got empty struct")
	}
}

func TestSplash(t *testing.T) {
	if client == nil {
		t.Skip("HYPRLAND_INSTANCE_SIGNATURE not set, skipping test")
	}

	got, err := client.Splash()
	if err != nil {
		t.Error(err)
	}
	if len(got) == 0 {
		t.Error("got empty response")
	}
}

func TestResponseValidation(t *testing.T) {
	if client == nil {
		t.Skip("HYPRLAND_INSTANCE_SIGNATURE not set, skipping test")
	}
	// Create its own client to avoid messing the global one
	client := MustClient()

	// With client.Validate = true, should fail this response
	client.Validate = true
	err := client.Dispatch("oops")
	if err == nil {
		t.Error("nil error")
	}

	// With client.Validate = false, should not fail this response
	client.Validate = false
	err = client.Dispatch("oops")
	if err != nil {
		t.Error(err)
	}
}
