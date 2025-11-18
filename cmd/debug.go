package main

import (
	"fmt"
	"net/http"
	"runtime"
)

type DebugInfo struct {
	Name     string
	Value    any
	Children []DebugInfo
}

func (server *Server) debugHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := MustLoadCommonData(ctx)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	info := []DebugInfo{
		{
			Name:  "Runtime",
			Value: "",
			Children: []DebugInfo{
				{Name: "Goroutines", Value: runtime.NumGoroutine()},
				{Name: "NumCPU", Value: runtime.NumCPU()},
				{Name: "NumCgoCall", Value: runtime.NumCgoCall()},
			},
		},
		{
			Name:  "Memory",
			Value: "",
			Children: []DebugInfo{
				{Name: "Alloc MB", Value: toMB(m.Alloc)},
				{Name: "TotalAlloc MB", Value: toMB(m.TotalAlloc)},
				{Name: "Sys MB", Value: toMB(m.Sys)},
			},
		},
	}

	_ = DebugPage(data, info).Render(ctx, w)
}

func toMB(v uint64) string {
	return fmt.Sprintf("%.0f", float64(v)/1024/1024)
}
