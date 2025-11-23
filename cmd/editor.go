package main

import "net/http"

func (server *Server) editor(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := MustLoadCommonData(ctx)

	_ = EditorPage(data).Render(ctx, w)
}
