package reader

import (
	"bytes"
	"io"
	"os"
)

type FileReader struct {
	Source string
	Buffer *bytes.Buffer
	fd     *os.File
	done   bool
}

func (re *FileReader) Open() error {
	fd, err := os.Open(re.Source)
	re.fd = fd
	re.Buffer = &bytes.Buffer{}
	return err
}

func (re *FileReader) Read() {
	if re.done {
		return
	}

	if re.fd == nil {
		panic("fd is nil")
	}

	buffer := make([]byte, 1024)
	size, err := re.fd.Read(buffer)
	if err == io.EOF || size <= 0 {
		re.fd.Close()
		re.done = true
		return
	}

	re.Buffer.Write(buffer[:size])
}

func (re *FileReader) Done() bool { return re.done }
