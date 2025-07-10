package baselib

import (
	"errors"
	"github.com/BurntSushi/toml"
	"os"
	"runtime"
	"strings"
)

var (
	WindowsSearchPath = [3]string{".\\", "..\\LocalConfig\\", "c:\\goservice\\"}
	UnixSearchPath    = [3]string{"./", "../LocalConfig/", "/etc/goservice/"}
)

var (
	FileNotExist = errors.New("file not exist")
)

func FileLocation(fname string) string {
	if isExist(fname) {
		return fname
	}

	switch runtime.GOOS {
	case "linux", "darwin":
		for _, path := range UnixSearchPath {
			fullPath := path + convert2UnixPath(fname)
			if isExist(fullPath) {
				return fullPath
			}
		}
	case "windows":
		for _, path := range WindowsSearchPath {
			fullPath := path + convert2WindowsPath(fname)
			if isExist(fullPath) {
				return fullPath
			}
		}
	}
	return ""
}

func isExist(file string) bool {
	if _, err := os.Stat(file); err != nil && !os.IsExist(err) {
		return false
	}
	return true
}

func convert2WindowsPath(rPath string) string {
	return strings.Replace(rPath, "/", "\\", 0)
}

func convert2UnixPath(rPath string) string {
	return strings.Replace(rPath, "\\", "/", 0)
}

func DecodeToml(fname string, v interface{}) error {
	fpath := FileLocation(fname)
	if len(fpath) == 0 {
		return FileNotExist
	}
	_, err := toml.DecodeFile(fpath, v)
	return err
}
