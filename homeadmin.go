package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jonathangjertsen/bino/sql"
)

func (server *Server) getHomesHandler(w http.ResponseWriter, r *http.Request, commonData *CommonData) {
	ctx := r.Context()

	homesDB, err := server.Queries.GetHomes(ctx)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	usersDB, err := server.Queries.GetAppusers(ctx)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	homes := make([]Home, len(homesDB))
	for i, home := range homesDB {
		homes[i] = Home{
			ID:    home.ID,
			Name:  home.Name,
			Users: nil,
		}
	}

	// todo(perf): make it not O(N^2)
	homeless := []User{}
	for _, user := range usersDB {
		found := false
		if user.HomeID.Valid {
			for i, home := range homesDB {
				if home.ID == user.HomeID.Int32 {
					homes[i].Users = append(homes[i].Users, User{
						ID:          user.ID,
						DisplayName: user.DisplayName,
						Email:       user.Email,
					})
					found = true
					break
				}
			}
		}
		if !found {
			homeless = append(homeless, User{
				ID:          user.ID,
				DisplayName: user.DisplayName,
				Email:       user.Email,
			})
		}
	}

	_ = HomesPage(commonData, homes, homeless).Render(ctx, w)
}

func (server *Server) postHomeHandler(w http.ResponseWriter, r *http.Request, commonData *CommonData) {
	formID, err := server.getFormValue(r, "form-id")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	switch formID {
	case "create-home":
		server.postHomeCreateHome(w, r, commonData)
	case "add-user":
		server.postHomeAddUser(w, r, commonData)
	default:
		server.renderError(w, r, commonData, fmt.Errorf("unknown form ID: '%s'", formID))
	}
}

func (server *Server) postHomeCreateHome(w http.ResponseWriter, r *http.Request, commonData *CommonData) {
	ctx := r.Context()

	name, err := server.getFormValue(r, "name")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	err = server.Queries.UpsertHome(ctx, name)
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	http.Redirect(w, r, "/homes", http.StatusFound)
}

func (server *Server) postHomeAddUser(w http.ResponseWriter, r *http.Request, commonData *CommonData) {
	ctx := r.Context()

	optionalFields, _ := server.getFormValues(r, "remove-from-current", "curr-home-id")
	currentStr := optionalFields["curr-home-id"]
	currentHomeID, err := strconv.ParseInt(currentStr, 10, 32)
	removeFromCurrent := err == nil && optionalFields["remove-from-current"] == "on"

	fields, err := server.getFormValues(r, "home-id", "user-id")
	if err != nil {
		server.renderError(w, r, commonData, err)
		return
	}

	userID, err := strconv.ParseInt(fields["user-id"], 10, 32)
	if err != nil {
		server.renderError(w, r, commonData, fmt.Errorf("invalid user ID '%s'", fields["user-id"]))
		return
	}

	homeID, err := strconv.ParseInt(fields["home-id"], 10, 32)
	if err != nil {
		server.renderError(w, r, commonData, fmt.Errorf("invalid home ID '%s'", fields["home-id"]))
		return
	}

	if server.Transaction(ctx, func(ctx context.Context, q *sql.Queries) error {
		if homeID > 0 {
			if err := server.Queries.AddUserToHome(ctx, sql.AddUserToHomeParams{
				HomeID:    int32(homeID),
				AppuserID: int32(userID),
			}); err != nil {
				return fmt.Errorf("adding user to home: %w", err)
			}
		}
		if removeFromCurrent {
			if err := server.Queries.RemoveUserFromHome(ctx, sql.RemoveUserFromHomeParams{
				HomeID:    int32(currentHomeID),
				AppuserID: int32(userID),
			}); err != nil {
				return fmt.Errorf("removing user from home: %w", err)
			}
		}
		return nil
	}); err != nil {
		server.renderError(w, r, commonData, fmt.Errorf("failed to add user: %w", err))
		return
	}

	http.Redirect(w, r, "/homes", http.StatusFound)
}

func (server *Server) getFormValues(r *http.Request, fields ...string) (map[string]string, error) {
	out := make(map[string]string)
	var err error
	for _, f := range fields {
		out[f], err = server.getFormValue(r, f)
	}
	return out, err
}

func (server *Server) getFormValue(r *http.Request, field string) (string, error) {
	if err := r.ParseForm(); err != nil {
		return "", fmt.Errorf("parsing form: %w", err)
	}
	values, ok := r.Form[field]
	if !ok {
		return "", fmt.Errorf("missing form value '%s'", field)
	}
	if len(values) != 1 {
		return "", fmt.Errorf("expect 1 form value for '%s', got %d", field, len(values))
	}
	return values[0], nil
}
