package orm

type Column struct {
	name string
}

func (c Column) expr() {}

type value struct {
	val any
}

func (v value) expr() {}

func valueOf(val any) value {
	return value{val}
}

func C(name string) Column {
	return Column{name}
}

func (c Column) EQ(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    EQ,
		right: exprOf(arg),
	}
}

func (c Column) LT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    LT,
		right: exprOf(arg),
	}
}

func (c Column) GT(arg any) Predicate {
	return Predicate{
		left:  c,
		op:    GT,
		right: exprOf(arg),
	}
}
