package services

import "log"

// NotificationMessage holds the content to be dispatched
type NotificationMessage struct {
	UserID  uint
	To      string // email address or phone number
	Title   string
	Body    string
	Channel string // "email" | "sms" | "inapp"
}

// Notifier is the interface all notification channels must implement
type Notifier interface {
	Send(msg NotificationMessage) error
}

// ---- In-App Notifier (no-op — handled by the DB notifications table) ----

type InAppNotifier struct{}

func (n *InAppNotifier) Send(msg NotificationMessage) error {
	// Notifications are already written to the DB by createNotif() in handlers.
	// This notifier is a no-op placeholder for the dispatcher pipeline.
	return nil
}

// ---- Email Notifier stub ----

type EmailNotifier struct {
	SMTPHost string
	SMTPPort int
	From     string
}

func (n *EmailNotifier) Send(msg NotificationMessage) error {
	// TODO: wire in net/smtp or a provider SDK (SendGrid, Mailgun, etc.)
	// Example with net/smtp:
	//   auth := smtp.PlainAuth("", user, pass, host)
	//   body := "Subject: " + msg.Title + "\r\n\r\n" + msg.Body
	//   smtp.SendMail(host:port, auth, from, []string{msg.To}, []byte(body))
	log.Printf("[EmailNotifier] STUB — to=%s subject=%q (configure SMTP to send real emails)", msg.To, msg.Title)
	return nil
}

// ---- SMS Notifier stub ----

type SMSNotifier struct {
	AccountSID string
	AuthToken  string
	From       string
}

func (n *SMSNotifier) Send(msg NotificationMessage) error {
	// TODO: wire in Twilio or Fast2SMS SDK
	// Example with Twilio REST API:
	//   POST https://api.twilio.com/2010-04-01/Accounts/{SID}/Messages.json
	//   Body: {"To": msg.To, "From": from, "Body": msg.Body}
	log.Printf("[SMSNotifier] STUB — to=%s body=%q (configure Twilio/Fast2SMS to send real SMS)", msg.To, msg.Body)
	return nil
}

// ---- Dispatcher — fan-out to all registered notifiers ----

type NotificationDispatcher struct {
	notifiers []Notifier
}

func NewNotificationDispatcher(notifiers ...Notifier) *NotificationDispatcher {
	return &NotificationDispatcher{notifiers: notifiers}
}

// Dispatch sends the message to all registered channels.
// Errors from individual notifiers are logged but do not stop the others.
func (d *NotificationDispatcher) Dispatch(msg NotificationMessage) {
	for _, n := range d.notifiers {
		if err := n.Send(msg); err != nil {
			log.Printf("[Dispatcher] notifier error: %v", err)
		}
	}
}

// DefaultDispatcher builds a dispatcher with all registered notifiers.
// Wire real credentials via env vars before using Email/SMS in production.
func DefaultDispatcher() *NotificationDispatcher {
	return NewNotificationDispatcher(
		&InAppNotifier{},
		&EmailNotifier{},
		&SMSNotifier{},
	)
}
