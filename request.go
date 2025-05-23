package hyprland

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/thiagokokada/hyprland-go/helpers"
	"github.com/thiagokokada/hyprland-go/internal/assert"
)

var (
	// Returned when the request command is too long, e.g., would not fit
	// the request buffer.
	ErrCommandTooLong = errors.New("command is too long")
	// Returned when the request is empty.
	ErrEmptyRequest = errors.New("empty request")
	// Returned when Hyprland returns an empty response.
	ErrEmptyResponse = errors.New("empty response")
	// Returned when the request is too big, e.g., would not fit the
	// request buffer.
	ErrRequestTooBig = errors.New("request too big")
)

// Initiate a new client or panic.
// This should be the preferred method for user scripts, since it will
// automatically find the proper socket to connect and use the
// HYPRLAND_INSTANCE_SIGNATURE for the current user.
// If you need to connect to arbitrary user instances or need a method that
// will not panic on error, use [NewClient] instead.
func MustClient() *RequestClient {
	return NewClient(
		assert.Must1(helpers.GetSocket(helpers.RequestSocket)),
	)
}

// Initiate a new client.
// Receive as parameters a requestSocket that is generally localised in
// '$XDG_RUNTIME_DIR/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket.sock'.
func NewClient(socket string) *RequestClient {
	return &RequestClient{
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
		return nil, ErrEmptyRequest
	}

	// Connect to the request socket
	conn, err := net.DialUnix("unix", nil, c.conn)
	if err != nil {
		return nil, fmt.Errorf("error while connecting to socket: %w", err)
	}

	defer func() {
		if e := conn.Close(); e != nil {
			err = errors.Join(err, fmt.Errorf("error while closing socket: %w", e))
		}
	}()

	if len(request) > bufSize {
		return nil, fmt.Errorf(
			"%w (%d>%d): %s",
			ErrRequestTooBig,
			len(request),
			bufSize,
			request,
		)
	}

	writer := bufio.NewWriter(conn)

	// Send the request to the socket
	_, err = writer.Write(request)
	if err != nil {
		return nil, fmt.Errorf("error while writing to socket: %w", err)
	}

	err = writer.Flush()
	if err != nil {
		return nil, fmt.Errorf("error while flushing to socket: %w", err)
	}

	// Get the response back
	rbuf := bytes.NewBuffer(nil)
	sbuf := make([]byte, bufSize)
	reader := bufio.NewReader(conn)

	for {
		n, err := reader.Read(sbuf)
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, fmt.Errorf("error while reading from socket: %w", err)
		}

		rbuf.Write(sbuf[:n])

		if n < bufSize {
			break
		}
	}

	return rbuf.Bytes(), err
}

// Active window command, similar to 'hyprctl activewindow'.
// Returns a [Window] object.
func (c *RequestClient) ActiveWindow() (w Window, err error) {
	response, err := c.doRequest("activewindow", nil, true)
	if err != nil {
		return w, err
	}

	return unmarshalResponse(response, &w)
}

// Get option command, similar to 'hyprctl activeworkspace'.
// Returns a [Workspace] object.
func (c *RequestClient) ActiveWorkspace() (w Workspace, err error) {
	response, err := c.doRequest("activeworkspace", nil, true)
	if err != nil {
		return w, err
	}

	return unmarshalResponse(response, &w)
}

// Animations command, similar to 'hyprctl animations'.
// Returns a [Animation] object.
func (c *RequestClient) Animations() (a [][]Animation, err error) {
	response, err := c.doRequest("animations", nil, true)
	if err != nil {
		return a, err
	}

	return unmarshalResponse(response, &a)
}

// Binds command, similar to 'hyprctl binds'.
// Returns a [Bind] object.
func (c *RequestClient) Binds() (b []Bind, err error) {
	response, err := c.doRequest("binds", nil, true)
	if err != nil {
		return b, err
	}

	return unmarshalResponse(response, &b)
}

// Clients command, similar to 'hyprctl clients'.
// Returns a [Client] object.
func (c *RequestClient) Clients() (cl []Client, err error) {
	response, err := c.doRequest("clients", nil, true)
	if err != nil {
		return cl, err
	}

	return unmarshalResponse(response, &cl)
}

// ConfigErrors command, similar to `hyprctl configerrors`.
// Returns a [ConfigError] object.
func (c *RequestClient) ConfigErrors() (ce []ConfigError, err error) {
	response, err := c.doRequest("configerrors", nil, true)
	if err != nil {
		return ce, err
	}

	return unmarshalResponse(response, &ce)
}

// Cursor position command, similar to 'hyprctl cursorpos'.
// Returns a [CursorPos] object.
func (c *RequestClient) CursorPos() (cu CursorPos, err error) {
	response, err := c.doRequest("cursorpos", nil, true)
	if err != nil {
		return cu, err
	}

	return unmarshalResponse(response, &cu)
}

// Decorations command, similar to `hyprctl decorations`.
// Returns a [Decoration] object.
func (c *RequestClient) Decorations(regex string) (d []Decoration, err error) {
	response, err := c.doRequest("decorations", []string{regex}, true)
	if err != nil {
		return d, err
	}
	// XXX: when no decoration is set, we get "none".
	// Is this something that also happen in other commands?
	if string(response) == "none" {
		return nil, nil
	}

	return unmarshalResponse(response, &d)
}

// Devices command, similar to `hyprctl devices`.
// Returns a [Devices] object.
func (c *RequestClient) Devices() (d Devices, err error) {
	response, err := c.doRequest("devices", nil, true)
	if err != nil {
		return d, err
	}

	return unmarshalResponse(response, &d)
}

// Dispatch commands, similar to 'hyprctl dispatch'.
// Accept multiple commands at the same time, in this case it will use batch
// mode, similar to 'hyprctl dispatch --batch'.
// Returns a [Response] list for each parameter, that may be useful for further
// validations.
func (c *RequestClient) Dispatch(params ...string) (r []Response, err error) {
	raw, err := c.doRequest("dispatch", params, false)
	if err != nil {
		return r, err
	}

	return parseAndValidateResponse(params, raw)
}

// Get option command, similar to 'hyprctl getoption'.
// Returns an [Option] object.
func (c *RequestClient) GetOption(name string) (o Option, err error) {
	response, err := c.doRequest("getoption", []string{name}, true)
	if err != nil {
		return o, err
	}

	return unmarshalResponse(response, &o)
}

// Keyword command, similar to 'hyprctl keyword'.
// Accept multiple commands at the same time, in this case it will use batch
// mode, similar to 'hyprctl keyword --batch'.
// Returns a [Response] list for each parameter, that may be useful for further
// validations.
func (c *RequestClient) Keyword(params ...string) (r []Response, err error) {
	raw, err := c.doRequest("keyword", params, false)
	if err != nil {
		return r, err
	}

	return parseAndValidateResponse(params, raw)
}

// Kill command, similar to 'hyprctl kill'.
// Kill an app by clicking on it, can exit with ESCAPE. Will NOT wait until the
// user to click in the window.
// Returns a [Response], that may be useful for further validations.
func (c *RequestClient) Kill() (r Response, err error) {
	raw, err := c.doRequest("kill", nil, true)
	if err != nil {
		return r, err
	}

	response, err := parseAndValidateResponse(nil, raw)

	return response[0], err // should return only one response
}

// Layer command, similar to 'hyprctl layers'.
// Returns a [Layer] object.
func (c *RequestClient) Layers() (l Layers, err error) {
	response, err := c.doRequest("layers", nil, true)
	if err != nil {
		return l, err
	}

	return unmarshalResponse(response, &l)
}

// Monitors command, similar to 'hyprctl monitors'.
// Returns a [Monitor] object.
func (c *RequestClient) Monitors() (m []Monitor, err error) {
	response, err := c.doRequest("monitors all", nil, true)
	if err != nil {
		return m, err
	}

	return unmarshalResponse(response, &m)
}

// Reload command, similar to 'hyprctl reload'.
// Returns a [Response], that may be useful for further validations.
func (c *RequestClient) Reload() (r Response, err error) {
	raw, err := c.doRequest("reload", nil, false)
	if err != nil {
		return r, err
	}

	response, err := parseAndValidateResponse(nil, raw)

	return response[0], err // should return only one response
}

// Set cursor command, similar to 'hyprctl setcursor'.
// Returns a [Response], that may be useful for further validations.
func (c *RequestClient) SetCursor(theme string, size int) (r Response, err error) {
	raw, err := c.doRequest("setcursor", []string{fmt.Sprintf("%s %d", theme, size)}, false)
	if err != nil {
		return r, err
	}

	response, err := parseAndValidateResponse(nil, raw)

	return response[0], err // should return only one response
}

// Set cursor command, similar to 'hyprctl switchxkblayout'.
// Returns a [Response], that may be useful for further validations.
// Param cmd can be either 'next', 'prev' or an ID (e.g: 0).
func (c *RequestClient) SwitchXkbLayout(device string, cmd string) (r Response, err error) {
	raw, err := c.doRequest("switchxkblayout", []string{fmt.Sprintf("%s %s", device, cmd)}, false)
	if err != nil {
		return r, err
	}

	response, err := parseAndValidateResponse(nil, raw)

	return response[0], err // should return only one response
}

// Splash command, similar to 'hyprctl splash'.
func (c *RequestClient) Splash() (s string, err error) {
	response, err := c.doRequest("splash", nil, false)
	if err != nil {
		return s, err
	}

	return string(response), nil
}

// Version command, similar to 'hyprctl version'.
// Returns a [Version] object.
func (c *RequestClient) Version() (v Version, err error) {
	response, err := c.doRequest("version", nil, true)
	if err != nil {
		return v, err
	}

	return unmarshalResponse(response, &v)
}

// Workspaces option command, similar to 'hyprctl workspaces'.
// Returns a [Workspace] object.
func (c *RequestClient) Workspaces() (w []Workspace, err error) {
	response, err := c.doRequest("workspaces", nil, true)
	if err != nil {
		return w, err
	}

	return unmarshalResponse(response, &w)
}

const (
	// https://github.com/hyprwm/Hyprland/blob/918d8340afd652b011b937d29d5eea0be08467f5/hyprctl/main.cpp#L278
	batch = "[[BATCH]]"
	// https://github.com/hyprwm/Hyprland/blob/918d8340afd652b011b937d29d5eea0be08467f5/hyprctl/main.cpp#L257
	bufSize = 8192
)

var (
	jsonReqHeader = []byte{'j', '/'}
	reqSep        = []byte{' ', ';'}
)

func prepareRequest(buf *bytes.Buffer, command string, param string, jsonResp bool) int {
	if jsonResp {
		buf.Write(jsonReqHeader)
	}

	buf.WriteString(command)
	buf.WriteByte(reqSep[0])
	buf.WriteString(param)
	buf.WriteByte(reqSep[1])

	return buf.Len()
}

func prepareRequests(command string, params []string, jsonResp bool) (requests []RawRequest, err error) {
	if command == "" {
		// Panic since this is not supposed to happen, i.e.: only by
		// misuse since this function is internal
		panic("empty command")
	}

	// Buffer that will store the temporary prepared request
	buf := bytes.NewBuffer(nil)
	bufErr := func() error {
		return fmt.Errorf(
			"%w (%d>=%d): %s",
			ErrCommandTooLong,
			buf.Len(),
			bufSize,
			buf.String(),
		)
	}

	switch len(params) {
	case 0:
		if jsonResp {
			buf.Write(jsonReqHeader)
		}

		buf.WriteString(command)

		if buf.Len() > bufSize {
			return nil, bufErr()
		}
	case 1:
		if jsonResp {
			buf.Write(jsonReqHeader)
		}

		buf.WriteString(command)
		buf.WriteByte(reqSep[0])
		buf.WriteString(params[0])

		if buf.Len() > bufSize {
			return nil, bufErr()
		}
	default:
		// Add [[BATCH]] to the buffer
		buf.WriteString(batch)
		// Initialise current length of buffer
		curLen := buf.Len()

		for _, param := range params {
			// Get the current command + param length + request
			// header and separators
			cmdLen := len(command) + len(param) + len(reqSep)
			if jsonResp {
				cmdLen += len(jsonReqHeader)
			}

			// If batch + command length is bigger than bufSize,
			// return an error since it will not fit the socket
			if len(batch)+cmdLen > bufSize {
				// Call prepare request for error
				prepareRequest(buf, command, param, jsonResp)

				return nil, bufErr()
			}

			// If the current length of the buffer + command +
			// param is bigger than bufSize, we will need to split
			// the request
			if curLen+cmdLen > bufSize {
				// Append current buffer contents to the
				// requests array
				requests = append(requests, buf.Bytes())

				// Reset the current buffer and add [[BATCH]]
				buf.Reset()
				buf.WriteString(batch)
			}

			// Add the contents of the request to the buffer
			curLen = prepareRequest(buf, command, param, jsonResp)
		}
	}
	// Append any remaining buffer content to requests array
	requests = append(requests, buf.Bytes())

	return requests, nil
}

func parseResponse(raw RawResponse) (response []Response, err error) {
	reader := bufio.NewReader(bytes.NewReader(raw))
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		resp := strings.TrimSpace(scanner.Text())
		if resp == "" {
			continue
		}

		response = append(response, Response(resp))
	}

	if err := scanner.Err(); err != nil {
		return response, err
	}

	return response, nil
}

func validateResponse(params []string, response []Response) ([]Response, error) {
	// Empty response, something went terrible wrong
	if len(response) == 0 {
		return []Response{}, fmt.Errorf("%w: empty response", ErrValidation)
	}

	// commands without parameters will have at least one return
	want := max(len(params), 1)

	// we have a different number of requests and responses
	if want != len(response) {
		return response, fmt.Errorf(
			"%w: want responses: %d, got: %d, responses: %v",
			ErrValidation,
			want,
			len(response),
			response,
		)
	}

	// validate that all responses are ok
	for i, r := range response {
		if r != "ok" {
			return response, fmt.Errorf(
				"%w: non-ok response from param: %s, response: %s",
				ErrValidation,
				params[i],
				r,
			)
		}
	}

	return response, nil
}

func parseAndValidateResponse(params []string, raw RawResponse) ([]Response, error) {
	response, err := parseResponse(raw)
	if err != nil {
		return response, err
	}

	return validateResponse(params, response)
}

func unmarshalResponse[T any](response RawResponse, v *T) (T, error) {
	if len(response) == 0 {
		return *v, ErrEmptyResponse
	}

	err := json.Unmarshal(response, &v)
	if err != nil {
		return *v, fmt.Errorf(
			"error while unmarshal: %w, response: %s",
			err,
			response,
		)
	}

	return *v, nil
}

func (c *RequestClient) doRequest(command string, params []string, jsonResp bool) (response RawResponse, err error) {
	requests, err := prepareRequests(command, params, jsonResp)
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
