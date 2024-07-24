package hyprland

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/thiagokokada/hyprland-go/internal/assert"
)

const BUF_SIZE = 8192

func prepareRequest(buf *bytes.Buffer, command string, param string) int {
	buf.WriteString(command)
	buf.WriteString(" ")
	buf.WriteString(param)
	buf.WriteString(";")

	return buf.Len()
}

func prepareRequests(command string, params []string) (requests []RawRequest, err error) {
	if command == "" {
		// Panic since this is not supposed to happen, i.e.: only by
		// misuse since this function is internal
		panic("empty command")
	}
	switch len(params) {
	case 0:
		if len(command) >= BUF_SIZE {
			return nil, fmt.Errorf(
				"command is too long (%d>%d): %s",
				BUF_SIZE,
				len(command),
				command,
			)
		}
		requests = append(requests, []byte(command))
	case 1:
		request := command + " " + params[0]
		if len(request) >= BUF_SIZE {
			return nil, fmt.Errorf(
				"command is too long (%d>%d): %s",
				BUF_SIZE,
				len(request),
				request,
			)
		}
		requests = append(requests, []byte(request))
	default:
		buf := bytes.NewBuffer(nil)

		const batch = "[[BATCH]]"
		// Add [[BATCH]] to the buffer
		buf.WriteString(batch)
		// Initialise current length of buffer
		curLen := buf.Len()

		for _, param := range params {
			// Get the current command + param length
			cmdLen := len(command) + len(param) + 2 // ; + <space>
			if len(batch)+cmdLen >= BUF_SIZE {
				// If batch + command + param length is bigger
				// than BUF_SIZE, return an error since it will
				// not fit the socket
				return nil, fmt.Errorf(
					"command is too long (%d>%d): %s %s",
					cmdLen,
					BUF_SIZE,
					command,
					param,
				)
			} else if curLen+cmdLen < BUF_SIZE {
				// If the current length of the buffer +
				// command + param is less than BUF_SIZE, the
				// request will fit
				curLen = prepareRequest(buf, command, param)
			} else {
				// If not, we will need to split the request,
				// so append current buffer contents to the
				// requests array
				requests = append(requests, buf.Bytes())

				// Reset the current buffer and add [[BATCH]]
				buf.Reset()
				buf.WriteString(batch)

				// And finally, add the contents of the request
				// to the buffer
				curLen = prepareRequest(buf, command, param)
			}
		}
		// Append any remaining buffer content to requests array
		requests = append(requests, buf.Bytes())
	}
	return requests, nil
}

func (c *RequestClient) validateResponse(params []string, response RawResponse) error {
	if !c.Validate {
		return nil
	}

	// Empty response
	if len(response) == 0 {
		return errors.New("empty response")
	}

	var resp = string(response)
	// Count the number of "ok" we got in response
	got := strings.Count(resp, "ok")
	want := len(params)
	// Commands without parameters still have a "ok" response
	if want == 0 {
		want = 1
	}
	// If we had less than expected number of "ok" results, it means
	// something went wrong
	if got < want {
		return fmt.Errorf(
			"got ok: %d, want: %d, response: %s",
			got,
			want,
			resp,
		)

	}
	return nil
}

func unmarshalResponse(response RawResponse, v any) (err error) {
	if len(response) == 0 {
		return errors.New("empty response")
	}

	err = json.Unmarshal(response, &v)
	if err != nil {
		return fmt.Errorf("error during unmarshal: %w", err)
	}
	return nil
}

func (c *RequestClient) doRequest(command string, params ...string) (response RawResponse, err error) {
	requests, err := prepareRequests(command, params)
	if err != nil {
		return nil, fmt.Errorf("error while preparing request: %w", err)
	}

	buf := bytes.NewBuffer(nil)
	for _, req := range requests {
		resp, err := c.RawRequest(req)
		if err != nil {
			return nil, fmt.Errorf("error while doing request: %w", err)
		}
		buf.Write(resp)
	}

	return buf.Bytes(), nil
}

func mustSocket(socket string) string {
	his := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if his == "" {
		panic("HYPRLAND_INSTANCE_SIGNATURE is empty, are you using Hyprland?")
	}

	// https://github.com/hyprwm/Hyprland/blob/83a5395eaa99fecef777827fff1de486c06b6180/hyprctl/main.cpp#L53-L62
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		user := assert.Must1(user.Current()).Uid
		runtimeDir = filepath.Join("/run/user", user)
	}
	return filepath.Join(runtimeDir, "hypr", his, socket)
}

// Initiate a new client or panic.
// This should be the preferred method for user scripts, since it will
// automatically find the proper socket to connect and use the
// HYPRLAND_INSTANCE_SIGNATURE for the current user.
// If you need to connect to arbitrary user instances or need a method that
// will not panic on error, use [NewClient] instead.
func MustClient() *RequestClient {
	return NewClient(mustSocket(".socket.sock"))
}

// Initiate a new client.
// Receive as parameters a requestSocket that is generally localised in
// '$XDG_RUNTIME_DIR/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket.sock'.
func NewClient(socket string) *RequestClient {
	return &RequestClient{
		Validate: true,
		conn: &net.UnixAddr{
			Net:  "unix",
			Name: socket,
		},
	}
}

// Low-level request method, should be avoided unless there is no alternative.
// Receives a byte array as parameter that should be a valid command similar to
// 'hyprctl' command, e.g.: 'hyprctl dispatch exec kitty' will be
// '[]byte("dispatch exec kitty")'.
// Keep in mind that there is no validation. In case of an invalid request, the
// response will generally be something different from "ok".
func (c *RequestClient) RawRequest(request RawRequest) (response RawResponse, err error) {
	if len(request) == 0 {
		return nil, errors.New("empty request")
	}

	// Connect to the request socket
	conn, err := net.DialUnix("unix", nil, c.conn)
	defer func() {
		if e := conn.Close(); e != nil {
			err = fmt.Errorf("error while closing socket: %w", e)
		}
	}()

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
	rbuf := bytes.NewBuffer(nil)
	sbuf := make([]byte, BUF_SIZE)
	for {
		n, err := conn.Read(sbuf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		rbuf.Write(sbuf[:n])
		if n < BUF_SIZE {
			break
		}
	}

	return rbuf.Bytes(), err
}

// Active window command, similar to 'hyprctl activewindow'.
// Returns a [Window] object.
func (c *RequestClient) ActiveWindow() (w Window, err error) {
	response, err := c.doRequest("activewindow")
	if err != nil {
		return w, err
	}
	return w, unmarshalResponse(response, &w)
}

// Get option command, similar to 'hyprctl activeworkspace'.
// Returns a [Workspace] object.
func (c *RequestClient) ActiveWorkspace() (w Workspace, err error) {
	response, err := c.doRequest("activeworkspace")
	if err != nil {
		return w, err
	}
	return w, unmarshalResponse(response, &w)
}

// Animations command, similar to 'hyprctl animations'.
// Returns a [Animation] object.
func (c *RequestClient) Animations() (a [][]Animation, err error) {
	response, err := c.doRequest("animations")
	if err != nil {
		return a, err
	}
	return a, unmarshalResponse(response, &a)
}

// Binds command, similar to 'hyprctl binds'.
// Returns a [Bind] object.
func (c *RequestClient) Binds() (b []Bind, err error) {
	response, err := c.doRequest("binds")
	if err != nil {
		return b, err
	}
	return b, unmarshalResponse(response, &b)
}

// Clients command, similar to 'hyprctl clients'.
// Returns a [Client] object.
func (c *RequestClient) Clients() (cl []Client, err error) {
	response, err := c.doRequest("clients")
	if err != nil {
		return cl, err
	}
	return cl, unmarshalResponse(response, &cl)
}

// ConfigErrors command, similar to `hyprctl configerrors`.
// Returns a [ConfigError] object.
func (c *RequestClient) ConfigErrors() (ce []ConfigError, err error) {
	response, err := c.doRequest("configerrors")
	if err != nil {
		return ce, err
	}
	return ce, unmarshalResponse(response, &ce)
}

// Cursor position command, similar to 'hyprctl cursorpos'.
// Returns a [CursorPos] object.
func (c *RequestClient) CursorPos() (cu CursorPos, err error) {
	response, err := c.doRequest("cursorpos")
	if err != nil {
		return cu, err
	}
	return cu, unmarshalResponse(response, &cu)
}

// Decorations command, similar to `hyprctl decorations`.
// Returns a [Decoration] object.
func (c *RequestClient) Decorations(regex string) (d []Decoration, err error) {
	response, err := c.doRequest("decorations", regex)
	if err != nil {
		return d, err
	}
	return d, unmarshalResponse(response, &d)
}

// Devices command, similar to `hyprctl devices`.
// Returns a [Devices] object.
func (c *RequestClient) Devices() (d Devices, err error) {
	response, err := c.doRequest("devices")
	if err != nil {
		return d, err
	}
	return d, unmarshalResponse(response, &d)
}

// Dispatch commands, similar to 'hyprctl dispatch'.
// Accept multiple commands at the same time, in this case it will use batch
// mode, similar to 'hyprctl dispatch --batch'.
// Returns the raw response, that may be useful for further validations,
// especially when [RequestClient] 'Validation' is set to false.
func (c *RequestClient) Dispatch(params ...string) (r RawResponse, err error) {
	response, err := c.doRequest("dispatch", params...)
	if err != nil {
		return response, err
	}
	return response, c.validateResponse(params, response)
}

// Get option command, similar to 'hyprctl getoption'.
// Returns an [Option] object.
func (c *RequestClient) GetOption(name string) (o Option, err error) {
	response, err := c.doRequest("getoption", name)
	if err != nil {
		return o, err
	}
	return o, unmarshalResponse(response, &o)
}

// Keyword command, similar to 'hyprctl keyword'.
// Accept multiple commands at the same time, in this case it will use batch
// mode, similar to 'hyprctl keyword --batch'.
// Returns the raw response, that may be useful for further validations,
// especially when [RequestClient] 'Validation' is set to false.
func (c *RequestClient) Keyword(params ...string) (r RawResponse, err error) {
	response, err := c.doRequest("keyword", params...)
	if err != nil {
		return response, err
	}
	return response, c.validateResponse(nil, response)
}

// Kill command, similar to 'hyprctl kill'.
// Kill an app by clicking on it, can exit with ESCAPE. Will NOT wait until the
// user to click in the window.
// Returns the raw response, that may be useful for further validations,
// especially when [RequestClient] 'Validation' is set to false.
func (c *RequestClient) Kill() (r RawResponse, err error) {
	response, err := c.doRequest("kill")
	if err != nil {
		return response, err
	}
	return response, c.validateResponse(nil, response)
}

// Layer command, similar to 'hyprctl layers'.
// Returns a [Layer] object.
func (c *RequestClient) Layers() (l Layers, err error) {
	response, err := c.doRequest("layers")
	if err != nil {
		return l, err
	}
	return l, unmarshalResponse(response, &l)
}

// Monitors command, similar to 'hyprctl monitors'.
// Returns a [Monitor] object.
func (c *RequestClient) Monitors() (m []Monitor, err error) {
	response, err := c.doRequest("monitors")
	if err != nil {
		return m, err
	}
	return m, unmarshalResponse(response, &m)
}

// Reload command, similar to 'hyprctl reload'.
// Returns the raw response, that may be useful for further validations,
// especially when [RequestClient] 'Validation' is set to false.
func (c *RequestClient) Reload() (r RawResponse, err error) {
	response, err := c.doRequest("reload")
	if err != nil {
		return response, err
	}
	return response, c.validateResponse(nil, response)
}

// Set cursor command, similar to 'hyprctl setcursor'.
// Returns the raw response, that may be useful for further validations,
// especially when [RequestClient] 'Validation' is set to false.
func (c *RequestClient) SetCursor(theme string, size int) (r RawResponse, err error) {
	response, err := c.doRequest("setcursor", fmt.Sprintf("%s %d", theme, size))
	if err != nil {
		return response, err
	}
	return response, c.validateResponse(nil, response)
}

// Splash command, similar to 'hyprctl splash'.
func (c *RequestClient) Splash() (s string, err error) {
	response, err := c.doRequest("splash")
	if err != nil {
		return "", err
	}
	return string(response), nil
}

// Version command, similar to 'hyprctl version'.
// Returns a [Version] object.
func (c *RequestClient) Version() (v Version, err error) {
	response, err := c.doRequest("version")
	if err != nil {
		return v, err
	}
	return v, unmarshalResponse(response, &v)
}

// Workspaces option command, similar to 'hyprctl workspaces'.
// Returns a [Workspace] object.
func (c *RequestClient) Workspaces() (w []Workspace, err error) {
	response, err := c.doRequest("workspaces")
	if err != nil {
		return w, err
	}
	return w, unmarshalResponse(response, &w)
}
