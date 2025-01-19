package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

func main() {
	store := NewKVStore()
	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		fmt.Println("failed to start the server:", err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("err acceping conns", err)
			continue
		}
		go handleConn(conn, store)
	}
}

func handleConn(conn net.Conn, store *KeyValueStore) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("err reading command:", err)
			return
		}

		response := handleCommand(strings.TrimSpace(command), store)
		conn.Write([]byte(response + "\n"))
	}
}

func handleCommand(command string, store *KeyValueStore) string {
	parts := strings.SplitN(command, " ", 3)
	if strings.ToUpper(parts[0]) == "SET" {
		if len(parts) < 3 {
			return "err: set requires key and a value"
		}
		store.Set(parts[1], parts[2])
		return "ok"
	}

	if strings.ToUpper(parts[0]) == "GET" {
		if len(parts) < 2 {
			return "err: get requires a key"
		}
		value, exists := store.Get(parts[1])
		if !exists {
			return "err: key does not exist"
		}
		return value
	}

	if strings.ToUpper(parts[0]) == "DEL" {
		if len(parts) < 2 {
			return "err: del requires a key"
		}
		store.Del(parts[1])
		return "ok"
	}

	return "err: command does not exist"
}
