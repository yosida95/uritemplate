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
	dense  []threadEntry
	sparse []uint32
}

type threadEntry struct {
	pc uint32
	t  *thread
}

type thread struct {
	op  *progOp
	cap map[string][]int
}

type machine struct {
	prog *prog

	list1   threadList
	list2   threadList
	matched bool

	input string
}

func (m *machine) at(pos int) (rune, int, bool) {
	if l := len(m.input); pos < l {
		c := m.input[pos]
		if c < utf8.RuneSelf {
			return rune(c), 1, pos+1 < l
		}
		r, size := utf8.DecodeRuneInString(m.input[pos:])
		return r, size, pos+size < l
	}
	return -1, 0, false
}

func (m *machine) add(list *threadList, pc uint32, pos int, next bool) {
	if i := list.sparse[pc]; i < uint32(len(list.dense)) && list.dense[i].pc == pc {
		return
	}

	n := len(list.dense)
	list.dense = list.dense[:n+1]
	list.sparse[pc] = uint32(n)

	e := &list.dense[n]
	e.pc = pc
	e.t = nil

	op := &m.prog.op[pc]
	switch op.code {
	default:
		panic("unhandled opcode")
	case opRune, opRuneClass, opEnd:
		e.t = &thread{
			op: &m.prog.op[pc],
		}
	case opLineBegin:
		if pos == 0 {
			m.add(list, pc+1, pos, next)
		}
	case opLineEnd:
		if !next {
			m.add(list, pc+1, pos, next)
		}
	case opCapStart, opCapEnd:
		m.add(list, pc+1, pos, next)
	case opSplit:
		m.add(list, pc+1, pos, next)
		m.add(list, op.i, pos, next)
	case opJmp:
		m.add(list, op.i, pos, next)
	case opJmpIfNotDefined:
		m.add(list, pc+1, pos, next)
		m.add(list, op.i, pos, next)
	case opJmpIfNotFirst:
		m.add(list, pc+1, pos, next)
		m.add(list, op.i, pos, next)
	case opJmpIfNotEmpty:
		m.add(list, op.i, pos, next)
		m.add(list, pc+1, pos, next)
	}
}

func (m *machine) step(clist *threadList, nlist *threadList, r rune, pos int, nextPos int, next bool) {
	for i := 0; i < len(clist.dense); i++ {
		e := clist.dense[i]
		if e.t == nil {
			continue
		}

		t := e.t
		op := t.op
		switch op.code {
		default:
			panic("unhandled opcode")
		case opRune:
			if op.r == r {
				m.add(nlist, e.pc+1, nextPos, next)
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
				m.add(nlist, e.pc+1, nextPos, next)
			}
		case opEnd:
			m.matched = true
			clist.dense = clist.dense[:0]
		}
	}
	clist.dense = clist.dense[:0]
}

func (m *machine) match() bool {
	pos := 0
	clist, nlist := &m.list1, &m.list2
	for {
		if len(clist.dense) == 0 && m.matched {
			break
		}
		r, width, next := m.at(pos)
		if !m.matched {
			m.add(clist, 0, pos, next)
		}
		m.step(clist, nlist, r, pos, pos+width, next)

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
			dense:  make([]threadEntry, 0, n),
			sparse: make([]uint32, n),
		},
		list2: threadList{
			dense:  make([]threadEntry, 0, n),
			sparse: make([]uint32, n),
		},
		input: expansion,
	}
	return m.match()
}
