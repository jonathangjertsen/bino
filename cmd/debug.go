package main

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"runtime/debug"

	"github.com/shirou/gopsutil/v3/load"
)

type DebugInfo struct {
	Name     string
	Value    any
	Children []DebugInfo
}

func fetchPublicIP() string {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return string(b)
}

func (server *Server) debugHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	data := MustLoadCommonData(ctx)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	avg, _ := load.Avg()

	buildInfo := []DebugInfo{
		{Name: "Build key", Value: data.BuildKey},
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			buildInfo = append(buildInfo, DebugInfo{Name: setting.Key, Value: setting.Value})
		}
	}

	info := []DebugInfo{
		{
			Name: "Runtime",
			Children: []DebugInfo{
				{Name: "Goroutines", Value: runtime.NumGoroutine()},
				{Name: "NumCPU", Value: runtime.NumCPU()},
				{Name: "Started", Value: data.User.Language.FormatTimeAbsWithRelParen(server.Runtime.TimeStarted)},
				{Name: "System total load avg (up to 100% * n cores)", Value: fmt.Sprintf("%.1f", avg.Load1*100)},
			},
		},
		{
			Name: "Memory",
			Children: []DebugInfo{
				{Name: "Alloc MB", Value: toMB(m.Alloc)},
				{Name: "Total MB", Value: toMB(m.Sys)},
			},
		},
		{
			Name: "Network",
			Children: []DebugInfo{
				{Name: "Public IP", Value: server.Runtime.PublicIP},
			},
		},
		{
			Name:     "Build",
			Children: buildInfo,
		},
	}

	_ = DebugPage(data, info).Render(ctx, w)
}

func toMB(v uint64) string {
	return fmt.Sprintf("%.0f", float64(v)/1024/1024)
}
