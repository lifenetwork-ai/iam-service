# Authentication Flows Documentation

The following sequence diagram illustrates the main authentication flows in the system:

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Identity
    
    Note over Client,Identity: Registration Flow
    Client->>API: POST /api/v1/users/register
    Note right of Client: Body: {email/phone}
    API->>Identity: Create Identity
    Identity-->>API: Identity Created
    API-->>Client: 200 OK {flowId, verification_needed: true}
    Note over Client,API: User receives verification code
    Client->>API: POST /api/v1/users/challenge-verify
    Note right of Client: Body: {flowId, code, type: "register"}
    API->>Identity: Verify Code
    Identity-->>API: Code Verified
    API-->>Client: 200 OK {session_token, user}
    
    Note over Client,Identity: Login Flow (Email)
    Client->>API: POST /api/v1/users/challenge-with-email
    Note right of Client: Body: {email}
    API->>Identity: Generate Challenge
    Identity-->>API: Challenge Created
    API-->>Client: 200 OK {flowId}
    Note over Client,API: User receives verification code
    Client->>API: POST /api/v1/users/challenge-verify
    Note right of Client: Body: {flowId, code, type: "login"}
    API->>Identity: Verify Challenge
    Identity-->>API: Challenge Verified
    API-->>Client: 200 OK {session_token, user}
    
    Note over Client,Identity: Login Flow (Phone)
    Client->>API: POST /api/v1/users/challenge-with-phone
    Note right of Client: Body: {phone}
    API->>Identity: Generate Challenge
    Identity-->>API: Challenge Created
    API-->>Client: 200 OK {flowId}
    Note over Client,API: User receives verification code
    Client->>API: POST /api/v1/users/challenge-verify
    Note right of Client: Body: {flowId, code, type: "login"}
    API->>Identity: Verify Challenge
    Identity-->>API: Challenge Verified
    API-->>Client: 200 OK {session_token, user}
    
    Note over Client,Identity: Profile & Logout
    Client->>API: GET /api/v1/users/me
    Note right of Client: Header: Authorization: Bearer {session_token}
    API->>Identity: Get Identity
    Identity-->>API: Identity Details
    API-->>Client: 200 OK {user}
    Client->>API: POST /api/v1/users/logout
    Note right of Client: Header: Authorization: Bearer {session_token}
    API->>Identity: Revoke Session
    Identity-->>API: Session Revoked
    API-->>Client: 200 OK
```

## API Endpoints

### Registration
- `POST /api/v1/users/register`
  - Headers:
    - `X-Tenant-Id`: string (required)
  - Request Body:
    ```json
    {
      "email": "string",  // Either email or phone must be provided
      "phone": "string"   // Cannot provide both
    }
    ```
  - Response:
    ```json
    {
      "data": {
        "verification_flow": {
          "flow_id": "string",
          "receiver": "string",
          "challenge_at": number
        },
        "verification_needed": true
      }
    }
    ```

### Login
- `POST /api/v1/users/challenge-with-email`
  - Headers:
    - `X-Tenant-Id`: string (required)
  - Request Body:
    ```json
    {
      "email": "string"
    }
    ```
  - Response:
    ```json
    {
      "data": {
        "flow_id": "string",
        "receiver": "string",
        "challenge_at": number
      }
    }
    ```

- `POST /api/v1/users/challenge-with-phone`
  - Headers:
    - `X-Tenant-Id`: string (required)
  - Request Body:
    ```json
    {
      "phone": "string"
    }
    ```
  - Response:
    ```json
    {
      "data": {
        "flow_id": "string",
        "receiver": "string",
        "challenge_at": number
      }
    }
    ```

### Verification
- `POST /api/v1/users/challenge-verify`
  - Headers:
    - `X-Tenant-Id`: string (required)
  - Request Body:
    ```json
    {
      "flow_id": "string",
      "code": "string",
      "type": "register" | "login"
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
          "phone": "string",
          "name": "string",
          "first_name": "string",
          "last_name": "string",
          "full_name": "string",
          "user_name": "string",
          "tenant": "string",
          "status": boolean,
          "created_at": number,
          "updated_at": number
        }
      }
    }
    ```

### Profile
- `GET /api/v1/users/me`
  - Headers:
    - `X-Tenant-Id`: string (required)
    - `Authorization`: Bearer {session_token} (required)
  - Response:
    ```json
    {
      "data": {
        "id": "string",
        "email": "string",
        "phone": "string",
        "name": "string",
        "first_name": "string",
        "last_name": "string",
        "full_name": "string",
        "user_name": "string",
        "tenant": "string",
        "status": boolean,
        "created_at": number,
        "updated_at": number
      }
    }
    ```

### Logout
- `POST /api/v1/users/logout`
  - Headers:
    - `X-Tenant-Id`: string (required)
    - `Authorization`: Bearer {session_token} (required)
  - Request Body: empty object `{}`
  - Response: 200 OK

## Error Responses

All endpoints may return the following error responses:

```json
{
  "status": number,    // HTTP status code
  "code": "string",    // Error code
  "message": "string", // Human-readable error message
  "errors": [         // Optional array of detailed errors
    {
      "field": "string",
      "error": "string"
    }
  ]
}
```

Common error codes:
- `MSG_INVALID_TENANT` - Invalid or missing tenant ID
- `MSG_INVALID_PAYLOAD` - Invalid request payload
- `MSG_INVALID_PHONE_NUMBER` - Phone number must be in international format (e.g., +1234567890)
- `MSG_UNAUTHORIZED` - Invalid or missing session token
- `MSG_CONTACT_METHOD_REQUIRED` - Either email or phone must be provided
- `MSG_ONLY_EMAIL_OR_PHONE_MUST_BE_PROVIDED` - Cannot provide both email and phone
- `MSG_EMAIL_IS_REQUIRED` - Email is required for email challenge
- `MSG_PHONE_NUMBER_IS_REQUIRED` - Phone number is required for phone challenge
- `MSG_FAILED_TO_MAKE_CHALLENGE` - Failed to create challenge
- `MSG_INVALID_VERIFICATION_TYPE` - Invalid verification type
- `MSG_FAILED_TO_GET_USER_PROFILE` - Failed to get user profile
- `MSG_IAM_LOOKUP_FAILED` - Failed to query IAM database for identity check
- `MSG_EMAIL_ALREADY_EXISTS` - Email has already been registered
- `MSG_PHONE_ALREADY_EXISTS` - Phone number has already been registered
- `MSG_RATE_LIMIT_CHECK_FAILED` - Could not check rate limit
- `MSG_RATE_LIMIT_EXCEEDED` - Too many attempts, please try again later
