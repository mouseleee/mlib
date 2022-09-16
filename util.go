package mouselib

import (
	"io"
	"os"
	"path/filepath"
)

// WriteFile 在指定路径写入数据
//
// path 文件路径 data 数据内容，[]byte类型
func WriteFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); err != nil {
		err = os.MkdirAll(dir, os.ModeDir|0770)
		if err != nil {
			return err
		}
	}

	var f *os.File
	if _, err := os.Stat(path); err != nil {
		f, err = os.Create(path)
		if err != nil {
			return err
		}
	} else {
		f, err = os.OpenFile(path, os.O_WRONLY, 0770)
		if err != nil {
			return err
		}
	}

	if _, err := io.WriteString(f, string(data)); err != nil {
		return err
	}

	return nil
}
