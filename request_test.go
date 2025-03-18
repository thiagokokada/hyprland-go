package hyprland

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/thiagokokada/hyprland-go/internal/assert"
)

var (
	c      *RequestClient
	reload = flag.Bool("reload", true, "reload configuration after tests end")
)

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

func testCommandR(t *testing.T, command func() (Response, error)) {
	testCommand(t, command, "")
}

func testCommandRs(t *testing.T, command func() ([]Response, error)) {
	testCommand(t, command, []Response{})
}

func testCommand[T any](t *testing.T, command func() (T, error), emptyValue any) {
	checkEnvironment(t)

	got, err := command()
	assert.NoError(t, err)
	assert.DeepNotEqual(t, got, emptyValue)
	t.Logf("got: %+v", got)
}

func setup() {
	if os.Getenv("HYPRLAND_INSTANCE_SIGNATURE") != "" {
		c = MustClient()
	}
}

func teardown() {
	if *reload && c != nil {
		// Make sure that the Hyprland config is in a sane state
		assert.Must1(c.Reload())
	}
}

func TestMain(m *testing.M) {
	flag.Parse()

	setup()

	exitCode := m.Run()

	teardown()

	os.Exit(exitCode)
}

func TestPrepareRequests(t *testing.T) {
	// test params
	tests := []struct {
		command  string
		params   []string
		jsonResp bool
		expected []string
	}{
		{"command", nil, true, []string{"j/command"}},
		{"command", []string{"param0"}, true, []string{"j/command param0"}},
		{"command", []string{"param0", "param1"}, true, []string{"[[BATCH]]j/command param0;j/command param1;"}},
		{"command", nil, false, []string{"command"}},
		{"command", []string{"param0"}, false, []string{"command param0"}},
		{"command", []string{"param0", "param1"}, false, []string{"[[BATCH]]command param0;command param1;"}},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("tests_%s-%s-%v", tt.command, tt.params, tt.jsonResp), func(t *testing.T) {
			requests, err := prepareRequests(tt.command, tt.params, tt.jsonResp)
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
		{"command", genParams("very big param list", 5000), 19},
		{"command", genParams("very big param list", 10000), 37},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("mass_tests_%s-%d", tt.command, len(tt.params)), func(t *testing.T) {
			requests, err := prepareRequests(tt.command, tt.params, true)
			assert.NoError(t, err)
			assert.Equal(t, len(requests), tt.expected)
		})
	}
}

func TestPrepareRequestsError(t *testing.T) {
	tests := []struct {
		lastSafeLen int
		paramsLen   int
		jsonResp    bool
	}{
		{bufSize - len(jsonReqHeader), 0, true},                                // '/j<command>'
		{bufSize - len(jsonReqHeader) - 2, 1, true},                            // '/j<command> p'
		{bufSize - len(jsonReqHeader) - len(batch) - len(reqSep) - 1, 2, true}, // '[[BATCH]]/j<command> p;'
		{bufSize, 0, false},                                // '<command>'
		{bufSize - 2, 1, false},                            // '<command> p'
		{bufSize - len(batch) - len(reqSep) - 1, 2, false}, // '[[BATCH]]<command> p;'
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("tests_%d-%d-%v", tt.lastSafeLen, tt.paramsLen, tt.jsonResp), func(t *testing.T) {
			// With last safe length, we should have no errors
			_, err := prepareRequests(
				strings.Repeat("c", tt.lastSafeLen),
				genParams("p", tt.paramsLen),
				tt.jsonResp,
			)
			assert.NoError(t, err)
			// Now we should have errors
			_, err = prepareRequests(
				strings.Repeat("c", tt.lastSafeLen+1),
				genParams("p", tt.paramsLen),
				tt.jsonResp,
			)
			assert.Error(t, err)
		})
	}
}

func BenchmarkPrepareRequests(b *testing.B) {
	params := genParams("param", 10000)

	for i := 0; i < b.N; i++ {
		prepareRequests("command", params, true)
	}
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		response RawResponse
		want     int
	}{
		{RawResponse("ok"), 1},
		{RawResponse("   ok  "), 1},
		{RawResponse("ok\r\nok"), 2},
		{RawResponse(" \r\nok\r\n \r\n ok"), 2},
		{RawResponse(strings.Repeat("ok\r\n", 5)), 5},
		{RawResponse(strings.Repeat("ok\r\n\r\n", 5)), 5},
		{RawResponse(strings.Repeat("ok\r\n\n", 10)), 10},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("tests_%s-%d", tt.response, tt.want), func(t *testing.T) {
			response, err := parseResponse(tt.response)
			assert.NoError(t, err)
			assert.Equal(t, len(response), tt.want)

			for _, r := range response {
				assert.Equal(t, r, "ok")
			}
		})
	}
}

func BenchmarkParseResponse(b *testing.B) {
	response := []byte(strings.Repeat("ok\r\n", 1000))

	for i := 0; i < b.N; i++ {
		parseResponse(response)
	}
}

func TestValidateResponse(t *testing.T) {
	tests := []struct {
		params   []string
		response []Response
		want     []Response
		wantErr  bool
	}{
		// empty response should error
		{genParams("param", 1), []Response{}, []Response{}, true},
		// happy path, nil param
		{nil, []Response{"ok"}, []Response{"ok"}, false},
		// happy path, 1 param
		{genParams("param", 1), []Response{"ok"}, []Response{"ok"}, false},
		// happy path, multiple params
		{genParams("param", 2), []Response{"ok", "ok"}, []Response{"ok", "ok"}, false},
		// missing response
		{genParams("param", 2), []Response{"ok"}, []Response{"ok"}, true},
		// non-ok response
		{genParams("param", 2), []Response{"ok", "Invalid command"}, []Response{"ok", "Invalid command"}, true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("tests_%v-%v", tt.params, tt.response), func(t *testing.T) {
			response, err := validateResponse(tt.params, tt.response)
			assert.DeepEqual(t, response, tt.want)

			if tt.wantErr {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, ErrorValidation))
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRawRequest(t *testing.T) {
	testCommand(t, func() (RawResponse, error) {
		return c.RawRequest([]byte("splash"))
	}, RawResponse{})
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
	if os.Getenv("NIX_CI") != "" {
		// https://github.com/NixOS/nixpkgs/issues/156067
		// https://github.com/hyprwm/Hyprland/discussions/1257
		t.Skip("skip test that always returns CursorPos{X:0, Y:0} in CI since we can't move cursor")
	}

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
	testCommandRs(t, func() ([]Response, error) {
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
		{"cursor:zoom_factor"},
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
	testCommandRs(t, func() ([]Response, error) {
		return c.Keyword("bind SUPER,K,exec,kitty", "general:border_size 5")
	})
}

func TestKill(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test that kill window")
	}

	testCommandR(t, c.Kill)
}

func TestLayers(t *testing.T) {
	testCommand(t, c.Layers, Layers{})
}

func TestMonitors(t *testing.T) {
	testCommand(t, c.Monitors, []Monitor{})
}

func TestReload(t *testing.T) {
	if testing.Short() {
		t.Skip("skip test that reload config")
	}

	testCommandR(t, c.Reload)
}

func TestSetCursor(t *testing.T) {
	testCommandR(t, func() (Response, error) {
		return c.SetCursor("Adwaita", 32)
	})
}

func TestSwitchXkbLayout(t *testing.T) {
	// Need to find a keyboard, call Devices()
	checkEnvironment(t)

	devices, err := c.Devices()
	assert.NoError(t, err)

	testCommandR(t, func() (Response, error) {
		return c.SwitchXkbLayout(devices.Keyboards[0].Name, "next")
	})
}

func TestSplash(t *testing.T) {
	testCommand(t, c.Splash, "")
}

func BenchmarkSplash(b *testing.B) {
	if c == nil {
		b.Skip("HYPRLAND_INSTANCE_SIGNATURE not set, skipping test")
	}

	for i := 0; i < b.N; i++ {
		c.Splash()
	}
}

func TestWorkspaces(t *testing.T) {
	testCommand(t, c.Workspaces, []Workspace{})
}

func TestVersion(t *testing.T) {
	testCommand(t, c.Version, Version{})

	if os.Getenv("NIX_CI") != "" {
		// make sure that we are running the same version of Hyprland
		// in NixOS VM test that we are declaring as compatible
		v, _ := c.Version()
		assert.Equal(t, v.Tag, "v"+HYPRLAND_VERSION)
	}
}
