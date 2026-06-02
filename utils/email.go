package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendOTPEmail(toEmail string, otp string) error {
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	address := fmt.Sprintf("%s:%s", host, port)

	subject := "Subject: Your LendoGo Verification Code\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"

	body := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<style>
  body { margin:0; padding:0; background:#EBF4FF; font-family:'Plus Jakarta Sans',Arial,sans-serif; }
  .wrap { background:#EBF4FF; padding:2rem 1rem; }
  .card { max-width:480px; margin:0 auto; background:#fff; border-radius:16px; overflow:hidden; border:1px solid #B5D4F4; }
  .hdr { background:#0C447C; padding:1.75rem 2rem 1.5rem; text-align:center; }
  .logo-row { display:flex; align-items:center; justify-content:center; gap:10px; margin-bottom:6px; }
  .logo-mark { width:36px; height:36px; background:#378ADD; border-radius:10px; display:inline-flex; align-items:center; justify-content:center; }
  .logo-name { font-size:20px; font-weight:600; color:#fff; }
  .hdr-sub { color:#85B7EB; font-size:11px; letter-spacing:1.5px; text-transform:uppercase; margin:6px 0 0; }
  .trust-row { display:flex; justify-content:center; gap:8px; margin-top:10px; }
  .trust-badge { display:flex; align-items:center; gap:4px; padding:3px 10px; background:#0C447C; border:1px solid #185FA5; border-radius:20px; font-size:10px; color:#B5D4F4; }
  .body { padding:2rem; }
  .greeting { font-size:15px; color:#042C53; font-weight:600; margin:0 0 6px; }
  .intro { font-size:13.5px; color:#185FA5; line-height:1.7; margin:0 0 1.5rem; }
  .otp-box { background:#E6F1FB; border-radius:14px; padding:1.75rem; text-align:center; margin-bottom:1.5rem; border:1.5px dashed #85B7EB; }
  .otp-label { font-size:11px; color:#378ADD; letter-spacing:1.2px; text-transform:uppercase; margin-bottom:1rem; font-weight:600; }
  .otp-code { font-family:'Courier New',monospace; font-size:36px; font-weight:700; letter-spacing:12px; color:#0C447C; margin-bottom:8px; padding-left:12px; }
  .otp-timer { font-size:12px; color:#993C1D; }
  .sec-box { background:#E6F1FB; border-left:3px solid #378ADD; border-radius:0 10px 10px 0; padding:1rem 1.25rem; margin-bottom:1.5rem; }
  .sec-title { font-size:12px; font-weight:600; color:#0C447C; margin:0 0 8px; }
  .sec-list { list-style:none; padding:0; margin:0; }
  .sec-list li { font-size:12px; color:#185FA5; padding:3px 0; padding-left:14px; position:relative; line-height:1.5; }
  .sec-list li::before { content:"›"; position:absolute; left:0; color:#378ADD; }
  .divider { height:1px; background:#B5D4F4; margin:1.25rem 0; }
  .note { font-size:12px; color:#185FA5; text-align:center; margin:0; }
  .note a { color:#0C447C; font-weight:600; }
  .badges { display:flex; justify-content:center; gap:8px; margin-top:1rem; flex-wrap:wrap; }
  .badge { padding:5px 12px; background:#E6F1FB; border:1px solid #B5D4F4; border-radius:20px; font-size:11px; color:#0C447C; font-weight:600; }
  .ftr { background:#E6F1FB; padding:1.25rem 2rem; border-top:1px solid #B5D4F4; }
  .ftr-links { display:flex; justify-content:center; gap:1.5rem; margin-bottom:8px; }
  .ftr-links a { font-size:12px; color:#185FA5; text-decoration:none; }
  .ftr-copy { text-align:center; font-size:11px; color:#378ADD; margin:0; }
</style>
</head>
<body>
<div class="wrap">
 <div class="card">

  <div class="hdr">
   <div class="logo-row">
    <div class="logo-mark">
     <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
      <path d="M4 10.5L8 14.5L16 6" stroke="white" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"/>
     </svg>
    </div>
    <span class="logo-name">LendoGo</span>
   </div>
   <p class="hdr-sub">Secure Identity Verification</p>
   <div class="trust-row">
    <span class="trust-badge">&#x1F512; Bank-grade security</span>
    <span class="trust-badge">&#x1F510; 256-bit TLS</span>
   </div>
  </div>

  <div class="body">
   <p class="greeting">Verify your identity</p>
   <p class="intro">Use the one-time code below to complete your sign-in. This code is valid for 5 minutes and can only be used once.</p>

   <div class="otp-box">
    <div class="otp-label">One-time passcode</div>
    <div class="otp-code">%s</div>
    <div class="otp-timer">&#x23F1; Expires in 5 minutes</div>
   </div>

   <div class="sec-box">
    <p class="sec-title">&#x1F6E1; Security notice</p>
    <ul class="sec-list">
     <li>LendoGo will never call or ask for this code</li>
     <li>Do not share this with anyone, including support staff</li>
     <li>This code becomes invalid after a single use</li>
     <li>If you didn't request this, secure your account now</li>
    </ul>
   </div>

   <div class="divider"></div>

   <p class="note">Didn't request this code? <a href="#">Report suspicious activity &rarr;</a></p>

   <div class="badges">
    <span class="badge">Encrypted</span>
    <span class="badge">PCI DSS</span>
    <span class="badge">Never stored</span>
    <span class="badge">RBI regulated</span>
   </div>
  </div>

  <div class="ftr">
   <div class="ftr-links">
    <a href="#">Privacy policy</a>
    <a href="#">Help centre</a>
    <a href="#">Unsubscribe</a>
   </div>
   <p class="ftr-copy">&copy; 2026 LendoGo Financial Technologies Pvt. Ltd. &middot; All rights reserved</p>
  </div>

 </div>
</div>
</body>
</html>`, otp)

	message := []byte(subject + mime + body)

	auth := smtp.PlainAuth("", from, password, host)
	err := smtp.SendMail(address, auth, from, []string{toEmail}, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}