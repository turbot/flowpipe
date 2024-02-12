package middleware

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

type ServeFileSystem interface {
	http.FileSystem
	Exists(prefix string, path string) bool
}

type localFileSystem struct {
	http.FileSystem
	fs      fs.FS
	root    string
	indexes bool
}

func LocalFile(root string, indexes bool, fs fs.FS) *localFileSystem {
	return &localFileSystem{
		FileSystem: http.FS(fs),
		fs:         fs,
		root:       root,
		indexes:    indexes,
	}
}

func (l *localFileSystem) Exists(prefix string, filepath string) bool {
	if p := strings.TrimPrefix(filepath, prefix); len(p) < len(filepath) {
		name := path.Join(l.root, p)

		fsFile, err := l.fs.Open(name)
		if err != nil {
			return false
		}
		stat, err := fsFile.Stat()
		if err != nil {
			return false
		}
		if stat.IsDir() {
			if !l.indexes {
				index := path.Join(name, "index.html")
				_, err := l.fs.Open(index)
				if err != nil {
					return false
				}
			}
		}
		return true
	}
	return false
}

// Static returns a middleware handler that serves static files in the given directory.
func Serve(urlPrefix string, fs ServeFileSystem) gin.HandlerFunc {
	fileserver := http.FileServer(fs)
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}
	return func(c *gin.Context) {
		if fs.Exists(urlPrefix, c.Request.URL.Path) {
			fileserver.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}
	}
}
