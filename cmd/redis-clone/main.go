package main

import (
	"fmt"
	"net"
	"strings"

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

	// read from connection
	for {
		resp := internal.NewReader(conn)
		value, err := resp.Read()
		if err != nil {
			fmt.Println(err)
			return
		}

		writer := internal.NewWriter(conn)

		command := strings.ToUpper(value.Array[0].Bulk)
		args := value.Array[1:]

		handler, ok := internal.Handlers[command]
		if !ok {
			writer.Write(internal.Value{Typ: "error", Str: "Invalid command: " + command})
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

// 	// read the value into a byte Array
// 	name := make([]byte, strSize)
// 	reader.Read(name)

// 	fmt.Println(string(name))
// }
