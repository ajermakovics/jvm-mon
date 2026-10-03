package jvmmon

import (
	"bufio"
	"errors"
	"log"
	"net"
	"sync"

	"github.com/asaskevich/EventBus"
)

// maxMessageSize limits a single metrics line to guard against runaway clients.
const maxMessageSize = 1 << 20

type Server struct {
	Port        int
	Messages    chan string
	Connections chan net.Addr

	listener net.Listener
	mu       sync.Mutex
	client   net.Conn
}

func NewServer(eb EventBus.Bus) (*Server, error) {
	// Bind to loopback only: the agent always connects to 127.0.0.1
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	log.Println("Server listening on port:", port)

	server := &Server{
		Port:        port,
		Messages:    make(chan string, 16),
		Connections: make(chan net.Addr, 4),
		listener:    listener,
	}
	go server.acceptConnections()

	err = eb.Subscribe("jvm-selected", func(pid string) {
		server.setClient(nil) // close existing
	})
	if err != nil {
		return nil, err
	}

	return server, nil
}

func (server *Server) acceptConnections() {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Println("Could not accept:", err)
			continue
		}
		log.Println("Accepted conn", conn.RemoteAddr())
		select {
		case server.Connections <- conn.RemoteAddr():
		default:
		}
		server.setClient(conn)
		go server.handleConn(conn)
	}
}

// setClient replaces the current client, closing the previous one.
func (server *Server) setClient(conn net.Conn) {
	server.mu.Lock()
	prev := server.client
	server.client = conn
	server.mu.Unlock()

	if prev != nil {
		log.Println("Closing existing connection")
		logErr("Error closing client", prev.Close())
	}
}

func (server *Server) isCurrent(conn net.Conn) bool {
	server.mu.Lock()
	defer server.mu.Unlock()
	return server.client == conn
}

func (server *Server) Close() error {
	server.setClient(nil)
	return server.listener.Close()
}

func (server *Server) handleConn(client net.Conn) {
	addr := client.RemoteAddr()
	scanner := bufio.NewScanner(client)
	scanner.Buffer(make([]byte, 64*1024), maxMessageSize)
	for scanner.Scan() {
		if !server.isCurrent(client) { // stale connection from previously selected JVM
			break
		}
		server.Messages <- scanner.Text()
	}
	if err := scanner.Err(); err != nil {
		log.Println("Connection read error", addr, err)
	}
	log.Println("Client disconnected", addr)

	server.mu.Lock()
	if server.client == client {
		server.client = nil
	}
	server.mu.Unlock()
	_ = client.Close()
}
