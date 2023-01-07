package orm

import (
	"github.com/gofaquan/go-http/orm/internal/errs"
	"reflect"
	"strings"
	"sync"
	"unicode"
)

var models = map[reflect.Type]*model{}

type registry struct {
	models sync.Map
}

func (r *registry) get(val any) (*model, error) {
	t := reflect.TypeOf(val)
	m, ok := r.models.Load(t)
	if !ok {
		var err error
		if m, err = r.parseModel(val); err != nil {
			return nil, err
		}
	}
	r.models.Store(t, m)
	return m.(*model), nil
}

func (r *registry) parseModel(val any) (*model, error) {
	typ := reflect.TypeOf(val)

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
		tags, err := r.parseTag(fdType.Tag)
		if err != nil {
			return nil, err
		}
		// 有 tag 就用 tag
		colName := tags[tagKeyColumn]
		if colName == "" {
			colName = underscoreName(fdType.Name)
		}

		fds[fdType.Name] = &field{
			colName: colName,
		}
	}

	// 获取 表名
	var tableName string

	if tn, ok := val.(TableName); ok {
		tableName = tn.TableName()
	}

	if tableName == "" {
		tableName = underscoreName(typ.Name())
	}

	return &model{
		tableName: tableName,
		fieldMap:  fds,
	}, nil
}

func (r *registry) parseTag(tag reflect.StructTag) (map[string]string, error) {
	ormTag := tag.Get("orm")
	if ormTag == "" {
		// 返回一个空的 map，这样调用者就不需要判断 nil 了
		return map[string]string{}, nil
	}
	// 这个初始化容量就是我们支持的 key 的数量，
	// 现在只有一个，所以我们初始化为 1
	res := make(map[string]string, 1)

	// 接下来就是字符串处理了
	pairs := strings.Split(ormTag, ",")
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) != 2 {
			return nil, errs.NewErrInvalidTagContent(pair)
		}
		res[kv[0]] = kv[1]
	}
	return res, nil
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
