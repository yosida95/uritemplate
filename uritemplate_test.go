// Copyright (C) 2016 Kohei YOSHIDA. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of The BSD 3-Clause License
// that can be found in the LICENSE file.

package uritemplate

import (
	"fmt"
)

func Example() {
	tmpl := MustNew("https://example.com/dictionary/{term:1}/{term}")

	vars := map[string]Value{
		"term": String("cat"),
	}
	ret, err := tmpl.Expand(vars)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(ret)

	// Output:
	// https://example.com/dictionary/c/cat
}
