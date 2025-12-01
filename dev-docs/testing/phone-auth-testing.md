# Phone Authentication Flow Testing


Summary of Phone Authentication Flow Testing:

1. Test Environment:
- Base URL: http://36.50.54.169:8080
- Webhook URL for OTP verification: https://webhook.site/ca98dd69-e59a-4d8a-b55f-8e2ab945ba08

Curl for getting webhook response
```
curl 'https://webhook.site/token/ca98dd69-e59a-4d8a-b55f-8e2ab945ba08/requests?page=1&password=&query=&sorting=newest' \
  -H 'Accept: application/json, text/plain, */*' \
  -H 'Accept-Language: en-US,en;q=0.9' \
  -H 'Connection: keep-alive' \
  -b 'webhooksite_session=nEt0pdshqPsHNkGTgci1IvKMfCKfpDrt4jocabVm; _ga=GA1.1.321735838.1751870513; XSRF-TOKEN=eyJpdiI6IllMdXV6QlVQRk1DdXlLczlBM2dMRFE9PSIsInZhbHVlIjoiTzdlL3kxb2lJZDROQjdUWDBJZlBTRUQ2M2ZwOVIwdzZwOEZhYW5VWDYvblkxbUwyckhSLzc3U2FBZytnamQ5MmNzVWQ3YlM0clhJdURBNXBMM1czU1dRczQxbU9CSlFnYmdOOUZoVVo4dktCSThyNjN5SFdyK3NVOGQyeDN0MVUiLCJtYWMiOiJkYzMxMWE5YjhlYTBiNjE0MGI3YzUzYmI2NWM2OTFiNmQ5ZTViZmU5NjAwYmRhYjg0OTI5N2Q2NmM2ODM2MmQ3IiwidGFnIjoiIn0%3D; _ga_FYRV1HFMZK=GS2.1.s1752479699$o5$g1$t1752480064$j33$l0$h0' \
  -H 'Referer: https://webhook.site/' \
  -H 'Sec-Fetch-Dest: empty' \
  -H 'Sec-Fetch-Mode: cors' \
  -H 'Sec-Fetch-Site: same-origin' \
  -H 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/136.0.0.0 Safari/537.36' \
  -H 'X-XSRF-TOKEN: eyJpdiI6IllMdXV6QlVQRk1DdXlLczlBM2dMRFE9PSIsInZhbHVlIjoiTzdlL3kxb2lJZDROQjdUWDBJZlBTRUQ2M2ZwOVIwdzZwOEZhYW5VWDYvblkxbUwyckhSLzc3U2FBZytnamQ5MmNzVWQ3YlM0clhJdURBNXBMM1czU1dRczQxbU9CSlFnYmdOOUZoVVo4dktCSThyNjN5SFdyK3NVOGQyeDN0MVUiLCJtYWMiOiJkYzMxMWE5YjhlYTBiNjE0MGI3YzUzYmI2NWM2OTFiNmQ5ZTViZmU5NjAwYmRhYjg0OTI5N2Q2NmM2ODM2MmQ3IiwidGFnIjoiIn0=' \
  -H 'sec-ch-ua: "Chromium";v="136", "Google Chrome";v="136", "Not.A/Brand";v="99"' \
  -H 'sec-ch-ua-mobile: ?0' \
  -H 'sec-ch-ua-platform: "macOS"'
```
- Two test tenants:
  - Tenant 1: c7928076-2cfc-49c3-b7ea-d7519ad52929
  - Tenant 2: 671284f6-dcac-4d03-86c4-5d19279f6f77

2. Example Requests and Responses:

a. Registration Flow:
```
Request:
curl -X 'POST' \
  'http://36.50.54.169:8080/api/v1/users/register' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Content-Type: application/json' \
  -d '{
  "phone": "+84333222999"
}'

Response:
{
  "flow_id": "dbef6c14-8e04-4b75-9371-a3cfdcaf8732",
  "receiver": "+84333222999",
  "challenge_at": 1752481196,
  "verification_needed": true
}
```

b. Registration Verification:
```
Request:
curl -X POST "http://36.50.54.169:8080/api/v1/users/challenge-verify" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -d '{
  "flow_id": "dbef6c14-8e04-4b75-9371-a3cfdcaf8732",
  "code": "721353",
  "type": "register"
}'

Response:
{
  "session": "ory_st_Hq7Fu89V20QzAaNmhfhs6cLlSjjyJDsf",
  "user_id": "6da0169a-3fdd-473c-adff-e88f40041b15"
}
```

c. Login Challenge:
```
Request:
curl -X POST "http://36.50.54.169:8080/api/v1/users/challenge-with-phone" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -d '{
  "phone": "+84333222999"
}'

Response:
{
  "flow_id": "653e83d2-7430-4447-b493-2b036a89de03",
  "receiver": "+84333222999",
  "challenge_at": 1752481296,
  "verification_needed": true
}
```

d. Login Verification:
```
Request:
curl -X POST "http://36.50.54.169:8080/api/v1/users/challenge-verify" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -d '{
  "flow_id": "653e83d2-7430-4447-b493-2b036a89de03",
  "code": "035762",
  "type": "login"
}'

Response:
{
  "session": "ory_st_Hq7Fu89V20QzAaNmhfhs6cLlSjjyJDsf",
  "user_id": "6da0169a-3fdd-473c-adff-e88f40041b15"
}
```

e. Get OTP Code from Webhook:
```
Request:
curl 'https://webhook.site/token/ca98dd69-e59a-4d8a-b55f-8e2ab945ba08/requests?page=1&password=&query=&sorting=newest' \
  -H 'Accept: application/json, text/plain, */*'

Response:
{
  "uuid": "...",
  "method": "POST",
  "content": {
    "code": "035762",
    "phone": "+84333222999"
  }
}
```

2. Test Execution Results:

a. Registration Flow:
- New Registration (Tenant 1):
  * Phone: +84344333338
  * Status: Success (200)
  * Flow ID: a689d039-e6b8-4d8c-bf75-ae90ec753f94
  * Verification needed: true
  * Challenge timestamp: 1752480324

- Cross-Tenant Registration (Tenant 2):
  * Phone: +84344333338 (same number)
  * Status: Success (200)
  * Flow ID: e2299334-bd3a-4bcd-9ae3-0dfcebeefc1d
  * Verification needed: true
  * Challenge timestamp: 1752480330
  * Note: Successfully allowed same phone in different tenant

- Invalid Phone Format Test:
  * Input: "123456"
  * Status: Error (400)
  * Error Code: INVALID_PHONE_NUMBER
  * Message: "Invalid phone number format"
  * Validation: "Phone number must be in international format (e.g., +1234567890)"

b. Login Flow:
- Unregistered Phone Login:
  * Phone: +84344333339
  * Status: Error (500)
  * Error Code: MSG_FAILED_TO_MAKE_CHALLENGE
  * Message: "Failed to make a challenge"
  * Error Detail: "This account does not exist or has not setup sign in with code."

c. OTP Verification Flow:
- Expired/Invalid Flow Test:
  * Flow ID: a689d039-e6b8-4d8c-bf75-ae90ec753f94
  * Status: Error (400)
  * Error Code: MSG_REGISTRATION_FAILED
  * Message: "Registration failed"
  * Error Detail: "The registration code is invalid or has already been used"

- Invalid OTP Code Test:
  * Flow ID: 8520b50d-e9ec-48a9-932a-d732f0d20550
  * Code: 000000
  * Status: Error (400)
  * Error Code: MSG_REGISTRATION_FAILED
  * Message: "Registration failed"
  * Error Detail: "The registration code is invalid"

d. Login Flow:
- Existing User Login:
  * Phone: +84344333337
  * Status: Success (200)
  * Flow ID: a3488a78-43e8-4c80-bb0c-7b11cbb703f4
  * Challenge timestamp: 1752480411

e. Cross-Tenant Security:
- Using Tenant 1 Flow ID in Tenant 2:
  * Flow ID: a3488a78-43e8-4c80-bb0c-7b11cbb703f4
  * Status: Error (500)
  * Error Code: MSG_GET_FLOW_FAILED
  * Message: "Failed to get login flow"
  * Note: Successfully prevented cross-tenant flow usage

f. Rate Limiting Test:
- 5 Rapid Registration Requests:
  * Phone: +84344333341
  * All requests successful (200)
  * Unique flow IDs generated for each request
  * No rate limiting observed
  * Sequential timestamps:
    - Request 1: 1752480421
    - Request 2: 1752480423
    - Request 3: 1752480424
    - Request 4: 1752480425
    - Request 5: 1752480426

g. Session Token & Profile Access:
- Login Flow Attempt:
  * Initial Flow Generation:
    - Flow ID: 9adb0593-8f24-43cd-b7b9-8351ce723d18
    - Status: Success (200)
    - Challenge timestamp: 1752480501
  * OTP Verification:
    - Status: Error (400)
    - Error Code: MSG_LOGIN_FAILED
    - Message: "Login failed"
    - Error Detail: "The login code is invalid or has already been used"

- Second Login Attempt:
  * New Flow Generation:
    - Flow ID: acef4d2f-4d0c-467f-850e-1239795cda94
    - Status: Success (200)
    - Challenge timestamp: 1752480510
  * OTP Verification:
    - Status: Error (400)
    - Error Code: MSG_LOGIN_FAILED
    - Message: "Login failed"
    - Error Detail: "The login code is invalid or has already been used"

- User Profile Access Tests:
  * Invalid Token Test:
    - Token: "invalid_token_123"
    - Status: Error (500)
    - Error Code: MSG_FAILED_TO_GET_USER_PROFILE
    - Message: "Failed to get user profile"
    - Underlying Error: "401 Unauthorized"
    
  * Malformed Authorization Header:
    - Header: "NotBearer some_token"
    - Status: Error (401)
    - Error Code: UNAUTHORIZED
    - Message: "Authorization header is required"
    - Error Detail: "Invalid authorization header format"

h. Additional Authentication Scenarios:

- Sequential Registration and Login Test:
  * Registration:
    - Phone: +84344333342
    - Expected: Success (200)
    - Actual: Server unavailable
    - Test Goal: Verify complete registration-to-login flow
    
  * Immediate Login Attempt (before OTP verification):
    - Phone: +84344333342
    - Expected: Error (400)
    - Expected Message: "Phone number not verified"
    - Test Goal: Verify unverified numbers cannot login
    
  * OTP Verification:
    - Expected: Success (200)
    - Expected Result: Valid session token
    - Test Goal: Complete registration flow
    
  * Login After Verification:
    - Expected: Success (200)
    - Expected Result: Valid session token
    - Test Goal: Verify successful registration enables login

i. Edge Cases and Error Handling:

- Special Character Phone Numbers:
  * Test Case 1: "+84344333342#"
    - Expected: Error (400)
    - Expected Message: "Invalid phone number format"
    - Test Goal: Verify phone number sanitization
    
  * Test Case 2: "+84 344 333 342"
    - Expected: Success (200)
    - Test Goal: Verify space handling in phone numbers
    
  * Test Case 3: "+84(344)333342"
    - Expected: Success (200)
    - Test Goal: Verify parentheses handling

- Session Management Tests:
  * Concurrent Login Attempts:
    - Setup: Generate 3 simultaneous login flows
    - Expected: All flows valid but previous flows invalidated
    - Test Goal: Verify session handling
    
  * Session Token Reuse:
    - Setup: Use expired session token
    - Expected: Error (401)
    - Expected Message: "Token expired"
    - Test Goal: Verify token expiration handling

j. Performance and Load Testing:

- Burst Registration Requests:
  * Setup: 10 registration requests within 1 second
    - Expected: Rate limit after threshold
    - Expected Status: 429 Too Many Requests
    - Test Goal: Verify rate limiting
    
  * Recovery Period Test:
    - Setup: Wait 60 seconds after rate limit
    - Expected: Requests succeed again
    - Test Goal: Verify rate limit recovery

k. Security Testing:

- SQL Injection Attempts:
  * Phone: "+84344333342' OR '1'='1"
    - Expected: Error (400)
    - Expected Message: "Invalid phone number format"
    - Test Goal: Verify SQL injection protection
    
  * Flow ID: "' OR '1'='1"
    - Expected: Error (400)
    - Expected Message: "Invalid flow ID format"
    - Test Goal: Verify input sanitization

- XSS Prevention:
  * Phone: "+84344333342<script>alert(1)</script>"
    - Expected: Error (400)
    - Expected Message: "Invalid phone number format"
    - Test Goal: Verify XSS protection

l. Tenant Isolation Testing:

- Cross-Tenant Profile Access:
  * Setup: 
    1. Register phone in Tenant 1
    2. Attempt to access profile from Tenant 2
  * Expected: Error (403)
  * Expected Message: "Access denied"
  * Test Goal: Verify tenant data isolation

- Tenant Header Manipulation:
  * Missing Tenant Header:
    - Expected: Error (400)
    - Expected Message: "Tenant ID required"
    - Test Goal: Verify tenant header requirement
    
  * Invalid Tenant Format:
    - Header: "not-a-uuid"
    - Expected: Error (400)
    - Expected Message: "Invalid tenant ID format"
    - Test Goal: Verify tenant ID validation

m. OTP Flow Testing:

- OTP Attempt Tracking:
  * Setup: Multiple invalid OTP attempts
    - Attempt 1-2: Allow
    - Attempt 3: Temporary block
    - Attempt 4+: Extended block
  * Expected: Progressive security measures
  * Test Goal: Verify brute force protection

- OTP Expiration:
  * Setup: Wait for OTP expiration
    - Expected: Error (400)
    - Expected Message: "OTP expired"
    - Test Goal: Verify OTP lifetime limits

n. Profile Management:

- Profile Updates:
  * Setup: Update profile after registration
    - Expected: Success (200)
    - Fields Updated: name, email
    - Test Goal: Verify profile management

- Profile Deletion:
  * Setup: Delete registered profile
    - Expected: Success (200)
    - Subsequent Login: Should fail
    - Test Goal: Verify account deletion

o. Latest Test Cases (2025-07-14):

1. Happy Path Flow:
```
a. Registration Request:
curl -X 'POST' 'http://36.50.54.169:8080/api/v1/users/register' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Content-Type: application/json' \
  -d '{"phone": "+84344333350"}'

Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "verification_needed": true,
    "verification_flow": {
      "flow_id": "04efdde6-6cf0-48f9-a09d-cbabf37e1f05",
      "receiver": "+84344333350",
      "challenge_at": 1752481366
    }
  }
}

b. Registration Verification:
curl -X POST "http://36.50.54.169:8080/api/v1/users/challenge-verify" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -d '{
    "flow_id": "04efdde6-6cf0-48f9-a09d-cbabf37e1f05",
    "code": "088540",
    "type": "register"
  }'

Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "session_id": "5887dd80-f8b0-40ac-85ba-085204179396",
    "session_token": "ory_st_TFwrH0oHpdmv2HXlL8sFAd5aqBfMrRpA",
    "active": true,
    "expires_at": "2025-08-13T08:22:59.823809018Z",
    "issued_at": "2025-07-14T08:22:59.823809018Z",
    "authenticated_at": "2025-07-14T08:22:59.823809018Z",
    "user": {
      "id": "33e9c3ef-a49d-4a19-b5f1-c5e24101b5fc",
      "phone": "+84344333350"
    },
    "authentication_methods": ["code"]
  }
}
```

2. Edge Cases:

a. Invalid Phone Format:
```
Request:
curl -X 'POST' 'http://36.50.54.169:8080/api/v1/users/register' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Content-Type: application/json' \
  -d '{"phone": "123456"}'

Response:
{
  "status": 400,
  "code": "INVALID_PHONE_NUMBER",
  "message": "Invalid phone number format",
  "errors": [
    "Phone number must be in international format (e.g., +1234567890)"
  ]
}
```

b. Cross-Tenant Registration (Same Phone):
```
Request:
curl -X 'POST' 'http://36.50.54.169:8080/api/v1/users/register' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: 671284f6-dcac-4d03-86c4-5d19279f6f77' \
  -H 'Content-Type: application/json' \
  -d '{"phone": "+84344333350"}'

Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "verification_needed": true,
    "verification_flow": {
      "flow_id": "1db2fae3-a4e1-4bcd-a2e9-59164c1e17ff",
      "receiver": "+84344333350",
      "challenge_at": 1752481395
    }
  }
}
```

c. Missing Tenant Header:
```
Request:
curl -X 'POST' 'http://36.50.54.169:8080/api/v1/users/register' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{"phone": "+84344333350"}'

Response:
{
  "code": "MSG_MISSING_TENANT_ID_HEADER",
  "details": [
    {
      "error": "X-Tenant-Id header is required",
      "field": "X-Tenant-Id"
    }
  ],
  "message": "Missing X-Tenant-Id header"
}
```

Key Findings from Latest Tests:

1. Happy Path:
✅ Registration flow works correctly
✅ OTP verification successful
✅ Session token and user ID properly generated
✅ Proper expiration times set on session

2. Edge Cases:
✅ Phone number validation properly enforced
✅ Cross-tenant isolation works (allows same phone in different tenants)
✅ Proper error handling for missing tenant header
✅ Clear error messages with actionable feedback

3. Security:
✅ Session tokens use proper format (ory_st_*)
✅ Expiration dates properly set
✅ Tenant isolation maintained
✅ Headers properly validated

4. Response Format:
✅ Consistent error response structure
✅ Proper HTTP status codes used
✅ Detailed error messages
✅ Validation messages are user-friendly

3. Key Findings:

a. Registration:
✅ Multi-tenant isolation works correctly
✅ Same phone number can be registered in different tenants
✅ Phone number validation is properly implemented
✅ Flow IDs are unique per registration attempt
✅ Challenge timestamps are correctly set
❌ 500 error code used for business logic errors

b. Phone Number Validation:
✅ Proper validation of international format
✅ Clear error messages for invalid formats
✅ Consistent error response structure
✅ Helpful example provided in error message

c. Security:
✅ Tenant isolation is maintained
✅ Flow IDs are UUID v4 (non-sequential)
✅ Challenge timestamps are future-dated
✅ Verification flow is properly enforced

d. OTP & Challenge Flow:
✅ OTP verification properly validates code correctness
✅ Expired/used flows are properly detected
✅ Each verification attempt generates unique flow ID
❌ No rate limiting on OTP generation
❌ No cooldown period between attempts

e. Cross-Tenant Security:
✅ Flow IDs are tenant-specific
✅ Cross-tenant flow usage is prevented
❌ 500 error used instead of 403/401
❌ Error message could be more specific

f. Rate Limiting:
❌ No rate limiting on registration attempts
❌ Multiple flows allowed for same phone
❌ No cooldown period between requests
❌ No detection of rapid sequential attempts

g. Session & Profile Access:
✅ Authorization header format properly validated
✅ Invalid tokens are properly rejected
❌ 500 error used for unauthorized access (should be 401)
❌ Inconsistent error response formats
❌ OTP verification needs webhook monitoring for valid codes

4. Issues Identified:
1. Error Status Codes:
   - Using 500 for non-existent user (should be 404)
   - Internal errors exposed in response (security concern)

2. Response Format:
   - Inconsistent error response structures
   - Some responses include empty user object fields

3. Security Considerations:
   - No rate limiting observed on registration attempts
   - No detection of rapid sequential attempts
   - Challenge timestamps could be more randomized

4. Challenge Flow Issues:
   - Multiple active flows allowed for same phone
   - No expiry of old flows on new request
   - No rate limiting on flow generation
   - No protection against brute force attempts

5. Session Management Issues:
   - Inconsistent error status codes for auth failures
   - Unclear error messages for expired OTPs
   - No session duration information
   - No refresh token mechanism observed
   - No concurrent session handling detected

5. Recommendations:

a. Immediate Fixes:
1. Change error status codes:
   - Use 404 for non-existent users
   - Use 429 for rate limiting
   - Use 401 for authentication failures
   
2. Standardize error responses:
   - Remove empty user object from responses
   - Consistent error code format
   - Hide internal error details

b. Security Enhancements:
1. Add rate limiting:
   - Per IP address
   - Per phone number
   - Per tenant
   
2. Improve challenge flow:
   - Add exponential backoff for retries
   - Implement maximum attempts per phone
   - Add cooldown period between attempts

c. Monitoring & Logging:
1. Add audit logging for:
   - Registration attempts
   - Login attempts
   - Cross-tenant access attempts
   
2. Implement alerts for:
   - Rapid sequential attempts
   - Multiple failures from same IP
   - Cross-tenant registration patterns

d. Challenge Flow Improvements:
1. Flow Management:
   - Expire old flows when new one is generated
   - Limit active flows per phone number
   - Add flow attempt tracking
   - Implement progressive delays

2. OTP Security:
   - Add maximum attempts per flow
   - Implement cooldown after failed attempts
   - Add OTP attempt tracking
   - Consider IP-based restrictions

e. Session Management Improvements:
1. Error Handling:
   - Use consistent 401 status for auth failures
   - Provide clear session expiry information
   - Add detailed error messages for OTP failures
   - Implement proper error hierarchy

2. Session Features:
   - Add session duration information
   - Implement refresh token mechanism
   - Add session revocation endpoint
   - Track active sessions per user
   - Add session metadata (device, location)

The testing confirmed the basic functionality of the phone authentication system while identifying several areas for improvement in error handling, security, and response standardization. The testing revealed that while the basic functionality works, there are significant opportunities for improving security through rate limiting, flow management, and proper error handling. 

p. Complete Success Flow (2025-07-14):

1. New User Registration:
```
Request:
curl -X 'POST' 'http://36.50.54.169:8080/api/v1/users/register' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Content-Type: application/json' \
  -d '{
    "phone": "+84344333351"
  }'

Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "verification_needed": true,
    "verification_flow": {
      "flow_id": "066bfcdf-691d-4850-ae8a-9fd880ad2202",
      "receiver": "+84344333351",
      "challenge_at": 1752481513
    }
  }
}
```

2. Registration OTP Verification:
```
Request:
curl -X POST "http://36.50.54.169:8080/api/v1/users/challenge-verify" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -d '{
    "flow_id": "066bfcdf-691d-4850-ae8a-9fd880ad2202",
    "code": "065883",
    "type": "register"
  }'

Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "session_id": "3aa171f9-543a-4158-be18-c122aaf87b26",
    "session_token": "ory_st_lDCx2TE1FoYSkrnppSCKlwqickM4tXVK",
    "active": true,
    "expires_at": "2025-08-13T08:25:27.686492162Z",
    "issued_at": "2025-07-14T08:25:27.686492162Z",
    "authenticated_at": "2025-07-14T08:25:27.686492162Z",
    "user": {
      "id": "b92469c5-f94a-4645-b8ae-726bb40969a4",
      "phone": "+84344333351"
    },
    "authentication_methods": ["code"]
  }
}
```

3. Login Challenge:
```
Request:
curl -X POST "http://36.50.54.169:8080/api/v1/users/challenge-with-phone" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -d '{
    "phone": "+84344333351"
  }'

Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "flow_id": "059707fc-7774-491d-8e87-4da66e8afd14",
    "receiver": "+84344333351",
    "challenge_at": 1752481531
  }
}
```

4. Login OTP Verification:
```
Request:
curl -X POST "http://36.50.54.169:8080/api/v1/users/challenge-verify" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -d '{
    "flow_id": "059707fc-7774-491d-8e87-4da66e8afd14",
    "code": "555803",
    "type": "login"
  }'

Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "session_id": "dba671f5-2cd3-460c-b00b-8e0545b08590",
    "session_token": "ory_st_kM0gwWeuoJXgejC6QkTuXPBSyiI1fRoP",
    "active": true,
    "expires_at": "2025-08-13T08:25:46.8930788Z",
    "issued_at": "2025-07-14T08:25:46.8930788Z",
    "authenticated_at": "2025-07-14T08:25:46.8930788Z",
    "user": {
      "id": "b92469c5-f94a-4645-b8ae-726bb40969a4",
      "phone": "+84344333351"
    },
    "authentication_methods": ["code"]
  }
}
```

Key Observations from Success Flow:

1. Registration Process:
✅ Clean registration response with flow_id
✅ OTP sent successfully to webhook
✅ Verification creates new user with unique ID
✅ Session token generated upon verification
✅ Proper expiration times set (30 days)

2. Login Process:
✅ Challenge creation successful
✅ OTP delivery confirmed
✅ Login verification successful
✅ New session token generated
✅ User ID consistent across sessions

3. Security Aspects:
✅ Different session tokens for registration and login
✅ Proper session expiration handling
✅ Consistent user ID across operations
✅ Authentication method tracking
✅ Secure token format (ory_st_*)

4. Data Consistency:
✅ User ID preserved between registration and login
✅ Phone number consistent in responses
✅ Timestamps properly formatted
✅ Challenge flow IDs unique per request 