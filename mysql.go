package mouselib

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type TableMetaData struct {
	TableName  string
	Comment    string
	EngineInfo string
}

type ColumnInfo struct {
	ColName    string
	ColType    string
	IsPrimary  bool
	IsUnique   bool
	IsAuto     bool
	IsNull     bool
	DefaultVal string
	Comment    string
}

type Create struct {
	TableMeta   TableMetaData
	TableCols   []ColumnInfo
	MaxColIdx   int
	MaxIndexIdx int
}

type Base struct {
	CreateTime time.Time `col:"create_time"`
	UpdateTime time.Time `col:"update_time"`
}

type Student struct {
	Id    int64     `col:"id" primary:"true" auto:"true"`
	Name  string    `col:"name" unique:"true"`
	Age   int       `col:"age"`
	Birth time.Time `col:"birth"`

	Base
}

const (
	Col     = "col"
	Primary = "primary"
	Unique  = "unique"
	Auto    = "auto"
	Null    = "null"
	Default = "default"
	Comment = "comment"
	StrLen  = "strlen"
	LongStr = "longstr"

	DEFAULT_STR_LEN = 200

	INNODB = "innodb"
)

type MouseMysqlErr struct {
	msg string
	e   error
}

func (e MouseMysqlErr) Error() string {
	return fmt.Sprintf("[mouse] -> mysql %s: %v", e.msg, e.e)
}

func NewMysqlErr(msg string, e error) MouseMysqlErr {
	return MouseMysqlErr{
		msg: msg,
		e:   e,
	}
}

// RegisterType 创建并执行建表语句，表不存在时创建，存在则更新
func RegisterTable(tb any, comment string) {
	tmpl, err := template.ParseFiles("./templates/create_table.template")
	if err != nil {
		logger.Error().AnErr("parse err", err).Msg("parse failed")
	}

	meta, err := ExtractTableInfo(tb, comment)
	if err != nil {
		logger.Error().AnErr("table extract err", err).Msg("table extract failed")
	}

	colsInfo, err := ExtractColFromTableType(tb)
	if err != nil {
		logger.Error().AnErr("col extract err", err).Msg("col extract failed")
	}

	c := Create{
		TableMeta:   *meta,
		TableCols:   colsInfo,
		MaxColIdx:   len(colsInfo) - 1,
		MaxIndexIdx: 0,
	}

	outLoc := "./templates/dup"
	f, err := os.Create(outLoc)
	if err != nil {
		logger.Error().AnErr("new file err", err).Msg("create file failed")
	}

	err = tmpl.Execute(f, c)
	if err != nil {
		logger.Error().AnErr("exec err", err).Msg("exec failed")
	}

}

func ExtractTableInfo(tb any, comment string) (*TableMetaData, error) {
	if t := reflect.TypeOf(tb); t != nil {
		return &TableMetaData{
			TableName:  CamelToUnderline(t.Name()),
			Comment:    comment,
			EngineInfo: INNODB,
		}, nil
	}
	return nil, NewMysqlErr("无效表对象实例", nil)
}

func ExtractColFromTableType(tb any) ([]ColumnInfo, error) {
	if t := reflect.TypeOf(tb); t != nil {
		return GroupColInfo(t)
	}

	return nil, NewMysqlErr("无效表对象实例", nil)
}

func GroupColInfo(t reflect.Type) ([]ColumnInfo, error) {
	r := make([]ColumnInfo, 0)

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag

		var (
			longStr bool
			strLen  int
		)
		if f.Type.Kind() == reflect.String {
			if _, ok := tag.Lookup(LongStr); ok {
				longStr = true
			}

			if v, ok := tag.Lookup(StrLen); ok {
				l, _ := strconv.ParseInt(v, 10, 32)
				strLen = int(l)
			} else {
				strLen = DEFAULT_STR_LEN
			}
		}
		colType := GoTypeToMysqlColType(f.Type, longStr, strLen)
		if colType == "" && f.Type.Kind() != reflect.Struct {
			return nil, NewMysqlErr("列 "+f.Name+" 不支持的类型", nil)
		}

		var primary, unique, auto, null bool
		if b, ok := tag.Lookup(Primary); ok {
			v, err := strconv.ParseBool(b)
			if err != nil {
				return nil, NewMysqlErr("列 "+f.Name+" 主键值类型错误，需要提供bool值", err)
			}
			primary = v
		}
		if b, ok := tag.Lookup(Unique); ok {
			v, err := strconv.ParseBool(b)
			if err != nil {
				return nil, NewMysqlErr("列 "+f.Name+" 唯一值类型错误，需要提供bool值", err)
			}
			unique = v
		}
		if b, ok := tag.Lookup(Auto); ok {
			v, err := strconv.ParseBool(b)
			if err != nil {
				return nil, NewMysqlErr("列 "+f.Name+" 自增值类型错误，需要提供bool值", err)
			}
			auto = v
		}
		if b, ok := tag.Lookup(Null); ok {
			v, err := strconv.ParseBool(b)
			if err != nil {
				return nil, NewMysqlErr("列 "+f.Name+" 是否为空值类型错误，需要提供bool值", err)
			}
			null = v
		}

		if colType == "" && f.Type.Kind() == reflect.Struct {
			sub, err := GroupColInfo(f.Type)
			if err != nil {
				return nil, err
			}
			r = append(r, sub...)

			continue
		}

		colName := tag.Get(Col)
		if colName == "" {
			return nil, NewMysqlErr("列 "+f.Name+" 需要指定列名", nil)
		}

		r = append(r, ColumnInfo{
			ColName:   colName,
			ColType:   colType,
			IsPrimary: primary,
			IsUnique:  unique,
			IsAuto:    auto,
			IsNull:    null,
		})
	}

	return r, nil
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
