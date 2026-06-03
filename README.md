# Go TCP Chat

Multi-room TCP chat server written in Go.

## Commands

- `/nick <name>` — Set your nickname
- `/join <room>` — Join or create a room
- `/rooms` — List available rooms
- `/msg <text>` — Send message to current room
- `/quit` — Leave current room

## Run

```bash
go run main.go
```

## Test

```bash
go test ./...
```
