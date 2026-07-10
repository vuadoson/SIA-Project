const express = require('express');
const app = express();
const PORT = process.env.PORT || 3000;

// Kho chứa mã bản quyền (Bạn có thể thay đổi các mã này)
const KHO_MA_BAN_QUYEN = ["LICENSE-TRONG-123", "LICENSE-TRONG-789", "LICENSE-TRONG-999"];

app.use(express.json());

// Giao diện hiển thị nút bấm và quét mã QR MB Bank của bạn
app.get('/', (req, res) => {
    res.send(`
    <!DOCTYPE html>
    <html lang="vi">
    <head>
        <meta charset="UTF-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        <title>Thanh Toán Bản Quyền - MB Bank</title>
        <style>
            body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #f4f6f8; display: flex; justify-content: center; align-items: center; min-height: 100vh; margin: 0; }
            .box { background: white; padding: 40px; border-radius: 12px; box-shadow: 0 4px 20px rgba(0,0,0,0.08); text-align: center; max-width: 450px; width: 100%; }
            h2 { color: #1e3a8a; margin-bottom: 5px; }
            .price { font-size: 26px; color: #0056b3; font-weight: bold; margin: 15px 0; }
            .btn { background-color: #0056b3; color: white; border: none; padding: 14px; font-size: 16px; font-weight: bold; border-radius: 6px; cursor: pointer; width: 100%; transition: 0.2s; }
            .btn:hover { background-color: #003d80; }
            .qr-zone { display: none; margin-top: 25px; padding-top: 20px; border-top: 1px dashed #ccc; }
            .qr-image { width: 250px; height: 250px; margin: 10px auto; background: #eee; }
            .info-text { font-size: 14px; color: #555; background: #f8f9fa; padding: 10px; border-radius: 6px; text-align: left; line-height: 1.6; }
        </style>
    </head>
    <body>
        <div class="box">
            <h2>Mua Mã Bản Quyền</h2>
            <p style="color: #666; margin: 5px 0;">Chủ tài khoản: <strong>VŨ VĂN TRỌNG</strong></p>
            <div class="price">500.000 VNĐ</div>
            
            <button id="pay-btn" class="btn">BẤM ĐỂ HIỂN THỊ MÃ QR THANH TOÁN</button>
            
            <!-- Vùng hiển thị QR sau khi khách bấm nút -->
            <div id="qr-zone" class="qr-zone">
                <p style="color: #ffc107; font-weight: bold;">Vui lòng dùng App Ngân hàng quét mã dưới đây:</p>
                <!-- Tự động tạo mã VietQR chuẩn xác đến tài khoản MB của bạn -->
                <img class="qr-image" src="https://img.vietqr.io/image/MB-1005071981-compact2.png?amount=500000&addInfo=BANQUYEN%20TRONG&accountName=VU%20VAN%20TRONG" alt="Mã QR Thanh Toán MB Bank">
                
                <div class="info-text">
                    <strong>Thông tin chuyển khoản dự phòng:</strong><br>
                    • Ngân hàng: **MB Bank (Ngân hàng Quân Đội)**<br>
                    • Số tài khoản VND: **1005071981**<br>
                    • Số tài khoản USD: **6801639330636**<br>
                    • Số tiền: **500.000 VND**<br>
                    • Nội dung ghi chú: <span style="color: red; font-weight: bold;">BANQUYEN TRONG</span>
                </div>
                <p style="font-size: 13px; color: #888; margin-top: 15px;">Sau khi chuyển khoản thành công, hệ thống sẽ xác thực và gửi mã vào màn hình/email của bạn.</p>
            </div>
        </div>

        <script>
            const payBtn = document.getElementById('pay-btn');
            const qrZone = document.getElementById('qr-zone');

            payBtn.addEventListener('click', () => {
                // Khi khách bấm nút, lập tức hiển thị vùng QR không bị đơ giật
                payBtn.style.display = 'none';
                qrZone.style.display = 'block';
            });
        </script>
    </body>
    </html>
    `);
});

app.listen(PORT, () => console.log(`Hệ thống hiển thị thanh toán MB Bank đang chạy trên cổng ${PORT}`));
