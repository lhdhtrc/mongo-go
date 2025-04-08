package internal

import (
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var pkgSourceDir string

func init() {
	_, file, _, _ := runtime.Caller(0)
	pkgSourceDir = sourceDir(file)
}

func sourceDir(file string) string {
	dir := filepath.Dir(file)
	dir = filepath.Dir(dir) // 获取项目根目录（github.com/lhdhtrc/mongo-go）
	return filepath.ToSlash(dir) + "/"
}

// FileWithLineNum 返回调用方的文件路径和行号（跳过当前包的内部调用）
func FileWithLineNum() string {
	// 从第3层调用栈开始检查（跳过runtime和当前包的调用）
	for i := 2; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && shouldInclude(file) {
			return file + ":" + strconv.Itoa(line)
		}
	}
	return ""
}

func shouldInclude(file string) bool {
	// 包含：非项目目录文件 或 测试文件，且排除生成的代码
	return (!strings.HasPrefix(file, pkgSourceDir) ||
		strings.HasSuffix(file, "_test.go")) &&
		!strings.HasSuffix(file, ".gen.go")
}
