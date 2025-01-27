package protocol

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

func EncodeSimpleString(value string) []byte {
	return []byte("+" + value + "\r\n")
}

func EncodeError(conn net.Conn, errorMsg string) []byte {
	return []byte("-" + errorMsg + "\r\n")
}

func EncodeInteger(value int64) []byte {
	return []byte(":" + strconv.FormatInt(value, 10) + "\r\n")
}

func EncodeBulkString(value string) []byte {
	if value == "" {
		return []byte("$-1\r\n")
	}
	return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(value), value))
}

func EncodeArray(values []string) []byte {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*%d\r\n", len(values)))
	for _, value := range values {
		sb.WriteString(string(EncodeBulkString(value)))
	}
	return []byte(sb.String())
}

func ParseRESP(reader *bufio.Reader) ([]string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	line = strings.TrimSpace(line)
	if len(line) == 0 || line[0] != '*' {
		return nil, fmt.Errorf("invalid array start")
	}

	count, err := strconv.Atoi(line[1:])
	if err != nil {
		return nil, fmt.Errorf("invalid array length")
	}

	parts := make([]string, count)
	for i := 0; i < count; i++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 || line[0] != '$' {
			return nil, fmt.Errorf("invalid bulk string start")
		}

		length, err := strconv.Atoi(line[1:])
		if err != nil || length < 0 {
			return nil, fmt.Errorf("invalid bulk string length")
		}

		part := make([]byte, length)
		if _, err := io.ReadFull(reader, part); err != nil {
			return nil, err
		}
		parts[i] = string(part)

		if _, err := reader.Discard(2); err != nil {
			return nil, err
		}
	}

	return parts, nil
}
