package event

type DummyEvHandler struct{}

func (e *DummyEvHandler) Workspace(WorkspaceName)        {}
func (e *DummyEvHandler) FocusedMonitor(FocusedMonitor)  {}
func (e *DummyEvHandler) ActiveWindow(ActiveWindow)      {}
func (e *DummyEvHandler) Fullscreen(bool)                {}
func (e *DummyEvHandler) MonitorRemoved(MonitorName)     {}
func (e *DummyEvHandler) MonitorAdded(MonitorName)       {}
func (e *DummyEvHandler) CreateWorkspace(WorkspaceName)  {}
func (e *DummyEvHandler) DestroyWorkspace(WorkspaceName) {}
func (e *DummyEvHandler) MoveWorkspace(MoveWorkspace)    {}
func (e *DummyEvHandler) ActiveLayout(ActiveLayout)      {}
func (e *DummyEvHandler) OpenWindow(OpenWindow)          {}
func (e *DummyEvHandler) CloseWindow(CloseWindow)        {}
func (e *DummyEvHandler) MoveWindow(MoveWindow)          {}
func (e *DummyEvHandler) OpenLayer(OpenLayer)            {}
func (e *DummyEvHandler) CloseLayer(CloseLayer)          {}
func (e *DummyEvHandler) SubMap(SubMap)                  {}
func (e *DummyEvHandler) Screencast(Screencast)          {}
