package event

// DefaultEventHandler is an implementation of [EventHandler] interface with
// all handlers doing nothing. It is a good starting point to be embedded your
// own struct to be extended.
type DefaultEventHandler struct{}

func (e *DefaultEventHandler) Workspace(WorkspaceName)        {}
func (e *DefaultEventHandler) FocusedMonitor(FocusedMonitor)  {}
func (e *DefaultEventHandler) ActiveWindow(ActiveWindow)      {}
func (e *DefaultEventHandler) Fullscreen(Fullscreen)          {}
func (e *DefaultEventHandler) MonitorRemoved(MonitorName)     {}
func (e *DefaultEventHandler) MonitorAdded(MonitorName)       {}
func (e *DefaultEventHandler) CreateWorkspace(WorkspaceName)  {}
func (e *DefaultEventHandler) DestroyWorkspace(WorkspaceName) {}
func (e *DefaultEventHandler) MoveWorkspace(MoveWorkspace)    {}
func (e *DefaultEventHandler) ActiveLayout(ActiveLayout)      {}
func (e *DefaultEventHandler) OpenWindow(OpenWindow)          {}
func (e *DefaultEventHandler) CloseWindow(CloseWindow)        {}
func (e *DefaultEventHandler) MoveWindow(MoveWindow)          {}
func (e *DefaultEventHandler) OpenLayer(OpenLayer)            {}
func (e *DefaultEventHandler) CloseLayer(CloseLayer)          {}
func (e *DefaultEventHandler) SubMap(SubMap)                  {}
func (e *DefaultEventHandler) Screencast(Screencast)          {}
