package orm

import (
	"github.com/gofaquan/go-http/orm/internal/errs"
	"reflect"
	"sync"
	"unicode"
)

type registry struct {
	models sync.Map
}

func (r *registry) get(val any) (*model, error) {
	t := reflect.TypeOf(val)
	m, ok := r.models.Load(t)
	if !ok {
		var err error
		if m, err = r.parseModel(t); err != nil {
			return nil, err
		}
	}
	r.models.Store(t, m)
	return m.(*model), nil
}

func (r *registry) parseModel(typ reflect.Type) (*model, error) {
	if typ.Kind() != reflect.Ptr ||
		typ.Elem().Kind() != reflect.Struct {
		return nil, errs.ErrPointerOnly
	}
	typ = typ.Elem()

	// 获得字段的数量
	numField := typ.NumField()
	fds := make(map[string]*field, numField)
	for i := 0; i < numField; i++ {
		fdType := typ.Field(i)
		fds[fdType.Name] = &field{
			colName: underscoreName(fdType.Name),
		}
	}
	return &model{
		tableName: underscoreName(typ.Name()),
		fieldMap:  fds,
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
