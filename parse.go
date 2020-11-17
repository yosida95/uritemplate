// Copyright (C) 2016 Kohei YOSHIDA. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of The BSD 3-Clause License
// that can be found in the LICENSE file.

package uritemplate

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type parseOp int

const (
	parseOpSimple parseOp = iota
	parseOpPlus
	parseOpCrosshatch
	parseOpDot
	parseOpSlash
	parseOpSemicolon
	parseOpQuestion
	parseOpAmpersand
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
	r     string
	start int
	stop  int
	state parseState
}

func (p *parser) errorf(format string, a ...interface{}) error {
	return errorf(p.stop, format, a...)
}

func (p *parser) rune() (rune, int) {
	r, size := utf8.DecodeRuneInString(p.r[p.stop:])
	p.stop += size
	return r, size
}

func (p *parser) unread(r rune) {
	p.stop -= utf8.RuneLen(r)
}

type parseState int

const (
	parseStateDefault = parseState(iota)
	parseStateOperator
	parseStateVarList
	parseStateVarName
	parseStatePrefix
)

func (p *parser) setState(state parseState) {
	p.state = state
	p.start = p.stop
}

func (p *parser) parseURITemplate() (*Template, error) {
	tmpl := Template{
		raw:   p.r,
		exprs: []template{},
	}

	var exp *expression
	for {
		r, size := p.rune()
		if r == utf8.RuneError {
			if size == 0 {
				if p.state != parseStateDefault {
					return nil, p.errorf("incomplete template")
				}
				if p.start < p.stop {
					tmpl.exprs = append(tmpl.exprs, literals(p.r[p.start:p.stop]))
				}
				return &tmpl, nil
			}
			return nil, p.errorf("invalid UTF-8 encoding")
		}

		switch p.state {
		case parseStateDefault:
			switch r {
			case '{':
				if stop := p.stop - size; stop > p.start {
					tmpl.exprs = append(tmpl.exprs, literals(p.r[p.start:stop]))
				}
				exp = &expression{}
				tmpl.exprs = append(tmpl.exprs, exp)
				p.setState(parseStateOperator)
			case '%':
				p.unread(r)
				if _, err := p.consumeTriplets(); err != nil {
					return nil, err
				}
			default:
				if !unicode.Is(rangeLiterals, r) {
					return nil, p.errorf("invalid literals")
				}
			}
		case parseStateOperator:
			switch r {
			default:
				p.unread(r)
				exp.op = parseOpSimple
			case '+':
				exp.op = parseOpPlus
			case '#':
				exp.op = parseOpCrosshatch
			case '.':
				exp.op = parseOpDot
			case '/':
				exp.op = parseOpSlash
			case ';':
				exp.op = parseOpSemicolon
			case '?':
				exp.op = parseOpQuestion
			case '&':
				exp.op = parseOpAmpersand
			case '=', ',', '!', '@', '|': // op-reserved
				return nil, p.errorf("unsupported operator")
			}
			p.setState(parseStateVarName)
		case parseStateVarList:
			switch r {
			case ',':
				p.setState(parseStateVarName)
			case '}':
				exp.init()
				p.setState(parseStateDefault)
			default:
				return nil, p.errorf("invalid variable-list")
			}
		case parseStateVarName:
			switch r {
			case ':':
				exp.vars = append(exp.vars, varspec{
					name: p.r[p.start : p.stop-size],
				})
				p.setState(parseStatePrefix)
			case '*':
				exp.vars = append(exp.vars, varspec{
					name:    p.r[p.start : p.stop-size],
					explode: true,
				})
				p.setState(parseStateVarList)
			case ',', '}':
				p.unread(r)
				exp.vars = append(exp.vars, varspec{
					name: p.r[p.start:p.stop],
				})
				p.setState(parseStateVarList)
			case '%':
				p.unread(r)
				if _, err := p.consumeTriplets(); err != nil {
					return nil, err
				}
			default:
				if !unicode.Is(rangeVarchar, r) {
					return nil, p.errorf("invalid varname")
				}
			}
		case parseStatePrefix:
			spec := &(exp.vars[len(exp.vars)-1])
			switch {
			case '0' <= r && r <= '9':
				spec.maxlen *= 10
				spec.maxlen += int(r - '0')
				if spec.maxlen == 0 || spec.maxlen > 9999 {
					return nil, p.errorf("max-length must be (0, 9999]")
				}
			default:
				if spec.maxlen == 0 {
					return nil, p.errorf("max-length must be (0, 9999]")
				}
				p.unread(r)
				p.setState(parseStateVarList)
			}
		default:
			panic(fmt.Errorf("unhandled parseState(%d)", p.state))
		}
	}
}

func (p *parser) consumeTriplets() (string, error) {
	if len(p.r)-p.stop < 3 || p.r[p.stop] != '%' || !ishex(p.r[p.stop+1]) || !ishex(p.r[p.stop+2]) {
		return "", errorf(p.stop, "incomplete pct-encodeed")
	}
	triplets := p.r[p.stop : p.stop+3]
	p.stop += 3
	return triplets, nil
}
