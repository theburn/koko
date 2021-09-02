package proxy

import (
	"context"
	"io"

	"github.com/gliderlabs/ssh"
)

type UserConnection interface {
	io.ReadWriteCloser
	ID() string
	WinCh() <-chan ssh.Window
	LoginFrom() string
	RemoteAddr() string
	Pty() ssh.Pty
	Context() context.Context
}

type SessionInfo struct {
	ID          string `json:"id"`
	EnableShare bool   `json:"enable_share"`
}
