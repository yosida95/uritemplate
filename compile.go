// Copyright (C) 2016 Kohei YOSHIDA. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of The BSD 3-Clause License
// that can be found in the LICENSE file.

package uritemplate

import (
	"strconv"
	"unicode/utf8"
)

type compiler struct {
	prog *prog
}

func (c *compiler) init() {
	c.prog = &prog{}
}

func (c *compiler) op(opcode progOpcode) uint32 {
	i := len(c.prog.op)
	c.prog.op = append(c.prog.op, progOp{code: opcode})
	return uint32(i)
}

func (c *compiler) opWithRune(opcode progOpcode, r rune) uint32 {
	addr := c.op(opcode)
	(&c.prog.op[addr]).r = r
	return addr
}

func (c *compiler) opWithRuneClass(opcode progOpcode, rc runeClass) uint32 {
	addr := c.op(opcode)
	(&c.prog.op[addr]).rc = rc
	return addr
}

func (c *compiler) opWithAddr(opcode progOpcode, absaddr uint32) uint32 {
	addr := c.op(opcode)
	(&c.prog.op[addr]).i = absaddr
	return addr
}

func (c *compiler) opWithAddrDelta(opcode progOpcode, delta uint32) uint32 {
	return c.opWithAddr(opcode, uint32(len(c.prog.op))+delta)
}

func (c *compiler) opWithName(opcode progOpcode, name string) uint32 {
	addr := c.op(opcode)
	(&c.prog.op[addr]).name = name
	return addr
}

func (c *compiler) sizeString(str string) uint32 {
	return uint32(utf8.RuneCountInString(str))
}

func (c *compiler) compileString(str string) {
	for i := 0; i < len(str); {
		// NOTE(yosida95): It is confirmed at parse time that literals
		// consist of only valid-UTF8 runes.
		r, size := utf8.DecodeRuneInString(str[i:])
		c.opWithRune(opRune, r)
		i += size
	}
}

func (c *compiler) sizeRuneClass(rc runeClass, maxlen int) uint32 {
	return 7*uint32(maxlen) - 1
}

func (c *compiler) compileRuneClass(rc runeClass, maxlen int) {
	// first rune
	c.opWithAddrDelta(opSplit, 3)                 // raw rune or pct-encoded
	c.opWithRuneClass(opRuneClass, rc)            // raw rune
	c.opWithAddrDelta(opJmp, 4)                   //
	c.opWithRune(opRune, '%')                     // pct-encoded
	c.opWithRuneClass(opRuneClass, runeClassPctE) //
	c.opWithRuneClass(opRuneClass, runeClassPctE) //
	// second and subsequent rune
	for i := 1; i < maxlen; i++ {
		c.opWithAddrDelta(opSplit, 7*uint32(maxlen-i))

		c.opWithAddrDelta(opSplit, 3)                 // raw rune or pct-encoded
		c.opWithRuneClass(opRuneClass, rc)            // raw rune
		c.opWithAddrDelta(opJmp, 4)                   //
		c.opWithRune(opRune, '%')                     // pct-encoded
		c.opWithRuneClass(opRuneClass, runeClassPctE) //
		c.opWithRuneClass(opRuneClass, runeClassPctE) //
	}
}

func (c *compiler) sizeRuneClassInfinite(rc runeClass) uint32 {
	return 8
}

func (c *compiler) compileRuneClassInfinite(rc runeClass) {
	// first rune
	addr := c.opWithAddrDelta(opSplit, 3)         // raw rune or pct-encoded
	c.opWithRuneClass(opRuneClass, rc)            // raw rune
	c.opWithAddrDelta(opJmp, 4)                   //
	c.opWithRune(opRune, '%')                     // pct-encoded
	c.opWithRuneClass(opRuneClass, runeClassPctE) //
	c.opWithRuneClass(opRuneClass, runeClassPctE) //
	c.opWithAddrDelta(opSplit, 2)                 // loop
	c.opWithAddr(opJmp, addr)                     //
}

func (c *compiler) sizeVarspecValue(spec varspec, expr *expression) uint32 {
	// opCapStart + opCapEnd
	size := uint32(2)
	if !spec.explode && spec.maxlen > 0 {
		size += c.sizeRuneClass(expr.allow, spec.maxlen)
	} else {
		size += c.sizeRuneClassInfinite(expr.allow)
	}
	return size
}

func (c *compiler) compileVarspecValue(spec varspec, expr *expression) {
	var specname string
	if !spec.explode && spec.maxlen > 0 {
		specname = spec.name + ":" + strconv.Itoa(spec.maxlen)
	} else {
		specname = spec.name
	}

	c.prog.numCap++
	c.opWithName(opCapStart, specname)
	if !spec.explode && spec.maxlen > 0 {
		c.compileRuneClass(expr.allow, spec.maxlen)
	} else {
		c.compileRuneClassInfinite(expr.allow)
	}
	c.opWithName(opCapEnd, specname)
}

func (c *compiler) sizeVarspec(spec varspec, expr *expression) uint32 {
	var sep string
	if spec.explode && spec.maxlen == 0 {
		sep = expr.sep
	} else {
		sep = ","
	}

	size := 5 + c.sizeString(expr.ifemp) + c.sizeVarspecValue(spec, expr) + c.sizeString(sep)
	if expr.named {
		size += c.sizeString(spec.name) + 1
		size += c.sizeString("=")
	}
	return size
}

func (c *compiler) compileVarspec(spec varspec, expr *expression) {
	var sep string
	if spec.explode && spec.maxlen == 0 {
		sep = expr.sep
	} else {
		sep = ","
	}

	var offset uint32
	var preval string
	var sizePreval uint32
	if expr.named { // begin loop
		preval = "="
		sizePreval = c.sizeString(preval)

		c.compileString(spec.name)
		if !spec.explode {
			offset += c.sizeString(spec.name)
			offset += sizePreval
		}
	}

	addr := c.opWithAddrDelta(opJmpIfNotEmpty, 2+c.sizeString(expr.ifemp)) // jmp to spec
	(&c.prog.op[addr]).name = spec.name                                    //
	c.compileString(expr.ifemp)                                            // expr.ifemp
	c.opWithAddrDelta(opJmp, 1+c.sizeVarspecValue(spec, expr)+sizePreval)  // skip spec
	c.compileString(preval)                                                // preval
	c.compileVarspecValue(spec, expr)                                      // spec
	c.opWithAddrDelta(opSplit, 2)                                          //
	c.opWithAddrDelta(opJmp, 2+c.sizeString(sep))                          // break loop
	c.compileString(sep)                                                   // sep
	size := c.sizeVarspec(spec, expr)                                      // continue
	c.opWithAddrDelta(opJmp, -size+offset+1)                               //
}

func (c *compiler) compileExpression(expr *expression) {
	if len(expr.vars) < 1 {
		return
	}
	sizeFirst := c.sizeString(expr.first)
	sizeSep := c.sizeString(expr.sep)

	{
		spec := expr.vars[0]
		size := 1 + sizeFirst + c.sizeVarspec(spec, expr)

		addr := c.opWithAddrDelta(opJmpIfNotDefined, size) // skip the spec
		(&c.prog.op[addr]).name = spec.name                //
		c.compileString(expr.first)                        // expr.first
		c.compileVarspec(spec, expr)                       // spec
	}

	for i := 1; i < len(expr.vars); i++ {
		spec := expr.vars[i]
		size := 3 + // opJmpIfNotDefined + opJmpIfNotFirst + opJmp
			sizeFirst + sizeSep + c.sizeVarspec(spec, expr)

		addr := c.opWithAddrDelta(opJmpIfNotDefined, size) // skip the spec
		(&c.prog.op[addr]).name = spec.name                //
		c.opWithAddrDelta(opJmpIfNotFirst, sizeFirst+2)    // jmp to expr.sep
		c.compileString(expr.first)                        //
		c.opWithAddrDelta(opJmp, sizeSep+1)                // jmp to spec
		c.compileString(expr.sep)                          // expr.sep
		c.compileVarspec(spec, expr)                       // spec
	}
}

func (c *compiler) compileLiterals(lt literals) {
	c.compileString(string(lt))
}

func (c *compiler) compile(tmpl *Template) {
	c.op(opLineBegin)
	for i := range tmpl.exprs {
		expr := tmpl.exprs[i]
		switch expr := expr.(type) {
		default:
			panic("unhandled expression")
		case *expression:
			c.compileExpression(expr)
		case literals:
			c.compileLiterals(expr)
		}
	}
	c.op(opLineEnd)
	c.op(opEnd)
}
