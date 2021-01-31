package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"unicode"

	"github.com/alecthomas/repr"
)

// Parser is responsible for parsing
type Parser struct {
	input *bufio.Reader
	line  int
	col   int
}

type expectedGot struct {
	expected string
	got      string
}

func (p *Parser) readRune() rune {
	r, _, err := p.input.ReadRune()
	if err != nil {
		panic(err)
	}
	p.col++
	if r == '\n' {
		p.col = 0
		p.line++
	}
	return r
}

func (p *Parser) unreadRune() {
	err := p.input.UnreadRune()
	if err != nil {
		panic(err)
	}
	p.col--
}

func (p *Parser) readTo(to rune) string {
	dat := strings.Builder{}

outer:
	for {
		r := p.readRune()
		switch r {
		case to:
			break outer
		case '\n':
			panic(expectedGot{expected: string(to), got: "newline"})
		default:
			dat.WriteRune(r)
		}
	}

	return dat.String()
}

func (p *Parser) readToColon() string {
	dat := p.readTo(':')

	r := p.readRune()
	for unicode.IsSpace(r) {
		r = p.readRune()
	}
	p.unreadRune()

	return dat
}

func (p *Parser) readToNewline() string {
	val := strings.Builder{}

	r := p.readRune()
	for r != '\n' {
		val.WriteRune(r)

		r = p.readRune()
	}

	return val.String()
}

func (p *Parser) skipToNewline() {
	for {
		r := p.readRune()
		switch {
		case r == '\n':
			return
		case unicode.IsSpace(r):
			continue
		default:
			panic(expectedGot{expected: "newline or whitespace", got: string(r)})
		}
	}
}

func (p *Parser) readToWhitespace() string {
	val := strings.Builder{}

	r := p.readRune()
	for !unicode.IsSpace(r) {
		val.WriteRune(r)

		r = p.readRune()
	}
	p.unreadRune()

	return val.String()
}

func (p *Parser) readOption() SetOption {
	return SetOption{
		Key:   p.readToColon(),
		Value: p.readToNewline(),
	}
}

func (p *Parser) readVariableOrString() Expression {
	r := p.readRune()
	switch r {
	case '@':
		defer p.readToWhitespace()
		return Variable(p.readToWhitespace())
	case '"':
		defer p.readToWhitespace()
		return Literal(p.readTo('"'))
	}

	panic("bad")
}

func (p *Parser) readExpression() Expression {
	r := p.readRune()
	switch r {
	case '@':
		defer p.readToWhitespace()
		return Variable(p.readToWhitespace())
	case '"':
		defer p.readToWhitespace()
		return Literal(p.readTo('"'))
	default:
		p.unreadRune()
		return Query(p.readToNewline())
	}
}

func (p *Parser) readSetVariable() SetVariable {
	return SetVariable{
		Name:  p.readToColon(),
		Value: p.readExpression(),
	}
}

func (p *Parser) readSetComponent() SetComponent {
	return SetComponent{
		Component: p.readToColon(),
		Query:     p.readExpression(),
	}
}

func (p *Parser) eatWhitespace() {
	r := p.readRune()
	for unicode.IsSpace(r) {
		r = p.readRune()
	}
	p.unreadRune()
}

func (p *Parser) readCall() Call {
	return Call{
		Name: p.readTo('('),
		Args: func() (e []Expression) {
			defer func() {
				p.readToColon()
				p.eatWhitespace()
			}()

			p.eatWhitespace()
			r := p.readRune()
			if r == ')' {
				return
			}
			p.unreadRune()

			mu := p.readTo(')')
			s := strings.Split(mu, ",")
			for _, str := range s {
				st := strings.TrimSpace(str)
				if !strings.HasPrefix(st, `"`) || !strings.HasPrefix(st, `"`) {
					panic("bad string")
				}
				e = append(e, Literal(strings.TrimSuffix(strings.TrimPrefix(st, `"`), `"`)))
			}

			return
		}(),
		On: p.readExpression(),
	}
}

func (p *Parser) readStatement() Statement {
	for {
		r := p.readRune()

		switch r {
		case '~':
			return p.readOption()
		case '$':
			return p.readSetVariable()
		case '@':
			return p.readCall()
		default:
			if unicode.IsSpace(r) {
				continue
			}

			p.unreadRune()
			return p.readSetComponent()
		}
	}
}

// Parse parses a document
func (p *Parser) Parse(r io.Reader) (d Document) {
	defer func() {
		if r := recover(); r != nil {
			if v, ok := r.(error); ok && errors.Is(v, io.EOF) {
				return
			}

			repr.Println(d)
			fmt.Printf("%d:%d\n", p.line, p.col)
			log.Fatal(r)
		}
	}()

	p.input = bufio.NewReader(r)
	p.line = 1
	p.col = 0

	for {
		d = append(d, p.readStatement())
	}
}
