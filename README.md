# Mini Redis

**Mini Redis** is a lightweight, in-memory key-value store inspired by Redis with custom built RESP protocol, implemented in Go. It supports basic Redis commands such as `SET`, `GET`, `DEL`, lists, sets, and hashes and TTL for `SET`.

## Features

- Basic key-value storage (`SET`, `GET`, `DEL`)
- List operations (`LPUSH`, `RPUSH`, `LPOP`, `RPOP`)
- Hash operations (`HSET`, `HGET`)
- Set operations (`SADD`, `SREM`, `SMEMBERS`)
- Data persistence using snapshots (`snapshot.json`)
- Concurrent connections handling

## Running the Server
Start the Redis-like server on port `6379`:
```
go run main.go
```
You should see:
```
Redis-like server is running on :6379...
```

## Usage
You can interact with mini-redis using Telnet
```
telnet localhost 6379
```
### Example commands:

**SET Command**
```
*3
$3
SET
$3
foo
$3
bar
```
Response:
```
+OK
```

**GET Command**
```
*2
$3
GET
$3
foo
```
Response:
```
$3
bar
```

**DEL Command**
```
*2
$3
DEL
$3
foo
```
Response:
```
:1
```
(`:1` indicates one key was deleted.)

**GET Command After Deletion**
```
*2
$3
GET
$3
foo
```
Response:
```
$-1
```
(`$-1` indicates the key does not exist.)

**LPUSH Command**
```
*3
$5
LPUSH
$4
list
$5
value
```
Response:
```
:1
```
(Indicates the new length of the list.)

**RPUSH Command**
```
*3
$5
RPUSH
$4
list
$5
value
```
Response:
```
:1
```
(New length of the list.)

**LPOP Command**
```
*2
$4
LPOP
$4
list
```
Response:
```
$5
value
```
(If the list is empty, the response is `$-1`.)

**RPOP Command**
```
*2
$4
RPOP
$4
list
```
Response:
```
$5
value
```
(If the list is empty, the response is `$-1`.)

**HSET Command**
```
*4
$4
HSET
$4
hash
$4
key1
$5
value
```
Response:
```
:1
```
(Indicates if a new field was created.)

**HGET Command**
```
*3
$4
HGET
$4
hash
$4
key1
```
Response:
```
$5
value
```
(If the field doesn't exist, the response is `$-1`.)

**SADD Command**
```
*3
$4
SADD
$3
set
$5
value
```
Response:
```
:1
```
(Indicates the number of members added.)

**SREM Command**
```
*3
$4
SREM
$3
set
$5
value
```
Response:
```
:1
```
(Indicates the number of members removed.)

**SMEMBERS Command**
```
*2
$8
SMEMBERS
$3
set
```
Response:
```
*2
$5
value
$5
value2
```
(The response is a RESP array of all members in the set.)

## Client Usage
A Go client is included (`client/client.go`)

## Persistence
Data is saved in `snapshot.json`. If the server crashes, it will restore data from the snapshot on restart
