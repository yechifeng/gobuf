package main

import (
	"fmt"

	"github.com/hotgo/gobuf"
)

func genSizer(o *writer, name string, t *gobuf.Type, n int) {
	if genArraySizer(o, name, t, n) {
		return
	}

	if genMapSizer(o, name, t, n) {
		return
	}

	if genPointerSizer(o, name, t) {
		return
	}

	genScalarSizer(o, name, t)
}

func genArraySizer(o *writer, name string, t *gobuf.Type, n int) bool {
	if t.Kind == gobuf.ARRAY {
		elemSize := t.Elem.Size()
		if elemSize != 0 {
			o.Writef("size += gobuf.UvarintSize(uint64(len(%s))) + len(%s) * %d", name, name, elemSize)
		} else {
			o.Writef("gobuf.UvarintSize(uint64(len(%s)))", name)
			o.Writef("for i%d := 0; i%d < len(%s); i%d ++ {", n, n, name, n)
			genSizer(o, fmt.Sprintf("%s[i%d]", name, n), t.Elem, n+1)
			o.Writef("}")
		}
		return true
	}
	return false
}

func genMapSizer(o *writer, name string, t *gobuf.Type, n int) bool {
	if t.Kind == gobuf.MAP {
		keySize := t.Key.Size()
		elemSize := t.Elem.Size()
		if keySize != 0 && elemSize != 0 {
			o.Writef("size += gobuf.UvarintSize(uint64(len(%s))) + len(%s) * (%d + %d)", name, name, keySize, elemSize)
		} else {
			o.Writef("size += gobuf.UvarintSize(uint64(len(%s)))", name)
			o.Writef("for key%d, val%d := range %s {", n, n, name)
			genScalarSizer(o, fmt.Sprintf("key%d", n), t.Key)
			genSizer(o, fmt.Sprintf("val%d", n), t.Elem, n+1)
			o.Writef("}")
		}
	}
	return false
}

func genPointerSizer(o *writer, name string, t *gobuf.Type) bool {
	if t.Kind == gobuf.POINTER {
		o.Writef("size += 1")
		o.Writef("if %s != nil {", name)
		genScalarSizer(o, "*"+name, t.Elem)
		o.Writef("}")
		return true
	}
	return false
}

func genScalarSizer(o *writer, name string, t *gobuf.Type) {
	if t.Size() != 0 {
		o.Writef("size += %d", t.Size())
		return
	}
	switch t.Kind {
	case gobuf.INT:
		o.Writef("size += gobuf.VarintSize(int64(%s))", name)
	case gobuf.UINT:
		o.Writef("size += gobuf.UvarintSize(uint64(%s))", name)
	case gobuf.BYTES, gobuf.STRING:
		o.Writef("size += gobuf.UvarintSize(uint64(len(%s))) + len(%s)", name, name)
	case gobuf.MESSAGE:
		if name[0] == '*' {
			name = name[1:]
		}
		o.Writef("size += %s.Size()", name)
	}
}