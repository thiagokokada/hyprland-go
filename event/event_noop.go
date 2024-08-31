package event

// NoopEventHandler is an implementation of [EventHandler] interface with all
// methods doing nothing. It is a good starting point to be embedded your own
// struct to be extended.
type NoopEventHandler struct{}

func (e *NoopEventHandler) Workspace(WorkspaceName)        {}
func (e *NoopEventHandler) FocusedMonitor(FocusedMonitor)  {}
func (e *NoopEventHandler) ActiveWindow(ActiveWindow)      {}
func (e *NoopEventHandler) Fullscreen(bool)                {}
func (e *NoopEventHandler) MonitorRemoved(MonitorName)     {}
func (e *NoopEventHandler) MonitorAdded(MonitorName)       {}
func (e *NoopEventHandler) CreateWorkspace(WorkspaceName)  {}
func (e *NoopEventHandler) DestroyWorkspace(WorkspaceName) {}
func (e *NoopEventHandler) MoveWorkspace(MoveWorkspace)    {}
func (e *NoopEventHandler) ActiveLayout(ActiveLayout)      {}
func (e *NoopEventHandler) OpenWindow(OpenWindow)          {}
func (e *NoopEventHandler) CloseWindow(CloseWindow)        {}
func (e *NoopEventHandler) MoveWindow(MoveWindow)          {}
func (e *NoopEventHandler) OpenLayer(OpenLayer)            {}
func (e *NoopEventHandler) CloseLayer(CloseLayer)          {}
func (e *NoopEventHandler) SubMap(SubMap)                  {}
func (e *NoopEventHandler) Screencast(Screencast)          {}
