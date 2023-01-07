package errs

import (
	"errors"
	"fmt"
)

var (
	ErrPointerOnly = errors.New("orm: 只支持一级指针作为输入，例如 *User")
)

// NewErrUnknownField 返回代表未知字段的错误
// 一般意味着你可能输入的是列名，或者输入了错误的字段名
func NewErrUnknownField(fd string) error {
	return fmt.Errorf("orm: 未知字段 %s", fd)
}

// NewErrUnsupportedExpressionType 返回一个不支持该 expression 错误信息
func NewErrUnsupportedExpressionType(exp any) error {
	return fmt.Errorf("orm: 不支持的表达式 %v", exp)
}
