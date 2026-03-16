package ipc

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/fulvian/verbalizer/daemon/pkg/api"
)

const (
	// DefaultSocketPath is the default Unix socket path.
	DefaultSocketPath = "/tmp/verbalizer.sock"
)

// CommandHandler handles incoming IPC commands.
type CommandHandler interface {
	HandleCommand(cmd *api.Command) (*api.Response, error)
}

// Server handles IPC communication via Unix socket.
type Server struct {
	socketPath string
	handler    CommandHandler
	listener   net.Listener
}

// NewServer creates a new IPC server.
func NewServer(socketPath string, handler CommandHandler) *Server {
	if socketPath == "" {
		socketPath = DefaultSocketPath
	}
	return &Server{
		socketPath: socketPath,
		handler:    handler,
	}
}

// Start starts the IPC server.
func (s *Server) Start() error {
	// Remove existing socket
	if err := os.Remove(s.socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing socket: %w", err)
	}

	// Create Unix socket listener
	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on unix socket: %w", err)
	}

	s.listener = listener

	go s.acceptLoop()

	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var cmd api.Command
		if err := decoder.Decode(&cmd); err != nil {
			if err != io.EOF {
				fmt.Printf("Error decoding command: %v\n", err)
			}
			return
		}

		resp, err := s.handler.HandleCommand(&cmd)
		if err != nil {
			resp = &api.Response{
				Success: false,
				Error:   err.Error(),
			}
		}

		if err := encoder.Encode(resp); err != nil {
			fmt.Printf("Error encoding response: %v\n", err)
			return
		}
	}
}

// Stop stops the IPC server.
func (s *Server) Stop() error {
	if s.listener != nil {
		err := s.listener.Close()
		os.Remove(s.socketPath)
		return err
	}
	return nil
}
