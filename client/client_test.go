package main

import (
	"testing"
)

func TestClient(t *testing.T) {
	client, err := NewClient("localhost:6379")
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer client.Close()

	res, err := client.Set("foo", "bar")
	if err != nil || res != "+OK" {
		t.Errorf("SET failed: %v, response: %s", err, res)
	}

	val, err := client.Get("foo")
	if err != nil || val != "bar" {
		t.Errorf("GET failed: %v, value: %s", err, val)
	}

	res, err = client.Del("foo")
	if err != nil || res != ":1" {
		t.Errorf("DEL failed: %v, response: %s", err, res)
	}

	res, err = client.LPush("mylist", []string{"item1", "item2"})
	if err != nil || res != ":2" {
		t.Errorf("LPUSH failed: %v, response: %s", err, res)
	}

	val, err = client.LPop("mylist")
	if err != nil || val != "item2" {
		t.Errorf("LPOP failed: %v, value: %s", err, val)
	}

	res, err = client.RPush("mylist", []string{"item3"})
	if err != nil || res != ":2" {
		t.Errorf("RPUSH failed: %v, response: %s", err, res)
	}

	val, err = client.RPop("mylist")
	if err != nil || val != "item3" {
		t.Errorf("RPOP failed: %v, value: %s", err, val)
	}

	res, err = client.SAdd("myset", []string{"member1", "member2"})
	if err != nil || res != ":2" {
		t.Errorf("SADD failed: %v, response: %s", err, res)
	}

	val, err = client.SMembers("myset")
	if err != nil || val != "member1\nmember2" {
		t.Errorf("SMEMBERS failed: %v, value: %s", err, val)
	}

	res, err = client.HSet("myhash", "field1", "value1")
	if err != nil || res != ":1" {
		t.Errorf("HSET failed: %v, response: %s", err, res)
	}

	val, err = client.HGet("myhash", "field1")
	if err != nil || val != "value1" {
		t.Errorf("HGET failed: %v, value: %s", err, val)
	}
}
