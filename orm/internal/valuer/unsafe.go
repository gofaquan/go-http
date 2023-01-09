package valuer

import (
	"database/sql"
	"github.com/gofaquan/go-http/orm/internal/errs"
	"github.com/gofaquan/go-http/orm/model"
	"reflect"
	"unsafe"
)

var _ Creator = NewUnsafeValue

type unsafeValue struct {
	addr unsafe.Pointer
	meta *model.Model
}

func (u unsafeValue) SetColumns(rows *sql.Rows) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	if len(columns) > len(u.meta.ColumnMap) {
		return errs.ErrTooManyReturnedColumns
	}

	colValues := make([]interface{}, len(columns))
	for i, column := range columns {
		cm, ok := u.meta.ColumnMap[column]
		if !ok {
			return errs.NewErrUnknownColumn(column)
		}

		ptr := unsafe.Pointer(uintptr(u.addr) + cm.Offset)
		val := reflect.NewAt(cm.Type, ptr)
		colValues[i] = val.Interface()
	}

	return rows.Scan(colValues)
}

func NewUnsafeValue(val any, meta *model.Model) Value {
	return unsafeValue{
		addr: unsafe.Pointer(reflect.ValueOf(val).Pointer()),
		meta: meta,
	}
}
