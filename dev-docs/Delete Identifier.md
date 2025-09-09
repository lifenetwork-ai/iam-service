# Delete Identifier Flow Documentation

---

## Sequence Diagram

```mermaid
sequenceDiagram
    participant Client
    participant API
    participant IAM
    participant DB
    participant Kratos

    Note over Client,Kratos: Delete Identifier Flow
    Client->>API: DELETE /api/v1/users/me/delete-identifier
    Note right of Client: Header: Authorization: Bearer {session_token}
    Note right of Client: Body: { "identifier_type": "email" }

    API->>IAM: ucase.DeleteIdentifier(...)
    IAM->>IAM: Validate identifier type
    IAM->>DB: List all user identifiers
    IAM->>IAM: Check identifier count > 1
    IAM->>IAM: Find identifier to delete
    IAM->>DB: Delete identifier from database
    IAM->>Kratos: Delete identifier from Kratos (best effort)
    IAM-->>API: Success or error
    API-->>Client: 200 OK or error
```

---

This section describes how an authenticated user can delete an identifier (email or phone) from their account. Deletion is only allowed if the user has more than one identifier.

---

## API Endpoint

### `DELETE /api/v1/users/me/delete-identifier`

Allows an authenticated user to delete an identifier (email or phone) from their account. Cannot delete the only identifier.

#### Headers

- `X-Tenant-Id`: `string` (required)
- `Authorization`: `Bearer {session_token}` (required)

#### Request Body

```json
{
  "identifier_type": "email|phone" // The type of identifier to delete
}
```

#### Response

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

## Flow Logic

| Step | Description |
|------|-------------|
| 1 | Validate identifier type (must be email or phone) |
| 2 | Ensure user has more than one identifier (cannot delete only identifier) |
| 3 | Ensure user has the identifier type to delete |
| 4 | Delete identifier from database |
| 5 | Delete identifier from Kratos (best effort) |
| 6 | Return success |

---

## Error Responses

All responses follow the standard error format:

```json
{
  "status": 400,
  "code": "MSG_INVALID_IDENTIFIER_TYPE",
  "message": "Invalid identifier type"
}
```

### Common Error Codes

| Code | Description |
|------|-------------|
| `MSG_INVALID_TENANT` | Invalid or missing tenant ID |
| `MSG_UNAUTHORIZED` | Missing or invalid session token |
| `MSG_INVALID_PAYLOAD` | Invalid request body |
| `MSG_INVALID_IDENTIFIER_TYPE` | Identifier must be email or phone |
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
