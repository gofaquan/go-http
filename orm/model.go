package orm

import (
	"github.com/gofaquan/go-http/orm/internal/errs"
	"reflect"
	"unicode"
)

type field struct {
	colName string
}

type model struct {
	tableName string
	fieldMap  map[string]*field
}

func parseModel(val any) (*model, error) {
	t := reflect.TypeOf(val)

	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}

	t = t.Elem()

	numField := t.NumField()
	fields := make(map[string]*field, numField)
	for i := 0; i < numField; i++ {
		fieldType := t.Field(i)
		fields[fieldType.Name] = &field{colName: underscoreName(fieldType.Name)}
	}
	return &model{
		tableName: underscoreName(t.Name()),
		fieldMap:  fields,
	}, nil

}

// underscoreName 驼峰转字符串命名
func underscoreName(tableName string) string {
	var buf []byte
	for i, v := range tableName {
		if unicode.IsUpper(v) {
			if i != 0 {
				buf = append(buf, '_')
			}
			buf = append(buf, byte(unicode.ToLower(v)))
		} else {
			buf = append(buf, byte(v))
		}

	}
	return string(buf)
}
