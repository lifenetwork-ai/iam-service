
# Ory Kratos Self-Service API Flows Documentation

## Overview

This document provides high-level documentation for Ory Kratos self-service flows using **API flows only** (not browser flows). API flows are designed for native applications like mobile apps, desktop applications, or IoT devices that don't run in a browser environment.

> **Important**: API flows are specifically for native applications. Never use API flows for browser-based applications as they lack CSRF protection and can expose security vulnerabilities.
> Swagger docs: https://www.postman.com/research-technologist-33381679/workspace/my-workspace/api/72668bb5-c52a-41c8-94c9-1bb9a6d781c7?action=share&creator=21575409&active-environment=21575409-bb8ad2a1-1708-49fd-9f5e-fdf73702a7dc
## Flow Types

Ory Kratos supports the following self-service flows for API clients:

1. **Registration Flow** - User sign-up
2. **Login Flow** - User sign-in
3. **Settings Flow** - Profile management and updates
4. **Recovery Flow** - Password/account recovery
5. **Verification Flow** - Email/phone verification

## Common Flow Pattern

All self-service flows follow a consistent pattern:

1. **Initialize Flow** - Call initialization endpoint to create flow
2. **Get Flow Data** - Retrieve flow configuration and form fields
3. **Submit Flow** - Submit user data to complete the flow

## Authentication Methods

Each flow supports multiple authentication methods:

- **`password`** - Username/email and password
- **`oidc`** - Social sign-in (Google, Facebook, GitHub, etc.)
- **`code`** - One-time password via email/SMS
- **`passkey`** - WebAuthn/FIDO2 passkeys
- **`profile`** - Profile trait updates (settings only)

---

## 1. Registration Flow


### Purpose

Create new user accounts in the system.
### API Endpoints

#### Initialize Registration Flow

```http
GET /self-service/registration/api
Accept: application/json
```

**Success Response**
```json
{
  "id": "flow-id-uuid",
  "type": "api",
  "expires_at": "2025-06-30T12:00:00Z",
  "issued_at": "2025-06-30T11:00:00Z",
  "request_url": "https://your-kratos.com/self-service/registration/api",
  "ui": {
    "action": "https://your-kratos.com/self-service/registration?flow=flow-id-uuid",
    "method": "POST",
    "nodes": [
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "traits.email",
          "type": "email",
          "required": true
        }
      },
      {
        "type": "input", 
        "group": "password",
        "attributes": {
          "name": "password",
          "type": "password",
          "required": true
        }
      }
    ]
  }
}
```


#### Submit Registration Flow

```http
POST /self-service/registration?flow=<flow-id>
Content-Type: application/json
```

```bash
curl -X POST \
  'https://auth.develop.lifenetwork.ai/self-service/registration?flow=84c15524-475b-43ca-9dc0-f23b2872bde2' \
  -H 'Content-Type: application/json' \
  -d '{
    "csrf_token": "",
    "method": "code",
    "traits.email": "user@example.com",
    "traits.phone_number": "+84987654321",
    "traits.tenant": "life_ai"
  }'
```

**Request Payload (Password Method)**:

```json
{
  "method": "password",
  "password": "secure-password",
  "traits.email": "user@example.com",
  "traits.name.first": "John",
  "traits.name.last": "Doe"
}
```

**Request Payload (OIDC Method)**:

```json
{
  "method": "oidc",
  "provider": "google"
}
```
**Request Payload (Code Method)**:
```json
{
	"csrf_token": "",
    "method": "code",
    "traits.email": "user@example.com",
    "traits.phone_number": "+84987654321",
    "traits.tenant": "life_ai"
    }
```
**Success Response**:

```json
{
  {
  "id": "84c15524-475b-43ca-9dc0-f23b2872bde2",
  "type": "api",
  "expires_at": "2025-06-30T17:09:29.187528Z",
  "issued_at": "2025-06-30T16:59:29.187528Z",
  "request_url": "https://auth.develop.lifenetwork.ai/self-service/registration/api",
  "active": "code",
  "ui": {
    "action": "https://auth.develop.lifenetwork.ai/self-service/registration?flow=84c15524-475b-43ca-9dc0-f23b2872bde2",
    "method": "POST",
    "nodes": [
      {
        "type": "input",
        "group": "default",
        "attributes": {
          "name": "csrf_token",
          "type": "hidden",
          "value": "",
          "required": true,
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {}
      },
      {
        "type": "input",
        "group": "default",
        "attributes": {
          "name": "traits.email",
          "type": "hidden",
          "value": "user@example.com",
          "autocomplete": "email",
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070002,
            "text": "E-Mail",
            "type": "info",
            "context": {
              "title": "E-Mail"
            }
          }
        }
      },
      {
        "type": "input",
        "group": "default",
        "attributes": {
          "name": "traits.phone_number",
          "type": "hidden",
          "value": "+84987654321",
          "autocomplete": "tel",
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070002,
            "text": "Phone Number",
            "type": "info",
            "context": {
              "title": "Phone Number"
            }
          }
        }
      },
      {
        "type": "input",
        "group": "default",
        "attributes": {
          "name": "traits.tenant",
          "type": "hidden",
          "value": "life_ai",
          "required": true,
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070002,
            "text": "Tenant",
            "type": "info",
            "context": {
              "title": "Tenant"
            }
          }
        }
      },
      {
        "type": "input",
        "group": "code",
        "attributes": {
          "name": "method",
          "type": "hidden",
          "value": "code",
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {}
      },
      {
        "type": "input",
        "group": "code",
        "attributes": {
          "name": "code",
          "type": "text",
          "required": true,
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070012,
            "text": "Registration code",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "code",
        "attributes": {
          "name": "method",
          "type": "submit",
          "value": "code",
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070009,
            "text": "Continue",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "code",
        "attributes": {
          "name": "resend",
          "type": "submit",
          "value": "code",
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1070008,
            "text": "Resend code",
            "type": "info"
          }
        }
      },
      {
        "type": "input",
        "group": "profile",
        "attributes": {
          "name": "screen",
          "type": "submit",
          "value": "credential-selection",
          "disabled": false,
          "node_type": "input"
        },
        "messages": [],
        "meta": {
          "label": {
            "id": 1040008,
            "text": "Back",
            "type": "info"
          }
        }
      }
    ],
    "messages": [
      {
        "id": 1040005,
        "text": "A code has been sent to the address(es) you provided. If you have not received a message, check the spelling of the address and retry the registration.",
        "type": "info"
      }
    ]
  },
  "organization_id": null,
  "transient_payload": {},
  "state": "sent_email"
}
}
```

**Success Response (with session hook enabled)**:

```json
{
  "session_token": "session-token-string",
  "session": {
    "id": "session-uuid",
    "active": true,
    "expires_at": "2025-07-01T11:00:00Z",
    "identity": { /* identity object */ }
  },
  "identity": { /* identity object */ }
}
```

---

## 2. Login Flow

### Purpose

Authenticate existing users and create sessions.

### API Endpoints

#### Initialize Login Flow

```http
GET /self-service/login/api
Accept: application/json
```

**Optional Query Parameters**:

- `refresh=true` - Force re-authentication
- `aal=aal2` - Request specific Authenticator Assurance Level

**Response**:

```json
{
  "id": "flow-id-uuid",
  "type": "api", 
  "expires_at": "2025-06-30T12:00:00Z",
  "issued_at": "2025-06-30T11:00:00Z",
  "request_url": "https://your-kratos.com/self-service/login/api",
  "requested_aal": "aal1",
  "ui": {
    "action": "https://your-kratos.com/self-service/login?flow=flow-id-uuid",
    "method": "POST",
    "nodes": [
      {
        "type": "input",
        "group": "password", 
        "attributes": {
          "name": "identifier",
          "type": "text",
          "required": true
        }
      },
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "password", 
          "type": "password",
          "required": true
        }
      }
    ]
  }
}
```

#### Submit Login Flow

```http
POST /self-service/login?flow=<flow-id>
Content-Type: application/json
```

**Request Payload (Password Method)**:

```json
{
  "method": "password",
  "identifier": "user@example.com",
  "password": "user-password"
}
```

**Request Payload (OIDC Method)**:

```json
{
  "method": "oidc", 
  "provider": "google"
}
```

**Success Response**:

```json
{
  "session_token": "session-token-string",
  "session": {
    "id": "session-uuid",
    "active": true,
    "expires_at": "2025-07-01T11:00:00Z",
    "authenticated_at": "2025-06-30T11:00:00Z",
    "authenticator_assurance_level": "aal1",
    "identity": {
      "id": "identity-uuid",
      "traits": {
        "email": "user@example.com"
      }
    }
  }
}
```

---

## 3. Settings Flow

### Purpose

Update user profile information, change passwords, manage 2FA, link/unlink social accounts.

### API Endpoints

#### Initialize Settings Flow

```http
GET /self-service/settings/api
Accept: application/json
Authorization: Bearer <session-token>
```

**Response**:

```json
{
  "id": "flow-id-uuid",
  "type": "api",
  "expires_at": "2025-06-30T12:00:00Z", 
  "issued_at": "2025-06-30T11:00:00Z",
  "identity": {
    "id": "identity-uuid",
    "traits": {
      "email": "user@example.com"
    }
  },
  "ui": {
    "action": "https://your-kratos.com/self-service/settings?flow=flow-id-uuid",
    "method": "POST",
    "nodes": [
      {
        "type": "input",
        "group": "profile",
        "attributes": {
          "name": "traits.email",
          "type": "email",
          "value": "user@example.com"
        }
      },
      {
        "type": "input",
        "group": "password", 
        "attributes": {
          "name": "password",
          "type": "password"
        }
      }
    ]
  }
}
```

#### Submit Settings Flow

```http
POST /self-service/settings?flow=<flow-id>
Content-Type: application/json
Authorization: Bearer <session-token>
```

**Request Payload (Profile Update)**:

```json
{
  "method": "profile",
  "traits.email": "newemail@example.com",
  "traits.name.first": "Jane"
}
```

**Request Payload (Password Change)**:

```json
{
  "method": "password",
  "password": "new-secure-password"
}
```

**Success Response**:

```json
{
  "identity": {
    "id": "identity-uuid",
    "traits": {
      "email": "newemail@example.com",
      "name": {
        "first": "Jane"
      }
    }
  }
}
```

---

## 4. Recovery Flow

### Purpose

Recover user accounts through email/SMS codes or recovery links.

### API Endpoints

#### Initialize Recovery Flow

```http
GET /self-service/recovery/api
Accept: application/json
```

**Response**:

```json
{
  "id": "flow-id-uuid",
  "type": "api",
  "expires_at": "2025-06-30T12:00:00Z",
  "issued_at": "2025-06-30T11:00:00Z",
  "ui": {
    "action": "https://your-kratos.com/self-service/recovery?flow=flow-id-uuid",
    "method": "POST",
    "nodes": [
      {
        "type": "input",
        "group": "code",
        "attributes": {
          "name": "email",
          "type": "email",
          "required": true
        }
      }
    ]
  }
}
```

#### Submit Recovery Flow

```http
POST /self-service/recovery?flow=<flow-id>
Content-Type: application/json
```

**Request Payload (Send Recovery Code)**:

```json
{
  "method": "code",
  "email": "user@example.com"
}
```

**Request Payload (Submit Recovery Code)**:

```json
{
  "method": "code", 
  "code": "123456",
  "email": "user@example.com"
}
```

**Success Response**:

```json
{
  "redirect_browser_to": "https://your-app.com/settings?flow=new-flow-id"
}
```

---

## 5. Verification Flow

### Purpose

Verify email addresses or phone numbers.

### API Endpoints

#### Initialize Verification Flow

```http
GET /self-service/verification/api
Accept: application/json
```

**Response**:

```json
{
  "id": "flow-id-uuid",
  "type": "api",
  "expires_at": "2025-06-30T12:00:00Z",
  "issued_at": "2025-06-30T11:00:00Z", 
  "ui": {
    "action": "https://your-kratos.com/self-service/verification?flow=flow-id-uuid",
    "method": "POST",
    "nodes": [
      {
        "type": "input",
        "group": "code",
        "attributes": {
          "name": "email",
          "type": "email",
          "required": true
        }
      }
    ]
  }
}
```

#### Submit Verification Flow

```http
POST /self-service/verification?flow=<flow-id>
Content-Type: application/json
```

**Request Payload (Send Verification Code)**:

```json
{
  "method": "code",
  "email": "user@example.com"
}
```

**Request Payload (Submit Verification Code)**:

```json
{
  "method": "code",
  "code": "123456", 
  "email": "user@example.com"
}
```

**Success Response**:

```json
{
  "redirect_browser_to": "https://your-app.com/verified"
}
```

---

## Session Management

### Get Current Session

```http
GET /sessions/whoami
Authorization: Bearer <session-token>
```

**Response**:

```json
{
  "id": "session-uuid",
  "active": true,
  "expires_at": "2025-07-01T11:00:00Z",
  "identity": {
    "id": "identity-uuid",
    "traits": {
      "email": "user@example.com"
    }
  }
}
```

### Logout

```http
		DELETE /self-service/logout/api
Authorization: Bearer <session-token>
```

---

## Error Handling

### Common Error Response Format

```json
{
  "id": "flow-id-uuid",
  "type": "api", 
  "ui": {
    "action": "https://your-kratos.com/self-service/login?flow=flow-id-uuid",
    "method": "POST",
    "messages": [
      {
        "id": 4000006,
        "text": "The provided credentials are invalid, check for spelling mistakes in your password or username, email address, or phone number.",
        "type": "error"
      }
    ],
    "nodes": [
      {
        "type": "input",
        "group": "password",
        "attributes": {
          "name": "identifier",
          "type": "text",
          "value": "user@example.com"
        },
        "messages": [
          {
            "id": 4000006,
            "text": "The provided credentials are invalid",
            "type": "error"
          }
        ]
      }
    ]
  }
}
```

### HTTP Status Codes

- **200**: Success or validation errors
- **400**: Bad request (e.g., flow expired)
- **401**: Unauthorized (invalid/missing session token)
- **403**: Forbidden (insufficient permissions)
- **404**: Flow not found
- **410**: Flow expired
- **422**: Validation errors

---

## Best Practices

1. **Flow Expiration**: Always check flow expiration and handle expired flows by initializing new ones
2. **Session Tokens**: Store session tokens securely and include them in Authorization headers
3. **Error Handling**: Parse error messages from the `ui.messages` and `ui.nodes[].messages` arrays
4. **OIDC Flows**: Handle browser redirects for social sign-in flows
5. **Validation**: Use the `ui.nodes` array to render dynamic forms based on available methods
6. **Security**: Never log or expose session tokens in client-side code

---

## Configuration Requirements

### Identity Schema

Define the required traits for your users:

```json
{
  "properties": {
    "traits": {
      "type": "object", 
      "properties": {
        "email": {
          "type": "string",
          "format": "email",
          "ory.sh/kratos": {
            "credentials": {
              "password": {
                "identifier": true
              }
            }
          }
        }
      },
      "required": ["email"]
    }
  }
}
```

### Session Hook (Auto-login)

Enable automatic session creation after registration:

```yaml
selfservice:
  flows:
    registration:
      after:
        password:
          hooks:
            - hook: session
        oidc:
          hooks:
            - hook: session
```