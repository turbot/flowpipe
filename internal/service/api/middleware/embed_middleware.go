package middleware

import (
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

// This code is based on a very early version of https://github.com/gin-contrib/static and modified to use embed.FS
//
// The later version seems to support embed.FS but I couldn't get it to work properly.

type ServeFileSystem interface {
	http.FileSystem
	Exists(prefix string, path string) bool
}

type localFileSystem struct {
	http.FileSystem
	fs            fs.FS
	root          string
	indexes       bool
	indexFileName string
}

func LocalFile(root string, indexes bool, indexFileName string, fs fs.FS) *localFileSystem {
	return &localFileSystem{
		FileSystem:    http.FS(fs),
		fs:            fs,
		root:          root,
		indexes:       indexes,
		indexFileName: indexFileName,
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

		// Is the path a directory?
		if stat.IsDir() {
			// If the l.indexes option is enabled, we check for an index file. If found return true, otherwise it's a 404
			if !l.indexes {
				index := path.Join(name, l.indexFileName)
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
		path := c.Request.URL.Path
		if fs.Exists(urlPrefix, path) {
			fileserver.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}
	}
}
