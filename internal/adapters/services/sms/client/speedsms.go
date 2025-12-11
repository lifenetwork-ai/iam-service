package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/lifenetwork-ai/iam-service/constants"
	"github.com/lifenetwork-ai/iam-service/packages/utils"
)

type SpeedSMSBrandname string

const (
	GeneticaBrandname SpeedSMSBrandname = "GENETICA"
	LifeBrandname     SpeedSMSBrandname = "LIFE AI"
)

// SpeedSMSResponse represents the API response structure
type SpeedSMSResponse struct {
	Status string `json:"status"`
	Code   string `json:"code"`
	Data   struct {
		TranId       int      `json:"tranId"`
		TotalSMS     int      `json:"totalSMS"`
		TotalPrice   float64  `json:"totalPrice"`
		InvalidPhone []string `json:"invalidPhone"`
	} `json:"data"`
	Message string `json:"message,omitempty"` // For error responses
}

// UserInfoResponse represents the user info API response
type UserInfoResponse struct {
	Status string `json:"status"`
	Code   string `json:"code"`
	Data   struct {
		Email    string `json:"email"`
		Balance  string `json:"balance"`
		Currency string `json:"currency"`
	} `json:"data"`
	Message string `json:"message,omitempty"` // For error responses
}

// SMS Types constants
const (
	SMSTypeRandom    = 2 // Random number
	SMSTypeBrandname = 3 // Custom brandname
	SMSTypeDefault   = 4 // Default brandname (Verify or Notify)
	SMSTypeAndroid   = 5 // Android app
)

type SpeedSMSClient struct {
	BaseURL     string
	AccessToken string
	HTTPClient  *http.Client
}

func NewSpeedSMSClient(accessToken string) *SpeedSMSClient {
	return &SpeedSMSClient{
		BaseURL:     "https://api.speedsms.vn",
		AccessToken: accessToken,
		HTTPClient:  &http.Client{},
	}
}

// GetUserInfo retrieves account information
func (c *SpeedSMSClient) GetUserInfo() (*UserInfoResponse, error) {
	endpoint := fmt.Sprintf("%s/index.php/user/info", c.BaseURL)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add access token as query parameter
	q := req.URL.Query()
	q.Add("access-token", c.AccessToken)
	req.URL.RawQuery = q.Encode()

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var userInfo UserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &userInfo, nil
}

// SendSMS sends SMS to multiple phone numbers
func (c *SpeedSMSClient) SendSMS(phones []string, content string, smsType int, sender SpeedSMSBrandname) (*SpeedSMSResponse, error) {
	endpoint := fmt.Sprintf("%s/index.php/sms/send", c.BaseURL)

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Build query parameters
	q := req.URL.Query()
	q.Add("access-token", c.AccessToken)
	q.Add("to", strings.Join(phones, ",")) // Join multiple phones with comma
	q.Add("content", content)
	q.Add("type", fmt.Sprintf("%d", smsType))
	if sender != "" {
		q.Add("sender", string(sender))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var smsResp SpeedSMSResponse
	if err := json.NewDecoder(resp.Body).Decode(&smsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &smsResp, nil
}

// SendSingleSMS sends SMS to a single phone number
func (c *SpeedSMSClient) SendSingleSMS(phone, content string, smsType int, sender SpeedSMSBrandname) (*SpeedSMSResponse, error) {
	// Normalize phone number
	phone, _, err := utils.NormalizePhoneE164(phone, constants.DefaultRegion)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize phone number: %w", err)
	}
	return c.SendSMS([]string{phone}, content, smsType, sender)
}

// SendOTP sends OTP message with custom content
func (c *SpeedSMSClient) SendOTP(phone, otp string, brandname SpeedSMSBrandname) (*SpeedSMSResponse, error) {
	content := otpContent(otp, brandname, "en")
	return c.SendSingleSMS(phone, content, SMSTypeBrandname, brandname)
}

func otpContent(otp string, brandname SpeedSMSBrandname, lang string) string {
	switch lang {
	case constants.EnglishLanguage:
		return fmt.Sprintf("Your OTP number at %s is %s.", brandname, otp)
	case constants.VietnameseLanguage:
		return fmt.Sprintf("%s la ma OTP cua ban tai %s.", otp, brandname)
	}
}
