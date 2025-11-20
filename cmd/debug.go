package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
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

	// Process info
	var procMem runtime.MemStats
	runtime.ReadMemStats(&procMem)

	// Machine info
	avg, err := load.Avg()
	if err != nil {
		LogCtx(ctx, "getting machine Avg: %v", err)
	}
	u, err := disk.Usage("/")
	if err != nil {
		LogCtx(ctx, "getting machine Disk usage: %v", err)
	}
	h, err := host.Info()
	if err != nil {
		LogCtx(ctx, "getting machine Info: %v", err)
	}
	users, err := host.Users()
	if err != nil {
		LogCtx(ctx, "getting machine Users: %v", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		LogCtx(ctx, "getting machine Getwd: %v", err)
	}
	mem, err := mem.VirtualMemory()
	if err != nil {
		LogCtx(ctx, "getting machine VirtualMemory: %v", err)
	}

	// Build info
	buildInfo := []DebugInfo{}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			buildInfo = append(buildInfo, DebugInfo{Name: setting.Key, Value: setting.Value})
		}
	}

	info := []DebugInfo{
		{
			Name: "Machine",
			Children: []DebugInfo{
				{Name: "NumCPU", Value: runtime.NumCPU()},
				{Name: "System total load avg (up to 100% * n cores)", Value: fmt.Sprintf("%.1f", avg.Load1*100)},
				{Name: "Disk total (GB)", Value: toGB(u.Total)},
				{Name: "Disk available (GB)", Value: toGB(u.Free)},
				{Name: "Disk used (GB)", Value: toGB(u.Used)},
				{Name: "Hostname", Value: h.Hostname},
				{Name: "Uptime", Value: data.User.Language.FormatTimeAbsWithRelParen(time.Unix(int64(h.BootTime), 0))},
				{Name: "Public IP", Value: server.Runtime.PublicIP},
				{Name: "Memory total (MB)", Value: toMB(mem.Total)},
				{Name: "Memory available (MB)", Value: toMB(mem.Available)},
				{Name: "Memory used (MB)", Value: toMB(mem.Used)},
				{Name: "Users", Children: SliceToSlice(users, func(in host.UserStat) DebugInfo {
					return DebugInfo{
						Name:  "Username",
						Value: in.User,
					}
				})},
			},
		},
		{
			Name: "Process",
			Children: []DebugInfo{
				{Name: "Goroutines", Value: runtime.NumGoroutine()},
				{Name: "Started", Value: data.User.Language.FormatTimeAbsWithRelParen(server.Runtime.TimeStarted)},
				{Name: "Alloc MB", Value: toMB(procMem.Alloc)},
				{Name: "Total MB", Value: toMB(procMem.Sys)},
				{Name: "Working directory", Value: cwd},
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

func toGB(v uint64) string {
	return fmt.Sprintf("%.0f", float64(v)/1024/1024/1024)
}
