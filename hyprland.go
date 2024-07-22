package hyprland

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const (
	BUF_SIZE     = 8192
	MAX_COMMANDS = 30
)

// IPCClient is the main struct from hyprland-go.
// You may want to set 'Validate' as false to avoid (possibly costly)
// validations, at the expense of not reporting some errors in the IPC.
type IPCClient struct {
	Validate    bool
	requestConn *net.UnixAddr
	eventConn   net.Conn
}

func must1[T any](v T, err error) T {
	must(err)
	return v
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func prepareRequests(command string, params []string) (requests [][]byte, err error) {
	if command == "" {
		return nil, errors.New("empty command")
	}

	if len(params) == 0 {
		requests = append(requests, []byte(command))
	} else if len(params) == 1 {
		requests = append(requests, []byte(fmt.Sprintf("%s %s", command, params[0])))
	} else {
		// Hyprland IPC has a hidden limit for commands, so we are
		// splitting the commands in multiple requests if the user pass
		// more commands that it is supported
		for i := 0; i < len(params); i += MAX_COMMANDS {
			end := i + MAX_COMMANDS
			if end > len(params) {
				end = len(params)
			}

			var buffer bytes.Buffer
			buffer.WriteString("[[BATCH]]")
			for j := i; j < end; j++ {
				buffer.WriteString(fmt.Sprintf("%s %s;", command, params[j]))
			}

			requests = append(requests, buffer.Bytes())
		}
	}

	return requests, nil
}

func (c *IPCClient) validateResponses(params []string, responses []byte) error {
	if !c.Validate {
		return nil
	}

	// Empty response
	if len(responses) == 0 {
		return errors.New("empty response")
	}
	// Count the number of "ok" we got in response
	got := strings.Count(string(responses), "ok")
	want := len(params)
	// Commands without parameters still have a "ok" response
	if want == 0 {
		want = 1
	}
	// If we had less than expected number of "ok" results, it means
	// something went wrong
	if got < want {
		return errors.New(
			fmt.Sprintf(
				"got ok: %d, want: %d, response: %s",
				got,
				want,
				responses,
			),
		)
	}
	return nil
}

func (c *IPCClient) doRequest(command string, params ...string) (responses []byte, err error) {
	requests, err := prepareRequests(command, params)
	if err != nil {
		return nil, fmt.Errorf("error while creating request: %w", err)
	}
	for _, r := range requests {
		response, err := c.Request(r)
		if err != nil {
			return nil, fmt.Errorf("error while doing request: %w", err)
		}
		responses = append(responses, response...)
	}
	return responses, nil
}

// Initiate a new client or panic.
// This should be the preferred method for user scripts, since it will
// automatically find the proper socket to connect and use the
// HYPRLAND_INSTANCE_SIGNATURE for the current user.
// If you need to connect to arbitrary user instances or need a method that
// will not panic on error, use [hyprland.NewClient] instead.
func MustClient() *IPCClient {
	his := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if his == "" {
		panic("HYPRLAND_INSTANCE_SIGNATURE is empty, are you using Hyprland?")
	}

	// https://github.com/hyprwm/Hyprland/blob/83a5395eaa99fecef777827fff1de486c06b6180/hyprctl/main.cpp#L53-L62
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		user := must1(user.Current()).Uid
		runtimeDir = filepath.Join("/run/user", user)
	}

	return must1(
		NewClient(
			filepath.Join(runtimeDir, "hypr", his, ".socket.sock"),
			filepath.Join(runtimeDir, "hypr", his, ".socket2.sock"),
		),
	)
}

// Initiate a new client.
// Receive as parameters a requestSocket that is generally localised in
// '$XDG_RUNTIME_DIR/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket.sock' and
// eventSocket that is generally localised in
// '$XDG_RUNTIME_DIR/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket2.sock'.
func NewClient(requestSocket, eventSocket string) (*IPCClient, error) {
	if requestSocket == "" || eventSocket == "" {
		return nil, errors.New("empty request or event socket")
	}

	conn, err := net.Dial("unix", eventSocket)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to socket: %w", err)
	}

	return &IPCClient{
		Validate: true,
		requestConn: &net.UnixAddr{
			Net:  "unix",
			Name: requestSocket,
		},
		eventConn: conn,
	}, nil
}

// Low-level request method, should be avoided unless there is no alternative.
// Receives a byte array as parameter that should be a valid command similar to
// 'hyprctl' command, e.g.: 'hyprctl dispatch exec kitty' will be
// '[]byte("dispatch exec kitty")'.
// Keep in mind that there is no validation. In case of an invalid request, the
// response will generally be something different from "ok".
func (c *IPCClient) Request(request []byte) (response []byte, err error) {
	if len(request) == 0 {
		return nil, errors.New("empty request")
	}

	// Connect to the request socket
	conn, err := net.DialUnix("unix", nil, c.requestConn)
	defer conn.Close()
	if err != nil {
		return nil, fmt.Errorf("error while connecting to socket: %w", err)
	}

	// Send the request to the socket
	request = append([]byte{'j', '/'}, request...)
	_, err = conn.Write(request)
	if err != nil {
		return nil, fmt.Errorf("error while writing to socket: %w", err)
	}

	// Get the response back
	buf := make([]byte, BUF_SIZE)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		response = append(response, buf[:n]...)
		if n < BUF_SIZE {
			break
		}
	}

	return response, nil
}

// Dispatch commands, similar to 'hyprctl dispatch'.
// Accept multiple commands at the same time, in this case it will use batch
// mode, similar to 'hyprctl dispatch --batch'.
func (c *IPCClient) Dispatch(params ...string) error {
	responses, err := c.doRequest("dispatch", params...)
	if err != nil {
		return err
	}
	return c.validateResponses(params, responses)
}

// Reload command, similar to 'hyprctl reload'.
func (c *IPCClient) Reload() error {
	responses, err := c.doRequest("reload")
	if err != nil {
		return err
	}
	return c.validateResponses(nil, responses)
}

// Kill command, similar to 'hyprctl kill'.
// Will NOT wait for the user to click in the window.
func (c *IPCClient) Kill() error {
	responses, err := c.doRequest("kill")
	if err != nil {
		return err
	}
	return c.validateResponses(nil, responses)
}

// Get option command, similar to 'hyprctl getoption'.
func (c *IPCClient) GetOption(name string) (string, error) {
	response, error := c.doRequest("getoption", name)
	if error != nil {
		return "", error
	}
	return string(response), error
}
