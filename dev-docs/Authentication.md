# Authentication Flows Documentation

The following sequence diagram illustrates the main authentication flows in the system:

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Identity
    
    Note over Client,Identity: Registration Flow
    Client->>API: POST /api/v1/users/register
    Note right of Client: Body: {email/phone, tenant}
    API->>Identity: Create Identity
    Identity-->>API: Identity Created
    API-->>Client: 200 OK {flowId}
    Note over Client,API: User receives verification code
    Client->>API: POST /api/v1/users/challenge-verify
    Note right of Client: Body: {flowId, code, type: "register"}
    API->>Identity: Verify Code
    Identity-->>API: Code Verified
    API-->>Client: 200 OK {access_token, refresh_token}
    
    Note over Client,Identity: Login Flow
    Client->>API: POST /api/v1/users/challenge-with-email
    Note right of Client: Body: {email, tenant}
    API->>Identity: Generate Challenge
    Identity-->>API: Challenge Created
    API-->>Client: 200 OK {flowId}
    Note over Client,API: User receives verification code
    Client->>API: POST /api/v1/users/challenge-verify
    Note right of Client: Body: {flowId, code, type: "login"}
    API->>Identity: Verify Challenge
    Identity-->>API: Challenge Verified
    API-->>Client: 200 OK {access_token, refresh_token}
    
    Note over Client,Identity: Profile & Logout
    Client->>API: GET /api/v1/users/me
    Note right of Client: Header: Authorization: Bearer {access_token}
    API->>Identity: Get Identity
    Identity-->>API: Identity Details
    API-->>Client: 200 OK {profile}
    Client->>API: POST /api/v1/users/logout
    Note right of Client: Header: Authorization: Bearer {access_token}
    API->>Identity: Revoke Tokens
    Identity-->>API: Tokens Revoked
    API-->>Client: 200 OK
```

## API Endpoints

### Registration
- `POST /api/v1/users/register`
  - Request: `{ email/phone: string, tenant: string }`
  - Response: `{ flowId: string }`

### Login
- `POST /api/v1/users/challenge-with-email`
  - Request: `{ email: string, tenant: string }`
  - Response: `{ flowId: string }`
- `POST /api/v1/users/challenge-with-phone`
  - Request: `{ phone: string, tenant: string }`
  - Response: `{ flowId: string }`

### Verification
- `POST /api/v1/users/challenge-verify`
  - Request: `{ flowId: string, code: string, type: "register" | "login" }`
  - Response: `{ access_token: string, refresh_token: string }`

### Profile
- `GET /api/v1/users/me`
  - Header: `Authorization: Bearer {access_token}`
  - Response: `{ profile: object }`

### Logout
- `POST /api/v1/users/logout`
  - Header: `Authorization: Bearer {access_token}`
  - Response: `200 OK`

## Error Responses

Common error responses across all endpoints:

- `400 Bad Request` - Invalid request format or missing required fields
- `401 Unauthorized` - Invalid or expired access token
- `404 Not Found` - Resource not found
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

### Example Error Response
```json
{
  "error": "invalid_request",
  "message": "Missing required field: tenant",
  "code": 400
}
```