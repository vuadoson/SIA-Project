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
	Currency   string  `json:"currency"` // Loại tiền tệ khách chọn: VND, USD, EUR
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

	// Xác định số tài khoản ngân hàng MB thật dựa trên loại tiền tệ
	targetAccount := "1005071981" // Mặc định là VND
	if data.Currency == "USD" {
		targetAccount = "6801639330636"
	} else if data.Currency == "EUR" {
		targetAccount = "5120703663032"
	}

	developerShare := data.Amount * 0.70
	infrastructureShare := data.Amount * 0.20
	reserveFund := data.Amount * 0.10
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Tạo mã QR thanh toán chuẩn VietQR tự động điền số tài khoản, ngân hàng và số tiền thực tế
	// Định dạng VietQR nhanh: https://img.vietqr.io/image/<BANK_ID>-<ACCOUNT_NO>-qr_only.png?amount=<AMOUNT>&addInfo=<MEMO>
	qrCodeURL := fmt.Sprintf("https://img.vietqr.io/image/MB-%s-qr_only.png?amount=%.0f&addInfo=SIA_ROYALTY_%s", 
		targetAccount, data.Amount, data.ActionType)

	logLine := fmt.Sprintf("[%s] %s | TỔNG: %.2f %s | TÀI KHOẢN NHẬN: MB-%s", 
		timestamp, data.ActionType, data.Amount, data.Currency, targetAccount)
	saveToLogFile(logLine)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":          "success",
		"message":         "Hệ thống SIA đã lập hóa đơn thực tế và khởi tạo cổng thanh toán trực tiếp qua cổng ngân hàng MB!",
		"timestamp":       timestamp,
		"currency":        data.Currency,
		"account_no":      targetAccount,
		"account_name":    "VU VAN TRONG",
		"qr_url":          qrCodeURL,
		"total_amount":    data.Amount,
		"share_dev":       developerShare,
		"share_hardware":  infrastructureShare,
		"share_reserve":   reserveFund,
	})
}

func handlerFetchLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	content, err := os.ReadFile("sia_money_flow.log")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"logs": "Chưa có dữ liệu nhật ký dòng tiền thực tế."})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"logs": string(content)})
}

func main() {
	http.HandleFunc("/api/v1/royalty", handlerRoyaltyIncoming)
	http.HandleFunc("/api/v1/logs", handlerFetchLogs)
	http.ListenAndServe(":8080", nil)
}
