// Package ipc provides Unix socket IPC server for the daemon.
package ipc

import (
	"net"
	"os"
)

const (
	// DefaultSocketPath is the default Unix socket path.
	DefaultSocketPath = "/tmp/verbalizer.sock"
)

// Server handles IPC communication via Unix socket.
type Server struct {
	socketPath string
	listener   net.Listener
}

// NewServer creates a new IPC server.
func NewServer(socketPath string) *Server {
	if socketPath == "" {
		socketPath = DefaultSocketPath
	}
	return &Server{
		socketPath: socketPath,
	}
}

// Start starts the IPC server.
func (s *Server) Start() error {
	// Remove existing socket
	os.Remove(s.socketPath)

	// Create Unix socket listener
	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return err
	}

	s.listener = listener
	return nil
}

// Stop stops the IPC server.
func (s *Server) Stop() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}
