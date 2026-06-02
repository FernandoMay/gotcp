package main

import (
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestNewServer(t *testing.T) {
	s := newServer()
	if s.rooms == nil {
		t.Error("expected rooms map to be initialized")
	}
	if s.commands == nil {
		t.Error("expected commands channel to be initialized")
	}
}

func TestServer_Nick(t *testing.T) {
	s := newServer()
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	c := &client{conn: serverConn, nick: "anonymous"}
	s.nick(c, []string{"/nick", "testuser"})

	buf := make([]byte, 1024)
	n, _ := clientConn.Read(buf)
	response := string(buf[:n])

	if !strings.Contains(response, "testuser") {
		t.Errorf("expected response to contain 'testuser', got: %s", response)
	}
	if c.nick != "testuser" {
		t.Errorf("expected nick 'testuser', got '%s'", c.nick)
	}
}

func TestServer_Nick_NoArg(t *testing.T) {
	s := newServer()
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	c := &client{conn: serverConn, nick: "anonymous"}
	s.nick(c, []string{"/nick"})

	buf := make([]byte, 1024)
	n, _ := clientConn.Read(buf)
	response := string(buf[:n])

	if !strings.Contains(response, "nick is required") {
		t.Errorf("expected error about missing nick, got: %s", response)
	}
	if c.nick != "anonymous" {
		t.Errorf("expected nick unchanged 'anonymous', got '%s'", c.nick)
	}
}

func TestServer_Join(t *testing.T) {
	s := newServer()
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	c := &client{conn: serverConn, nick: "testuser"}
	s.join(c, []string{"/join", "general"})

	buf := make([]byte, 1024)
	n, _ := clientConn.Read(buf)
	response := string(buf[:n])

	if !strings.Contains(response, "welcome to general") {
		t.Errorf("expected welcome message, got: %s", response)
	}
	if c.room == nil {
		t.Fatal("expected client to be in a room")
	}
	if c.room.name != "general" {
		t.Errorf("expected room name 'general', got '%s'", c.room.name)
	}
}

func TestServer_Join_NoArg(t *testing.T) {
	s := newServer()
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	c := &client{conn: serverConn}
	s.join(c, []string{"/join"})

	buf := make([]byte, 1024)
	n, _ := clientConn.Read(buf)
	response := string(buf[:n])

	if !strings.Contains(response, "room name is required") {
		t.Errorf("expected error about missing room name, got: %s", response)
	}
}

func TestServer_Join_CreatesNewRoom(t *testing.T) {
	s := newServer()
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	c := &client{conn: serverConn, nick: "testuser"}
	s.join(c, []string{"/join", "newroom"})

	clientConn.Read(make([]byte, 1024))

	if _, ok := s.rooms["newroom"]; !ok {
		t.Error("expected room 'newroom' to be created")
	}
}

func TestServer_ListRooms(t *testing.T) {
	s := newServer()
	conn, reader := net.Pipe()
	defer conn.Close()

	s.rooms["general"] = &room{name: "general", members: make(map[net.Addr]*client)}
	s.rooms["random"] = &room{name: "random", members: make(map[net.Addr]*client)}

	c := &client{conn: conn}
	s.listRooms(c)

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	response := string(buf[:n])

	if !strings.Contains(response, "general") {
		t.Errorf("expected 'general' in response, got: %s", response)
	}
	if !strings.Contains(response, "random") {
		t.Errorf("expected 'random' in response, got: %s", response)
	}
}

func TestServer_ListRooms_Empty(t *testing.T) {
	s := newServer()
	conn, reader := net.Pipe()
	defer conn.Close()

	c := &client{conn: conn}
	s.listRooms(c)

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	response := string(buf[:n])

	if !strings.Contains(response, "available rooms") {
		t.Errorf("expected room list header, got: %s", response)
	}
}

func TestServer_Msg(t *testing.T) {
	s := newServer()

	conn1, reader1 := net.Pipe()
	defer conn1.Close()
	c1 := &client{conn: conn1, nick: "alice"}

	conn2, reader2 := net.Pipe()
	defer conn2.Close()
	c2 := &client{conn: conn2, nick: "bob"}

	s.join(c1, []string{"/join", "general"})
	reader1.Read(make([]byte, 1024))

	s.join(c2, []string{"/join", "general"})
	reader1.Read(make([]byte, 1024))
	reader2.Read(make([]byte, 1024))

	s.msg(c2, []string{"/msg", "hello", "everyone"})

	buf := make([]byte, 1024)
	n, _ := reader1.Read(buf)
	response := string(buf[:n])

	if !strings.Contains(response, "bob: hello everyone") {
		t.Errorf("expected 'bob: hello everyone', got: %s", response)
	}
}

func TestServer_Msg_NoArg(t *testing.T) {
	s := newServer()
	conn, reader := net.Pipe()
	defer conn.Close()

	c := &client{conn: conn, nick: "testuser"}
	s.msg(c, []string{"/msg"})

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	response := string(buf[:n])

	if !strings.Contains(response, "message is required") {
		t.Errorf("expected error about missing message, got: %s", response)
	}
}

func TestServer_Quit(t *testing.T) {
	s := newServer()
	conn, reader := net.Pipe()
	defer conn.Close()

	c := &client{conn: conn, nick: "testuser"}
	s.join(c, []string{"/join", "general"})
	reader.Read(make([]byte, 1024))

	addr := conn.RemoteAddr()
	s.quit(c)

	buf := make([]byte, 1024)
	n, _ := reader.Read(buf)
	response := string(buf[:n])

	if !strings.Contains(response, "sad to see you go") {
		t.Errorf("expected goodbye message, got: %s", response)
	}

	if c.room != nil {
		t.Error("expected client room to be nil after quit")
	}

	if _, ok := s.rooms["general"].members[addr]; ok {
		t.Error("expected client to be removed from room members")
	}
}

func TestServer_QuitCurrentRoom(t *testing.T) {
	s := newServer()
	conn, reader := net.Pipe()
	defer conn.Close()

	c := &client{conn: conn, nick: "testuser"}
	s.join(c, []string{"/join", "general"})
	reader.Read(make([]byte, 1024))

	s.quitCurrentRoom(c)
	if c.room != nil {
		t.Error("expected client room to be nil after quitCurrentRoom")
	}
}

func TestServer_QuitCurrentRoom_NoRoom(t *testing.T) {
	s := newServer()
	conn, _ := net.Pipe()
	defer conn.Close()

	c := &client{conn: conn, nick: "testuser"}
	s.quitCurrentRoom(c)

	if c.room != nil {
		t.Error("expected client room to stay nil")
	}
}

func TestServer_Run_Dispatch(t *testing.T) {
	s := newServer()
	go s.run()

	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	c := &client{
		conn:     serverConn,
		nick:     "anonymous",
		commands: s.commands,
	}

	s.commands <- command{
		id:     CMD_NICK,
		client: c,
		args:   []string{"/nick", "dispatch_user"},
	}

	buf := make([]byte, 1024)
	n, _ := clientConn.Read(buf)
	response := string(buf[:n])

	if !strings.Contains(response, "dispatch_user") {
		t.Errorf("expected nick response, got: %s", response)
	}
}

func TestServer_Run_AllCommands(t *testing.T) {
	tests := []struct {
		name     string
		cmdID    commandID
		args     []string
		expected string
	}{
		{"/nick", CMD_NICK, []string{"/nick", "bot"}, "call you bot"},
		{"/rooms", CMD_ROOMS, []string{"/rooms"}, "available rooms"},
		{"/quit", CMD_QUIT, []string{"/quit"}, "sad to see you go"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newServer()
			go s.run()

			serverConn, clientConn := net.Pipe()
			defer serverConn.Close()
			defer clientConn.Close()
			defer func() {
				s.commands <- command{id: CMD_QUIT, client: &client{conn: serverConn}}
			}()

			if tt.name == "/quit" {
				s.rooms["lobby"] = &room{name: "lobby", members: make(map[net.Addr]*client)}
			}

			c := &client{conn: serverConn, nick: "test", commands: s.commands}

			if tt.name == "/rooms" {
				s.rooms["lobby"] = &room{name: "lobby", members: make(map[net.Addr]*client)}
			}

			s.commands <- command{
				id:     tt.cmdID,
				client: c,
				args:   tt.args,
			}

			buf := make([]byte, 1024)
			n, _ := clientConn.Read(buf)
			response := string(buf[:n])

			if !strings.Contains(response, tt.expected) {
				t.Errorf("expected response containing '%s', got: %s", tt.expected, response)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	s := newServer()
	go s.run()

	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	go s.newClient(serverConn)

	clientConn.Write([]byte("/nick testuser\n"))

	cmd := <-s.commands
	if cmd.id != CMD_NICK {
		t.Errorf("expected CMD_NICK, got %v", cmd.id)
	}
	if len(cmd.args) < 2 || cmd.args[1] != "testuser" {
		t.Errorf("expected args 'testuser', got %v", cmd.args)
	}
}
