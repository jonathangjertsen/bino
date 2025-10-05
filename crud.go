package main

type CRUDPage struct {
	Header          string
	LangIndependent []string
}

type Names map[int32]string

func (n Names) Name(i LanguageID) (string, bool) {
	name, found := n[int32(i)]
	return name, found
}
