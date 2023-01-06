package orm

import (
	"reflect"
	"strings"
)

type Selector[T any] struct {
	table string
}

func (s *Selector[T]) From(table string) *Selector[T] {
	s.table = table
	return s
}

func (s *Selector[T]) Build() (*Query, error) {
	var sb strings.Builder
	sb.WriteString("SELECT * FROM ")
	if s.table == "" {
		sb.WriteByte('`')
		var t T
		sb.WriteString(reflect.TypeOf(t).Name())
		sb.WriteByte('`')
	} else {
		sb.WriteString(s.table)
	}
	sb.WriteString(";")
	return &Query{
		SQL:  sb.String(),
		Args: nil,
	}, nil
}

func NewSelector[T any]() *Selector[T] {
	return &Selector[T]{}
}
