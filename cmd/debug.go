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
	u, _ := disk.Usage("/")
	h, _ := host.Info()
	users, _ := host.Users()
	cwd, _ := os.Getwd()

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
				{Name: "Disk total GB", Value: toGB(u.Total)},
				{Name: "Disk free GB", Value: toGB(u.Free)},
				{Name: "Disk used GB", Value: toGB(u.Used)},
				{Name: "Hostname", Value: h.Hostname},
				{Name: "Uptime", Value: data.User.Language.FormatTimeAbsWithRelParen(time.Unix(int64(h.BootTime), 0))},
				{Name: "Public IP", Value: server.Runtime.PublicIP},
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
				{Name: "Alloc MB", Value: toMB(m.Alloc)},
				{Name: "Total MB", Value: toMB(m.Sys)},
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
