package email

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

// Config holds SMTP connection settings loaded from environment variables.
// For local Postfix: Host=localhost, Port=25, User/Pass empty.
// For Gmail: Host=smtp.gmail.com, Port=587, User=your@gmail.com, Pass=app-password.
type Config struct {
	Host     string
	Port     string
	User     string
	Pass     string
	From     string
	FromName string
	Enabled  bool
}

// EmailData contains all the variables needed to render any email template.
type EmailData struct {
	To           string
	ToName       string
	Subject      string
	BookingID    uint
	Status       string
	GeneratorName string
	VendorName   string
	CustomerName string
	Amount       float64
	StartDate    string
	EndDate      string
	OTP          string
	Message      string // extra contextual info
}

// Send sends an HTML email. Non-blocking — runs in a goroutine.
// A failed email never causes the request to fail.
func Send(cfg Config, data EmailData, htmlBody string) {
	if !cfg.Enabled || data.To == "" {
		return
	}
	go func() {
		if err := sendEmail(cfg, data.To, data.ToName, data.Subject, htmlBody); err != nil {
			log.Printf("[Email] Failed to send to %s: %v", data.To, err)
		} else {
			log.Printf("[Email] Sent '%s' to %s", data.Subject, data.To)
		}
	}()
}

func sendEmail(cfg Config, to, toName, subject, htmlBody string) error {
	addr := cfg.Host + ":" + cfg.Port

	from := fmt.Sprintf("%s <%s>", cfg.FromName, cfg.From)
	recipient := fmt.Sprintf("%s <%s>", toName, to)

	var msg bytes.Buffer
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", recipient))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)

	var auth smtp.Auth
	if cfg.User != "" && cfg.Pass != "" {
		auth = smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)
	}

	return smtp.SendMail(addr, auth, cfg.From, []string{to}, msg.Bytes())
}

// ---- Template Builders ----

func BookingRequested(data EmailData) string {
	return renderTemplate("New Booking Request", data, `
	<p>Hello <strong>`+data.VendorName+`</strong>,</p>
	<p>You have a new booking request on <strong>GenRent</strong>.</p>
	`+bookingBox(data)+`
	<div style="text-align:center;margin:2rem 0;">
		<a href="http://localhost:8080/vendor-dashboard" `+ctaStyle("primary")+`>View &amp; Accept Booking</a>
	</div>
	<p style="color:#6b7280;font-size:0.85rem;">You have <strong>2 hours</strong> to accept or reject. After that the booking will be auto-cancelled.</p>
	`)
}

func BookingAccepted(data EmailData) string {
	return renderTemplate("Booking Accepted!", data, `
	<p>Hello <strong>`+data.CustomerName+`</strong>,</p>
	<p>Great news! Your booking has been <strong style="color:#16a34a;">accepted</strong> by the vendor.</p>
	`+bookingBox(data)+`
	<div style="background:#f0fdf4;border:1px solid #bbf7d0;border-radius:8px;padding:1rem;margin:1.5rem 0;">
		<p style="margin:0;color:#15803d;"><strong>Next step:</strong> Complete the advance payment to secure your booking.</p>
	</div>
	<div style="text-align:center;margin:2rem 0;">
		<a href="http://localhost:8080/my-bookings" `+ctaStyle("primary")+`>Pay Advance &amp; Confirm</a>
	</div>
	`)
}

func BookingRejected(data EmailData) string {
	return renderTemplate("Booking Update", data, `
	<p>Hello <strong>`+data.CustomerName+`</strong>,</p>
	<p>Unfortunately, the vendor was unable to accept your booking at this time.</p>
	`+bookingBox(data)+`
	<div style="background:#fef2f2;border:1px solid #fecaca;border-radius:8px;padding:1rem;margin:1.5rem 0;">
		<p style="margin:0;color:#dc2626;"><strong>Booking Rejected</strong>`+reason(data.Message)+`</p>
	</div>
	<p>No payment has been charged. You can search for other available generators.</p>
	<div style="text-align:center;margin:2rem 0;">
		<a href="http://localhost:8080" `+ctaStyle("secondary")+`>Find Another Generator</a>
	</div>
	`)
}

func PaymentReceived(data EmailData) string {
	return renderTemplate("Advance Payment Received", data, `
	<p>Hello <strong>`+data.VendorName+`</strong>,</p>
	<p>The customer has completed the advance payment. Please prepare the generator for dispatch.</p>
	`+bookingBox(data)+`
	<div style="background:#f0fdf4;border:1px solid #bbf7d0;border-radius:8px;padding:1rem;margin:1.5rem 0;">
		<p style="margin:0;color:#15803d;"><strong>Amount held in escrow:</strong> ₹`+fmt.Sprintf("%.0f", data.Amount)+`</p>
		<p style="margin:0.5rem 0 0;color:#6b7280;font-size:0.85rem;">This will be released to your wallet after the customer confirms delivery.</p>
	</div>
	<div style="text-align:center;margin:2rem 0;">
		<a href="http://localhost:8080/vendor-dashboard" `+ctaStyle("primary")+`>Go to Dashboard &amp; Dispatch</a>
	</div>
	`)
}

func GeneratorDispatched(data EmailData) string {
	otpSection := ""
	if data.OTP != "" {
		otpSection = `
		<div style="background:#fffbeb;border:2px solid #f59e0b;border-radius:12px;padding:1.5rem;margin:1.5rem 0;text-align:center;">
			<p style="margin:0 0 0.5rem;color:#92400e;font-weight:600;font-size:0.9rem;">DELIVERY CONFIRMATION OTP</p>
			<div style="font-size:2.5rem;font-weight:800;letter-spacing:0.5rem;color:#d97706;">` + data.OTP + `</div>
			<p style="margin:0.5rem 0 0;color:#92400e;font-size:0.8rem;">Share this OTP only when the generator is delivered to you.</p>
		</div>`
	}
	return renderTemplate("Your Generator is On the Way!", data, `
	<p>Hello <strong>`+data.CustomerName+`</strong>,</p>
	<p>Your generator has been dispatched and is on its way to you!</p>
	`+bookingBox(data)+`
	`+otpSection+`
	<div style="text-align:center;margin:2rem 0;">
		<a href="http://localhost:8080/my-bookings" `+ctaStyle("primary")+`>Track Your Booking</a>
	</div>
	`)
}

func DeliveryConfirmed(data EmailData) string {
	return renderTemplate("Delivery Confirmed — Payment Released", data, `
	<p>Hello <strong>`+data.VendorName+`</strong>,</p>
	<p>The customer has confirmed delivery. Your payment has been <strong style="color:#16a34a;">released from escrow to your wallet</strong>.</p>
	`+bookingBox(data)+`
	<div style="background:#f0fdf4;border:1px solid #bbf7d0;border-radius:8px;padding:1rem;margin:1.5rem 0;">
		<p style="margin:0;color:#15803d;"><strong>Amount credited:</strong> ₹`+fmt.Sprintf("%.0f", data.Amount)+`</p>
	</div>
	<div style="text-align:center;margin:2rem 0;">
		<a href="http://localhost:8080/vendor-dashboard" `+ctaStyle("primary")+`>View Wallet</a>
	</div>
	`)
}

func BookingCancelled(data EmailData) string {
	return renderTemplate("Booking Cancelled", data, `
	<p>Hello <strong>`+data.CustomerName+`</strong>,</p>
	<p>Your booking has been cancelled.</p>
	`+bookingBox(data)+`
	<div style="background:#fef2f2;border:1px solid #fecaca;border-radius:8px;padding:1rem;margin:1.5rem 0;">
		<p style="margin:0;color:#dc2626;"><strong>Cancellation reason:</strong> `+ifEmpty(data.Message, "No reason provided")+`</p>
	</div>
	<p>If an advance was paid, it will be refunded within 3–5 business days.</p>
	<div style="text-align:center;margin:2rem 0;">
		<a href="http://localhost:8080" `+ctaStyle("secondary")+`>Browse Generators</a>
	</div>
	`)
}

func BookingCompleted(data EmailData) string {
	return renderTemplate("Booking Completed — Thank You!", data, `
	<p>Hello <strong>`+data.CustomerName+`</strong>,</p>
	<p>Your rental has been completed. We hope everything went smoothly!</p>
	`+bookingBox(data)+`
	<div style="text-align:center;margin:2rem 0;">
		<a href="http://localhost:8080/my-bookings" `+ctaStyle("primary")+`>Rate Your Experience</a>
	</div>
	<p style="color:#6b7280;font-size:0.85rem;text-align:center;">Your feedback helps other customers make better decisions.</p>
	`)
}

func WithdrawalOTPEmail(data EmailData) string {
	return renderTemplate("Confirm Your Withdrawal — OTP", data, `
	<p>Hello <strong>`+data.VendorName+`</strong>,</p>
	<p>You requested a withdrawal of <strong>₹`+fmt.Sprintf("%.0f", data.Amount)+`</strong> from your GenRent wallet.</p>
	<div style="background:#fffbeb;border:2px solid #f59e0b;border-radius:12px;padding:1.5rem;margin:1.5rem 0;text-align:center;">
		<p style="margin:0 0 0.5rem;color:#92400e;font-weight:600;font-size:0.9rem;">YOUR WITHDRAWAL OTP</p>
		<div style="font-size:2.5rem;font-weight:800;letter-spacing:0.5rem;color:#d97706;">`+data.OTP+`</div>
		<p style="margin:0.5rem 0 0;color:#92400e;font-size:0.8rem;">This OTP expires in <strong>10 minutes</strong>. Do not share it with anyone.</p>
	</div>
	<p style="color:#6b7280;font-size:0.85rem;">If you did not request this withdrawal, please ignore this email and your funds are safe.</p>
	`)
}

// ---- Shared helpers ----

func bookingBox(data EmailData) string {
	rows := []string{}
	if data.BookingID > 0 {
		rows = append(rows, row("Booking ID", fmt.Sprintf("#%d", data.BookingID)))
	}
	if data.GeneratorName != "" {
		rows = append(rows, row("Generator", data.GeneratorName))
	}
	if data.VendorName != "" && data.CustomerName != "" {
		rows = append(rows, row("Vendor", data.VendorName))
	}
	if data.StartDate != "" {
		rows = append(rows, row("Rental Period", data.StartDate+" → "+data.EndDate))
	}
	if data.Amount > 0 {
		rows = append(rows, row("Amount", fmt.Sprintf("₹%.0f", data.Amount)))
	}
	return `<div style="background:#f9fafb;border:1px solid #e5e7eb;border-radius:8px;padding:1rem;margin:1.5rem 0;">` +
		strings.Join(rows, "") + `</div>`
}

func row(label, value string) string {
	return fmt.Sprintf(`<div style="display:flex;justify-content:space-between;padding:0.4rem 0;border-bottom:1px solid #e5e7eb;">
		<span style="color:#6b7280;font-size:0.875rem;">%s</span>
		<span style="font-weight:600;font-size:0.875rem;">%s</span>
	</div>`, label, value)
}

func ctaStyle(kind string) string {
	bg := "#2563eb"
	if kind == "secondary" {
		bg = "#6b7280"
	}
	return fmt.Sprintf(`style="display:inline-block;background:%s;color:white;padding:0.75rem 2rem;border-radius:8px;text-decoration:none;font-weight:600;font-size:1rem;"`, bg)
}

func reason(r string) string {
	if r == "" {
		return ""
	}
	return ": " + r
}

func ifEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

func renderTemplate(title string, data EmailData, content string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>` + title + `</title>
</head>
<body style="margin:0;padding:0;background:#f3f4f6;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;">
<table width="100%" cellpadding="0" cellspacing="0" style="background:#f3f4f6;padding:2rem 1rem;">
  <tr><td align="center">
    <table width="600" cellpadding="0" cellspacing="0" style="max-width:600px;width:100%;">

      <!-- Header -->
      <tr><td style="background:#1d4ed8;border-radius:12px 12px 0 0;padding:1.5rem 2rem;text-align:center;">
        <div style="font-size:1.5rem;font-weight:800;color:white;letter-spacing:-0.5px;">⚡ GenRent</div>
        <div style="color:#bfdbfe;font-size:0.85rem;margin-top:0.25rem;">Generator Rental Platform</div>
      </td></tr>

      <!-- Body -->
      <tr><td style="background:white;padding:2rem;border-left:1px solid #e5e7eb;border-right:1px solid #e5e7eb;">
        ` + content + `
      </td></tr>

      <!-- Footer -->
      <tr><td style="background:#f9fafb;border:1px solid #e5e7eb;border-radius:0 0 12px 12px;padding:1.25rem 2rem;text-align:center;">
        <p style="margin:0;color:#9ca3af;font-size:0.8rem;">© 2026 GenRent · This is an automated notification, please do not reply.</p>
        <p style="margin:0.5rem 0 0;color:#9ca3af;font-size:0.8rem;">
          <a href="http://localhost:8080" style="color:#6b7280;">Visit GenRent</a>
        </p>
      </td></tr>

    </table>
  </td></tr>
</table>
</body>
</html>`
}
