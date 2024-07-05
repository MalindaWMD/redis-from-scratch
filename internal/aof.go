package internal

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
)

type Aof struct {
	file   *os.File
	rd     *bufio.Reader
	mu     sync.Mutex
	config *Config
}

var defaultFileName = "database.aof"

func NewAof(config Config) (*Aof, error) {
	f, err := createDefaultFile(config)
	if err != nil {
		return nil, err
	}

	aof := &Aof{
		file:   f,
		rd:     bufio.NewReader(f),
		config: &config,
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

	files, err := os.ReadDir(aof.config.AOFDir)
	if err != nil {
		return err
	}

	fmt.Println(files)

	for _, filename := range files {
		path := getFilename(aof.config.AOFDir, filename.Name())
		file, err := os.OpenFile(path, os.O_RDONLY, 0644)
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

	size := int64(aof.config.AOFMaxSize)
	if fi.Size() >= size {
		fmt.Printf("AOF chunk size of %d bytes exceeded. Writing to a new file...\n", size)

		f, err := aof.reCreate()
		if err != nil {
			fmt.Println("Error re-creating AOF:", err)
			return err
		}

		// assing newly created file
		aof.file = f
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

func (aof *Aof) reCreate() (*os.File, error) {
	files, err := os.ReadDir(aof.config.AOFDir)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// rename existing file to a numbered file
	fmt.Println("Renaming old AOF.")
	newFilename := fmt.Sprintf("database-%d.aof", len(files))
	path := getFilename(aof.config.AOFDir, newFilename)
	err = os.Rename(aof.file.Name(), path)
	if err != nil {
		fmt.Println("Rename error:", err)
		return nil, err
	}

	// re-create aof with a new default file. Otherwise it will point to the renamed file.
	f, err := createDefaultFile(*aof.config)
	if err != nil {
		return nil, err
	}

	return f, nil
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

func createDefaultFile(config Config) (*os.File, error) {
	path := getFilename(config.AOFDir, defaultFileName)
	return os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)
}

func getFilename(dir string, name string) string {
	return dir + name
}
