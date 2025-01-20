package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

type Client struct {
	conn net.Conn
}

func NewClient(address string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server :%w", err)
	}

	return &Client{conn: conn}, nil
}

func (c *Client) sendComamand(command string) (string, error) {
	_, err := c.conn.Write([]byte(command))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	reader := bufio.NewReader(c.conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to send res: %w", err)
	}

	return strings.TrimSpace(response), nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Set(key, value string) (string, error) {
	command := fmt.Sprintf("SET %s %s\n", key, value)
	return c.sendComamand(command)
}

func (c *Client) Get(key string) (string, error) {
	command := fmt.Sprintf("GET %s\n", key)
	return c.sendComamand(command)
}

func (c *Client) Del(key string) (string, error) {
	command := fmt.Sprintf("DEL %s\n", key)
	return c.sendComamand(command)
}

func main() {
	c, err := NewClient("localhost:12345")
	if err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}
	defer c.Close()

	resp, err := c.Set("name", "Alice")
	if err != nil {
		log.Fatalf("Set failed: %v", err)
	}
	fmt.Println("SET response:", resp)

	value, err := c.Get("name")
	if err != nil {
		log.Fatalf("Get failed: %v", err)
	}
	fmt.Println("GET response:", value)

	resp, err = c.Del("name")
	if err != nil {
		log.Fatalf("Del failed: %v", err)
	}
	fmt.Println("DEL response:", resp)
}
