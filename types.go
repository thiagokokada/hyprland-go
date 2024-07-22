package hyprland

// Try to keep this in the same order as the output for `hyprctl` for sanity.

type Option struct {
	Option string `json:"option"`
	Int    int    `json:"int"`
	Set    bool   `json:"set"`
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

type Window struct {
	Address        string        `json:"address"`
	Mapped         bool          `json:"mapped"`
	Hidden         bool          `json:"hidden"`
	Size           []int         `json:"size"`
	Workspace      WorkspaceType `json:"workspace"`
	At             []int         `json:"at"`
	Floating       bool          `json:"floating"`
	Pseudo         bool          `json:"pseudo"`
	Monitor        int           `json:"monitor"`
	Class          string        `json:"class"`
	Title          string        `json:"title"`
	InitialClass   string        `json:"initialClass"`
	InitialTitle   string        `json:"initialTitle"`
	Pid            int           `json:"pid"`
	Xwayland       bool          `json:"xwayland"`
	Pinned         bool          `json:"pinned"`
	Fullscreen     bool          `json:"fullscreen"`
	FullscreenMode int           `json:"fullscreenMode"`
	FakeFullscreen bool          `json:"fakeFullscreen"`
	Grouped        []string      `json:"grouped"`
	Tags           []string      `json:"tags"`
	Swallowing     string        `json:"swallowing"`
	FocusHistoryId int           `json:"focusHistoryID"`
}

type WorkspaceType struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
