package main

import (
	"testing"
)

func TestClient(t *testing.T) {
	client, err := NewClient("localhost:12345")
	if err != nil {
		t.Fatalf("failed to connect to server: %v", err)
	}
	defer client.Close()

	if res, err := client.Set("foo", "bar"); err != nil || res != "ok" {
		t.Errorf("set failed: %v, res: %s", err, res)
	}

	if val, err := client.Get("foo"); err != nil || val != "bar" {
		t.Errorf("get failed: %v, value: %s", err, val)
	}

	if res, err := client.Del("foo"); err != nil || res != "ok" {
		t.Errorf("del failed: %v, value: %s", err, res)
	}
}
