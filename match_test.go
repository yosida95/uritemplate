// Copyright (C) 2016 Kohei YOSHIDA. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of The BSD 3-Clause License
// that can be found in the LICENSE file.

package uritemplate

import (
	"testing"
)

func TestMatcher(t *testing.T) {
	for _, c := range testTemplateCases {
		tmpl, err := New(c.raw)
		if err != nil {
			t.Errorf("unexpected error on %q: %#v", c.raw, err)
			continue
		}

		m, err := CompileMatcher(tmpl)
		if err != nil {
			t.Errorf("unexpected error on %q: %#v", c.raw, err)
			return
		}

		if !m.Match(c.expected, nil) {
			t.Errorf("failed to match %q against %q", c.raw, c.expected)
			t.Log(m.prog.String())
			continue
		}
	}
}
