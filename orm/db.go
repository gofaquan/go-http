package orm

type DB struct {
	r *registry
}

type DBOption func(db *DB)

// NewDB 多返回 error 留作后续拓展
func NewDB(opts ...DBOption) (*DB, error) {
	db := &DB{r: &registry{}}
	for _, opt := range opts {
		opt(db)
	}

	return db, nil
}
