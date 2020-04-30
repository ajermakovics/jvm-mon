package jvmmon

import (
	"net"
	"log"
	"bufio"
)

type Server struct {
	Port     int
	Messages chan string

	listener net.Listener
	client   net.Conn
	Connections chan net.Addr
}

func NewServer() (*Server, error) {
	listener, err := net.Listen("tcp", ":0")
	if listener == nil {
		log.Fatal("Cannot listen", err)
		return nil, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	log.Println("Server listening on port:", port)
	messages := make(chan string)
	connections := make(chan net.Addr)
	server := Server{port, messages, listener, nil, connections}
	go server.acceptConnections()

	return &server, nil
}

func (server *Server) acceptConnections() {
	conns := server.clientConns()
	for {
		conn := <-conns
		server.closeClient()
		server.client = conn
		go server.handleConn(conn)
	}
}

func (server *Server) closeClient() {
	if server.client != nil {
		log.Println("Closing existing connection")
		err := server.client.Close()
		logErr("Error closing client", err)
		server.client = nil
		log.Println("Closed connection")
	}
}

func (server *Server) clientConns() chan net.Conn {
	ch := make(chan net.Conn)

	go func() {
		for {
			client, err := server.listener.Accept()
			if err != nil {
				log.Fatal("Could not accept", err)
				continue
			}
			log.Println("Accepted conn", client.RemoteAddr())
			server.Connections <- client.RemoteAddr()
			ch <- client
		}
	}()
	return ch
}

func (server *Server) readLine(b *bufio.Reader) (string, error) {
	message := ""
	for {
		lineBytes, isPrefix, err := b.ReadLine()
		line := string(lineBytes)
		if err != nil { // EOF, or worse
			return message, err
		}
		message += line
		if !isPrefix { // read until '\n'
			return message, nil
		}
	}
	return message, nil
}

func (server *Server) handleConn(client net.Conn) {
	addr := client.RemoteAddr()
	b := bufio.NewReader(client)
	for {
		message, err := server.readLine(b)
		if err != nil { // EOF, or worse
			log.Println("Connection read error ", addr, " ", err)
			break
		}
		server.Messages <- message
	}
	log.Println("Client disconnected ", addr)
	if server.client != nil && client.RemoteAddr() == server.client.RemoteAddr() {
		log.Println("Clearing cur client ", addr)
		server.client = nil
	}
}
