package main

import "github.com/jonathangjertsen/bino/ln"

type CRUDPage struct {
	Header          ln.L
	LangIndependent []ln.L
}

type Names map[int32]string

func (n Names) Name(i int32) (string, bool) {
	name, found := n[i]
	return name, found
}
