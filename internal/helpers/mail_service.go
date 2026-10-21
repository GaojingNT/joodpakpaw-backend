package helpers

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"time"

	"gopkg.in/gomail.v2"
)

type MailService struct {
	SmtpHost     string
	SmtpPort     int
	SmtpUsername string
	SmtpPassword string
}

// NewMailService creates a new instance of MailService
func NewMailService(host string, port int, username, password string) *MailService {
	return &MailService{
		SmtpHost:     host,
		SmtpPort:     port,
		SmtpUsername: username,
		SmtpPassword: password,
	}
}

// GenerateOTP generates a random 6-digit OTP
func (m *MailService) GenerateOTP() string {
	rand.Seed(time.Now().UnixNano())
	otp := rand.Intn(900000) + 100000
	return fmt.Sprintf("%d", otp)
}

// SendBookingCreatedEmail sends a booking confirmation email with a link button
func (m *MailService) SendBookingCreatedEmail(to, recipientName, propertyName, checkIn, checkOut string, numberOfPets int, bookingURL string, isHost bool) error {
	message := gomail.NewMessage()
	message.SetHeader("From", m.SmtpUsername)
	message.SetHeader("To", to)
	message.SetHeader("Subject", "Booking Confirmation - JoodPakPaw")

	intro := fmt.Sprintf("Your booking at <strong>%s</strong> from <strong>%s</strong> to <strong>%s</strong> has been successfully created.", propertyName, checkIn, checkOut)
	btnLabel := "View My Booking"
	if isHost {
		intro = fmt.Sprintf("A new booking has been made at <strong>%s</strong> from <strong>%s</strong> to <strong>%s</strong> with <strong>%d pet(s)</strong>.", propertyName, checkIn, checkOut, numberOfPets)
		btnLabel = "View Booking"
	}

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 40px; border-radius: 10px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
        <h1 style="color: #333333; text-align: center;">JoodPakPaw</h1>
        <p style="color: #666666; font-size: 16px; line-height: 1.5;">Hi <strong>%s</strong>,</p>
        <p style="color: #666666; font-size: 16px; line-height: 1.5;">%s</p>

        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #007bff; color: #ffffff; padding: 14px 28px; border-radius: 6px; text-decoration: none; font-size: 16px; font-weight: bold; display: inline-block;">%s</a>
        </div>

        <hr style="border: none; border-top: 1px solid #eeeeee; margin: 30px 0;">
        <p style="color: #999999; font-size: 12px; text-align: center;">© 2025 JoodPakPaw. All rights reserved.</p>
    </div>
</body>
</html>
`, recipientName, intro, bookingURL, btnLabel)

	message.SetBody("text/html", htmlBody)

	dialer := gomail.NewDialer(m.SmtpHost, m.SmtpPort, m.SmtpUsername, m.SmtpPassword)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := dialer.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

// SendPaymentConfirmedEmail sends a payment confirmation email with a link button
func (m *MailService) SendPaymentConfirmedEmail(to, recipientName, propertyName, checkIn, checkOut string, numberOfPets int, bookingURL string, isHost bool) error {
	message := gomail.NewMessage()
	message.SetHeader("From", m.SmtpUsername)
	message.SetHeader("To", to)
	message.SetHeader("Subject", "Payment Confirmed - JoodPakPaw")

	var intro string
	var btnLabel string
	if isHost {
		intro = fmt.Sprintf("Payment has been received for a booking at <strong>%s</strong> from <strong>%s</strong> to <strong>%s</strong> with <strong>%d pet(s)</strong>.", propertyName, checkIn, checkOut, numberOfPets)
		btnLabel = "View Booking"
	} else {
		intro = fmt.Sprintf("Your payment for the booking at <strong>%s</strong> from <strong>%s</strong> to <strong>%s</strong> has been confirmed.", propertyName, checkIn, checkOut)
		btnLabel = "View My Booking"
	}

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 40px; border-radius: 10px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
        <h1 style="color: #333333; text-align: center;">JoodPakPaw</h1>
        <p style="color: #666666; font-size: 16px; line-height: 1.5;">Hi <strong>%s</strong>,</p>
        <p style="color: #666666; font-size: 16px; line-height: 1.5;">%s</p>

        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #28a745; color: #ffffff; padding: 14px 28px; border-radius: 6px; text-decoration: none; font-size: 16px; font-weight: bold; display: inline-block;">%s</a>
        </div>

        <hr style="border: none; border-top: 1px solid #eeeeee; margin: 30px 0;">
        <p style="color: #999999; font-size: 12px; text-align: center;">© 2025 JoodPakPaw. All rights reserved.</p>
    </div>
</body>
</html>
`, recipientName, intro, bookingURL, btnLabel)

	message.SetBody("text/html", htmlBody)

	dialer := gomail.NewDialer(m.SmtpHost, m.SmtpPort, m.SmtpUsername, m.SmtpPassword)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := dialer.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

// SendReminderEmail sends a check-in reminder email (1 day before check-in) with a link button
func (m *MailService) SendReminderEmail(to, recipientName, propertyName, checkIn string, numberOfPets int, bookingURL string, isHost bool) error {
	message := gomail.NewMessage()
	message.SetHeader("From", m.SmtpUsername)
	message.SetHeader("To", to)
	message.SetHeader("Subject", "Check-in Reminder Tomorrow - JoodPakPaw")

	var intro string
	var btnLabel string
	if isHost {
		intro = fmt.Sprintf("A guest is checking in tomorrow at <strong>%s</strong> on <strong>%s</strong> with <strong>%d pet(s)</strong>. Please prepare for their arrival.", propertyName, checkIn, numberOfPets)
		btnLabel = "View Booking"
	} else {
		intro = fmt.Sprintf("Your pets' stay at <strong>%s</strong> starts tomorrow on <strong>%s</strong>. Get ready for a wonderful stay!", propertyName, checkIn)
		btnLabel = "View My Booking"
	}

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 40px; border-radius: 10px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
        <h1 style="color: #333333; text-align: center;">JoodPakPaw</h1>
        <p style="color: #666666; font-size: 16px; line-height: 1.5;">Hi <strong>%s</strong>,</p>
        <p style="color: #666666; font-size: 16px; line-height: 1.5;">%s</p>

        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #fd7e14; color: #ffffff; padding: 14px 28px; border-radius: 6px; text-decoration: none; font-size: 16px; font-weight: bold; display: inline-block;">%s</a>
        </div>

        <hr style="border: none; border-top: 1px solid #eeeeee; margin: 30px 0;">
        <p style="color: #999999; font-size: 12px; text-align: center;">© 2025 JoodPakPaw. All rights reserved.</p>
    </div>
</body>
</html>
`, recipientName, intro, bookingURL, btnLabel)

	message.SetBody("text/html", htmlBody)

	dialer := gomail.NewDialer(m.SmtpHost, m.SmtpPort, m.SmtpUsername, m.SmtpPassword)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := dialer.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

// SendStayCompletedEmail sends a stay completed email with a link button
func (m *MailService) SendStayCompletedEmail(to, recipientName, propertyName string, bookingURL string, isHost bool) error {
	message := gomail.NewMessage()
	message.SetHeader("From", m.SmtpUsername)
	message.SetHeader("To", to)
	message.SetHeader("Subject", "Stay Completed - JoodPakPaw")

	var intro string
	var btnLabel string
	if isHost {
		intro = fmt.Sprintf("Check-out at <strong>%s</strong> has been completed. Thank you for hosting with JoodPakPaw!", propertyName)
		btnLabel = "View Booking"
	} else {
		intro = fmt.Sprintf("Your pets' stay at <strong>%s</strong> is complete! We hope you had a wonderful experience. We'd love to hear your feedback.", propertyName)
		btnLabel = "Leave a Review"
	}

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 40px; border-radius: 10px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
        <h1 style="color: #333333; text-align: center;">JoodPakPaw</h1>
        <p style="color: #666666; font-size: 16px; line-height: 1.5;">Hi <strong>%s</strong>,</p>
        <p style="color: #666666; font-size: 16px; line-height: 1.5;">%s</p>

        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #6f42c1; color: #ffffff; padding: 14px 28px; border-radius: 6px; text-decoration: none; font-size: 16px; font-weight: bold; display: inline-block;">%s</a>
        </div>

        <hr style="border: none; border-top: 1px solid #eeeeee; margin: 30px 0;">
        <p style="color: #999999; font-size: 12px; text-align: center;">© 2025 JoodPakPaw. All rights reserved.</p>
    </div>
</body>
</html>
`, recipientName, intro, bookingURL, btnLabel)

	message.SetBody("text/html", htmlBody)

	dialer := gomail.NewDialer(m.SmtpHost, m.SmtpPort, m.SmtpUsername, m.SmtpPassword)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := dialer.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

// SendBookingCancelledEmail sends a booking cancellation email with a link button
func (m *MailService) SendBookingCancelledEmail(to, recipientName, propertyName, checkIn, checkOut string, numberOfPets int, bookingURL string, isHost bool) error {
	message := gomail.NewMessage()
	message.SetHeader("From", m.SmtpUsername)
	message.SetHeader("To", to)
	message.SetHeader("Subject", "Booking Cancelled - JoodPakPaw")

	var intro string
	var btnLabel string
	if isHost {
		intro = fmt.Sprintf("A booking at <strong>%s</strong> from <strong>%s</strong> to <strong>%s</strong> with <strong>%d pet(s)</strong> has been cancelled.", propertyName, checkIn, checkOut, numberOfPets)
		btnLabel = "View Booking"
	} else {
		intro = fmt.Sprintf("Your booking at <strong>%s</strong> from <strong>%s</strong> to <strong>%s</strong> has been cancelled.", propertyName, checkIn, checkOut)
		btnLabel = "View My Booking"
	}

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 40px; border-radius: 10px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
        <h1 style="color: #333333; text-align: center;">JoodPakPaw</h1>
        <p style="color: #666666; font-size: 16px; line-height: 1.5;">Hi <strong>%s</strong>,</p>
        <p style="color: #666666; font-size: 16px; line-height: 1.5;">%s</p>

        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #dc3545; color: #ffffff; padding: 14px 28px; border-radius: 6px; text-decoration: none; font-size: 16px; font-weight: bold; display: inline-block;">%s</a>
        </div>

        <hr style="border: none; border-top: 1px solid #eeeeee; margin: 30px 0;">
        <p style="color: #999999; font-size: 12px; text-align: center;">© 2025 JoodPakPaw. All rights reserved.</p>
    </div>
</body>
</html>
`, recipientName, intro, bookingURL, btnLabel)

	message.SetBody("text/html", htmlBody)

	dialer := gomail.NewDialer(m.SmtpHost, m.SmtpPort, m.SmtpUsername, m.SmtpPassword)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := dialer.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

// SendOTP sends OTP to the specified email address
func (m *MailService) SendOTP(to, otp string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", m.SmtpUsername)
	message.SetHeader("To", to)
	message.SetHeader("Subject", "Your OTP Code - JoodPakPaw")

	// Create HTML body
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
    <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; padding: 40px; border-radius: 10px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
        <h1 style="color: #333333; text-align: center;">Welcome to JoodPakPaw!</h1>
        <p style="color: #666666; font-size: 16px; line-height: 1.5;">Thank you for registering with us. To complete your registration, please use the following OTP code:</p>

        <div style="background-color: #f8f9fa; padding: 20px; border-radius: 5px; text-align: center; margin: 30px 0;">
            <h2 style="color: #007bff; font-size: 36px; margin: 0; letter-spacing: 5px;">%s</h2>
        </div>

        <p style="color: #666666; font-size: 14px; line-height: 1.5;">This OTP will expire in <strong>10 minutes</strong>.</p>
        <p style="color: #666666; font-size: 14px; line-height: 1.5;">If you didn't request this code, please ignore this email.</p>

        <hr style="border: none; border-top: 1px solid #eeeeee; margin: 30px 0;">

        <p style="color: #999999; font-size: 12px; text-align: center;">© 2025 JoodPakPaw. All rights reserved.</p>
    </div>
</body>
</html>
`, otp)

	message.SetBody("text/html", htmlBody)

	// Create dialer
	dialer := gomail.NewDialer(m.SmtpHost, m.SmtpPort, m.SmtpUsername, m.SmtpPassword)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Send email
	if err := dialer.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
