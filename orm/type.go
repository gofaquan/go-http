package orm

import (
	"context"
	"database/sql"
)

type Querier[T any] interface {
	Get(ctx context.Context) (*T, error)
	GetMulti(ctx context.Context) (*[]T, error)
}

type Excecutor interface {
	Exec(ct context.Context) (sql.Result, error)
}

type Query struct {
	SQL  string
	Args []any
}

type QueryBuilder interface {
	Build() (*Query, error)
}
