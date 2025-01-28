package main

import (
	"bufio"
	"fmt"
	"io"
	"mini-redis/protocol"
	"mini-redis/store"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const snapshotFile = "snapshot.json"

func main() {
	kvStore := store.NewKVStore()

	if err := kvStore.LoadSnapshot(snapshotFile); err != nil {
		fmt.Printf("Error loading snapshot: %v\n", err)
	} else {
		fmt.Println("Snapshot loaded successfully.")
	}

	go periodicSnapshot(kvStore, snapshotFile, 30*time.Second)

	go handleSignals(kvStore, snapshotFile)

	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Redis-like server is running on :6379...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn, kvStore)
	}
}

func periodicSnapshot(kvStore *store.KeyValueStore, file string, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if err := kvStore.SaveSnapshot(file); err != nil {
			fmt.Printf("Error saving snapshot: %v\n", err)
		} else {
			fmt.Println("Snapshot saved successfully.")
		}
	}
}

func handleSignals(kvStore *store.KeyValueStore, file string) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nReceived termination signal, saving snapshot...")
	if err := kvStore.SaveSnapshot(file); err != nil {
		fmt.Printf("Error saving snapshot during shutdown: %v\n", err)
	} else {
		fmt.Println("Snapshot saved successfully.")
	}
	os.Exit(0)
}

func handleConnection(conn net.Conn, kvStore *store.KeyValueStore) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		command, err := protocol.ParseRESP(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			protocol.EncodeError(conn, "ERR invalid input")
			continue
		}
		response := executeCommand(kvStore, command)
		conn.Write(response)
	}
}

func executeCommand(kvStore *store.KeyValueStore, command []string) []byte {
	if len(command) == 0 {
		return protocol.EncodeError(nil, "ERR empty command")
	}

	cmd := strings.ToUpper(command[0])
	args := command[1:]

	switch cmd {
	case "SET":
		if len(args) < 2 || len(args) > 3 {
			return protocol.EncodeError(nil, "ERR wrong number of arguments for 'SET' command")
		}
		ttl := 0
		if len(args) == 3 {
			parsedTTL, err := strconv.Atoi(args[2])
			if err != nil {
				return protocol.EncodeError(nil, "ERR invalid TTL value")
			}
			ttl = parsedTTL
		}
		kvStore.Set(args[0], args[1], ttl)
		return protocol.EncodeSimpleString("OK")
	case "GET":
		if len(args) != 1 {
			return protocol.EncodeError(nil, "ERR wrong number of arguments for 'GET' command")
		}
		value, exists := kvStore.Get(args[0])
		if !exists {
			return protocol.EncodeBulkString("")
		}
		return protocol.EncodeBulkString(value)
	case "DEL":
		if len(args) == 0 {
			return protocol.EncodeError(nil, "ERR wrong number of arguments for 'DEL' command")
		}
		for _, key := range args {
			kvStore.Del(key)
		}
		return protocol.EncodeInteger(int64(len(args)))
	case "RPUSH":
		if len(args) < 2 {
			return protocol.EncodeError(nil, "ERR wrong number of arguments for 'RPUSH' command")
		}
		length := kvStore.RPush(args[0], args[1:]...)
		return protocol.EncodeInteger(int64(length))
	case "LPUSH":
		if len(args) < 2 {
			return protocol.EncodeError(nil, "ERR wrong number of arguments for 'LPUSH' command")
		}
		length := kvStore.LPush(args[0], args[1:]...)
		return protocol.EncodeInteger(int64(length))
	case "LPOP":
		if len(args) != 1 {
			return protocol.EncodeError(nil, "ERR wrong number of arguments for 'LPOP' command")
		}
		value, exists := kvStore.LPop(args[0])
		if !exists {
			return protocol.EncodeBulkString("")
		}
		return protocol.EncodeBulkString(value)
	case "RPOP":
		if len(args) != 1 {
			return protocol.EncodeError(nil, "ERR wrong number of arguments for 'RPOP' command")
		}
		value, exists := kvStore.RPop(args[0])
		if !exists {
			return protocol.EncodeBulkString("")
		}
		return protocol.EncodeBulkString(value)
	case "HSET":
		if len(args) != 3 {
			return protocol.EncodeError(nil, "ERR wrong number of arguments for 'HSET' command")
		}
		kvStore.HSet(args[0], args[1], args[2])
		return protocol.EncodeInteger(1)
	case "HGET":
		if len(args) != 2 {
			return protocol.EncodeError(nil, "ERR wrong number of arguments for 'HGET' command")
		}
		value, exists := kvStore.HGet(args[0], args[1])
		if !exists {
			return protocol.EncodeBulkString("")
		}
		return protocol.EncodeBulkString(value)
	case "SADD":
		if len(args) < 2 {
			return protocol.EncodeError(nil, "ERR wrong number of arguments for 'SADD' command")
		}
		count := kvStore.SAdd(args[0], args[1:]...)
		return protocol.EncodeInteger(int64(count))
	case "SREM":
		if len(args) < 2 {
			return protocol.EncodeError(nil, "ERR wrong number of arguments for 'SREM' command")
		}
		count := kvStore.SRem(args[0], args[1:]...)
		return protocol.EncodeInteger(int64(count))
	case "SMEMBERS":
		if len(args) != 1 {
			return protocol.EncodeError(nil, "ERR wrong number of arguments for 'SMEMBERS' command")
		}
		members := kvStore.SMember(args[0])
		return protocol.EncodeArray(members)
	default:
		return protocol.EncodeError(nil, fmt.Sprintf("ERR unknown command '%s'", cmd))
	}
}
