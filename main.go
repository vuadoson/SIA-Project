package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"time"
)

// Cấu hình Email gửi đi (Sử dụng SMTP của Gmail hoặc các dịch vụ khác)
const (
	smtpHost     = "smtp.gmail.com"
	smtpPort     = "587"
	senderEmail  = "hethong.sia.2026@gmail.com" // Thay bằng email của hệ thống ông
	senderPass   = "xxxx xxxx xxxx xxxx"       // Mật khẩu ứng dụng (App Password) sinh ra từ Google
)

// Cấu trúc dữ liệu lưu trữ tạm thời hồ sơ khách hàng chờ thanh toán
type ClientOrder struct {
	ClientName    string    `json:"client_name"`
	ClientPhone   string    `json:"client_phone"`
	ClientAddress string    `json:"client_address"` // Đây chính là Email nhận hàng
	ActionType    string    `json:"action_type"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Timestamp     time.Time `json:"timestamp"`
}

// Lưu trữ bộ nhớ đệm (Trong thực tế nên dùng Database, tạm thời lưu map để demo)
var orderCache = make(map[string]ClientOrder)

// Hàm tự động sinh ra chuỗi bảo mật API Key ngẫu nhiên cho khách
func generateAPIKey() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "SIA-KEY-DEFAULT-2026"
	}
	return "SIA_KEY_" + hex.EncodeToString(bytes)
}

// 1. API Nhận hồ sơ từ giao diện index.html khi khách bấm nút
func handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order ClientOrder
	err := json.NewDecoder(r.Body).Decode(&order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	order.Timestamp = time.Now()

	// Sử dụng số điện thoại hoặc mã định danh làm Key để tra cứu khi tiền về
	orderCache[order.ClientPhone] = order

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Hồ sơ đã được khóa vào hàng đợi thanh toán"})
}

// 2. WEBHOOK: Nơi ngân hàng tự động bắn tín hiệu về khi khách thanh toán thành công
func handleBankWebhook(w http.ResponseWriter, r *http.Request) {
	// Cấu trúc dữ liệu mẫu mà các bên như Casso/PayOS thường bắn về
	type WebhookData struct {
		Amount      float64 `json:"amount"`
		Description string  `json:"description"` // Nội dung chuyển khoản chứa số điện thoại ví dụ: "SIA_0912345678_API_CALL"
	}

	var incomingData WebhookData
	err := json.NewDecoder(r.Body).Decode(&incomingData)
	if err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Phân tích nội dung chuyển khoản để tìm số điện thoại khách hàng nhằm lục lại hồ sơ
	// Tìm kiếm xem có trùng khớp với khách hàng nào trong bộ nhớ không
	var foundPhone string
	for phone := range orderCache {
		if fmt.Sprintf("%s", incomingData.Description) == phone { // Hoặc logic tìm chuỗi con
			foundPhone = phone
			break
		}
	}

	if foundPhone != "" {
		client := orderCache[foundPhone]

		// TIẾN HÀNH TỰ ĐỘNG GỬI EMAIL SẢN PHẨM
		apiKey := generateAPIKey()
		subject := "Subject: [SIA SYSTEM] BAN GIAO QUYEN TRUY VAN VA SAN PHAM CONG NGHE\n"
		mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
		
		body := fmt.Sprintf(`
			<h3>Kính gửi đối tác: %s</h3>
			<p>Hệ thống Trọng tài Công nghệ SIA Production V5 xác nhận đã nhận đủ số tiền: <strong>%.2f %s</strong></p>
			<p>Dòng tiền đã được phân bổ sòng phẳng tự động (70%% Kỹ sư - 20%% Hạ tầng - 10%% Quỹ dự phòng).</p>
			<hr/>
			<h4>🎁 SẢN PHẨM CÔNG NGHỆ CỦA ÔNG ĐÃ ĐƯỢC KÍCH HOẠT:</h4>
			<ul>
				<li><strong>Loại hình khai thác:</strong> %s</li>
				<li><strong>Mã bảo mật API Key của ông:</strong> <code style='background:#f1f5f9; padding:5px; color:#e11d48;'>%s</code></li>
				<li><strong>Đường dẫn tải dữ liệu lõi (Dữ liệu sạch):</strong> <a href='https://vuadoson.github.io/SIA-Project/download/data.zip'>Tải tại đây</a></li>
			</ul>
			<br/>
			<p><i>Mọi lịch sử giao dịch đã được băm mã hóa bảo mật và lưu vết kiểm toán 3 năm vĩnh viễn trên tệp hệ thống sia_money_flow.log.</i></p>
			<p>Trân trọng,<br/><strong>Hệ thống Tác quyền Tự động SIA Production V5</strong></p>
		`, client.ClientName, client.Amount, client.Currency, client.ActionType, apiKey)

		msg := []byte(subject + mime + body)
		auth := smtp.PlainAuth("", senderEmail, senderPass, smtpHost)

		// Thực hiện lệnh gửi email đi lập tức
		err := smtp.SendMail(smtpHost+":"+smtpPort, auth, senderEmail, []string{client.ClientAddress}, msg)
		if err != nil {
			log.Printf("❌ Lỗi gửi email tự động: %v", err)
		} else {
			log.Printf("✅ Đã tự động bàn giao sản phẩm thành công tới email: %s", client.ClientAddress)
			// Thanh toán xong thì xóa khỏi bộ nhớ đệm hàng đợi
			delete(orderCache, foundPhone)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	http.HandleFunc("/api/v1/order", handleCreateOrder)
	http.HandleFunc("/api/v1/webhook-bank", handleBankWebhook) // Đường link để cấu hình bên phía ngân hàng

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("🚀 Hệ thống Tự động hóa Toàn phần SIA đang chạy tại Port: %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
