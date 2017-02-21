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

// threadList implements https://research.swtch.com/sparse.
type threadList struct {
	dense  []thread
	sparse []uint32
}

type thread struct {
	op *progOp
	pc uint32
}

type machine struct {
	prog *prog

	list1   threadList
	list2   threadList
	matched bool

	input string
}

func (m *machine) at(pos int) (rune, int) {
	if pos < len(m.input) {
		r := m.input[pos]
		if r < utf8.RuneSelf {
			return rune(r), 1
		}
		return utf8.DecodeRuneInString(m.input[pos:])
	}
	return -1, 0
}

func (m *machine) clear(list *threadList) {
	list.dense = list.dense[:0]
}

func (m *machine) add(list *threadList, pc uint32) {
	if i := list.sparse[pc]; i < uint32(len(list.dense)) && list.dense[i].pc == pc {
		return
	}

	n := len(list.dense)
	list.dense = list.dense[:n+1]
	list.sparse[pc] = uint32(n)

	t := &list.dense[n]
	t.pc = pc
	t.op = &m.prog.op[pc]
}

func (m *machine) step(clist *threadList, nlist *threadList, r rune, pos int) {
	for i := 0; i < len(clist.dense); i++ {
		t := clist.dense[i]
		op := t.op

		switch op.code {
		default:
			panic("unhandled opcode")
		case opRune:
			if op.r == r {
				m.add(nlist, t.pc+1)
			}
		case opRuneClass:
			ret := false
			if !ret && op.rc&runeClassU == runeClassU {
				ret = ret || unicode.Is(rangeUnreserved, r)
			}
			if !ret && op.rc&runeClassR == runeClassR {
				ret = ret || unicode.Is(rangeReserved, r)
			}
			if !ret && op.rc&runeClassPctE == runeClassPctE {
				ret = ret || unicode.Is(unicode.ASCII_Hex_Digit, r)
			}
			if ret {
				m.add(nlist, t.pc+1)
			}
		case opLineBegin:
			if pos == 0 {
				m.add(clist, t.pc+1)
			}
		case opLineEnd:
			if r == -1 {
				m.add(clist, t.pc+1)
			}
		case opCapStart:
			m.add(clist, t.pc+1)
		case opCapEnd:
			m.add(clist, t.pc+1)
		case opSplit:
			m.add(clist, t.pc+1)
			m.add(clist, op.i)
		case opJmp, opJmpIfNotDefined, opJmpIfNotEmpty, opJmpIfNotFirst:
			m.add(clist, t.pc+1)
			m.add(clist, op.i)
		case opEnd:
			m.matched = true
		}
	}
	m.clear(clist)
}

func (m *machine) match() bool {
	pos := 0
	clist, nlist := &m.list1, &m.list2
	for {
		if len(clist.dense) == 0 {
			if m.matched {
				break
			}
		}
		if !m.matched {
			m.add(clist, 0)
		}

		r, width := m.at(pos)
		m.step(clist, nlist, r, pos)
		if width < 1 {
			break
		}
		pos += width

		clist, nlist = nlist, clist
	}
	return m.matched
}

type Matcher struct {
	prog prog
}

func CompileMatcher(tmpl *Template) (*Matcher, error) {
	c := compiler{}
	c.init()

	c.compile(tmpl)

	m := Matcher{
		prog: *c.prog,
	}
	return &m, nil
}

func (match *Matcher) Match(expansion string, vals map[string]Value) bool {
	n := len(match.prog.op)
	m := machine{
		prog: &match.prog,
		list1: threadList{
			dense:  make([]thread, 0, n),
			sparse: make([]uint32, n),
		},
		list2: threadList{
			dense:  make([]thread, 0, n),
			sparse: make([]uint32, n),
		},
		input: expansion,
	}
	return m.match()
}
