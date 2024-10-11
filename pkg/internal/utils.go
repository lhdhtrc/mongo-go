package internal

import (
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var mongoSourceDir string

func init() {
	_, file, _, _ := runtime.Caller(0)
	// compatible solution to get gorm source directory with various operating systems
	mongoSourceDir = sourceDir(file)
}

func sourceDir(file string) string {
	dir := filepath.Dir(file)
	dir = filepath.Dir(dir)

	s := filepath.Dir(dir)
	if filepath.Base(s) != "mongo-go" {
		s = dir
	}
	return filepath.ToSlash(s) + "/"
}

func FileWithLineNum() string {
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && (!strings.HasPrefix(file, mongoSourceDir) || strings.HasSuffix(file, "_test.go")) &&
			!strings.HasSuffix(file, ".gen.go") {
			return file + ":" + strconv.FormatInt(int64(line), 10)
		}
	}

	return ""
}
