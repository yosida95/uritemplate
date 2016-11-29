// Copyright (C) 2016 Kohei YOSHIDA. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of The BSD 3-Clause License
// that can be found in the LICENSE file.

package uritemplate

const (
	parseOpSimple = iota
	parseOpPlus
	parseOpCrosshatch
	parseOpDot
	parseOpSlash
	parseOpSemicolon
	parseOpQuestion
	parseOpAmpersand
)

type parser struct {
	r    string
	read int
}

func (p *parser) dropN(n int) {
	p.read += n
	p.r = p.r[n:]
}

func (p *parser) consumeOp() (int, error) {
	debug.Printf("consumeOp: %q", p.r)
	switch p.r[0] {
	case '+':
		p.dropN(1)
		return parseOpPlus, nil
	case '#':
		p.dropN(1)
		return parseOpCrosshatch, nil
	case '.':
		p.dropN(1)
		return parseOpDot, nil
	case '/':
		p.dropN(1)
		return parseOpSlash, nil
	case ';':
		p.dropN(1)
		return parseOpSemicolon, nil
	case '?':
		p.dropN(1)
		return parseOpQuestion, nil
	case '&':
		p.dropN(1)
		return parseOpAmpersand, nil
	case '=', ',', '!', '@', '|': // op-reserved
		return 0, errorf(p.read+1, "unsupported operator")
	default:
		return parseOpSimple, nil
	}
}

func (p *parser) consumeMaxLength() (int, error) {
	debug.Printf("consumeMaxLength: %q", p.r)
	c := p.r[0]
	if c < '1' || c > '9' {
		return 0, errorf(p.read+1, "max-length must be integer")
	}
	p.dropN(1)
	maxlen := int(c - '0')

	for {
		c := p.r[0]
		if c < '0' || c > '9' {
			break
		}
		p.dropN(1)

		maxlen *= 10
		maxlen += int(c - '0')
		if maxlen >= 1000 || len(p.r) == 0 {
			break
		}
	}
	return maxlen, nil
}

func (p *parser) consumeVarspec() (varspec, error) {
	debug.Printf("consumeVarspec: %q", p.r)
	var ret varspec
	for i := 0; i < len(p.r); {
		switch p.r[i] {
		default:
			i++
			continue
		case ':':
			ret := varspec{
				name: p.r[:i],
			}
			p.dropN(i + 1)
			debug.Printf("consumeVarspec: name=%q", ret.name)

			maxlen, err := p.consumeMaxLength()
			if err != nil {
				return ret, err
			}
			ret.maxlen = maxlen
			debug.Printf("consumeVarspec: maxlen=%d", ret.maxlen)
			return ret, nil
		case '*':
			ret := varspec{
				name:    p.r[:i],
				explode: true,
			}
			p.dropN(i + 1)
			debug.Printf("consumeVarspec: name=%q", ret.name)
			debug.Printf("consumeVarspec: explode=true")
			return ret, nil
		case ',', '}':
			ret := varspec{
				name: p.r[:i],
			}
			p.dropN(i)
			debug.Printf("consumeVarspec: name=%q", ret.name)
			return ret, nil
		}
	}
	debug.Printf("consumeVarspec: remains=%q", p.r)
	return ret, errorf(p.read+1, "incomplete template")
}

func (p *parser) consumeVariableList() ([]varspec, error) {
	debug.Printf("consumeVariableList: %q", p.r)
	varspecs := []varspec{}
	for {
		varspec, err := p.consumeVarspec()
		if err != nil {
			return nil, err
		}
		varspecs = append(varspecs, varspec)

		if len(p.r) == 0 {
			return nil, errorf(p.read+1, "incomplete template")
		}
		switch p.r[0] {
		case ',':
			p.dropN(1)
			continue
		case '}':
			return varspecs, nil
		default:
			return nil, errorf(p.read+1, "unrecognized variable-list")
		}
	}
}

func (p *parser) consumeExpression() (expression, error) {
	debug.Printf("consumeExpression: %q", p.r)

	p.dropN(1) // '{'
	if len(p.r) == 0 {
		return nil, errorf(p.read+1, "incomplete template")
	}

	op, err := p.consumeOp()
	if err != nil {
		return nil, err
	}
	if len(p.r) == 0 {
		return nil, errorf(p.read+1, "incomplete template")
	}

	varspecs, err := p.consumeVariableList()
	if err != nil {
		return nil, err
	}
	p.dropN(1) // '}'

	switch op {
	case parseOpSimple:
		return &expSimple{vars: varspecs}, nil
	case parseOpPlus:
		return &expPlus{vars: varspecs}, nil
	case parseOpCrosshatch:
		return &expCrosshatch{vars: varspecs}, nil
	case parseOpDot:
		return &expDot{vars: varspecs}, nil
	case parseOpSlash:
		return &expSlash{vars: varspecs}, nil
	case parseOpSemicolon:
		return &expSemicolon{vars: varspecs}, nil
	case parseOpQuestion:
		return &expQuestion{vars: varspecs}, nil
	case parseOpAmpersand:
		return &expAmpersand{vars: varspecs}, nil
	default:
		return nil, errorf(p.read, "unsupported operator")
	}
}

func (p *parser) consumeLiterals() (expression, error) {
	debug.Printf("consumeLiterals: %q", p.r)
	i := 0
	for {
		c := p.r[i]
		if c == '{' {
			break
		}
		// TODO(yosida95): is c in literals?
		i++
	}
	exp := expNoop(p.r[:i])
	p.dropN(i)
	return exp, nil
}

func (p *parser) parseURITemplate() (*Template, error) {
	debug.Printf("parseURITemplate: %q", p.r)
	tmpl := Template{
		exprs: []expression{},
	}
	for {
		if len(p.r) == 0 {
			break
		}

		var expr expression
		var err error
		if p.r[0] == '{' {
			expr, err = p.consumeExpression()
		} else {
			expr, err = p.consumeLiterals()
		}
		if err != nil {
			return nil, err
		}
		tmpl.exprs = append(tmpl.exprs, expr)
	}
	return &tmpl, nil
}
