//go:generate go tool go-enum --no-iota
package main

import "net/http"

// ENUM(
//
//	Unknown                        = 0,
//	Registered                     = 1,
//	Admitted                       = 2,
//	Adopted                        = 3,
//	Released                       = 4,
//	TransferredToOtherHome         = 5,
//	TransferredOutsideOrganization = 6,
//	Died                           = 7,
//	Euthanized                     = 8,
//	TagAdded                       = 9,  // Associated ID is tag ID
//	TagRemoved                     = 10, // Associated ID is tag ID
//	StatusChanged                  = 11, // Associated ID is status
//	Deleted                        = 12,
//	NameChanged                    = 13,
//
// )
type Event int32

func (server *Server) postEventSetNoteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commonData := MustLoadCommonData(ctx)

	event, err := server.getPathID(r, "event")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	note, err := server.getFormValue(r, "value")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	if err := server.Queries.SetEventNote(ctx, SetEventNoteParams{
		ID:   event,
		Note: note,
	}); err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	server.redirectToReferer(w, r)
}
