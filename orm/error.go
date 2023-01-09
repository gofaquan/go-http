package orm

import (
	"github.com/gofaquan/go-http/orm/internal/errs"
)

// 将内部的 sentinel error 暴露出去
var (
	// ErrNoRows 代表没有找到数据
	ErrNoRows = errs.ErrNoRows
)
