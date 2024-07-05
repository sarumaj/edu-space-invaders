// Code generated on 2024-07-05 13:56:22.502
package dist

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed *.html *.css *.js *.wasm *.ico
var embeddedFsys embed.FS

var _ http.File = httpFile{}

var _ fs.FileInfo = httpFileInfo{}

var Hashes = func() map[string]string {
	entries, err := embeddedFsys.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	hashes := make(map[string]string)
	for _, entry := range entries {
		content, err := embeddedFsys.ReadFile(entry.Name())
		if err != nil {
			log.Fatal(err)
		}

		hash := sha256.Sum256(content)
		hashes[entry.Name()] = hex.EncodeToString(hash[:])
	}

	return hashes
}()

var HttpFS http.FileSystem = httpFS{http.FS(embeddedFsys)}

type httpFile struct {
	name   string
	size   int64
	reader io.ReadSeeker
}

func (h httpFile) Close() error                                 { return nil }
func (h httpFile) Readdir(_ int) ([]fs.FileInfo, error)         { return nil, fs.ErrNotExist }
func (h httpFile) Read(p []byte) (n int, err error)             { return h.reader.Read(p) }
func (h httpFile) Seek(offset int64, whence int) (int64, error) { return h.reader.Seek(offset, whence) }
func (h httpFile) Stat() (fs.FileInfo, error)                   { return httpFileInfo{name: h.name, size: h.size}, nil }

type httpFileInfo struct {
	name string
	size int64
}

func (h httpFileInfo) Name() string     { return filepath.Base(h.name) }
func (h httpFileInfo) Size() int64      { return h.size }
func (httpFileInfo) Mode() fs.FileMode  { return fs.FileMode(os.ModePerm) }
func (httpFileInfo) ModTime() time.Time { return time.Time{} }
func (httpFileInfo) IsDir() bool        { return false }
func (httpFileInfo) Sys() any           { return nil }

type httpFS struct{ fsys http.FileSystem }

func (h httpFS) Open(name string) (http.File, error) {
	if !strings.HasSuffix(name, ".sha256") {
		return h.fsys.Open(name)
	}

	if hash, ok := Hashes[strings.TrimSuffix(filepath.Base(name), ".sha256")]; ok {
		return httpFile{
			name:   name,
			size:   int64(len(hash)),
			reader: strings.NewReader(hash),
		}, nil
	}

	return nil, fs.ErrNotExist
}
