package main

import (
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

		command := strings.ToUpper(value.array[0].bulk)
		args := value.array[1:]

		handler, ok := Handlers[command]
		if !ok {
			writer.Write(Value{typ: "error", str: "Invalid command: " + command})
			continue
		}

		res := handler(args)
		writer.Write(res)
	}
}

// func read() {
// 	input := "$5\r\nAhmed\r\n"

// 	reader := bufio.NewReader(strings.NewReader(input))

// 	b, _ := reader.ReadByte()

// 	if b != '$' {
// 		fmt.Println("invalid type, expecting strings only")
// 		os.Exit(1)
// 	}

// 	size, _ := reader.ReadByte()
// 	strSize, _ := strconv.ParseInt(string(size), 10, 64)

// 	// read \r\n
// 	reader.ReadByte()
// 	reader.ReadByte()

// 	// read the value into a byte array
// 	name := make([]byte, strSize)
// 	reader.Read(name)

// 	fmt.Println(string(name))
// }
