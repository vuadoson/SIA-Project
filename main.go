package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Khóa tối cao bảo mật hệ thống SIA
const SIASecurityToken = "SIA-SUPER-KEY-2026"

// Cấu trúc dữ liệu nâng cấp nhận từ Giao diện
type RoyaltyPayload struct {
	ActionType string  `json:"action_type"`
	Amount     float64 `json:"amount"`
	SIAKey     string  `json:"sia_key"`
}

// Hàm lưu nhật ký dòng tiền vào tệp hệ thống vĩnh viễn
func saveToLogFile(logLine string) {
	// Mở hoặc tạo tệp sia_money_flow.log, cho phép ghi nối tiếp vào cuối tệp
	file, err := os.OpenFile("sia_money_flow.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("[LỖI] Không thể ghi nhật ký vào tệp: %v\n", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(logLine + "\n"); err != nil {
		fmt.Printf("[LỖI] Không thể viết dữ liệu vào tệp: %v\n", err)
	}
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
	developerShare := data.Amount * 0.70
	infrastructureShare := data.Amount * 0.20
	reserveFund := data.Amount * 0.10

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// 💾 LƯU TRỮ VĨNH VIỄN: Tạo dòng nhật ký chuẩn hóa dữ liệu
	logLine := fmt.Sprintf("[%s] KHAI THÁC: %s | TỔNG: %.2f USD | LẬP TRÌNH VIÊN (70%%): %.2f USD | HẠ TẦNG (20%%): %.2f USD | DỰ PHÒNG (10%%): %.2f USD | KIỂM ĐỊNH: PASSED", 
		timestamp, data.ActionType, data.Amount, developerShare, infrastructureShare, reserveFund)
	
	// Gọi hàm ghi trực tiếp vào tệp lưu trữ trên máy chủ
	saveToLogFile(logLine)

	// In lịch sử xử lý sòng phẳng lên màn hình máy chủ Render
	fmt.Println(logLine)

	// Phản hồi hóa đơn phân bổ chi tiết về giao diện người dùng
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":               "success",
		"message":              "Trọng tài SIA đã phê duyệt, phân bổ và LƯU TRỮ NHẬT KÝ 3 NĂM vĩnh viễn thành công!",
		"timestamp":            timestamp,
		"total_amount_usd":     data.Amount,
		"share_developer_70":   developerShare,
		"share_hardware_20":    infrastructureShare,
		"share_reserve_10":     reserveFund,
		"security_audit":       "PASSED & ARCHIVED (Đã lưu kho nhật ký lõi)",
	})
}

func main() {
	fmt.Println("========================================")
	fmt.Println("       SIA CORE ENGINE v3 - LONG-TERM LOG  ")
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
				<title>SIA System v3</title>
				<style>
					body { font-family: sans-serif; text-align: center; padding: 50px; background: #0f172a; color: #e2e8f0; }
					h1 { color: #a78bfa; }
					.status { display: inline-block; padding: 10px 20px; background: #10b981; color: white; border-radius: 5px; font-weight: bold; }
				</style>
			</head>
			<body>
				<h1>Security ID Automatic (SIA) - Core v3</h1>
				<p>Hệ thống tích hợp Tự động ghi tệp nhật ký dòng tiền vĩnh viễn.</p>
				<div class="status">STATUS: LOGGING ACTIVE</div>
			</body>
			</html>
		`))
	})

	fmt.Println("[HỆ THỐNG] Máy chủ SIA v3 đang mở cổng 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("[LỖI] Không thể mở cổng máy chủ: %v\n", err)
	}
}
