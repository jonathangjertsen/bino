package main

import (
	"context"
	"log"
	"time"
)

func backgroundDeleteExpiredSessions(
	ctx context.Context,
	queries *Queries,
) {
	for {
		log.Printf("running background job: delete expired sessions")
		if result, err := queries.DeleteStaleSessions(ctx); err != nil {
			log.Printf("error deleting stale sessions: %v", err)
		} else {
			log.Printf("deleted stale sessions (%d)", result.RowsAffected())
		}
		time.Sleep(time.Hour)
	}
}
