uritemplate.go
==============

`uritemplate.go`_ is a Go implementation of `URI Template`_ [RFC6570] with
full functionality of URI Template Level 4.

Example
-------

.. code-block:: go

   package main

   import (
       "fmt"
       "github.com/yosida95/uritemplate.go"
   )

   var (
       tmpl = uritemplate.MustNew("https://example.com/dictionary/{term:1}/{term}")
   )

   func main() {
       vars := map[string]uritemplate.Value{
           "term": uritemplate.String("cat"),
       }
       ret, err := tmpl.Expand(vars)
       if err != nil{
           fmt.Println(err)
           return
       }
       fmt.Println(ret)

       // Output:
       // https://example.com/dictionary/c/cat
   }

Getting Started
---------------

Installation
~~~~~~~~~~~~

.. code-block:: sh

   $ go get -u github.com/yosida95/uritemplate.go

Documentation
~~~~~~~~~~~~~

The documentation is available on GoDoc_.

License
-------

`uritemplate.go`_ is distributed under the BSD 3-Clause license.
PLEASE read ./LICENSE carefully and follow its clauses to use this software.


Author
------

yosida95_


.. _`URI TEmplate`: https://tools.ietf.org/html/rfc6570
.. _Godoc: https://godoc.org/github.com/yosida95/uritemplate.go
.. _yosida95: https://yosida95.com/
.. _`uritemplate.go`: https://github.com/yosida95/uritemplate.go
