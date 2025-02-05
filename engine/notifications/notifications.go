package notifications

import (
	"ecommerce-service/models"
	"fmt"
	"log"
	"os"

	"gopkg.in/mail.v2"
)

type AfricasTalkingGateway struct {
	Username string
	APIKey   string
	Sender   string
}

var gateway = &AfricasTalkingGateway{
	Username: os.Getenv("AT_USERNAME"),
	APIKey:   os.Getenv("AT_API_KEY"),
	Sender:   "Store",
}

func SendOrderConfirmationSMS(order *models.Order) error {
	message := fmt.Sprintf(
		"Thank you for your order #%s. Total: $%.2f. We'll process it right away!",
		order.ID.String()[:8],
		order.Total,
	)

	// In production, use the actual Africa's Talking API
	// For now, we'll just log the message
	log.Printf("SMS to %s: %s", order.Customer.PhoneNumber, message)

	return nil
}

func SendOrderNotificationEmail(order *models.Order) error {
	// Configure SMTP settings
	smtpHost := os.Getenv("SMTP_HOST")
	// smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	adminEmail := os.Getenv("ADMIN_EMAIL")

	// Create email message
	m := mail.NewMessage()
	m.SetHeader("From", smtpUser)
	m.SetHeader("To", adminEmail)
	m.SetHeader("Subject", fmt.Sprintf("New Order #%s", order.ID.String()[:8]))

	// Build email body
	body := fmt.Sprintf(`
New order received:

Order ID: %s
Customer: %s (%s)
Total: $%.2f

Items:
`, order.ID.String()[:8], order.Customer.Names, order.Customer.Email, order.Total)

	for _, item := range order.Items {
		body += fmt.Sprintf("- %dx %s ($%.2f each)\n",
			item.Quantity, item.Product.Name, item.UnitPrice)
	}

	m.SetBody("text/plain", body)

	// Send email
	d := mail.NewDialer(smtpHost, 587, smtpUser, smtpPass)
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	return nil
}
func SendPasswordResetEmail(email, resetToken string) error {
	// Configure SMTP settings
	smtpHost := os.Getenv("SMTP_HOST")
	// smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	// Create email message
	m := mail.NewMessage()
	m.SetHeader("From", smtpUser)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Password Reset Request")

	// Build email body
	body := fmt.Sprintf(`
You have requested to reset your password. Click the link below to reset your password:

%s/reset-password/%s
If you did not request a password reset, please ignore this email.
`, os.Getenv("CLIENT_URL"), resetToken)

	m.SetBody("text/plain", body)
	// Send email
	d := mail.NewDialer(smtpHost, 587, smtpUser, smtpPass)
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}
	return nil
}
