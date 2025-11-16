//go:generate go tool go-enum --no-iota --values
package main

import "net/http"

// ENUM(
//
//	None        = 0,
//	Rehabber    = 1,
//	Coordinator = 2,
//	Admin       = 3,
//
// )
type AccessLevel int32

func (server *Server) accessHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)
	_ = AccessLevelPage(commonData).Render(ctx, w)
}
