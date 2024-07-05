package main

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

func main() {
	// create tcp listener
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}

	aof, err := NewAof("database.aof")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer aof.Close()

	aof.Read(func(value Value) {
		handleCommand(value)
	})

	// start accepting requests
	conn, err := l.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	fmt.Println("Listening on port:6379")

	// read from connection
	for {
		resp := NewReader(conn)
		value, err := resp.Read()
		if err != nil {
			fmt.Println(err)
			return
		}

		writer := NewWriter(conn)

		res, err := handleCommand(value)
		if err != nil {
			writer.Write(Value{typ: "error", str: err.Error()})
			continue
		}

		// Not good. Format
		if strings.ToUpper(value.array[0].bulk) == "SET" {
			aof.Write(value)
		}

		writer.Write(res)
	}
}

func handleCommand(value Value) (Value, error) {
	if len(value.array) == 0 {
		return Value{}, nil
	}

	command := strings.ToUpper(value.array[0].bulk)
	args := value.array[1:]

	handler, ok := Handlers[command]
	if !ok {
		fmt.Println("Invalid command")
		return Value{}, errors.New("invalid command: " + command)
	}

	res := handler(args)

	return res, nil
}
