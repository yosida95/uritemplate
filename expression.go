// Copyright (C) 2016 Kohei YOSHIDA. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of The BSD 3-Clause License
// that can be found in the LICENSE file.

package uritemplate

import (
	"io"
)

type varspec struct {
	name    string
	maxlen  int
	explode bool
}

type expression interface {
	expand(io.Writer, map[string]Value) error
}

type expLiterals string

func (exp expLiterals) expand(w io.Writer, _ map[string]Value) error {
	w.Write([]byte(exp))
	return nil
}

type expSimple struct {
	vars []varspec
}

func (exp *expSimple) expand(w io.Writer, varmap map[string]Value) error {
	// TODO(yosida95): implement here
	return nil
}

type expPlus struct {
	vars []varspec
}

func (exp *expPlus) expand(w io.Writer, varmap map[string]Value) error {
	// TODO(yosida95): implement here
	return nil
}

type expCrosshatch struct {
	vars []varspec
}

func (exp *expCrosshatch) expand(w io.Writer, varmap map[string]Value) error {
	// TODO(yosida95): implement here
	return nil
}

type expDot struct {
	vars []varspec
}

func (exp *expDot) expand(w io.Writer, varmap map[string]Value) error {
	// TODO(yosida95): implement here
	return nil
}

type expSlash struct {
	vars []varspec
}

func (exp *expSlash) expand(w io.Writer, varmap map[string]Value) error {
	// TODO(yosida95): implement here
	return nil
}

type expSemicolon struct {
	vars []varspec
}

func (exp *expSemicolon) expand(w io.Writer, varmap map[string]Value) error {
	// TODO(yosida95): implement here
	return nil
}

type expQuestion struct {
	vars []varspec
}

func (exp *expQuestion) expand(w io.Writer, varmap map[string]Value) error {
	// TODO(yosida95): implement here
	return nil
}

type expAmpersand struct {
	vars []varspec
}

func (exp *expAmpersand) expand(w io.Writer, varmap map[string]Value) error {
	// TODO(yosida95): implement here
	return nil
}
