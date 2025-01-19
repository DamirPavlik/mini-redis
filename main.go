package main

import (
	"bufio"
	"fmt"
	"mini-redis/store"
	"net"
	"strconv"
	"strings"
	"time"
)

func main() {
	store := store.NewKVStore()
	snapshotFile := "snapshot.json"

	err := store.LoadSnapshot(snapshotFile)
	if err != nil {
		fmt.Println("Failed to load snapshot:", err)
		return
	}

	go func() {
		for range time.Tick(10 * time.Second) {
			if err := store.SaveSnapshot(snapshotFile); err != nil {
				fmt.Println("failed to save snapshot:", err)
			}
		}
	}()

	go func() {
		for range time.Tick(1 * time.Second) {
			store.CleanupExpiredKeys()
		}
	}()

	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		fmt.Println("failed to start the server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("server is running at port 12345")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("err accepting conns", err)
			continue
		}
		go handleConn(conn, store)
	}
}

func handleConn(conn net.Conn, store *store.KeyValueStore) {
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

func handleCommand(command string, store *store.KeyValueStore) string {
	parts := strings.SplitN(command, " ", 5)
	if strings.ToUpper(parts[0]) == "SET" {
		if len(parts) < 3 {
			return "err: set requires key and a value"
		}

		ttl := 0
		if len(parts) == 5 && strings.ToUpper(parts[3]) == "EX" {
			ttlValue, err := strconv.Atoi(parts[4])
			if err != nil {
				return "err: invalid ttl value"
			}
			ttl = ttlValue
		}
		store.Set(parts[1], parts[2], ttl)
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
