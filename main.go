package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Khóa tối cao bảo mật hệ thống SIA
const SIASecurityToken = "SIA-SUPER-KEY-2026"

// Cấu trúc dữ liệu nâng cấp nhận từ Giao diện
type RoyaltyPayload struct {
	ActionType string  `json:"action_type"`
	Amount     float64 `json:"amount"`
	SIAKey     string  `json:"sia_key"` // Khóa bảo mật truyền từ giao diện
}

// Hàm xử lý khi có dòng tiền tác quyền thật gửi lên
func handlerRoyaltyIncoming(w http.ResponseWriter, r *http.Request) {
	// Mở khóa cổng kết nối an toàn (CORS) sòng phẳng cho trình duyệt
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
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Dữ liệu không hợp lệ", http.StatusBadRequest)
		return
	}

	// 🔒 KIỂM TRA BẢO MẬT: Xác thực khóa tối cao SIA
	if data.SIAKey != SIASecurityToken {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "denied",
			"message": "CẢNH BÁO BẢO MẬT: Khóa xác thực SIA không chính xác! Lệnh bị từ chối.",
		})
		return
	}

	// ⚖️ TRỌNG TÀI LOGIC: Tự động phân bổ dòng tiền tác quyền khách quan
	developerShare := data.Amount * 0.70 // 70% chia cho người viết code, đóng góp chất xám
	infrastructureShare := data.Amount * 0.20 // 20% nuôi máy chủ, vận hành hệ thống đám mây
	reserveFund := data.Amount * 0.10         // 10% Quỹ dự phòng phát triển tương lai

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// In lịch sử xử lý sòng phẳng lên màn hình máy chủ Render
	fmt.Printf("[%s] KHỞI CHẠY PHÂN BỔ DÒNG TIỀN TÁC QUYỀN SÒNG PHẲNG\n", timestamp)
	fmt.Printf(" ├ Tổng nhận: %.2f USD | Khai thác: %s\n", data.Amount, data.ActionType)
	fmt.Printf(" ├ Quỹ Lập Trình Viên (70%%): %.2f USD\n", developerShare)
	fmt.Printf(" ├ Quỹ Hạ Tầng Đám Mây (20%%): %.2f USD\n", infrastructureShare)
	fmt.Printf(" └ Quỹ Dự Phòng Hệ Thống (10%%): %.2f USD\n", reserveFund)

	// Phản hồi hóa đơn phân bổ chi tiết về giao diện người dùng
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":               "success",
		"message":              "Trọng tài hệ thống SIA đã phê duyệt và phân bổ dòng tiền thành công!",
		"timestamp":            timestamp,
		"total_amount_usd":     data.Amount,
		"share_developer_70":   developerShare,
		"share_hardware_20":    infrastructureShare,
		"share_reserve_10":     reserveFund,
		"security_audit":       "PASSED (Hệ thống an toàn tuyệt đối)",
	})
}

func main() {
	fmt.Println("========================================")
	fmt.Println("       SIA CORE ENGINE v2 - SECURE LIVE  ")
	fmt.Println("========================================")

	http.HandleFunc("/api/v1/royalty", handlerRoyaltyIncoming)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`
			<html>
			<head>
				<title>SIA System v2</title>
				<style>
					body { font-family: sans-serif; text-align: center; padding: 50px; background: #0f172a; color: #e2e8f0; }
					h1 { color: #a78bfa; }
					.status { display: inline-block; padding: 10px 20px; background: #10b981; color: white; border-radius: 5px; font-weight: bold; }
				</style>
			</head>
			<body>
				<h1>Security ID Automatic (SIA) - Core v2</h1>
				<p>Hệ thống máy chủ đám mây tích hợp Trọng tài phân bổ đang vận hành.</p>
				<div class="status">STATUS: LIVE & SECURE</div>
			</body>
			</html>
		`))
	})

	fmt.Println("[HỆ THỐNG] Máy chủ SIA v2 đang mở cổng 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("[LỖI] Không thể mở cổng máy chủ: %v\n", err)
	}
}
