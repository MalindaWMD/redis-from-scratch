package internal

import (
	"bufio"
	"io"
	"strconv"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

type Value struct {
	Typ   string  // data type of the value
	Str   string  // value of the string
	Bulk  string  // value of the bulk string
	Array []Value // value of the array
}

type Reader struct {
	reader *bufio.Reader
}

type Writer struct {
	writer io.Writer
}

func NewReader(rd io.Reader) *Reader {
	return &Reader{
		reader: bufio.NewReader(rd),
	}
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: w,
	}
}

func (r *Reader) Read() (Value, error) {
	_type, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch _type {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	default:
		return Value{}, nil
	}
}

func (w *Writer) Write(v Value) error {
	var bytes = v.Marshal()

	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}

func (r *Reader) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}

		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}

	return line[:len(line)-2], n, nil
}

func (r *Reader) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}

	i, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}

	return int(i), n, nil
}

func (r *Reader) readArray() (Value, error) {
	v := Value{}
	v.Typ = "array"

	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	v.Array = make([]Value, 0)
	for i := 0; i < len; i++ {
		val, err := r.Read()
		if err != nil {
			return v, err
		}

		v.Array = append(v.Array, val)
	}

	return v, nil
}

func (r *Reader) readBulk() (Value, error) {
	v := Value{}
	v.Typ = "bulk"

	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	bulk := make([]byte, len)

	r.reader.Read(bulk)

	v.Bulk = string(bulk)

	// Read remaining \r\n or CRLF
	r.readLine()

	return v, nil
}

func (v *Value) Marshal() []byte {
	switch v.Typ {
	case "array":
		return v.marshalArray()
	case "bulk":
		return v.marshalBulk()
	case "string":
		return v.marshalString()
	case "null":
		return v.marshalNull()
	case "error":
		return v.marshalError()
	default:
		return []byte{}
	}
}

func (v *Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, STRING)
	bytes = append(bytes, v.Str...) // v.Str... treat string as a slice
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v *Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(len(v.Bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.Bulk...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v *Value) marshalArray() []byte {
	var bytes []byte
	bytes = append(bytes, ARRAY)
	bytes = append(bytes, strconv.Itoa(len(v.Array))...)
	bytes = append(bytes, '\r', '\n')

	for _, value := range v.Array {
		bytes = append(bytes, value.Marshal()...)
	}

	return bytes
}

func (v *Value) marshalError() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.Str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v *Value) marshalNull() []byte {
	return []byte("$-1\r\n")
}
