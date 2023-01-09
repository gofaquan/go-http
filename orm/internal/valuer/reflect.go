package valuer

import (
	"database/sql"
	"github.com/gofaquan/go-http/orm/internal/errs"
	"github.com/gofaquan/go-http/orm/model"
	"reflect"
)

var _ Creator = NewReflectValue

type reflectValue struct {
	val  reflect.Value
	meta *model.Model
}

func NewReflectValue(val any, meta *model.Model) Value {
	return reflectValue{
		val:  reflect.ValueOf(val).Elem(),
		meta: meta,
	}
}

func (r reflectValue) SetColumns(rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	colValues := make([]interface{}, len(columns))
	colElemVal := make([]reflect.Value, len(columns))

	for i, c := range columns {
		cm, ok := r.meta.ColumnMap[c]
		if !ok {
			return errs.NewErrUnknownColumn(c)
		}
		val := reflect.New(cm.Type)
		colValues[i] = val.Interface()
		colElemVal[i] = val.Elem()
	}

	if err = rows.Scan(colValues); err != nil {
		return err
	}

	for i, c := range columns {
		cm := r.meta.ColumnMap[c]
		fd := r.val.FieldByName(cm.GoName)
		fd.Set(colElemVal[i])
	}

	return nil
}
