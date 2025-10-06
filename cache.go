package main

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Server) CacheGet(ctx context.Context, key string, obj any) bool {
	text, err := s.Queries.CacheGet(ctx, key)
	if err != nil {
		return false
	}
	if !text.Valid {
		return false
	}
	if err := json.Unmarshal([]byte(text.String), obj); err != nil {
		return false
	}

	return true
}

func (s *Server) CacheSet(ctx context.Context, key string, obj any) error {
	text, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return s.Queries.CacheSet(ctx, CacheSetParams{
		Key:   key,
		Value: pgtype.Text{String: string(text), Valid: true},
	})
}
