package mouselib

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
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

// GoTypeToMysqlColType 返回mysql中对应golang类型的类型，如果不存在则返回空字符串
func GoTypeToMysqlColType(t reflect.Type, longText bool, strLen int) string {
	// println(t.Kind().String())
	// println(longText)
	// println(strLen)
	switch t.Kind() {
	case reflect.Bool:
		return "BOOLEAN"
	case reflect.Int:
		return "INT"
	case reflect.Int64:
		return "BIGINT"
	case reflect.Float32, reflect.Float64:
		return "FLOAT"
	case reflect.String:
		if longText {
			return "TEXT"
		} else {
			return "VARCHAR(" + strconv.FormatInt(int64(strLen), 10) + ")"
		}
	case reflect.Struct:
		if t.AssignableTo(reflect.TypeOf(time.Time{})) {
			return "DATETIME"
		} else {
			return ""
		}
	default:
		return ""
	}
}

// CamelToUnderline 驼峰命名转换为下划线命名
func CamelToUnderline(title string) string {
	reg := "[A-Z]"
	r, _ := regexp.Compile(reg)

	if indices := r.FindAllIndex([]byte(title), -1); indices != nil {
		res := strings.Builder{}
		l := len(indices)
		if l == 1 {
			res.WriteString(strings.ToLower(string(title)))
			return res.String()
		}

		for i, idx := range indices {
			if i == l-1 {
				res.WriteString(strings.ToLower(title[idx[0]:]))
			} else {
				res.WriteString(strings.ToLower(title[idx[0]:indices[i+1][0]]) + "_")
			}
		}
		return res.String()
	}

	return title
}
