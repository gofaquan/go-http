package orm

const (
	EQ  = "="
	LT  = "<"
	GT  = ">"
	AND = "AND"
	OR  = "OR"
	NOT = "NOT"
)

type operator string

func (o operator) String() string {
	return string(o)
}

// Expression 标记是语句用的接口
type Expression interface {
	expr()
}

func exprOf(e any) Expression {
	switch expr := e.(type) {
	case Expression:
		return expr

	default:
		return valueOf(expr)
	}
}

type Predicate struct {
	left  Expression
	op    operator
	right Expression
}

// 用来实现 Express 接口
func (p Predicate) expr() {}

func (p Predicate) And(r Predicate) Predicate {
	return Predicate{
		left:  p,
		op:    AND,
		right: r,
	}
}

func (p Predicate) OR(r Predicate) Predicate {
	return Predicate{
		left:  p,
		op:    OR,
		right: r,
	}
}
func Not(p Predicate) Predicate {
	return Predicate{
		op:    NOT,
		right: p,
	}
}
