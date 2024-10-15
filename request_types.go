package hyprland

import (
	"errors"
	"net"
)

// Indicates the version where the structs are up-to-date.
const HYPRLAND_VERSION = "0.44.1"

// Represents a raw request that is passed for Hyprland's socket.
type RawRequest []byte

// Represents a raw response returned from the Hyprland's socket.
type RawResponse []byte

// Represents a parsed response returned from the Hyprland's socket.
type Response string

// RequestClient is the main struct from hyprland-go.
type RequestClient struct {
	conn *net.UnixAddr
}

// ErrorValidation is used to return errors from response validation. In some
// cases you may want to ignore those errors, in this case you can use
// [errors.Is] to compare the errors returned with this type.
var ErrorValidation = errors.New("validation error")

// Unmarshal structs for requests.
// Try to keep struct fields in the same order as the output for `hyprctl -j`
// for sanity.

type Animation struct {
	Name       string  `json:"name"`
	Overridden bool    `json:"overridden"`
	Bezier     string  `json:"bezier"`
	Enabled    bool    `json:"enabled"`
	Speed      float64 `json:"speed"`
	Style      string  `json:"style"`
}

type Bind struct {
	Locked         bool   `json:"locked"`
	Mouse          bool   `json:"mouse"`
	Release        bool   `json:"release"`
	Repeat         bool   `json:"repeat"`
	NonConsuming   bool   `json:"non_consuming"`
	HasDescription bool   `json:"has_description"`
	ModMask        int    `json:"modmask"`
	SubMap         string `json:"submap"`
	Key            string `json:"key"`
	KeyCode        int    `json:"keycode"`
	CatchAll       bool   `json:"catch_all"`
	Description    string `json:"description"`
	Dispatcher     string `json:"dispatcher"`
	Arg            string `json:"arg"`
}

type FullscreenState int

const (
	None FullscreenState = iota
	Maximized
	Fullscreen
	MaximizedFullscreen
)

type Client struct {
	Address          string          `json:"address"`
	Mapped           bool            `json:"mapped"`
	Hidden           bool            `json:"hidden"`
	At               []int           `json:"at"`
	Size             []int           `json:"size"`
	Workspace        WorkspaceType   `json:"workspace"`
	Floating         bool            `json:"floating"`
	Pseudo           bool            `json:"pseudo"`
	Monitor          int             `json:"monitor"`
	Class            string          `json:"class"`
	Title            string          `json:"title"`
	InitialClass     string          `json:"initialClass"`
	InitialTitle     string          `json:"initialTitle"`
	Pid              int             `json:"pid"`
	Xwayland         bool            `json:"xwayland"`
	Pinned           bool            `json:"pinned"`
	Fullscreen       FullscreenState `json:"fullscreen"`
	FullscreenClient FullscreenState `json:"fullscreenClient"`
	Grouped          []string        `json:"grouped"`
	Tags             []string        `json:"tags"`
	Swallowing       string          `json:"swallowing"`
	FocusHistoryId   int             `json:"focusHistoryID"`
}

type ConfigError string

type CursorPos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Decoration struct {
	DecorationName string `json:"decorationName"`
	Priority       int    `json:"priority"`
}

type Devices struct {
	Mice []struct {
		Address      string  `json:"address"`
		Name         string  `json:"name"`
		DefaultSpeed float64 `json:"defaultSpeed"`
	} `json:"mice"`
	Keyboards []struct {
		Address      string `json:"address"`
		Name         string `json:"name"`
		Rules        string `json:"rules"`
		Model        string `json:"model"`
		Layout       string `json:"layout"`
		Variant      string `json:"variant"`
		Options      string `json:"options"`
		ActiveKeymap string `json:"active_keymap"`
		Main         bool   `json:"main"`
	} `json:"keyboards"`
	Tablets  []interface{} `json:"tablets"` // TODO: need a tablet to test
	Touch    []interface{} `json:"touch"`   // TODO: need a touchscreen to test
	Switches []struct {
		Address string `json:"address"`
		Name    string `json:"name"`
	} `json:"switches"`
}

type Output string

type Layers map[Output]Layer

type Layer struct {
	Levels map[int][]LayerField `json:"levels"`
}

type LayerField struct {
	Address   string `json:"address"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	W         int    `json:"w"`
	H         int    `json:"h"`
	Namespace string `json:"namespace"`
}

type Monitor struct {
	Id               int           `json:"id"`
	Name             string        `json:"name"`
	Description      string        `json:"description"`
	Make             string        `json:"make"`
	Model            string        `json:"model"`
	Serial           string        `json:"serial"`
	Width            int           `json:"width"`
	Height           int           `json:"height"`
	RefreshRate      float64       `json:"refreshRate"`
	X                int           `json:"x"`
	Y                int           `json:"y"`
	ActiveWorkspace  WorkspaceType `json:"activeWorkspace"`
	SpecialWorkspace WorkspaceType `json:"specialWorkspace"`
	Reserved         []int         `json:"reserved"`
	Scale            float64       `json:"scale"`
	Transform        int           `json:"transform"`
	Focused          bool          `json:"focused"`
	DpmsStatus       bool          `json:"dpmsStatus"`
	Vrr              bool          `json:"vrr"`
	ActivelyTearing  bool          `json:"activelyTearing"`
	CurrentFormat    string        `json:"currentFormat"`
	AvailableModes   []string      `json:"availableModes"`
}

type Option struct {
	Option string `json:"option"`
	Int    int    `json:"int"`
	Set    bool   `json:"set"`
}

type Version struct {
	Branch        string   `json:"branch"`
	Commit        string   `json:"commit"`
	Dirty         bool     `json:"dirty"`
	CommitMessage string   `json:"commit_message"`
	CommitDate    string   `json:"commit_date"`
	Tag           string   `json:"tag"`
	Commits       string   `json:"commits"`
	Flags         []string `json:"flags"`
}

type Window struct {
	Client
}

type Workspace struct {
	WorkspaceType
	Monitor         string `json:"monitor"`
	MonitorID       int    `json:"monitorID"`
	Windows         int    `json:"windows"`
	HasFullScreen   bool   `json:"hasfullscreen"`
	LastWindow      string `json:"lastwindow"`
	LastWindowTitle string `json:"lastwindowtitle"`
}

type WorkspaceType struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
