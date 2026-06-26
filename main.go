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
	// 🚨 ÔNG ĐIỀN 16 KÝ TỰ MẬT KHẨU ỨNG DỤNG GOOGLE CỦA ÔNG VÀO ĐÂY:
	senderPass  = "xxxx xxxx xxxx xxxx" 
)

type ClientOrder struct {
	ClientName    string    `json:"client_name"`
	ClientPhone   string    `json:"client_phone"`
	ClientAddress string    `json:"client_address"`
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

func handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	if r.Method == http.MethodOptions {
		return
	}

	var order ClientOrder
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	order.Timestamp = time.Now()
	orderCache[order.ClientPhone] = order
	
	log.Printf("📝 Đã tạo đơn hàng chờ: Khách %s - SĐT: %s", order.ClientName, order.ClientPhone)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "pending_payment"})
}

func handlePhoneNotification(w http.ResponseWriter, r *http.Request) {
	smsContent := r.URL.Query().Get("sms_content")
	if smsContent == "" {
		http.Error(w, "Content is empty", http.StatusBadRequest)
		return
	}

	log.Printf("📩 Nhận thông báo giao dịch từ MB Bank: %s", smsContent)

	isMatched := false
	auth := smtp.PlainAuth("", senderEmail, senderPass, smtpHost)

	for phone, client := range orderCache {
		if strings.Contains(smsContent, phone) {
			isMatched = true
			apiKey := generateAPIKey()

			subject := "Subject: [SIA SYSTEM] KICH HOAT TAC QUYEN TU DONG\n"
			mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
			body := fmt.Sprintf(`
				<h3>Xin chào %s,</h3>
				<p>Hệ thống bảo hộ SIA xác nhận đã nhận đủ tiền thanh toán thông qua cổng quét mã QR tự động MB Bank.</p>
				<hr/>
				<h4>🎁 SẢN PHẨM CỦA ÔNG ĐÃ ĐƯỢC KÍCH HOẠT THÀNH CÔNG:</h4>
				<ul>
					<li><strong>Gói bản quyền:</strong> %s</li>
					<li><strong>Mã bảo mật API Key:</strong> <code>%s</code></li>
					<li><strong>Link tải dữ liệu nguồn lõi:</strong> <a href='https://vuadoson.github.io/SIA-Project/download/data.zip'>Tải xuống tại đây</a></li>
				</ul>
				<p>Hệ thống vận hành tự động toàn cầu 24/7. Cảm ơn ông đã hợp tác sòng phẳng!</p>
			`, client.ClientName, client.ActionType, apiKey)

			msg := []byte(subject + mime + body)
			
			err := smtp.SendMail(smtpHost+":"+smtpPort, auth, senderEmail, []string{client.ClientAddress}, msg)
			if err == nil {
				delete(orderCache, phone)
				log.Printf("✅ [Thành công] Đã tự động gửi API Key đến khách: %s", client.ClientAddress)
			} else {
				log.Printf("❌ Lỗi gửi mail cho khách: %v", err)
			}
			break
		}
	}

	// 🚨 HỆ THỐNG HOÀN TIỀN: Khách gõ sai nội dung -> Bắn mail báo động cho ông
	if !isMatched {
		log.Printf("⚠️ CẢNH BÁO: Giao dịch sai nội dung: %s", smsContent)
		
		adminSubject := "Subject: [🚨 SIA ALERT] CO GIAO DICH SAI NOI DUNG - CAN HOAN TIEN\n"
		adminMime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
		adminBody := fmt.Sprintf(`
			<h2 style="color: #ef4444; font-family: sans-serif;">🚨 CẢNH BÁO HỆ THỐNG SIA</h2>
			<p>Hệ thống nhận biến động số dư từ MB Bank nhưng <strong>KHÔNG KHỚP</strong> với bất kỳ SĐT đơn hàng nào đang chờ.</p>
			<hr/>
			<div style="background: #f8fafc; padding: 15px; border: 1px solid #cbd5e1; border-radius: 6px;">
				<strong>Chi tiết thông báo ngân hàng nhận được:</strong><br/>
				<p style="font-family: monospace; color: #1e293b; background: #e2e8f0; padding: 10px; border-radius: 4px;">%s</p>
			</div>
			<br/>
			<p style="font-size: 16px; color: #2563eb;">👉 <strong>Hành động của ông:</strong> Hãy mở app MB Bank, đối chiếu số tiền và chuyển khoản ngược lại <strong>HOÀN TIỀN TRẢ LẠI</strong> cho người gửi này nhé sếp!</p>
		`, smsContent)

		adminMsg := []byte(adminSubject + adminMime + adminBody)
		_ = smtp.SendMail(smtpHost+":"+smtpPort, auth, senderEmail, []string{senderEmail}, adminMsg)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	http.HandleFunc("/api/v1/order", handleCreateOrder)
	http.HandleFunc("/api/v1/sms-trigger", handlePhoneNotification)
	
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
