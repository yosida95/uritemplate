// Copyright (C) 2016 Kohei YOSHIDA. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of The BSD 3-Clause License
// that can be found in the LICENSE file.

package uritemplate

import (
	"unicode"
	"unicode/utf8"
)

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

const (
	parseStateDefault = iota
	parseStateHexDigit1
	parseStateHexDigit2
)

var (
	rangeVarchar = &unicode.RangeTable{
		R16: []unicode.Range16{
			{Lo: 0x0030, Hi: 0x0039, Stride: 1}, // '0' - '9'
			{Lo: 0x0041, Hi: 0x005A, Stride: 1}, // 'A' - 'Z'
			{Lo: 0x005F, Hi: 0x005F, Stride: 1}, // '_'
			{Lo: 0x0061, Hi: 0x007A, Stride: 1}, // 'a' - 'z'
		},
		LatinOffset: 4,
	}
	rangeLiterals = &unicode.RangeTable{
		R16: []unicode.Range16{
			{Lo: 0x0021, Hi: 0x0021, Stride: 1}, // '!'
			{Lo: 0x0023, Hi: 0x0024, Stride: 1}, // '#' - '$'
			{Lo: 0x0026, Hi: 0x0026, Stride: 1}, // '&'
			{Lo: 0x0028, Hi: 0x003B, Stride: 1}, // '(' - ';'
			{Lo: 0x003D, Hi: 0x003D, Stride: 1}, // '='
			{Lo: 0x003F, Hi: 0x005B, Stride: 1}, // '?' - '['
			{Lo: 0x005D, Hi: 0x005D, Stride: 1}, // ']'
			{Lo: 0x005F, Hi: 0x005F, Stride: 1}, // '_'
			{Lo: 0x0061, Hi: 0x007A, Stride: 1}, // 'a' - 'z'
			{Lo: 0x007E, Hi: 0x007E, Stride: 1}, // '~'
			{Lo: 0x00A0, Hi: 0xD7FF, Stride: 1}, // ucschar
			{Lo: 0xE000, Hi: 0xF8FF, Stride: 1}, // iprivate
			{Lo: 0xF900, Hi: 0xFDCF, Stride: 1}, // ucschar
			{Lo: 0xFDF0, Hi: 0xFFEF, Stride: 1}, // ucschar
		},
		R32: []unicode.Range32{
			{Lo: 0x00010000, Hi: 0x0001FFFD, Stride: 1}, // ucschar
			{Lo: 0x00020000, Hi: 0x0002FFFD, Stride: 1}, // ucschar
			{Lo: 0x00030000, Hi: 0x0003FFFD, Stride: 1}, // ucschar
			{Lo: 0x00040000, Hi: 0x0004FFFD, Stride: 1}, // ucschar
			{Lo: 0x00050000, Hi: 0x0005FFFD, Stride: 1}, // ucschar
			{Lo: 0x00060000, Hi: 0x0006FFFD, Stride: 1}, // ucschar
			{Lo: 0x00070000, Hi: 0x0007FFFD, Stride: 1}, // ucschar
			{Lo: 0x00080000, Hi: 0x0008FFFD, Stride: 1}, // ucschar
			{Lo: 0x00090000, Hi: 0x0009FFFD, Stride: 1}, // ucschar
			{Lo: 0x000A0000, Hi: 0x000AFFFD, Stride: 1}, // ucschar
			{Lo: 0x000B0000, Hi: 0x000BFFFD, Stride: 1}, // ucschar
			{Lo: 0x000C0000, Hi: 0x000CFFFD, Stride: 1}, // ucschar
			{Lo: 0x000D0000, Hi: 0x000DFFFD, Stride: 1}, // ucschar
			{Lo: 0x000E1000, Hi: 0x000EFFFD, Stride: 1}, // ucschar
			{Lo: 0x000F0000, Hi: 0x000FFFFD, Stride: 1}, // iprivate
			{Lo: 0x00100000, Hi: 0x0010FFFD, Stride: 1}, // iprivate
		},
		LatinOffset: 10,
	}
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
	var state int
	var ret varspec
	var err error
	for i := 0; i < len(p.r); {
		r, size := utf8.DecodeRuneInString(p.r[i:])
		if r == utf8.RuneError {
			return ret, errorf(p.read+i, "invalid encoding")
		}
		switch state {
		case parseStateDefault:
			switch r {
			default:
				if !unicode.Is(rangeVarchar, r) {
					debug.Printf("consumeVarspec: found=%q at=%d", string(r), i)
					return ret, errorf(p.read+i, "invalid varname")
				}
				i += size
				continue
			case '%':
				state = parseStateHexDigit1
			case ':':
				ret.name = p.r[:i]
				debug.Printf("consumeVarspec: name=%q", ret.name)
				p.dropN(i + 1) // name + ':'
				ret.maxlen, err = p.consumeMaxLength()
				if err != nil {
					return ret, err
				}
				debug.Printf("consumeVarspec: maxlen=%d", ret.maxlen)
				return ret, nil
			case '*':
				ret.name = p.r[:i]
				debug.Printf("consumeVarspec: name=%q", ret.name)
				ret.explode = true
				debug.Printf("consumeVarspec: explode=true")
				p.dropN(i + 1) // name + '*'
				return ret, nil
			case ',', '}':
				ret.name = p.r[:i]
				debug.Printf("consumeVarspec: name=%q", ret.name)
				p.dropN(i)
				return ret, nil
			}
		case parseStateHexDigit1:
			if !unicode.Is(unicode.ASCII_Hex_Digit, r) {
				debug.Printf("consumeVarspec: found=%q at=%d", string(r), i)
				return ret, errorf(p.read+i, "invalid pct-encoded")
			}
			state = parseStateHexDigit2
		case parseStateHexDigit2:
			if !unicode.Is(unicode.ASCII_Hex_Digit, r) {
				debug.Printf("consumeVarspec: found=%q at=%d", string(r), i)
				return ret, errorf(p.read+i, "invalid pct-encoded")
			}
			state = parseStateDefault
		}
	}
	debug.Printf("consumeVarspec: unexpected end of URI Template")
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

func (p *parser) consumeExpression() (template, error) {
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

	ret := expression{
		vars: varspecs,
		op:   op,
	}
	ret.init()
	return &ret, nil
}

func (p *parser) consumeLiterals() (template, error) {
	debug.Printf("consumeLiterals: %q", p.r)
	state := parseStateDefault
	i := 0
Loop:
	for i < len(p.r) {
		r, size := utf8.DecodeRuneInString(p.r[i:])
		if r == utf8.RuneError {
			return nil, errorf(p.read+i, "invalid encoding")
		}
		switch state {
		case parseStateDefault:
			switch r {
			case '{':
				break Loop
			case '%':
				state = parseStateHexDigit1
			default:
				if !unicode.Is(rangeLiterals, r) {
					debug.Printf("consumeLiterals: found=%q at=%d", string(r), i)
					return nil, errorf(p.read+i, "invalid literals")
				}
			}
		case parseStateHexDigit1:
			if !unicode.Is(unicode.ASCII_Hex_Digit, r) {
				debug.Printf("consumeLiterals: found=%q at=%d", string(r), i)
				return nil, errorf(p.read+i, "invalid pct-encoded")
			}
			state = parseStateHexDigit2
		case parseStateHexDigit2:
			if !unicode.Is(unicode.ASCII_Hex_Digit, r) {
				debug.Printf("consumeLiterals: found=%q at=%d", string(r), i)
				return nil, errorf(p.read+i, "invalid pct-encoded")
			}
			state = parseStateDefault
		}
		i += size
	}
	if state != parseStateDefault {
		return nil, errorf(p.read+i, "invalid pct-encoded")
	}
	exp := literals(p.r[:i])
	p.dropN(i)
	return exp, nil
}

func (p *parser) parseURITemplate() (*Template, error) {
	debug.Printf("parseURITemplate: %q", p.r)
	tmpl := Template{
		raw:   p.r,
		exprs: []template{},
	}
	for {
		if len(p.r) == 0 {
			break
		}

		var expr template
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
