package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

const listenPort = "8088"
const masterProcessMarker = "nginx: master process"

// UptimeResponse 定义了 JSON 响应的结构
type UptimeResponse struct {
	Status        string `json:"status"`
	Timestamp     int64  `json:"timestamp"`
	UptimeSeconds int64  `json:"uptime_seconds"`
	Days          int64  `json:"days"`
	Hours         int64  `json:"hours"`
	Minutes       int64  `json:"minutes"`
	Seconds       int64  `json:"seconds"`
	Formatted     string `json:"formatted_uptime"`
}

// findNginxMasterProcess 遍历所有进程，查找 Nginx Master 进程
func findNginxMasterProcess() (*process.Process, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, fmt.Errorf("无法获取进程列表: %v", err)
	}

	for _, p := range procs {
		cmdline, err := p.Cmdline()
		if err == nil && strings.Contains(cmdline, masterProcessMarker) {
			return p, nil
		}
	}

	return nil, fmt.Errorf("未找到 Nginx Master 进程 (%s)", masterProcessMarker)
}

// getNginxUptimeData 计算并返回 Nginx Master 进程的运行时间数据结构
func getNginxUptimeData() (*UptimeResponse, error) {
	masterProc, err := findNginxMasterProcess()
	if err != nil {
		return nil, err
	}

	// 获取进程创建时间 (毫秒级 Unix 时间戳)
	createTimeMillis, err := masterProc.CreateTime()
	if err != nil {
		return nil, fmt.Errorf("无法获取 Nginx 进程创建时间: %v", err)
	}

	now := time.Now()
	uptimeDuration := now.Sub(time.UnixMilli(createTimeMillis))

	totalSeconds := int64(uptimeDuration.Seconds())

	days := totalSeconds / 86400
	totalSeconds %= 86400
	hours := totalSeconds / 3600
	totalSeconds %= 3600
	minutes := totalSeconds / 60
	seconds := totalSeconds % 60

	// 格式化输出字符串
	formattedUptime := fmt.Sprintf(
		"%d days, %02d:%02d:%02d",
		days, hours, minutes, seconds,
	)

	return &UptimeResponse{
		Status:        "ok",
		Timestamp:     now.Unix(),
		UptimeSeconds: int64(uptimeDuration.Seconds()),
		Days:          days,
		Hours:         hours,
		Minutes:       minutes,
		Seconds:       seconds,
		Formatted:     formattedUptime,
	}, nil
}

// uptimeHandler 处理 HTTP 请求，返回 Nginx Uptime JSON
func uptimeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	data, err := getNginxUptimeData()
	if err != nil {
		log.Printf("获取 Uptime 失败: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		// 构造一个错误 JSON 响应
		errorResponse := map[string]string{
			"status":  "error",
			"message": err.Error(),
		}
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	w.WriteHeader(http.StatusOK)
	// 将 UptimeResponse 结构体编码为 JSON 并写入响应
	json.NewEncoder(w).Encode(data)
}

func main() {
	http.HandleFunc("/uptime", uptimeHandler)

	addr := fmt.Sprintf("127.0.0.1:%s", listenPort)
	log.Printf("Go Uptime 服务启动中，监听地址: http://%s/uptime", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("无法启动服务: %v", err)
	}
}
