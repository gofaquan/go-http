package valuer

import (
	"database/sql"
	"github.com/gofaquan/go-http/orm/model"
)

type Value interface {
	// SetColumns 设置新值
	SetColumns(rows *sql.Rows) error
}

type Creator func(val interface{}, meta *model.Model) Value
