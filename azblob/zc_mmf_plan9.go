// +build plan9

package azblob

import (
	"os"
	"sync"
)

// Plan 9 does not have native memory-mapping of files
// Derived from: https://play.golang.org/p/Bl-C2Ku51Z0

type mmf []byte

// Use 2 maps and a type since mmf type must be []byte for compatibility
type mfile struct {
	*os.File
	offset		int64
}

var (
	buffers		map[*byte][]byte
	files		map[*byte]mfile
	mapLock		sync.Mutex
)

func newMMF(file *os.File, writable bool, offset int64, length int) (mmf, error) {
	if buffers == nil {
		buffers = make(map[*byte][]byte)
	}
	if files == nil {
		files = make(map[*byte]mfile)
	}

	b := make([]byte, length)

	p := &b[cap(b)-1]

	mapLock.Lock()
	defer mapLock.Unlock()

	buffers[p] = b
	files[p] = mfile{file, offset}

	return b, nil
}

func (m *mmf) unmap() {
	data := []byte(*m)

	p := &data[cap(data)-1]

	mapLock.Lock()
	defer mapLock.Unlock()

	b := buffers[p]
	if b == nil || &b[0] != &data[0] {
		panic("if we are unable to unmap the memory-mapped file, there is serious concern for memory corruption")
	}

	// Write out the buffer to the file
	mf := files[p]
	_, err := mf.WriteAt(b, mf.offset)
	if err != nil {
		panic("cannot write out file from memory")
	}

	delete(buffers, p)
	delete(files, p)
}
