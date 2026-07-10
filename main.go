package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	// 1. Route cho trang chủ Tiếng Việt (/)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Ngăn chặn các request đường dẫn lạ (không tồn tại) tự động nhảy vào index.html
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "index.html")
	})

	// 2. Route phục vụ trang Tiếng Anh (/en) - Sửa triệt để lỗi 404
	http.HandleFunc("/en", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "en.html")
	})

	// Lấy PORT từ môi trường của Render (mặc định nếu thiếu là 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("SIA Server đang chạy mượt mà trên port: " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
