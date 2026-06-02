package main

import (
	"errors"
	"net"
	"strings"
	"testing"
)

func readConn(t *testing.T, conn net.Conn) string {
	t.Helper()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	return string(buf[:n])
}

func TestClient_Err(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	c := &client{conn: serverConn}
	c.err(errors.New("test error"))

	response := readConn(t, clientConn)
	expected := "err: test error\n"
	if response != expected {
		t.Errorf("expected '%q', got '%q'", expected, response)
	}
}

func TestClient_Msg(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	c := &client{conn: serverConn}
	c.msg("hello world")

	response := readConn(t, clientConn)
	expected := "> hello world\n"
	if response != expected {
		t.Errorf("expected '%q', got '%q'", expected, response)
	}
}

func TestClient_ReadInput_Nick(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	commands := make(chan command, 10)
	c := &client{conn: serverConn, nick: "anonymous", commands: commands}

	go c.readInput()

	clientConn.Write([]byte("/nick testuser\n"))

	cmd := <-commands
	if cmd.id != CMD_NICK {
		t.Errorf("expected CMD_NICK, got %v", cmd.id)
	}
	if len(cmd.args) < 2 || cmd.args[1] != "testuser" {
		t.Errorf("expected args[1] = 'testuser', got %v", cmd.args)
	}
}

func TestClient_ReadInput_Join(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	commands := make(chan command, 10)
	c := &client{conn: serverConn, nick: "anonymous", commands: commands}

	go c.readInput()

	clientConn.Write([]byte("/join general\n"))

	cmd := <-commands
	if cmd.id != CMD_JOIN {
		t.Errorf("expected CMD_JOIN, got %v", cmd.id)
	}
	if len(cmd.args) < 2 || cmd.args[1] != "general" {
		t.Errorf("expected args[1] = 'general', got %v", cmd.args)
	}
}

func TestClient_ReadInput_Rooms(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	commands := make(chan command, 10)
	c := &client{conn: serverConn, nick: "anonymous", commands: commands}

	go c.readInput()

	clientConn.Write([]byte("/rooms\n"))

	cmd := <-commands
	if cmd.id != CMD_ROOMS {
		t.Errorf("expected CMD_ROOMS, got %v", cmd.id)
	}
}

func TestClient_ReadInput_Msg(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	commands := make(chan command, 10)
	c := &client{conn: serverConn, nick: "anonymous", commands: commands}

	go c.readInput()

	clientConn.Write([]byte("/msg hello everyone\n"))

	cmd := <-commands
	if cmd.id != CMD_MSG {
		t.Errorf("expected CMD_MSG, got %v", cmd.id)
	}
	if len(cmd.args) < 3 || strings.Join(cmd.args[1:], " ") != "hello everyone" {
		t.Errorf("expected args 'hello everyone', got %v", cmd.args)
	}
}

func TestClient_ReadInput_Quit(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	commands := make(chan command, 10)
	c := &client{conn: serverConn, nick: "anonymous", commands: commands}

	go c.readInput()

	clientConn.Write([]byte("/quit\n"))

	cmd := <-commands
	if cmd.id != CMD_QUIT {
		t.Errorf("expected CMD_QUIT, got %v", cmd.id)
	}
}

func TestClient_ReadInput_Unknown(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	commands := make(chan command, 10)
	c := &client{conn: serverConn, nick: "anonymous", commands: commands}

	go c.readInput()

	clientConn.Write([]byte("/unknown\n"))

	response := readConn(t, clientConn)
	if !strings.Contains(response, "unknown command") {
		t.Errorf("expected 'unknown command' error, got: %s", response)
	}

	select {
	case <-commands:
		t.Error("expected no command for unknown input")
	default:
	}
}

func TestClient_ReadInput_TrimsWhitespace(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	commands := make(chan command, 10)
	c := &client{conn: serverConn, nick: "anonymous", commands: commands}

	go c.readInput()

	clientConn.Write([]byte("  /nick  spaced  \n"))

	cmd := <-commands
	if cmd.id != CMD_NICK {
		t.Errorf("expected CMD_NICK, got %v", cmd.id)
	}
}
