package main

import (
	"fmt"
	"net"

	"github.com/MalindaWMD/redis-from-scratch/internal"
)

func main() {
	// create tcp listener
	l, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println(err)
		return
	}

	// start accepting requests
	conn, err := l.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	fmt.Println("Listening on port:6379")

	config, err := internal.LoadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	// open AOF
	aof, err := internal.NewAof(config)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer aof.Close()

	// read AOF
	fmt.Println("Reading persisted data from AOF")
	go aof.Read(func(value internal.Value) {
		internal.HandleCommand(value)
	})

	// read from connection
	for {
		resp := internal.NewReader(conn)
		value, err := resp.Read()
		if err != nil {
			fmt.Println(err)
			return
		}

		if value.Typ != "array" {
			fmt.Println("Invalid request, expected array")
			continue
		}

		if len(value.Array) == 0 {
			fmt.Println("Invalid request, expected array length > 0")
			continue
		}

		writer := internal.NewWriter(conn)

		res, command, err := internal.HandleCommand(value)
		if err != nil {
			writer.Write(internal.Value{Typ: "error", Str: err.Error()})
			continue
		}

		if command == "SET" || command == "HSET" || command == "DEL" {
			aof.Write(value)
		}

		writer.Write(res)
	}
}
