# Update Identifier Flow Testing

## Test Environment
- Base URL: http://localhost:8080
- Test Tenant: c7928076-2cfc-49c3-b7ea-d7519ad52929 (Genetica)
- Test User Email: testuser@example.com
- Test User Phone: +84321339333
- New Email for Update: updated@example.com
- New Phone for Update: +84321339334
- Webhook URL: https://webhook.site/78e3b174-fe28-40ba-93d6-2fde8adc290f

## Test Plan: Full Happy Case Flow

### Step 1: User Registration (Initial Setup) ✅ COMPLETED
First, we need to register a user with an initial identifier (email or phone).

```bash
# Register user with phone
curl -X POST 'http://localhost:8080/api/v1/users/register' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Content-Type: application/json' \
  -d '{
    "phone": "+84321339333"
  }'

# Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "verification_needed": true,
    "verification_flow": {
      "flow_id": "4407c8e3-a494-4f52-900c-99bbf0261a7a",
      "receiver": "+84321339333",
      "challenge_at": 1753597774
    }
  }
}
```

### Step 2: Verify Registration with OTP ❌ BLOCKED
**ISSUE FOUND**: Challenge sessions are expiring very quickly (within seconds instead of 5 minutes)

```bash
# Get OTP from webhook
curl 'https://webhook.site/token/78e3b174-fe28-40ba-93d6-2fde8adc290f/requests?page=1&password=&query=&sorting=newest' \
  -H 'Accept: application/json, text/plain, */*'

# OTP Code: 962609

# Verify registration
curl -X POST "http://localhost:8080/api/v1/users/challenge-verify" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -d '{
    "flow_id": "4407c8e3-a494-4f52-900c-99bbf0261a7a",
    "code": "962609",
    "type": "register"
  }'

# Response: Challenge session not found (expired too quickly)
```

### Step 3: Login to Get Session Token ❌ BLOCKED
Since registration verification is blocked, we cannot proceed to login.

### Step 4: Initiate Update Identifier Flow ❌ BLOCKED
Cannot proceed without successful registration and login.

### Step 5: Verify Update Identifier with OTP ❌ BLOCKED
Cannot proceed without successful update identifier initiation.

## Issues Identified

### 1. Challenge Session Expiration Issue
**Problem**: Challenge sessions are expiring within seconds instead of the configured 5 minutes.

**Root Cause Analysis**:
- `DefaultChallengeDuration = 5 * time.Minute` is set correctly
- Cache configuration shows `DefaultExpiration = 30 * time.Second` 
- Challenge sessions are using the wrong TTL value

**Configuration Issues**:
- Cache type: Using in-memory cache (default)
- Cache TTL: 30 seconds instead of 5 minutes
- Challenge session TTL: Not being applied correctly

**Recommendations**:
1. Fix the cache TTL configuration
2. Ensure challenge sessions use `DefaultChallengeDuration` instead of `DefaultExpiration`
3. Add proper error handling for expired sessions
4. Add session duration information in responses

### 2. Webhook Integration Working ✅
**Status**: Webhook integration is working correctly
- OTP messages are being sent to webhook URL
- Messages contain correct tenant name and OTP codes
- Webhook retrieval is functioning properly

### 3. Registration Flow Working ✅
**Status**: Registration initiation is working correctly
- Flow IDs are generated properly
- Tenant validation is working
- Phone number validation is working

## Test Results Summary

### ✅ Working Components:
1. **Server Health**: Server is running on port 8080
2. **Registration Initiation**: Creates flow IDs correctly
3. **Webhook Integration**: OTP delivery to webhook is working
4. **Tenant Validation**: Correct tenant ID validation
5. **Phone Number Validation**: Proper phone format validation

### ❌ Blocking Issues:
1. **Challenge Session Expiration**: Sessions expire too quickly (seconds vs 5 minutes)
2. **Registration Verification**: Cannot complete due to session expiration
3. **Login Flow**: Cannot test due to incomplete registration
4. **Update Identifier Flow**: Cannot test due to incomplete authentication

## Next Steps

### Immediate Fixes Required:
1. **Fix Cache TTL Configuration**:
   ```bash
   # Set environment variables
   export CACHE_TYPE="redis"  # or fix in-memory cache TTL
   export REDIS_TTL="5m"
   ```

2. **Fix Challenge Session TTL**:
   - Ensure challenge sessions use `DefaultChallengeDuration` (5 minutes)
   - Override `DefaultExpiration` for challenge sessions

3. **Add Better Error Handling**:
   - Provide clear error messages for expired sessions
   - Add session duration information in responses
   - Add retry mechanisms

### Once Fixed, Continue Testing:
1. Complete user registration
2. Test login flow
3. Test update identifier flow
4. Test full happy case scenario

## Configuration Recommendations

### Environment Variables:
```bash
export CACHE_TYPE="redis"
export REDIS_TTL="5m"
export MOCK_WEBHOOK_URL="https://webhook.site/78e3b174-fe28-40ba-93d6-2fde8adc290f"
```

### Code Fixes:
1. Update cache configuration to use correct TTL for challenge sessions
2. Add session duration information in API responses
3. Improve error handling for expired sessions
4. Add retry mechanisms for failed verifications

## Conclusion

The update identifier flow testing revealed a critical issue with challenge session expiration that prevents the complete testing of the authentication flows. While the basic components (registration initiation, webhook integration, tenant validation) are working correctly, the challenge session management needs to be fixed before the full update identifier flow can be tested.

The webhook integration and OTP delivery system are working properly, which is a positive finding. Once the challenge session expiration issue is resolved, the update identifier flow should work as designed. 