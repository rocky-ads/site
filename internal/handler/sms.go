package handler

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rocky-ads/site/internal/accountrecovery"
	"github.com/rocky-ads/site/internal/logger"
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

	bodyTrimmed := strings.TrimSpace(webhookData.Body)
	_, isRecover := accountrecovery.ParseRecoverCode(bodyTrimmed)
	logBody := bodyTrimmed
	if isRecover {
		logBody = "RECOVER [redacted]"
	}

	logger.Debug("Webhook received",
		"component", "SMS",
		"hasFrom", webhookData.From != "",
		"hasTo", webhookData.To != "",
		"body", logBody,
		"status", webhookData.MessageStatus,
		"messageSid", webhookData.MessageSid)

	// Update the status tracker if this is a status update
	if webhookData.MessageStatus != "" && webhookData.MessageSid != "" {
		status := sms.SMSStatus(webhookData.MessageStatus)
		sms.SetMessageStatus(webhookData.MessageSid, status)
	}

	// Incoming STOP: OTP codes are owned by Twilio Verify (short TTL). Carrier
	// STOP may also block Programmable Messaging delivery. App notification
	// preference remains sms_opted_out in settings.
	body := strings.ToUpper(strings.TrimSpace(webhookData.Body))
	phone := webhookData.From
	if body == "STOP" {
		logger.Info("STOP received; Verify OTPs expire on their own",
			"component", "SMS", "phone", phone)
		return c.JSON(fiber.Map{"status": "success"})
	}

	if isRecover {
		logger.Info("Recovery SMS received", "component", "SMS")
		err := accountrecovery.CompleteFromSMS(phone, bodyTrimmed)
		if err != nil {
			if errors.Is(err, accountrecovery.ErrNoUser) ||
				errors.Is(err, accountrecovery.ErrNotFound) ||
				errors.Is(err, accountrecovery.ErrInvalidSMS) {
				logger.Info("Recovery SMS not applied",
					"component", "SMS", "reason", err.Error())
			} else {
				logger.Error("Failed to complete recovery from SMS",
					"error", err, "component", "SMS")
			}
		}
	}

	return c.JSON(fiber.Map{
		"status": "success",
	})
}
