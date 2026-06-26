package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Cấu hình tài khoản đối lưu của Nhà Sáng Lập
type FounderConfig struct {
	FounderName   string
	BankName      string
	AccountNumber string
	Currency      string
}

// Cấu trúc dữ liệu Tác quyền THẬT nhận từ đối tác gửi về
type RoyaltyPayload struct {
	Partner    string  `json:"partner"`     // Tên tập đoàn gửi dữ liệu
	DeviceID   string  `json:"device_id"`   // Mã định danh thiết bị quét
	ActionType string  `json:"action_type"` // Loại hình đối lưu (0-chạm, mã hóa...)
	Amount     float64 `json:"amount"`      // Số tiền tác quyền ($0.01)
}

// Khởi tạo thông tin cố định của ông
var siaFounder = FounderConfig{
	FounderName:   "VU VAN TRONG",
	BankName:      "MB Bank (Ngân hàng Quân đội)",
	AccountNumber: "6801639330636",
	Currency:      "USD",
}

// Hàm xử lý khi có dòng tiền tác quyền thật đổ về qua Internet
func handlerRoyaltyIncoming(w http.ResponseWriter, r *http.Request) {
	// 1. Mở khóa cổng kết nối an toàn (CORS)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	// 2. Chấp nhận lệnh kiểm tra an toàn (Preflight Request) của trình duyệt
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 3. Chỉ cho phép xử lý nếu là lệnh POST gửi dữ liệu thật
	if r.Method != http.MethodPost {
		http.Error(w, "Phương thức không được hỗ trợ", http.StatusMethodNotAllowed)
		return
	}


	var data RoyaltyPayload
	// Giải mã dữ liệu JSON thực tế gửi từ Internet
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	// Ghi nhận nhật ký xử lý thời gian thực trên máy chủ
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Printf("[%s] NHẬN TÁC QUYỀN THỰC TẾ từ: %s (Thiết bị: %s)\n", timestamp, data.Partner, data.DeviceID)
	fmt.Printf("    └─ Hành động: %s | Số tiền: +$%0.3f -> Trạng thái: Sẵn sàng lệnh kết chuyển về STK %s\n", 
		data.ActionType, data.Amount, siaFounder.AccountNumber)

	// Phản hồi sòng phẳng lại cho phía Big Tech là máy chủ SIA đã ghi nhận thành công
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"success","message":"Ghi nhận tác quyền đối lưu thành công"}`))
}

func main() {
	fmt.Println("==================================================")
	fmt.Println("        SIA CORE ENGINE - PRODUCTION SERVER       ")
	fmt.Println("==================================================")
	fmt.Printf("[XÁC THỰC] Chủ sở hữu tối cao: %s\n", siaFounder.FounderName)
	fmt.Printf("[KẾT NỐI] Cổng thanh toán MB Bank: %s - STK: %s\n", siaFounder.BankName, siaFounder.AccountNumber)
	
	// Thiết lập đường dẫn API thực tế để tiếp nhận dữ liệu toàn cầu
	http.HandleFunc("/api/v1/royalty", handleRoyaltyIncoming)
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
                body { font-family: sans-serif; text-align: center; padding: 50px; background: #f4f6f9; color: #333; }
                h1 { color: #5b21b6; }
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


	// Mở cổng mạng 8080 để lắng nghe kết nối Internet thực tế
	fmt.Println("[HỆ THỐNG] Máy chủ SIA đang mở cổng :8080 và lắng nghe toàn cầu...")
	fmt.Println("--------------------------------------------------")
	
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("[LỖI] Không thể mở cổng máy chủ: %v\n", err)
	}
}
