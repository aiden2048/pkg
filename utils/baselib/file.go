package baselib

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetFileLastModifyTimeStr returs 获取文件最后修改时间
func GetFileLastModifyTimeStr(file string) string {
	if fi, err := os.Stat(file); err != nil {
		return ""
	} else {
		return fi.ModTime().String()
	}
}

// Dir returns dir
func Dir(file string) string {
	file, _ = exec.LookPath(file)
	// 绝对路径
	path, _ := filepath.Abs(file)
	rst := filepath.Dir(path)
	return rst
}

// CheckFileExist return if file exists
func CheckFileExist(file string) bool {
	var exist = true
	if _, err := os.Stat(file); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

// SimpleWriteFile do write text to file
// append: if false, will re-create file, clear old contents
func SimpleWriteFile(file string, text string, append bool) error {
	var err error
	var f *os.File
	var flag int

	flag = os.O_CREATE | os.O_RDWR
	if append {
		flag |= os.O_APPEND
	} else {
		if CheckFileExist(file) {
			errRem := os.Remove(file)
			if errRem != nil {
				return errRem
			}
		}
	}

	f, err = os.OpenFile(file, flag, 0666) //打开文件
	if err != nil {
		return err
	}
	defer f.Close()
	_ /*n*/, err = f.WriteString(text)
	return err
}

// SimpleCopyFile is
func SimpleCopyFile(dstName string, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

// SimpleRemoveFile is
func SimpleRemoveFile(srcName string) error {
	return os.Remove(srcName)
}

// ListFilesBySuffix
func ListFilesBySuffix(dirName string, suffix string) ([]string, error) {
	files := make([]string, 0, 10)
	dir, errDir := ioutil.ReadDir(dirName)
	if errDir != nil {
		return files, errDir
	}
	for _, fi := range dir {
		if fi.IsDir() {
			continue
		} else {
			if strings.HasSuffix(fi.Name(), suffix) {
				files = append(files, fi.Name())
			}
		}
	}
	return files, nil
}
