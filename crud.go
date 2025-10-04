package main

type CRUDPage struct {
	Header          string
	LangIndependent []string
}

type Names map[int32]string

func (n Names) Name(i int32) (string, bool) {
	name, found := n[i]
	return name, found
}
