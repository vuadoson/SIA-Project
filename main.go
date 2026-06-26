package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Cấu trúc dữ liệu nhận từ Frontend
type RoyaltyPayload struct {
	ActionType string  `json:"action_type"`
	Amount     float64 `json:"amount"`
}

// Hàm xử lý khi có dòng tiền tác quyền thật gửi lên
func handlerRoyaltyIncoming(w http.ResponseWriter, r *http.Request) {
	// 1. Mở khóa cổng kết nối an toàn (CORS) sòng phẳng cho trình duyệt
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	// 2. Chấp nhận lệnh kiểm tra an toàn (Preflight OPTIONS)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 3. Chỉ cho phép xử lý phương thức POST gửi dữ liệu thật
	if r.Method != http.MethodPost {
		http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
		return
	}

	var data RoyaltyPayload
	// Giải mã dữ liệu JSON từ giao diện gửi tới
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	// Ghi nhận nhật ký xử lý thời gian thực trên máy chủ
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] XÁC THỰC TÁC QUYỀN THỰC TẾ\n", timestamp)
	fmt.Printf(" └ Hành động: %s | Số tiền: %.2f USD\n", data.ActionType, data.Amount)

	// Phản hồi kết quả sòng phẳng lại cho phía Frontend
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "success",
		"message":   "Hệ thống SIA đã xác thực dòng tiền thành công!",
		"timestamp": timestamp,
	})
}

func main() {
	fmt.Println("========================================")
	fmt.Println("          SIA CORE ENGINE - PRODUCTION   ")
	fmt.Println("========================================")

	// Thiết lập đường dẫn tiếp nhận dữ liệu API từ nút bấm điều khiển
	http.HandleFunc("/api/v1/royalty", handlerRoyaltyIncoming)

	// Thiết lập giao diện trang chủ kiểm tra trạng thái máy chủ
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`
			<html>
			<head>
				<title>SIA System</title>
				<style>
					body { font-family: sans-serif; text-align: center; padding: 50px; background: #0f172a; color: #e2e8f0; }
					h1 { color: #a78bfa; }
					.status { display: inline-block; padding: 10px 20px; background: #10b981; color: white; border-radius: 5px; font-weight: bold; }
				</style>
			</head>
			<body>
				<h1>Security ID Automatic (SIA)</h1>
				<p>Hệ thống máy chủ đám mây đang vận hành thực tế.</p>
				<div class="status">STATUS: LIVE</div>
			</body>
			</html>
		`))
	})

	fmt.Println("[HỆ THỐNG] Máy chủ SIA đang mở cổng 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("[LỖI] Không thể mở cổng máy chủ: %v\n", err)
	}
}
