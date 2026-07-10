package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
	"os"
)

//go:embed index.html en.html
var templateFiles embed.FS

func main() {
	// Khởi tạo và parse các template đã nhúng sẵn
	tmpl, err := template.ParseFS(templateFiles, "index.html", "en.html")
	if err != nil {
		log.Fatalf("Lỗi nạp file HTML: %v", err)
	}

	// Route bản tiếng Việt (Trang chủ)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Ngăn chặn các request sai đường dẫn tĩnh rơi vào đây gây lỗi 404
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(w, "index.html", nil)
	})

	// Route bản tiếng Anh
	http.HandleFunc("/en", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl.ExecuteTemplate(w, "en.html", nil)
	})

	// Lấy cổng Port từ hệ thống Render (mặc định là 8080 nếu chạy local)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server SIA đang chạy mượt mà tại cổng :%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Không thể khởi động server: %v", err)
	}
}
