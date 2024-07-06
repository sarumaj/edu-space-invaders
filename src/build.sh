#!/bin/bash

set -e

print_usage() {
	echo "Usage: $0 [-d target_directory]"
	echo "  -d target_directory   Directory where the project will be created (default: current directory)"
}

log_message() {
	local message
	local timestamp
	message="$1"
	timestamp="$(date +"%Y-%m-%d %H:%M:%S.%3N")"
	echo "[$timestamp] $message"
}

SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"

# Default values
TARGET_DIR="."

# Parse command line options
while getopts 'n:d:' flag; do
	case "${flag}" in
	d) TARGET_DIR="${OPTARG}" ;;
	*)
		print_usage
		exit 1
		;;
	esac
done

# Create target directory if it doesn't exist
log_message "Creating target directory: $TARGET_DIR"
mkdir -p "$TARGET_DIR"

# Build the Go program
log_message "Building the Go program"
GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o "$TARGET_DIR/main.wasm" "$SCRIPT_DIR/main.go"
if [ -f "$TARGET_DIR/main.wasm" ]; then
	log_message "Go program built successfully"
else
	log_message "Failed to build the Go program"
	exit 1
fi

# Download the Go runtime for WebAssembly
log_message "Downloading the Go runtime for WebAssembly"
curl \
	--retry 3 \
	--retry-all-errors \
	--retry-delay 5 \
	-sL https://raw.githubusercontent.com/golang/go/master/misc/wasm/wasm_exec.js \
	-o "$TARGET_DIR/wasm_exec.js"
log_message "Go runtime downloaded successfully"

# Copy the static files
log_message "Copying static files"
cp -fr "$SCRIPT_DIR/static/"* "$TARGET_DIR/"
log_message "Static files copied successfully"

# Create the fs.go file
log_message "Creating fs.go file"
cat <<EOL >"$TARGET_DIR/fs.go"
// Code generated on $(date +"%Y-%m-%d %H:%M:%S.%3N")
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
EOL
log_message "fs.go file created successfully"
