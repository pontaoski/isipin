package main

type Expression interface {
	is_Expression()
}
type Variable string

func (v Variable) is_Expression() {}

type Query string

func (v Query) is_Expression() {}

type Literal string

func (v Literal) is_Expression() {}

type Statement interface {
	is_Statement()
}
type SetOption struct {
	Key   string
	Value string
}

func (v SetOption) is_Statement() {}

type SetComponent struct {
	Component string
	Query     Expression
}

func (v SetComponent) is_Statement() {}

type Call struct {
	Name string
	Args []Expression
	On   Expression
}

func (v Call) is_Statement() {}

type SetVariable struct {
	Name  string
	Value Expression
}

func (v SetVariable) is_Statement() {}
