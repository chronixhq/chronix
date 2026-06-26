package systemdata

import (
	"log/slog"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
)

type (
	SystemData struct {
		Alloc         uint64  `json:"alloc"`
		SystemAlloc   uint64  `json:"systemAlloc"`
		NumGoRoutines int     `json:"numGoRoutines"`
		NumCPUs       int     `json:"numCPUs"`
		CPUPercent    float64 `json:"CPUPercent"`
	}
)

var (
	systemData = SystemData{}
)

func GetSystemData() SystemData {
	return systemData
}

func StartSystemDataUpdates() {
	slog.Debug("Starting system data updates")
	go func() {
		updateSystemData()
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			updateSystemData()
		}
	}()

}

func updateSystemData() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	cpuPerc, _ := cpu.Percent(time.Duration(0), false)
	systemData.Alloc = m.Alloc
	systemData.SystemAlloc = m.Sys
	systemData.NumGoRoutines = runtime.NumGoroutine()
	systemData.NumCPUs = runtime.NumCPU()
	if len(cpuPerc) > 0 {
		systemData.CPUPercent = cpuPerc[0]
	} else {
		systemData.CPUPercent = 0
	}
}
