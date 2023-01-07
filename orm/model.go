package orm

type field struct {
	colName string
}

type model struct {
	tableName string
	fieldMap  map[string]*field
}
