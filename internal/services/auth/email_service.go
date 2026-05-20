package auth

import (
	"fmt"
	"net/smtp"
	"os"
)

// SendOTPEmail physically connects to Gmail's servers to mail out the 6-digit code
func SendOTPEmail(toEmail string, otp string) error {
	// 1. Pull SMTP credentials securely from your .env configuration
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD") // Your 16-character Google App Password
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	// 2. Draft the raw cryptographic email headers and body message string
	subject := "Subject: LendoGo Account Verification Code\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; color: #333;">
			<h2>Welcome to LendoGo!</h2>
			<p>Thank you for registering. Please use the secure 6-digit verification code below to activate your account:</p>
			<h1 style="color: #4F46E5; letter-spacing: 2px;">%s</h1>
			<p>This code will expire in <strong>5 minutes</strong>.</p>
			<p>If you did not request this code, please ignore this email.</p>
		</body>
		</html>
	`, otp)

	msg := []byte(subject + mime + body)

	// 3. Set up the automated Plain Authentication handshake for the Google servers
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// 4. Fire the email out into the internet mesh network!
	addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
	err := smtp.SendMail(addr, auth, from, []string{toEmail}, msg)
	if err != nil {
		return fmt.Errorf("SMTP gateway dropped connection: %w", err)
	}

	return nil
}