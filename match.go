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
	cap     map[string][]int

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

func (m *machine) add(list *threadList, pc uint32, pos int, next bool, cap map[string][]int) {
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
			op:  &m.prog.op[pc],
			cap: make(map[string][]int, len(m.cap)),
		}
		for k, v := range cap {
			e.t.cap[k] = make([]int, len(v))
			copy(e.t.cap[k], v)
		}
	case opLineBegin:
		if pos == 0 {
			m.add(list, pc+1, pos, next, cap)
		}
	case opLineEnd:
		if !next {
			m.add(list, pc+1, pos, next, cap)
		}
	case opCapStart, opCapEnd:
		ocap := make(map[string][]int, len(m.cap))
		for k, v := range cap {
			ocap[k] = make([]int, len(v))
			copy(ocap[k], v)
		}
		ocap[op.name] = append(ocap[op.name], pos)
		m.add(list, pc+1, pos, next, ocap)
	case opSplit:
		m.add(list, pc+1, pos, next, cap)
		m.add(list, op.i, pos, next, cap)
	case opJmp:
		m.add(list, op.i, pos, next, cap)
	case opJmpIfNotDefined:
		m.add(list, pc+1, pos, next, cap)
		m.add(list, op.i, pos, next, cap)
	case opJmpIfNotFirst:
		m.add(list, pc+1, pos, next, cap)
		m.add(list, op.i, pos, next, cap)
	case opJmpIfNotEmpty:
		m.add(list, op.i, pos, next, cap)
		m.add(list, pc+1, pos, next, cap)
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
				m.add(nlist, e.pc+1, nextPos, next, t.cap)
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
				m.add(nlist, e.pc+1, nextPos, next, t.cap)
			}
		case opEnd:
			m.matched = true
			for k, v := range t.cap {
				m.cap[k] = make([]int, len(v))
				copy(m.cap[k], v)
			}
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
			m.add(clist, 0, pos, next, m.cap)
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

func (matcher *Matcher) Match(expansion string) Values {
	n := len(matcher.prog.op)
	m := machine{
		prog: &matcher.prog,
		list1: threadList{
			dense:  make([]threadEntry, 0, n),
			sparse: make([]uint32, n),
		},
		list2: threadList{
			dense:  make([]threadEntry, 0, n),
			sparse: make([]uint32, n),
		},
		cap:   make(map[string][]int, matcher.prog.numCap),
		input: expansion,
	}
	if !m.match() {
		return nil
	}

	match := make(Values, len(m.cap))
	for name, indices := range m.cap {
		v := Value{V: make([]string, len(indices)/2)}
		for i := range v.V {
			v.V[i] = pctDecode(expansion[indices[2*i]:indices[2*i+1]])
		}
		if len(v.V) == 1 {
			v.T = ValueTypeString
		} else {
			v.T = ValueTypeList
		}
		match[name] = v
	}

	return match
}
