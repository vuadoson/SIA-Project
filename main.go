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
	"strings"
	"time"
)

const (
	smtpHost    = "smtp.gmail.com"
	smtpPort    = "587"
	senderEmail = "vuadoson@gmail.com"
	senderPass  = "xxxx xxxx xxxx xxxx" // Nhớ giữ nguyên 16 ký tự mật khẩu ứng dụng Google của ông
)

type ClientOrder struct {
	ClientName    string    `json:"client_name"`
	ClientPhone   string    `json:"client_phone"`
	ClientAddress string    `json:"client_address"` // Email khách nhận hàng
	ActionType    string    `json:"action_type"`
	Amount        float64   `json:"amount"`
	Timestamp     time.Time `json:"timestamp"`
}

var orderCache = make(map[string]ClientOrder)

func generateAPIKey() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "SIA_KEY_" + hex.EncodeToString(bytes)
}

// 1. API nhận thông tin từ giao diện web khi khách bấm nút
func handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	if r.Method == http.MethodOptions { return }

	var order ClientOrder
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	order.Timestamp = time.Now()
	
	// Lưu vào bộ nhớ đệm, lấy Số điện thoại làm chìa khóa khóa đơn
	orderCache[order.ClientPhone] = order
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "pending_payment"})
}

// 2. API ĐÓN THÔNG BÁO TỪ MACRODROID TRÊN ĐIỆN THOẠI ÔNG BẮN LÊN
func handlePhoneNotification(w http.ResponseWriter, r *http.Request) {
	smsContent := r.URL.Query().Get("sms_content")
	if smsContent == "" {
		http.Error(w, "Empty content", http.StatusBadRequest)
		return
	}

	log.Printf("📩 Nhận thông báo từ MBBank: %s", smsContent)

	// Duyệt qua danh sách khách đang chờ tiền
	for phone, client := range orderCache {
		// Nếu trong nội dung thông báo app MBBank có chứa số điện thoại của khách
		if strings.Contains(smsContent, phone) {
			apiKey := generateAPIKey()

			// Tiến hành soạn mail bàn giao tự động 100%
			subject := "Subject: [SIA SYSTEM] KICH HOAT TAC QUYEN TU DONG\n"
			mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
			body := fmt.Sprintf(`
				<h3>Xin chào %s,</h3>
				<p>Hệ thống bảo hộ SIA xác nhận đã nhận đủ tiền thanh toán thông qua cổng quét mã QR tự động MB Bank.</p>
				<hr/>
				<h4>🎁 SẢN PHẨM SẠCH CỦA ÔNG ĐÃ ĐƯỢC KÍCH HOẠT THÀNH CÔNG:</h4>
				<ul>
					<li><strong>Gói bản quyền:</strong> %s</li>
					<li><strong>Mã bảo mật API Key:</strong> <code>%s</code></li>
					<li><strong>Link tải dữ liệu nguồn lõi:</strong> <a href='https://vuadoson.github.io/SIA-Project/download/data.zip'>Tải xuống tại đây</a></li>
				</ul>
				<p>Hệ thống vận hành tự động toàn cầu 24/7. Cảm ơn ông đã hợp tác sòng phẳng!</p>
			`, client.ClientName, client.ActionType, apiKey)

			msg := []byte(subject + mime + body)
			auth := smtp.PlainAuth("", senderEmail, senderPass, smtpHost)
			
			err := smtp.SendMail(smtpHost+":"+smtpPort, auth, senderEmail, []string{client.ClientAddress}, msg)
			if err == nil {
				delete(orderCache, phone) // Gửi xong thì xóa khách khỏi hàng đợi
				log.Printf("✅ [Thành công] Đã gửi mail bàn giao đến: %s", client.ClientAddress)
			} else {
				log.Printf("❌ Lỗi gửi mail: %v", err)
			}
			break
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	http.HandleFunc("/api/v1/order", handleCreateOrder)
	http.HandleFunc("/api/v1/sms-trigger", handlePhoneNotification)
	
	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
