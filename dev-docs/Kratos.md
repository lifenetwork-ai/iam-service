### Authentication Flows Documentation

The following sequence diagram illustrates the main authentication flows in the system:

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Identity
    
    %% Registration Flow
    rect rgb(200, 220, 240)
        Note over Client,API: Registration Flow
        Client->>API: POST /api/v1/users/register<br/>Body: {email/phone, tenant}
        API->>Identity: Create Identity
        Identity-->>API: Identity Created
        API-->>Client: 200 OK {flowId}
        Note over Client,API: User receives verification code
        Client->>API: POST /api/v1/users/challenge-verify<br/>Body: {flowId, code, type: "registration"}
        API->>Identity: Verify Code
        Identity-->>API: Code Verified
        API-->>Client: 200 OK {access_token, refresh_token}
    end

    %% Login Flow
    rect rgb(220, 240, 200)
        Note over Client,API: Login Flow
        Client->>API: POST /api/v1/users/challenge-with-email<br/>Body: {email, tenant}<br/>or<br/>POST /api/v1/users/challenge-with-phone<br/>Body: {phone, tenant}
        API->>Identity: Generate Challenge
        Identity-->>API: Challenge Created
        API-->>Client: 200 OK {flowId}
        Note over Client,API: User receives verification code
        Client->>API: POST /api/v1/users/challenge-verify<br/>Body: {flowId, code, type: "login"}
        API->>Identity: Verify Challenge
        Identity-->>API: Challenge Verified
        API-->>Client: 200 OK {access_token, refresh_token}
    end

    %% Profile & Logout
    rect rgb(240, 220, 220)
        Note over Client,API: Profile & Logout
        Client->>API: GET /api/v1/users/me<br/>Header: Authorization: Bearer {access_token}
        API->>Identity: Get Identity
        Identity-->>API: Identity Details
        API-->>Client: 200 OK {profile}
        Client->>API: POST /api/v1/users/logout<br/>Header: Authorization: Bearer {access_token}
        API->>Identity: Revoke Tokens
        Identity-->>API: Tokens Revoked
        API-->>Client: 200 OK
    end
```

### API Endpoints

#### Registration
- `POST /api/v1/users/register`
  - Request: `{ email/phone: string, tenant: string }`
  - Response: `{ flowId: string }`

#### Login
- `POST /api/v1/users/challenge-with-email`
  - Request: `{ email: string, tenant: string }`
  - Response: `{ flowId: string }`
- `POST /api/v1/users/challenge-with-phone`
  - Request: `{ phone: string, tenant: string }`
  - Response: `{ flowId: string }`

#### Verification
- `POST /api/v1/users/challenge-verify`
  - Request: `{ flowId: string, code: string, type: "registration" | "login" }`
  - Response: `{ access_token: string, refresh_token: string }`

#### Profile
- `GET /api/v1/users/me`
  - Header: `Authorization: Bearer {access_token}`
  - Response: `{ profile: object }`

#### Logout
- `POST /api/v1/users/logout`
  - Header: `Authorization: Bearer {access_token}`
  - Response: `200 OK`

