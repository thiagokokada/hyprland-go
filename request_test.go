package hyprland

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/thiagokokada/hyprland-go/internal/assert"
)

var c *RequestClient

type DummyClient struct {
	RequestClient
}

func genParams(param string, n int) (params []string) {
	for i := 0; i < n; i++ {
		params = append(params, param)
	}
	return params
}

func checkEnvironment(t *testing.T) {
	if c == nil {
		t.Skip("HYPRLAND_INSTANCE_SIGNATURE not set, skipping test")
	}
}

func testCommandRR(t *testing.T, command func() (RawResponse, error)) {
	testCommand(t, command, RawResponse(""))
}

func testCommand[T any](t *testing.T, command func() (T, error), emptyValue any) {
	checkEnvironment(t)
	got, err := command()
	assert.NoError(t, err)
	assert.Equal(t, reflect.TypeOf(got), reflect.TypeOf(emptyValue))
	assert.True(t, reflect.DeepEqual(got, emptyValue))
	t.Logf("got: %+v", got)
}

func setup() {
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		c = MustClient()
	}
}

func teardown() {
	if c != nil {
		// Make sure that the Hyprland config is in a sane state
		assert.Must1(c.Reload())
	}
}

func TestMain(m *testing.M) {
	setup()
	defer teardown()

	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestPrepareRequests(t *testing.T) {
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
			assert.NoError(t, err)
			for i, e := range tt.expected {
				assert.Equal(t, string(requests[i]), e)
			}
		})
	}
}

func TestPrepareRequestsMass(t *testing.T) {
	// test massive amount of parameters
	tests := []struct {
		command  string
		params   []string
		expected int
	}{
		{"command", genParams("very big param list", 1), 1},
		{"command", genParams("very big param list", 50), 1},
		{"command", genParams("very big param list", 100), 1},
		{"command", genParams("very big param list", 500), 2},
		{"command", genParams("very big param list", 1000), 4},
		{"command", genParams("very big param list", 5000), 18},
		{"command", genParams("very big param list", 10000), 35},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("mass_tests_%v-%d", tt.command, len(tt.params)), func(t *testing.T) {
			requests, err := prepareRequests(tt.command, tt.params)
			assert.NoError(t, err)
			assert.Equal(t, len(requests), tt.expected)
		})
	}
}

func TestPrepareRequestsError(t *testing.T) {
	_, err := prepareRequests(
		strings.Repeat("c", BUF_SIZE-1),
		nil,
	)
	assert.Error(t, err)

	_, err = prepareRequests(
		strings.Repeat("c", BUF_SIZE-len("p ")-1),
		genParams("p", 1),
	)
	assert.Error(t, err)

	_, err = prepareRequests(
		strings.Repeat("c", BUF_SIZE-len("[[BATCH]] p;")-1),
		genParams("p", 5),
	)
	assert.Error(t, err)
}

func BenchmarkPrepareRequests(b *testing.B) {
	params := genParams("param", 10000)

	for i := 0; i < b.N; i++ {
		prepareRequests("command", params)
	}
}

func TestValidateResponse(t *testing.T) {
	// Dummy client to allow this test to run without Hyprland
	c := DummyClient{}

	tests := []struct {
		params    []string
		response  RawResponse
		validate  bool
		expectErr bool
	}{
		{genParams("param", 1), RawResponse("   ok   "), true, false},
		{genParams("param", 2), RawResponse("ok"), true, true},
		{genParams("param", 2), RawResponse("ok"), false, false},
		{genParams("param", 1), RawResponse("ok ok"), true, false}, // not sure about this case, will leave like this for now
		{genParams("param", 5), RawResponse(strings.Repeat("ok", 5)), true, false},
		{genParams("param", 6), RawResponse(strings.Repeat("ok", 5)), true, true},
		{genParams("param", 6), RawResponse(strings.Repeat("ok", 5)), false, false},
		{genParams("param", 10), RawResponse(strings.Repeat(" ok ", 10)), true, false},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("tests_%v-%v", tt.params, tt.response), func(t *testing.T) {
			c.Validate = tt.validate
			err := c.validateResponse(tt.params, tt.response)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func BenchmarkValidateResponse(b *testing.B) {
	// Dummy client to allow this test to run without Hyprland
	c := DummyClient{}
	params := genParams("param", 10000)
	response := strings.Repeat("ok"+strings.Repeat(" ", 1000), 10000)

	for i := 0; i < b.N; i++ {
		c.validateResponse(params, RawResponse(response))
	}
}

func TestRawRequest(t *testing.T) {
	testCommandRR(t, func() (RawResponse, error) {
		return c.RawRequest([]byte("splash"))
	})
}

func TestActiveWindow(t *testing.T) {
	testCommand(t, c.ActiveWindow, Window{})
}

func TestActiveWorkspace(t *testing.T) {
	testCommand(t, c.ActiveWorkspace, Workspace{})
}

func TestAnimations(t *testing.T) {
	testCommand(t, c.Animations, [][]Animation{})
}

func TestBinds(t *testing.T) {
	testCommand(t, c.Binds, []Bind{})
}

func TestClients(t *testing.T) {
	testCommand(t, c.Clients, []Client{})
}

func TestConfigErrors(t *testing.T) {
	testCommand(t, c.ConfigErrors, []ConfigError{})
}

func TestCursorPos(t *testing.T) {
	testCommand(t, c.CursorPos, CursorPos{})
}

func TestDecorations(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test that depends in kitty running")
	}
	testCommand(t, func() ([]Decoration, error) {
		return c.Decorations("kitty")
	}, []Decoration{})
}

func TestDevices(t *testing.T) {
	testCommand(t, c.Devices, Devices{})
}

func TestDispatch(t *testing.T) {
	testCommandRR(t, func() (RawResponse, error) {
		return c.Dispatch("exec kitty sh -c 'echo Testing hyprland-go && sleep 1 && exit 0'")
	})

	if testing.Short() {
		t.Skip("skip slow test")
	}
	// Testing if we can open at least the amount of instances we asked
	// Dispatch() to open.
	// The reason this test exist is because Hyprland has a hidden
	// limitation in the total amount of batch commands you can trigger,
	// but this is not documented and it also fails silently.
	// So this test allows us to validate that the current split of
	// batch commands is working as expected.
	// See also: prepareRequests function and MAX_COMMANDS const
	const want = 35
	const retries = 15
	t.Run(fmt.Sprintf("test_opening_%d_kitty_instances", want), func(t *testing.T) {
		_, err := c.Dispatch(genParams(fmt.Sprintf("exec kitty sh -c 'sleep %d && exit 0'", retries), want)...)
		assert.NoError(t, err)

		aw, err := c.ActiveWorkspace()
		assert.NoError(t, err)

		got := 0
		for i := 0; i < retries; i++ {
			got = 0
			time.Sleep(1 * time.Second)
			cls, err := c.Clients()
			assert.NoError(t, err)

			for _, cl := range cls {
				if cl.Workspace.Id == aw.Id && cl.Class == "kitty" {
					got += 1
				}
			}
			if got >= want {
				t.Logf("after retries: %d, got kitty: %d, finishing test", i+1, got)
				return
			}
		}
		// after maximum amount of retries, give up
		t.Errorf("after retries: %d, got kitty: %d, want at least: %d", retries, got, want)
	})
}

func TestGetOption(t *testing.T) {
	tests := []struct{ option string }{
		{"general:border_size"},
		{"gestures:workspace_swipe"},
		{"misc:vrr"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("mass_tests_%v", tt.option), func(t *testing.T) {
			testCommand(t, func() (Option, error) {
				return c.GetOption(tt.option)
			}, Option{})
		})
	}
}

func TestKeyword(t *testing.T) {
	testCommandRR(t, func() (RawResponse, error) {
		return c.Keyword("general:border_size 1", "general:border_size 5")
	})
}

func TestKill(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test that kill window")
	}
	testCommandRR(t, c.Kill)
}

func TestLayers(t *testing.T) {
	testCommand(t, c.Layers, Layers{})
}

func TestMonitors(t *testing.T) {
	testCommand(t, c.Monitors, []Monitor{})
}

func TestReload(t *testing.T) {
	testCommandRR(t, c.Reload)
}

func TestSetCursor(t *testing.T) {
	testCommandRR(t, func() (RawResponse, error) {
		return c.SetCursor("Adwaita", 32)
	})
}

func TestSplash(t *testing.T) {
	testCommand(t, c.Splash, "")
}

func TestWorkspaces(t *testing.T) {
	testCommand(t, c.Workspaces, []Workspace{})
}

func TestVersion(t *testing.T) {
	testCommand(t, c.Version, Version{})
}
