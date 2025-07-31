package sms

import (
	"context"
	"time"

	"github.com/lifenetwork-ai/iam-service/conf"
	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/packages/logger"
)

type smsProvider struct {
	config *conf.SmsConfiguration

	twillioClient  *TwilioClient
	whatsappClient *WhatsAppClient
}

func NewSMSProvider(config *conf.SmsConfiguration) *smsProvider {
	return &smsProvider{
		config:         config,
		twillioClient:  NewTwilioClient(config.Twilio.TwilioAccountSID, config.Twilio.TwilioAuthToken, config.Twilio.TwilioBaseURL),
		whatsappClient: NewWhatsAppClient(config.Whatsapp.WhatsappAccessToken, config.Whatsapp.WhatsappPhoneID, config.Whatsapp.WhatsappBaseURL),
	}
}

func (s *smsProvider) SendOTP(ctx context.Context, tenantName, receiver, channel, otp string, ttl time.Duration) error {
	logger.GetLogger().Infof("Sending OTP to %s", receiver)
	otpMessage := GetOTPMessage(tenantName, otp, ttl)
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

func (s *smsProvider) sendToWebhook(_ context.Context, tenantName, receiver, message string) error {
	logger.GetLogger().Infof("Call to webhook url: %s to send OTP to %s", tenantName, receiver)
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
