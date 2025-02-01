package main

import (
	"bufio"
	"bytes"
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
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	return &Client{conn: conn}, nil
}

func (c *Client) readResponse() (string, error) {
	reader := bufio.NewReader(c.conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	line = strings.TrimSpace(line)

	switch line[0] {
	case '+':
		return line[1:], nil
	case '-':
		return "", fmt.Errorf("redis error: %s", line[1:])
	case ':':
		return line[1:], nil
	case '$':
		var length int
		fmt.Sscanf(line, "$%d", &length)
		if length == -1 {
			return "", nil
		}
		data := make([]byte, length+2)
		_, err := reader.Read(data)
		if err != nil {
			return "", fmt.Errorf("failed to read bulk string: %w", err)
		}
		return string(data[:length]), nil
	case '*':
		var count int
		fmt.Sscanf(line, "*%d", &count)
		var result []string
		for i := 0; i < count; i++ {
			resp, err := c.readResponse()
			if err != nil {
				return "", err
			}
			result = append(result, resp)
		}
		return strings.Join(result, "\n"), nil
	}
	return "", fmt.Errorf("unknown response type: %s", line)
}

func (c *Client) sendCommand(command []string) (string, error) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("*%d\r\n", len(command)))
	for _, arg := range command {
		buffer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
	}

	_, err := c.conn.Write(buffer.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to send command: %w", err)
	}

	return c.readResponse()
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Set(key, value string) (string, error) {
	return c.sendCommand([]string{"SET", key, value})
}

func (c *Client) Get(key string) (string, error) {
	return c.sendCommand([]string{"GET", key})
}

func (c *Client) Del(key string) (string, error) {
	return c.sendCommand([]string{"DEL", key})
}

func (c *Client) LPush(key string, values []string) (string, error) {
	args := append([]string{"LPUSH", key}, values...)
	return c.sendCommand(args)
}

func (c *Client) RPush(key string, values []string) (string, error) {
	args := append([]string{"RPUSH", key}, values...)
	return c.sendCommand(args)
}

func (c *Client) SAdd(key string, members []string) (string, error) {
	args := append([]string{"SADD", key}, members...)
	return c.sendCommand(args)
}

func (c *Client) HSet(key, field, value string) (string, error) {
	return c.sendCommand([]string{"HSET", key, field, value})
}

func (c *Client) LPop(key string) (string, error) {
	return c.sendCommand([]string{"LPOP", key})
}

func (c *Client) RPop(key string) (string, error) {
	return c.sendCommand([]string{"RPOP", key})
}

func (c *Client) HGet(key, field string) (string, error) {
	return c.sendCommand([]string{"HGET", key, field})
}

func (c *Client) SMembers(key string) (string, error) {
	return c.sendCommand([]string{"SMEMBERS", key})
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

	resp, err = c.LPush("mylist", []string{"item1", "item2"})
	if err != nil {
		log.Fatalf("LPush failed: %v", err)
	}
	fmt.Println("LPUSH response:", resp)

	resp, err = c.RPush("mylist", []string{"item3"})
	if err != nil {
		log.Fatalf("RPush failed: %v", err)
	}
	fmt.Println("RPUSH response:", resp)

	resp, err = c.SAdd("myset", []string{"member1", "member2"})
	if err != nil {
		log.Fatalf("SAdd failed: %v", err)
	}
	fmt.Println("SADD response:", resp)

	resp, err = c.HSet("myhash", "field1", "value1")
	if err != nil {
		log.Fatalf("HSet failed: %v", err)
	}
	fmt.Println("HSET response:", resp)

	resp, err = c.LPop("mylist")
	if err != nil {
		log.Fatalf("LPop failed: %v", err)
	}
	fmt.Println("LPOP response:", resp)

	resp, err = c.RPop("mylist")
	if err != nil {
		log.Fatalf("RPop failed: %v", err)
	}
	fmt.Println("RPOP response:", resp)

	resp, err = c.Del("name")
	if err != nil {
		log.Fatalf("Del failed: %v", err)
	}
	fmt.Println("DEL response:", resp)

	resp, err = c.SMembers("myset")
	if err != nil {
		log.Fatalf("SMembers failed: %v", err)
	}
	fmt.Println("SMEMBERS response:", resp)
}
