package httpd

import "time"

type Message struct {
	Id   string `json:"id"`
	Type string `json:"type"`
	Data string `json:"data"`
	Raw  []byte `json:"-"`
}

const (
	PING           = "PING"
	PONG           = "PONG"
	CONNECT        = "CONNECT"
	CLOSE          = "CLOSE"
	TERMINALINIT   = "TERMINAL_INIT"
	TERMINALDATA   = "TERMINAL_DATA"
	TERMINALRESIZE = "TERMINAL_RESIZE"
	TERMINALBINARY = "TERMINAL_BINARY"
	TERMINALSESSION = "TERMINAL_SESSION"
)

type WindowSize struct {
	Cols int `json:"cols"`
	Rows int `json:"rows"`
}

type SessionData struct {
	ID string `json:"id"`
}

const (
	TargetTypeAsset = "asset"

	// TargetTypeMonitor todo: 前端参数将 统一修改成 monitor
	TargetTypeMonitor = "shareroom"

	TargetTypeShare = "share"
)

const (
	maxReadTimeout  = 5 * time.Minute
	maxWriteTimeOut = 5 * time.Minute
)

const (
	TTYName       = "terminal"
	WebFolderName = "web_folder"
)
