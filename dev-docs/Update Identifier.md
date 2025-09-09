# Update Identifier Flow Documentation

This document describes how an authenticated user updates their primary login identifier (email or phone) using an OTP-based verification flow.

---

## Sequence Diagram

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Identity

    Note over Client,Identity: Update Identifier (OTP)
    Client->>API: POST /api/v1/users/me/update-identifier {new_identifier}
    API->>Identity: Change Identifier (initiate verification)
    Identity-->>API: {flow_id, receiver, challenge_at}
    API-->>Client: 200 OK {flow}
    Note over Client,API: User must verify the new identifier
    Client->>API: POST /api/v1/users/challenge-verify {flow_id, code, type: "register"}
    API->>Identity: Verify & finalize update
    API-->>Client: 200 OK {session, user}
```

---

## API Endpoints

### `POST /api/v1/users/me/update-identifier`

Allows an authenticated user to update their primary identifier (email or phone). OTP verification is required.

#### Headers

- `X-Tenant-Id`: `string` (required)
- `Authorization`: `Bearer {session_token}` (required)

#### Request Body

```json
{
  "new_identifier": "string" // The new identifier value (email or phone)
}
```

#### Response

```json
{
  "data": {
    "flow_id": "string",
    "receiver": "string",
    "challenge_at": number
  }
}
```

---

### Verification

- `POST /api/v1/users/challenge-verify`
  - Headers:
    - `X-Tenant-Id`: string (required)
  - Request Body:

    ```json
    {
      "flow_id": "string",
      "code": "string",
      "type": "register"
    }
    ```

  - Response:

    ```json
    {
      "data": {
        "session_id": "string",
        "session_token": "string",
        "issued_at": "string",
        "expires_at": "string",
        "authenticated_at": "string",
        "authentication_methods": ["string"],
        "active": boolean,
        "user": {
          "id": "string",
          "email": "string",
          "phone": "string"
        }
      }
    }
    ```

---

## Error Responses

All responses follow the standard error format:

```json
{
  "status": number,
  "code": "string",
  "message": "string",
  "errors": [
    {
      "field": "string",
      "error": "string"
    }
  ]
}
```

### Common Error Codes

| Code | Description |
|------|-------------|
| `MSG_INVALID_TENANT` | Invalid or missing tenant ID |
| `MSG_UNAUTHORIZED` | Missing or invalid session token |
| `MSG_INVALID_PAYLOAD` | Invalid request body |
| `MSG_INVALID_IDENTIFIER_TYPE` | Identifier must be email or phone |
| `MSG_INVALID_EMAIL` | Email format is invalid |
| `MSG_INVALID_PHONE_NUMBER` | Phone number is invalid |
| `MSG_IDENTIFIER_ALREADY_EXISTS` | Identifier already exists in system |
| `MSG_IDENTIFIER_TYPE_NOT_EXISTS` | User does not have an identifier of this type to update |
| `MSG_MULTIPLE_IDENTIFIERS_EXISTS` | User has multiple identifiers, cross-type change not allowed |
| `MSG_RATE_LIMIT_EXCEEDED` | Too many OTP attempts |
| `MSG_INIT_REG_FLOW_FAILED` | Failed to initialize verification flow |
| `MSG_REGISTRATION_FAILED` | Verification submission failed |
| `MSG_SAVE_CHALLENGE_FAILED` | Could not persist challenge session |
---