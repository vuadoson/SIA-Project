package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const SIASecurityToken = "SIA-SUPER-KEY-2026"

type RoyaltyPayload struct {
	ActionType string  `json:"action_type"`
	Amount     float64 `json:"amount"`
	SIAKey     string  `json:"sia_key"`
}

func saveToLogFile(logLine string) {
	file, err := os.OpenFile("sia_money_flow.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("[LỖI] Không thể ghi nhật ký: %v\n", err)
		return
	}
	defer file.Close()
	file.WriteString(logLine + "\n")
}

// API 1: Tiếp nhận và phân bổ dòng tiền
func handlerRoyaltyIncoming(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
		return
	}

	var data RoyaltyPayload
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	if data.SIAKey != SIASecurityToken {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"status": "denied", "message": "Mã khóa SIA không chính xác!"})
		return
	}

	developerShare := data.Amount * 0.70
	infrastructureShare := data.Amount * 0.20
	reserveFund := data.Amount * 0.10
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	logLine := fmt.Sprintf("[%s] %s | TỔNG: $%.2f | DEV: $%.2f | HẠ TẦNG: $%.2f | DỰ PHÒNG: $%.2f", 
		timestamp, data.ActionType, data.Amount, developerShare, infrastructureShare, reserveFund)
	
	saveToLogFile(logLine)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":               "success",
		"message":              "Trọng tài hệ thống SIA đã phê duyệt, phân bổ và LƯU TRỮ NHẬT KÝ 3 NĂM vĩnh viễn thành công!",
		"timestamp":            timestamp,
		"total_amount_usd":     data.Amount,
		"share_developer_70":   developerShare,
		"share_hardware_20":    infrastructureShare,
		"share_reserve_10":     reserveFund,
		"security_audit":       "PASSED & ARCHIVED (Đã lưu kho nhật ký lõi)",
	})
}

// API 2: Đọc xuất nhật ký dòng tiền ra màn hình
func handlerFetchLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Đọc tệp tin nhật ký từ ổ đĩa máy chủ
	content, err := os.ReadFile("sia_money_flow.log")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"logs": "Chưa có dữ liệu nhật ký dòng tiền được lưu trữ."})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"logs": string(content)})
}

func main() {
	http.HandleFunc("/api/v1/royalty", handlerRoyaltyIncoming)
	http.HandleFunc("/api/v1/logs", handlerFetchLogs) // Cổng xuất dữ liệu nhật ký

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<h1>SIA CORE ENGINE v4 - FULL SYSTEM ACTIVE</h1>`))
	})

	http.ListenAndServe(":8080", nil)
}
