package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/constants"
	cachetypes "github.com/lifenetwork-ai/iam-service/infrastructures/caching/types"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type smsProvider struct {
	config    *conf.SmsConfiguration
	cacheRepo cachetypes.CacheRepository

	twillioClient  *TwilioClient
	whatsappClient *WhatsAppClient
}

func NewSMSProvider(config *conf.SmsConfiguration, cache cachetypes.CacheRepository) *smsProvider {
	return &smsProvider{
		config:         config,
		cacheRepo:      cache,
		twillioClient:  NewTwilioClient(config.Twilio.TwilioAccountSID, config.Twilio.TwilioAuthToken, config.Twilio.TwilioBaseURL),
		whatsappClient: NewWhatsAppClient(config.Whatsapp.WhatsappAccessToken, config.Whatsapp.WhatsappPhoneID, config.Whatsapp.WhatsappBaseURL),
	}
}

func (s *smsProvider) SendOTP(ctx context.Context, tenantName, receiver, channel, otp string, ttl time.Duration) error {
	logger.GetLogger().Infof("Preparing to send OTP %s to %s via %s", otp, receiver, channel)
	otpMessage := GetOTPMessage(tenantName, otp, ttl)

	if conf.IsDevReviewerBypassEnabled() && receiver == conf.DevReviewerIdentifier() {
		// Capture the OTP in cache for dev reviewer
		key := &cachetypes.Keyer{Raw: fmt.Sprintf("otp:%s:%s", tenantName, receiver)}
		_ = s.cacheRepo.SaveItem(key, otp, ttl)

		logger.GetLogger().Infof("Dev bypass enabled: skip sending OTP",
			"receiver", receiver,
			"tenant", tenantName,
			"ttl", ttl,
		)
		return nil
	}

	logger.GetLogger().Infof("Sending OTP to %s", receiver)
	switch channel {
	case constants.ChannelSMS:
		return s.sendSMS(ctx, tenantName, receiver, otpMessage)
	case constants.ChannelWhatsApp:
		return s.sendToWhatsapp(ctx, tenantName, receiver, otpMessage)
	case constants.ChannelZalo:
		return s.sendToZalo(ctx, tenantName, receiver, otpMessage)
	default:
		return s.sendToWebhook(ctx, tenantName, receiver, otpMessage)
	}
}

func (s *smsProvider) sendToWebhook(ctx context.Context, tenantName, receiver, message string) error {
	url := conf.GetMockWebhookURL()
	if url == "" {
		return errors.New("MOCK_WEBHOOK_URL is not set")
	}

	type webhookPayload struct {
		Tenant  string `json:"tenant"`
		To      string `json:"to"`
		Message string `json:"message"`
	}

	payload := webhookPayload{
		Tenant:  tenantName,
		To:      receiver,
		Message: message,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set(constants.HeaderKeyContentType, constants.HeaderContentTypeJson)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return errors.New("webhook returned non-2xx status: " + resp.Status)
	}

	logger.GetLogger().Infof("Webhook sent successfully to %s", receiver)
	return nil
}

func (s *smsProvider) sendSMS(_ context.Context, tenantName, receiver, message string) error {
	logger.GetLogger().Infof("Sending SMS to %s", receiver)
	// TODO: Twilio phone number should be dynamic and not set in the config
	resp, err := s.twillioClient.SendSMS(tenantName, s.config.Twilio.TwilioFrom, receiver, message)
	if err != nil {
		return err
	}
	logger.GetLogger().Infof("SMS sent successfully: %+v", resp)
	return nil
}

func (s *smsProvider) sendToWhatsapp(_ context.Context, tenantName, receiver, message string) error {
	logger.GetLogger().Infof("Sending via WhatsApp to %s", receiver)
	resp, err := s.whatsappClient.SendMessage(tenantName, receiver, message)
	if err != nil {
		return err
	}
	logger.GetLogger().Infof("WhatsApp sent successfully: %+v", resp)
	return nil
}

func (s *smsProvider) sendToZalo(_ context.Context, tenantName, receiver, message string) error {
	logger.GetLogger().Infof("Sending Zalo to %s", receiver)
	return nil
}
