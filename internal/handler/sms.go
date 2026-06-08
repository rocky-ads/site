package handler

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/phoneverification"
	"github.com/rocky-ads/site/internal/service/sms"
)

// SMSWebhookHandler processes Twilio webhook callbacks for SMS status updates
// and incoming messages
func SMSWebhookHandler(c *fiber.Ctx) error {
	// Verify Twilio signature to ensure request is authentic
	if !sms.VerifyTwilioSignature(c) {
		logger.Warn("Webhook rejected: invalid signature",
			"component", "SMS", "ip", c.IP())
		return c.Status(403).JSON(fiber.Map{
			"error": "Invalid signature",
		})
	}

	// Parse the webhook data from Twilio
	webhookData, err := sms.ParseWebhook(c)
	if err != nil {
		logger.Error("Failed to parse webhook",
			"error", err, "component", "SMS")
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid webhook data",
		})
	}

	// Log all incoming webhook data for debugging
	logger.Debug("Webhook received",
		"component", "SMS", "from", webhookData.From, "to",
		webhookData.To, "body", webhookData.Body, "status",
		webhookData.MessageStatus, "messageSid",
		webhookData.MessageSid)

	// Update the status tracker if this is a status update
	if webhookData.MessageStatus != "" && webhookData.MessageSid != "" {
		status := sms.SMSStatus(webhookData.MessageStatus)
		sms.SetMessageStatus(webhookData.MessageSid, status)
	}

	// Incoming STOP invalidates pending verification codes only. SMS notification
	// preference is controlled in app settings (sms_opted_out), not via carrier STOP.
	body := strings.ToUpper(strings.TrimSpace(webhookData.Body))
	phone := webhookData.From
	if body == "STOP" {
		logger.Info("STOP received; invalidating verification codes only",
			"component", "SMS", "from", phone)
		if err := phoneverification.InvalidateCodes(phone); err != nil {
			logger.Error("Failed to invalidate verification codes",
				"error", err, "component", "SMS", "phoneNumber", phone)
			return err
		}
	}

	// Return success to Twilio
	return c.JSON(fiber.Map{
		"status": "success",
	})
}
