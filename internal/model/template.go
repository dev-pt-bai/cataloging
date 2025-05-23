package model

import "time"

const (
	emailVerification = `
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>Verification Code</title>
  <style>
    body {
      font-family: Arial, sans-serif;
      background-color: #f4f4f7;
      padding: 0;
      margin: 0;
    }
    .email-container {
      max-width: 600px;
      margin: 30px auto;
      background-color: #ffffff;
      padding: 30px;
      border-radius: 8px;
      box-shadow: 0 2px 5px rgba(0,0,0,0.1);
    }
    h2 {
      color: #333333;
    }
    p {
      font-size: 16px;
      color: #555555;
    }
    .code-box {
      background-color: #f0f0f0;
      padding: 15px;
      text-align: center;
      font-size: 24px;
      letter-spacing: 4px;
      border-radius: 6px;
      margin: 20px 0;
      font-weight: bold;
      color: #2d3748;
    }
    .footer {
      text-align: center;
      font-size: 12px;
      color: #999999;
      margin-top: 20px;
    }
  </style>
</head>
<body>
  <div class="email-container">
    <h2>Verifikasi email Anda</h2>
    <p>Halo, %s</p>
    <p>Terima kasih telah menggunakan aplikasi Cataloging. Silakan gunakan kode One-Time-Password (OTP) berikut untuk memverifikasi email Anda:</p>
    <div class="code-box">%s</div>
    <p>Kode ini hanya berlaku sampai %v WIB. Jika Anda merasa tidak sedang melakukan pendaftaran akun atau verifikasi email pada aplikasi Cataloging, abaikan saja email ini.</p>
    <p>Salam,<br>Aplikasi Cataloging</p>
    <div class="footer">&copy; 2025 PT Borneo Alumina Indonesia. Seluruh hak cipta dilindungi undang-undang.</div>
  </div>
</body>
</html>`
	emailWelcome = `
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Account Verified</title>
  <style>
    body {
      font-family: Arial, sans-serif;
      background-color: #f4f4f7;
      padding: 0;
      margin: 0;
    }
    .email-container {
      max-width: 600px;
      margin: 30px auto;
      background-color: #ffffff;
      padding: 30px;
      border-radius: 8px;
      box-shadow: 0 2px 5px rgba(0,0,0,0.1);
    }
    h2 {
      color: #333333;
    }
    p {
      font-size: 16px;
      color: #555555;
    }
    .cta-button {
      display: inline-block;
      background-color: #2b6cb0;
      color: #ffffff;
      text-decoration: none;
      padding: 12px 20px;
      border-radius: 5px;
      font-weight: bold;
    }
    .footer {
      text-align: center;
      font-size: 12px;
      color: #999999;
      margin-top: 20px;
    }
  </style>
</head>
<body>
  <div class="email-container">
    <h2>Selamat bergabung, akun Anda telah terverifikasi!</h2>
    <p>Halo, %s</p>
    <p>Terima kasih telah memverifikasi email Anda. Kini, akun Anda telah berhasil teraktivasi dan Anda dapat mulai mengajukan permohonan katalog melalui aplikasi ini.</p>
    <p>Klik untuk mengeksplorasi lebih lanjut:</p>
    <a href="%s" class="cta-button">Ajukan katalog</a>
    <p>Jika Anda merasa tidak membuat akun ini, silakan kontak tim pendukung (IT).</p>
    <p>Salam,<br>Aplikasi Cataloging</p>
    <div class="footer">&copy; 2025 PT Borneo Alumina Indonesia. Seluruh hak cipta dilindungi undang-undang.</div>
  </div>
</body>
</html>`
)

var indonesianMonth = map[time.Month]string{
	time.January:   "Januari",
	time.February:  "Februari",
	time.March:     "Maret",
	time.April:     "April",
	time.May:       "Mei",
	time.June:      "Juni",
	time.July:      "Juli",
	time.August:    "Agustus",
	time.September: "September",
	time.October:   "Oktober",
	time.November:  "November",
	time.December:  "Desember",
}
