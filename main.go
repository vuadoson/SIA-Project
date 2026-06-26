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
	// 🚨 ÔNG CHỈ CẦN SỬA ĐÚNG DÒNG NÀY: Thay bằng 16 ký tự mật khẩu ứng dụng Google của ông
	senderPass  = "woug ejkp ndmr ttri" 
)

// Cấu trúc dữ liệu đơn hàng của khách hàng
type ClientOrder struct {
	ClientName    string    `json:"client_name"`
	ClientPhone   string    `json:"client_phone"`
	ClientAddress string    `json:"client_address"` // Email của khách nhận hàng
	ActionType    string    `json:"action_type"`
	Amount        float64   `json:"amount"`
	Timestamp     time.Time `json:"timestamp"`
}

// Bộ nhớ đệm chạy tạm trên RAM để lưu danh sách đơn hàng đang chờ tiền
var orderCache = make(map[string]ClientOrder)

// Hàm tự động "đẻ" ra mã API Key bảo mật ngẫu nhiên không trùng lặp
func generateAPIKey() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "SIA_KEY_" + hex.EncodeToString(bytes)
}

// 1. API TIẾP NHẬN ĐƠN HÀNG KHI KHÁCH BẤM NÚT TRÊN WEB
func handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	// Cấu hình CORS để giao diện Web gọi được vào API Render công khai
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
	
	// Lưu đơn hàng vào hàng đợi, lấy Số điện thoại khách làm chìa khóa định danh
	orderCache[order.ClientPhone] = order
	
	log.Printf("📝 Đã tạo đơn hàng chờ: Khách %s - SĐT: %s", order.ClientName, order.ClientPhone)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "pending_payment"})
}

// 2. API ĐÓN THÔNG BÁO TIỀN VỀ TỪ APP MACRODROID TRÊN ĐIỆN THOẠI ÔNG
func handlePhoneNotification(w http.ResponseWriter, r *http.Request) {
	smsContent := r.URL.Query().Get("sms_content")
	if smsContent == "" {
		http.Error(w, "Content is empty", http.StatusBadRequest)
		return
	}

	log.Printf("📩 Hệ thống nhận thông báo giao dịch từ MB Bank: %s", smsContent)

	isMatched := false
	auth := smtp.PlainAuth("", senderEmail, senderPass, smtpHost)

	// Vòng lặp quét xem nội dung tin nhắn ngân hàng có chứa SĐT của khách nào không
	for phone, client := range orderCache {
		if strings.Contains(smsContent, phone) {
			isMatched = true
			
			// Máy tự sinh mã API Key bảo mật cho khách
			apiKey := generateAPIKey()

			// Soạn email tự động gửi hàng cho khách chuẩn chỉ
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
			
			// Tiến hành gửi email tự động cho khách
			err := smtp.SendMail(smtpHost+":"+smtpPort, auth, senderEmail, []string{client.ClientAddress}, msg)
			if err == nil {
				delete(orderCache, phone) // Gửi xong xóa khách khỏi hàng đợi
				log.Printf("✅ [Thành công] Đã tự động gửi API Key đến khách: %s", client.ClientAddress)
			} else {
				log.Printf("❌ Lỗi gửi mail cho khách: %v", err)
			}
			break
		}
	}

	// 🚨 NẾU KHÁCH GÕ SAI NỘI DUNG (Quét hết hàng đợi mà không tìm thấy SĐT nào khớp)
	if !isMatched {
		log.Printf("⚠️ CẢNH BÁO: Giao dịch không hợp lệ hoặc sai nội dung: %s", smsContent)
		
		// Hệ thống tự soạn email báo động gửi thẳng về hộp thư riêng vuadoson@gmail.com của ông
		adminSubject := "Subject: [🚨 SIA ALERT] CO GIAO DICH SAI NOI DUNG - CAN HOAN TIEN\n"
		adminMime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
		adminBody := fmt.Sprintf(`
			<h2 style="color: #ef4444; font-family: sans-serif;">🚨 CẢNH BÁO HỆ THỐNG SIA</h2>
			<p style="font-size: 15px;">Hệ thống vừa nhận được thông báo tiền vào tài khoản MB Bank nhưng <strong>KHÔNG KHỚP</strong> với số điện thoại của bất kỳ đơn hàng nào đang chờ.</p>
			<p style="font-size: 15px;">Khách hàng này đã cố tình hoặc vô tình gõ sai nội dung chuyển khoản mặc định.</p>
			<hr/>
			<div style="background: #f8fafc; padding: 15px; border: 1px solid #cbd5e1; border-radius: 6px;">
				<strong>Chi tiết thông báo ngân hàng nhận được từ MacroDroid:</strong><br/>
				<p style="font-family: monospace; color: #1e293b; background: #e2e8f0; padding: 10px; border-radius: 4px;">%s</p>
			</div>
			<br/>
			<p style="font-size: 16px; color: #2563eb;">👉 <strong>Hành động của ông:</strong> Ông hãy mở ngay app MB Bank, đối chiếu số tiền trong lịch sử giao dịch và tiến hành <strong>HOÀN TIỀN TRẢ LẠI</strong> cho người gửi này nhé sếp!</p>
		`, smsContent)

		adminMsg := []byte(adminSubject + adminMime + adminBody)
		
		// Bắn mail báo động về chính mình để mở app ra bank tiền trả lại khách
		err := smtp.SendMail(smtpHost+":"+smtpPort, auth, senderEmail, []string{senderEmail}, adminMsg)
		if err == nil {
			log.Printf("🚨 Đã gửi email cảnh báo hoàn tiền về hộp thư của Admin!")
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	// Khởi tạo các tuyến đường API kết nối
	http.HandleFunc("/api/v1/order", handleCreateOrder)
	http.HandleFunc("/api/v1/sms-trigger", handlePhoneNotification)
	
	// Cấu hình cổng chạy mặc định cho môi trường Render
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("🚀 Máy chủ SIA Backend đang chạy ổn định tại cổng: %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
