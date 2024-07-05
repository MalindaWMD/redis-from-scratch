package internal

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RedisEntry struct {
	value  string
	expire time.Time // 0 -> indefinitely
}

var Handlers = map[string]func([]Value) Value{
	"PING":    ping,
	"SET":     set,
	"GET":     get,
	"HSET":    hset,
	"HGET":    hget,
	"HGETALL": hgetAll,
	"DEL":     del,
	"EXPIRE":  expire,
	"TTL":     ttl,
}

var SETs = map[string]RedisEntry{}
var SETsMu = sync.RWMutex{}

var HSETs = map[string]map[string]string{}
var HSEtsMu = sync.RWMutex{}

func ping(args []Value) Value {
	if len(args) > 0 {
		return Value{Typ: "string", Str: args[0].Bulk}
	}

	return Value{Typ: "string", Str: "PONG"}
}

func set(args []Value) Value {
	if len(args) < 2 {
		return Value{Typ: "error", Str: "SET command requires at least 2 arguments. Key and value."}
	}

	key := args[0].Bulk
	value := args[1].Bulk

	entry := RedisEntry{value: value}

	if len(args) == 4 {
		expire, _ := strconv.Atoi(args[3].Bulk)
		entry.expire = time.Now().Add(time.Duration(expire) * time.Second)
	}

	// Here we nee to use a mutex to handle the concurrent requests.
	SETsMu.Lock()
	SETs[key] = entry
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

	if !value.expire.IsZero() && time.Now().After(value.expire) {
		delete(SETs, key)
		return Value{Typ: "null"}
	}

	return Value{Typ: "bulk", Bulk: value.value}
}

func hset(args []Value) Value {
	if len(args) < 3 {
		return Value{Typ: "error", Str: "HSET command requires at least 3 arguments."}
	}

	hash := args[0].Bulk

	HSEtsMu.Lock()
	if _, ok := HSETs[hash]; !ok {
		HSETs[hash] = map[string]string{}
	}

	for i := 1; i < len(args); i = i + 2 {
		key := args[i].Bulk
		value := ""
		if i+1 < len(args) {
			value = args[i+1].Bulk
		}

		if value == "" {
			continue
		}

		HSETs[hash][key] = value
	}
	HSEtsMu.Unlock()

	return Value{Typ: "string", Str: "OK"}
}

func hget(args []Value) Value {
	if len(args) != 2 {
		return Value{Typ: "error", Str: "HGET command requires 2 arguments."}
	}

	hash := args[0].Bulk
	key := args[1].Bulk

	HSEtsMu.RLock()
	value, ok := HSETs[hash][key]
	HSEtsMu.RUnlock()
	if !ok {
		return Value{Typ: "null"}
	}

	return Value{Typ: "bulk", Bulk: value}
}

func hgetAll(args []Value) Value {
	if len(args) != 1 {
		return Value{Typ: "error", Str: "HGETALL command requires 1 argument."}
	}

	hash := args[0].Bulk

	HSEtsMu.RLock()
	values, ok := HSETs[hash]
	HSEtsMu.RUnlock()

	if !ok {
		return Value{Typ: "null"}
	}

	valueSlice := []Value{}
	for _, v := range values {
		valueSlice = append(valueSlice, Value{
			Typ: "bulk", Bulk: v,
		})
	}

	return Value{Typ: "array", Array: valueSlice}
}

func del(args []Value) Value {
	if len(args) != 1 {
		return Value{Typ: "error", Str: "DEL command requires 1 argument."}
	}

	for i := 0; i < len(args); i++ {
		key := args[i].Bulk

		delete(SETs, key)
		delete(HSETs, key)
	}

	return Value{Typ: "string", Str: "OK"}
}

func expire(args []Value) Value {
	if len(args) != 2 {
		return Value{Typ: "error", Str: "EXPIRE command requires 2 arguments."}
	}

	key := args[0].Bulk
	ttl, _ := strconv.Atoi(args[1].Bulk)

	SETsMu.Lock()
	value, ok := SETs[key]
	if !ok {
		SETsMu.Unlock()
		return Value{Typ: "null"}
	}

	value.expire = time.Now().Add(time.Duration(ttl) * time.Second)
	SETs[key] = value
	SETsMu.Unlock()

	return Value{Typ: "string", Str: "OK"}
}

func ttl(args []Value) Value {
	if len(args) != 1 {
		return Value{Typ: "error", Str: "TTL command requires 1 argument."}
	}

	key := args[0].Bulk

	SETsMu.RLock()
	value, ok := SETs[key]
	SETsMu.RUnlock()

	if !ok {
		return Value{Typ: "null"}
	}

	seconds := time.Since(value.expire).Abs().Seconds()
	remainingSeconds := int(math.Ceil(seconds))

	return Value{Typ: "string", Str: strconv.Itoa(remainingSeconds)}
}

func HandleCommand(value Value) (Value, string, error) {
	if len(value.Array) == 0 {
		return Value{}, "", nil
	}

	command := strings.ToUpper(value.Array[0].Bulk)
	args := value.Array[1:]

	handler, ok := Handlers[command]
	if !ok {
		return Value{}, "", errors.New("Invalid command: " + command)
	}

	res := handler(args)

	return res, command, nil
}
