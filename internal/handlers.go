package internal

import (
	"sync"
)

var Handlers = map[string]func([]Value) Value{
	"PING": ping,
	"SET":  set,
	"GET":  get,
}

var SETs = map[string]string{}
var SETsMu = sync.RWMutex{}

func ping(args []Value) Value {
	if len(args) > 0 {
		return Value{Typ: "string", Str: args[0].Bulk}
	}

	return Value{Typ: "string", Str: "PONG"}
}

func set(args []Value) Value {
	if len(args) != 2 {
		return Value{Typ: "error", Str: "SET command requires 2 arguments. Key and value."}
	}

	key := args[0].Bulk
	value := args[1].Bulk

	// Here we nee to use a mutex to handle the concurrent requests.
	SETsMu.Lock()
	SETs[key] = value
	SETsMu.Unlock()

	return Value{Typ: "string", Str: "OK"}
}

func get(args []Value) Value {
	if len(args) != 1 {
		return Value{Typ: "error", Str: "GET command requires 1 argument."}
	}

	key := args[0].Bulk

	SETsMu.RLock()
	value, ok := SETs[key]
	SETsMu.RUnlock()

	if !ok {
		return Value{Typ: "null"}
	}

	return Value{Typ: "bulk", Bulk: value}
}
