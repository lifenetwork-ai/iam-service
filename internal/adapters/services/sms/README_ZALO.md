# Zalo SMS Client Implementation

This document describes the ZaloClient implementation for sending SMS messages via the Zalo Business API.

## Overview

The ZaloClient provides functionality to send template-based SMS messages using the Zalo Business API. It supports sending OTP messages and custom template messages.

## Configuration

Add the following environment variables to your `.env` file:

```env
ZALO_ACCESS_TOKEN=your_zalo_access_token_here
ZALO_BASE_URL=https://business.openapi.zalo.me
ZALO_TEMPLATE_ID=473988
```

## API Endpoint

The client uses the Zalo Business API endpoint:
- **URL**: `https://business.openapi.zalo.me/message/template`
- **Method**: POST
- **Headers**: 
  - `access_token`: Your Zalo access token
  - `Content-Type`: application/json

## Request Format

```json
{
    "phone": "84346840626",
    "template_id": 473988,
    "template_data": {
        "otp": "170202"
    }
}
```

## Response Format

```json
{
    "error": 0,
    "message": "Success",
    "data": {
        "sent_time": "1754380391326",
        "sending_mode": "1",
        "quota": {
            "remainingQuota": "4998",
            "dailyQuota": "5000"
        },
        "msg_id": "55940651e1fb48a311ec"
    }
}
```

## Usage

### Basic OTP Sending

```go
import (
    "context"
    "github.com/lifenetwork-ai/iam-service/internal/adapters/services/sms"
)

// Initialize client
client := sms.NewZaloClient(accessToken, baseURL)

// Send OTP
ctx := context.Background()
response, err := client.SendOTP(ctx, "84346840626", "170202", 473988)
if err != nil {
    log.Printf("Error: %v", err)
    return
}

fmt.Printf("Message ID: %s\n", response.Data.MsgID)
```

### Custom Template Message

```go
// Send custom template message
templateData := map[string]interface{}{
    "otp": "170202",
    "user_name": "John Doe",
    // Add other template variables as needed
}

response, err := client.SendTemplateMessage(ctx, "84346840626", 473988, templateData)
if err != nil {
    log.Printf("Error: %v", err)
    return
}
```

### Integration with SMS Service

The ZaloClient is automatically integrated into the SMS service. When using the `zalo` channel, the service will:

1. Extract the OTP from the message
2. Send it using the configured Zalo template
3. Log the response

```go
// The SMS service will automatically use Zalo when channel is "zalo"
err := smsProvider.SendOTP(ctx, "tenant", "84346840626", "zalo", "170202", 5*time.Minute)
```

## Error Handling

The client handles various error scenarios:

- **API Errors**: When the Zalo API returns an error (error code != 0)
- **HTTP Errors**: When the HTTP request fails (status code >= 400)
- **Network Errors**: When the request cannot be completed
- **JSON Errors**: When the response cannot be parsed

## Testing

Run the tests to verify the implementation:

```bash
go test ./internal/adapters/services/sms -v
```

## Notes

- The client uses a 30-second timeout for HTTP requests
- OTP extraction from messages is done using a simple algorithm that looks for 6-digit sequences
- The template ID is configurable via environment variables
- The client supports context cancellation for request timeouts

## Example cURL Request

```bash
curl 'https://business.openapi.zalo.me/message/template' \
--header 'access_token: <access_token>' \
--header 'Content-Type: application/json' \
--data '{
    "phone": "84346840626",
    "template_id": 473988,
    "template_data": {
        "otp": "170202"
    }
}'
``` 