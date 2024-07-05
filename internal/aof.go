package internal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
)

type Aof struct {
	file *os.File
	rd   *bufio.Reader
	mu   sync.Mutex
}

var defaultFilePath = "./internal/data/"
var defaultFileName = "database.aof"

func NewAof() (*Aof, error) {
	f, err := os.OpenFile(defaultFilePath+defaultFileName, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	aof := &Aof{
		file: f,
		rd:   bufio.NewReader(f),
	}

	return aof, nil
}

// We use a callback function here.
// What if we directly call the handleCommand() function here?
// Does it makes this depend on handlers?
// A: Yes it makes it coupled. Best way to go with the callback.
// one other way is to use a interface but here we have only one implementation, so no point.
func (aof *Aof) Read(fn func(value Value)) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	files, err := os.ReadDir(defaultFilePath)
	if err != nil {
		return err
	}

	fmt.Println(files)

	for _, filename := range files {
		file, err := os.OpenFile(defaultFilePath+filename.Name(), os.O_RDONLY, 0644)
		if err != nil {
			return err
		}

		aof.readFile(file, fn)
		file.Close()
	}

	return nil
}

func (aof *Aof) Write(value Value) error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	fi, err := aof.file.Stat()
	if err != nil {
		return err
	}

	mb := int64(1 * 1000000) // 1mb in bytes
	if fi.Size() >= mb {
		fmt.Println("AOF chunk size exceeded. Writing to a new file...")

		aof, err = aof.reCreate()
		if err != nil {
			fmt.Println("Error re-creating AOF:", err)
			return err
		}
	}

	_, err = aof.file.Write(value.Marshal())
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (aof *Aof) Close() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	return aof.file.Close()
}

func (aof *Aof) reCreate() (*Aof, error) {
	files, err := os.ReadDir(defaultFilePath)
	if err != nil {
		fmt.Println(err)
		return aof, err
	}

	// rename existing file to a numbered file
	fmt.Println("Renaming old AOF.")
	newFilename := fmt.Sprintf(defaultFilePath+"database-%d.aof", len(files))
	err = os.Rename(aof.file.Name(), newFilename)
	if err != nil {
		fmt.Println("Rename error:", err)
		return aof, err
	}

	// re-create aof with a new default file. Otherwise it will point to the renamed file.
	aof, err = NewAof()
	if err != nil {
		fmt.Println(err)
		return aof, err
	}

	return aof, nil
}

func (aof *Aof) readFile(file *os.File, fn func(value Value)) error {
	fmt.Println("Reading file:", file.Name())

	file.Seek(0, io.SeekStart)
	reader := NewReader(file)

	for {
		value, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		fn(value)
	}

	return nil
}
