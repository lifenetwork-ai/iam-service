### DeleteIdentifier and Register flows (soft delete + orphan cleanup)

```mermaid
sequenceDiagram
    participant Client
    participant UseCase as IdentityUserUseCase
    participant Kratos
    participant IAMDB as IAM DB

    rect rgb(245,245,245)
    Note over Client,UseCase: DeleteIdentifier
    Client->>UseCase: DeleteIdentifier(globalUserID, tenantID, kratosUserID, type)
    UseCase->>Kratos: DeleteIdentity(kratosUserID)
    alt Kratos delete fails
        Kratos-->>UseCase: error
        UseCase-->>Client: error (abort)
    else Kratos delete succeeds
        Kratos-->>UseCase: OK
        UseCase->>IAMDB: Soft delete identity (best-effort)
        alt DB delete fails
            IAMDB-->>UseCase: error
            UseCase->>UseCase: Log error
            UseCase-->>Client: success
        else DB delete succeeds
            IAMDB-->>UseCase: OK
            UseCase-->>Client: success
        end
    end
    end

    rect rgb(245,245,245)
    Note over Client,UseCase: Register(identifier)
    Client->>UseCase: Register(email/phone)
    UseCase->>IAMDB: Check existence (deleted_at IS NULL)
    alt Exists
        UseCase->>IAMDB: Load existing identity
        UseCase->>Kratos: GetIdentity(kratos_user_id)
        alt Kratos missing
            UseCase->>IAMDB: Soft delete orphan
            UseCase->>Kratos: SubmitRegistrationFlow
            Kratos-->>UseCase: OK
            UseCase->>IAMDB: Insert new identity
            UseCase-->>Client: VerificationNeeded
        else Kratos present
            UseCase-->>Client: Conflict (already registered)
        end
    else Not exists
        UseCase->>Kratos: SubmitRegistrationFlow
        Kratos-->>UseCase: OK
        UseCase->>IAMDB: Insert new identity
        UseCase-->>Client: VerificationNeeded
    end
    end
```

# Delete Identifier Flow Documentation

---

## Sequence Diagram

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant Identity

    Note over Client,Identity: Delete Identifier
    Client->>API: DELETE /api/v1/users/me/delete-identifier {identifier_type}
    API->>Identity: Delete Identifier
    Identity->>Identity: Validate, ensure >1 identifier
    Identity-->>API: Success
    API-->>Client: 200 OK
    Note over Client,API: If a new identifier was recently added/updated, user must verify it via challenge-verify
```

---

This document describes how an authenticated user deletes an identifier (email or phone). Deletion is only allowed if the user has more than one identifier. If a user recently updated or added a new identifier, they must complete verification via `challenge-verify` first; deletion does not bypass verification.

---

## API Endpoints

### `DELETE /api/v1/users/me/delete-identifier`

Allows an authenticated user to delete an identifier (email or phone) from their account. Cannot delete the only identifier (must retain at least one of email or phone).

#### Headers

- `X-Tenant-Id`: `string` (required)
- `Authorization`: `Bearer {session_token}` (required)

#### Request Body

```json
{
  "identifier_type": "email|phone_number" // The identifier type to delete
}
```

#### Response

```json
{
  "data": {
    "message": "Identifier deleted successfully"
  }
}
```

---

---

## Error Responses

All responses follow the standard error format:

```json
{
  "status": number,
  "code": "string",
  "message": "string"
}
```

### Common Error Codes

| Code | Description |
|------|-------------|
| `MSG_INVALID_TENANT` | Invalid or missing tenant ID |
| `MSG_UNAUTHORIZED` | Missing or invalid session token |
| `MSG_INVALID_PAYLOAD` | Invalid request body |
| `MSG_INVALID_IDENTIFIER_TYPE` | Identifier must be `email` or `phone_number` |
| `MSG_IDENTIFIER_TYPE_NOT_EXISTS` | User does not have an identifier of this type |
| `MSG_CANNOT_DELETE_ONLY_IDENTIFIER` | Cannot delete the only identifier |
| `MSG_GET_IDENTIFIERS_FAILED` | Failed to get user identifiers |
| `MSG_DELETE_IDENTIFIER_FAILED` | Failed to delete identifier |

---

## Example Request

```http
DELETE /api/v1/users/me/delete-identifier
Authorization: Bearer ory_abc.def.ghi
X-Tenant-Id: tenant-123

{
  "identifier_type": "email"
}
```

**Response:**

```json
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "message": "Identifier deleted successfully"
  }
}
```

---

> **Important**: You cannot delete your only identifier. At least one email or phone must remain on your account.
