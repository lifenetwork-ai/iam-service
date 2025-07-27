# Update Identifier Flow Testing

## Test Environment
- Base URL: http://localhost:8080
- Test Tenant: c7928076-2cfc-49c3-b7ea-d7519ad52929 (Genetica)
- Test User Email: testuser@example.com
- Test User Phone: +84321339333
- New Email for Update: updated@example.com
- New Phone for Update: +84321339334
- Webhook URL: https://webhook.site/50c8ecda-46f3-439f-ab7b-9ae576ecde23

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
    "phone": "+84321339666"
  }'

# Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "verification_needed": true,
    "verification_flow": {
      "flow_id": "4413975e-2c3a-45fe-9202-6f1082aa7ba5",
      "receiver": "+84321339333",
      "challenge_at": 1753604648
    }
  }
}
```

### Step 2: Verify Registration with OTP ✅ COMPLETED
**STATUS**: Challenge sessions are now working correctly with proper TTL

```bash
# Get OTP from webhook
curl 'https://webhook.site/token/78e3b174-fe28-40ba-93d6-2fde8adc290f/requests?page=1&password=&query=&sorting=newest' \
  -H 'Accept: application/json, text/plain, */*'

# OTP Code: 469346

# Verify registration
curl -X POST "http://localhost:8080/api/v1/users/challenge-verify" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -d '{
    "flow_id": "02e8078e-66c7-4ae5-899c-87d202594725",
    "code": "469346",
    "type": "register"
  }'

# Response: ✅ SUCCESS
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "session_id": "bf378437-836c-4643-947f-ac30850ae16e",
    "session_token": "ory_st_g1exfvdt4fudEAGQ0mmMudlcJc2Flv4I",
    "active": true,
    "expires_at": "2025-07-27T08:39:31.498452308Z",
    "issued_at": "2025-07-27T08:24:31.498452308Z",
    "authenticated_at": "2025-07-27T08:24:31.498452308Z",
    "user": {
      "id": "dccaea90-875f-411e-be82-b787be0d1609",
      "phone": "+84321339333"
    },
    "authentication_methods": ["code"]
  }
}
```

### Step 3: Login to Get Session Token ✅ COMPLETED
**STATUS**: Login flow working correctly

```bash
# Login challenge
curl -X POST 'http://localhost:8080/api/v1/users/challenge-with-phone' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Content-Type: application/json' \
  -d '{"phone": "+84321339333"}'

# Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "flow_id": "20305310-6693-4978-8f8c-e2da827e1956",
    "receiver": "+84321339333",
    "challenge_at": 1753604685
  }
}

# Get OTP and verify login
# OTP Code: 550053

curl -X POST "http://localhost:8080/api/v1/users/challenge-verify" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -d '{
    "flow_id": "20305310-6693-4978-8f8c-e2da827e1956",
    "code": "550053",
    "type": "login"
  }'

# Response: ✅ SUCCESS
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "session_id": "b1fec427-7fb8-4b96-ae95-7f8c4b6ba714",
    "session_token": "ory_st_o8TOhvej1M5McSDb88Zb73es4ihhnOkj",
    "active": true,
    "expires_at": "2025-07-27T08:39:57.704853552Z",
    "issued_at": "2025-07-27T08:24:57.704853552Z",
    "authenticated_at": "2025-07-27T08:24:57.704853552Z",
    "user": {
      "id": "dccaea90-875f-411e-be82-b787be0d1609",
      "phone": "+84321339333"
    },
    "authentication_methods": ["code"]
  }
}
```

### Step 4: Initiate Update Identifier Flow ✅ COMPLETED
**STATUS**: Update identifier flow working correctly

```bash
# Update phone number
curl -X POST 'http://localhost:8080/api/v1/users/me/update-identifier' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Authorization: Bearer ory_st_o8TOhvej1M5McSDb88Zb73es4ihhnOkj' \
  -H 'Content-Type: application/json' \
  -d '{
    "new_identifier": "+84321339334",
    "identifier_type": "phone_number"
  }'

# Response: ✅ SUCCESS
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "flow_id": "b252329d-39cf-4088-a0c1-efde2ffe3d8f",
    "receiver": "+84321339334",
    "challenge_at": 1753604706
  }
}
```

### Step 5: Verify Update Identifier with OTP ✅ COMPLETED
**STATUS**: Update identifier verification working correctly

```bash
# Get OTP from webhook
# OTP Code: 273837

# Verify update identifier
curl -X POST "http://localhost:8080/api/v1/users/challenge-verify" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -d '{
    "flow_id": "b252329d-39cf-4088-a0c1-efde2ffe3d8f",
    "code": "273837",
    "type": "register"
  }'

# Response: ✅ SUCCESS
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "session_id": "fc34ff1a-bbb7-418d-bbec-ce4a81bf174f",
    "session_token": "ory_st_YppSGSAEnZoDQptFOOiMirRi47WTG2o9",
    "active": true,
    "expires_at": "2025-07-27T08:40:17.488897884Z",
    "issued_at": "2025-07-27T08:25:17.488897884Z",
    "authenticated_at": "2025-07-27T08:25:17.488897884Z",
    "user": {
      "id": "26643070-9634-4dbc-a4d2-3153c2e65234",
      "phone": "+84321339334"
    },
    "authentication_methods": ["code"]
  }
}
```

### Step 6: Verify the Update Worked ✅ COMPLETED
**STATUS**: New phone number works for login

```bash
# Login with new phone number
curl -X POST 'http://localhost:8080/api/v1/users/challenge-with-phone' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Content-Type: application/json' \
  -d '{"phone": "+84321339334"}'

# Response: ✅ SUCCESS
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "flow_id": "58df42b5-0580-4021-b387-57a179e2ae33",
    "receiver": "+84321339334",
    "challenge_at": 1753604723
  }
}

# Get OTP and verify login with new phone
# OTP Code: 100927

curl -X POST "http://localhost:8080/api/v1/users/challenge-verify" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -d '{
    "flow_id": "58df42b5-0580-4021-b387-57a179e2ae33",
    "code": "100927",
    "type": "login"
  }'

# Response: ✅ SUCCESS
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "session_id": "49f2c475-b76b-4684-8411-2e33fb9a7ff7",
    "session_token": "ory_st_sO5nmIqtzim3Q5vxUg5poYqUjdqVbQVx",
    "active": true,
    "expires_at": "2025-07-27T08:40:44.436021186Z",
    "issued_at": "2025-07-27T08:25:44.436021186Z",
    "authenticated_at": "2025-07-27T08:25:44.436021186Z",
    "user": {
      "id": "26643070-9634-4dbc-a4d2-3153c2e65234",
      "phone": "+84321339334"
    },
    "authentication_methods": ["code"]
  }
}
```

### Step 7: Verify User Profile ✅ COMPLETED
**STATUS**: User profile shows updated phone number

```bash
# Get user profile
curl -X GET 'http://localhost:8080/api/v1/users/me' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Authorization: Bearer ory_st_sO5nmIqtzim3Q5vxUg5poYqUjdqVbQVx'

# Response: ✅ SUCCESS
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "global_user_id": "ba788764-8762-49dd-b9c6-a015d75e9608",
    "id": "26643070-9634-4dbc-a4d2-3153c2e65234",
    "phone": "+84321339334",
    "tenant": "genetica"
  }
}
```

## Test Results Summary

### ✅ All Components Working:
1. **Server Health**: Server is running on port 8080 ✅
2. **Registration Initiation**: Creates flow IDs correctly ✅
3. **Registration Verification**: OTP verification working ✅
4. **Webhook Integration**: OTP delivery to webhook is working ✅
5. **Tenant Validation**: Correct tenant ID validation ✅
6. **Phone Number Validation**: Proper phone format validation ✅
7. **Login Flow**: Login with phone and OTP working ✅
8. **Update Identifier Flow**: Update identifier initiation working ✅
9. **Update Identifier Verification**: OTP verification for update working ✅
10. **New Identifier Login**: Login with updated phone number working ✅
11. **User Profile**: Profile shows updated phone number ✅

### ❌ Issues Identified:

#### 1. Old Phone Number Still Works After Update
**Problem**: After updating the phone number from `+84321339333` to `+84321339334`, the old phone number still works for login challenges.

**Test Evidence**:
```bash
# After updating phone number, old phone still works for login
curl -X POST 'http://localhost:8080/api/v1/users/challenge-with-phone' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Content-Type: application/json' \
  -d '{"phone": "+84321339333"}'

# Response: ✅ SUCCESS (This should NOT work after update)
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "flow_id": "ed3355cd-5036-43a8-bff2-26c8130e3241",
    "receiver": "+84321339333",
    "challenge_at": 1753604749
  }
}
```

**Expected Behavior**: After updating a phone number, the old phone number should no longer work for login challenges.

**Root Cause Analysis**:
- The update identifier flow appears to be adding the new identifier without properly removing the old one
- The system may be treating this as an "add identifier" rather than a "replace identifier" operation
- The old identifier remains active in the authentication system

**Recommendations**:
1. **Fix Update Identifier Logic**: Ensure the update identifier flow properly removes the old identifier
2. **Add Identifier Replacement Logic**: Implement proper replacement instead of addition
3. **Update Kratos Integration**: Ensure Kratos identity traits are properly updated
4. **Add Validation**: Prevent login with old identifiers after update
5. **Database Cleanup**: Ensure old identifier records are properly removed from the database

## Conclusion

The update identifier flow testing revealed that while most components are working correctly, there is a **CRITICAL SECURITY ISSUE** that needs to be addressed:

### ✅ **Working Components:**
1. **Registration Flow**: User registration and verification working
2. **Login Flow**: Login with phone and OTP working
3. **Update Identifier Flow**: Update identifier initiation and verification working
4. **New Identifier Login**: Login with updated phone number working
5. **User Profile**: Profile correctly shows updated phone number
6. **Webhook Integration**: OTP delivery and retrieval working perfectly

### ❌ **Critical Security Issue:**
**Old Phone Number Still Works After Update**: After updating a phone number, the old phone number remains active for login challenges. This is a security vulnerability as users who have updated their phone numbers should not be able to login with their old phone numbers.

### 🔧 **Required Fixes:**
1. **Implement Proper Identifier Replacement**: The update identifier flow should replace the old identifier, not add a new one
2. **Remove Old Identifier from Authentication System**: Ensure old identifiers are properly deactivated
3. **Update Kratos Identity Traits**: Properly update the identity traits in Kratos to reflect the change
4. **Database Cleanup**: Remove old identifier records from the database
5. **Add Validation**: Prevent login attempts with old identifiers after update

### 📊 **Overall Status:**
- **Functionality**: ✅ MOSTLY FUNCTIONAL
- **Security**: ❌ CRITICAL ISSUE - Old identifiers remain active
- **User Experience**: ⚠️ CONFUSING - Users can login with old phone numbers

**Priority**: **HIGH** - This security issue should be fixed before production deployment. 

## Retest Results - Bug Verification (2025-01-27)

### Test Summary
**STATUS**: ❌ **BUG STILL EXISTS** - The critical security issue where old phone numbers remain active after update has NOT been fixed.

### Test Flow Executed

#### Step 1: Login with Existing Phone ✅
```bash
# Login with phone +84321339334
curl -X POST 'http://localhost:8080/api/v1/users/challenge-with-phone' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Content-Type: application/json' \
  -d '{"phone": "+84321339334"}'

# Response: ✅ SUCCESS
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "flow_id": "89428587-6187-4ac5-a689-ef6fa9e13e0b",
    "receiver": "+84321339334",
    "challenge_at": 1753605688
  }
}

# OTP: 917947
# Login verification successful
# Session token: ory_st_1Rt4NYPvIOA2wFYQ7rRqAGkcKzmKOdZq
```

#### Step 2: Update Identifier ✅
```bash
# Update phone to +84321339338
curl -X POST 'http://localhost:8080/api/v1/users/me/update-identifier' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Authorization: Bearer ory_st_1Rt4NYPvIOA2wFYQ7rRqAGkcKzmKOdZq' \
  -H 'Content-Type: application/json' \
  -d '{
    "new_identifier": "+84321339338",
    "identifier_type": "phone_number"
  }'

# Response: ✅ SUCCESS
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "flow_id": "0721b86e-d2c1-4448-af14-d7a5ec7179fb",
    "receiver": "+84321339338",
    "challenge_at": 1753605707
  }
}

# OTP: 483453
# Update verification successful
# New session token: ory_st_AdsyQtlgdj4N13inNsA4OPDKe4JFwRst
```

#### Step 3: Verify User Profile ✅
```bash
# Get user profile
curl -X GET 'http://localhost:8080/api/v1/users/me' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Authorization: Bearer ory_st_AdsyQtlgdj4N13inNsA4OPDKe4JFwRst'

# Response: ✅ SUCCESS
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "global_user_id": "ba788764-8762-49dd-b9c6-a015d75e9608",
    "id": "ab1c352b-180c-4000-9fb9-96626de603b6",
    "phone": "+84321339338",
    "tenant": "genetica"
  }
}
```

#### Step 4: Test New Phone Number ✅
```bash
# Login with new phone number
curl -X POST 'http://localhost:8080/api/v1/users/challenge-with-phone' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Content-Type: application/json' \
  -d '{"phone": "+84321339338"}'

# Response: ✅ SUCCESS
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "flow_id": "9f7ebba1-9df2-4fdc-8e20-efc6624e8c78",
    "receiver": "+84321339338",
    "challenge_at": 1753605729
  }
}
```

#### Step 5: Test Old Phone Number ❌ **BUG CONFIRMED**
```bash
# Login with OLD phone number (should NOT work)
curl -X POST 'http://localhost:8080/api/v1/users/challenge-with-phone' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Content-Type: application/json' \
  -d '{"phone": "+84321339334"}'

# Response: ❌ BUG - Old phone still works!
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "flow_id": "1ef0cd7a-b431-4179-a8e4-e65f188246e5",
    "receiver": "+84321339334",
    "challenge_at": 1753605723
  }
}
```

### Retest Results Summary

#### ✅ **Working Components:**
1. **Login Flow**: Login with existing phone number working
2. **Update Identifier Flow**: Update identifier initiation and verification working
3. **User Profile**: Profile correctly shows updated phone number
4. **New Identifier Login**: Login with updated phone number working
5. **Webhook Integration**: OTP delivery and retrieval working

#### ❌ **Critical Security Issue Still Present:**
**Old Phone Number Still Works After Update**: After updating the phone number from `+84321339334` to `+84321339338`, the old phone number `+84321339334` still works for login challenges.

**Evidence**:
- User profile shows: `"phone": "+84321339338"`
- New phone login: ✅ Works
- Old phone login: ❌ **Still works** (should be blocked)

### Conclusion

**🚨 CRITICAL SECURITY ISSUE NOT FIXED**

The update identifier flow testing confirms that the **critical security vulnerability** identified in the previous test is still present:

1. **Update Process**: The update identifier flow works correctly and updates the user profile
2. **New Identifier**: The new phone number works for login as expected
3. **Old Identifier**: ❌ **The old phone number still works for login** - This is a security vulnerability

**Impact**: Users who have updated their phone numbers can still login with their old phone numbers, which defeats the purpose of the update and creates a security risk.

**Recommendation**: This issue needs to be fixed immediately before production deployment. The update identifier flow should properly remove/deactivate old identifiers when updating to new ones. 

## Latest Retest Results - Bug Verification (2025-01-27 - Second Test)

### Test Summary
**STATUS**: ❌ **BUG STILL EXISTS** - The critical security issue where old phone numbers remain active after update has NOT been fixed in the latest retest.

### Test Flow Executed

#### Step 1: Verify Existing Session ✅
```bash
# Check if previous session token is still valid
curl -X GET 'http://localhost:8080/api/v1/users/me' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Authorization: Bearer ory_st_AdsyQtlgdj4N13inNsA4OPDKe4JFwRst'

# Response: ✅ SUCCESS - Session still valid
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "global_user_id": "ba788764-8762-49dd-b9c6-a015d75e9608",
    "id": "ab1c352b-180c-4000-9fb9-96626de603b6",
    "phone": "+84321339338",
    "tenant": "genetica"
  }
}
```

#### Step 2: Test Old Phone Number ❌ **BUG CONFIRMED**
```bash
# Login with OLD phone number (should NOT work)
curl -X POST 'http://localhost:8080/api/v1/users/challenge-with-phone' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Content-Type: application/json' \
  -d '{"phone": "+84321339334"}'

# Response: ❌ BUG - Old phone still works!
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "flow_id": "d94661d6-2721-467b-8ccc-2b1962edf85f",
    "receiver": "+84321339334",
    "challenge_at": 1753606013
  }
}
```

#### Step 3: Test New Phone Number ✅
```bash
# Login with new phone number
curl -X POST 'http://localhost:8080/api/v1/users/challenge-with-phone' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Content-Type: application/json' \
  -d '{"phone": "+84321339338"}'

# Response: ✅ SUCCESS
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "flow_id": "9420e62b-643d-481a-8b5d-716eed1e2760",
    "receiver": "+84321339338",
    "challenge_at": 1753606019
  }
}
```

### Latest Retest Results Summary

#### ✅ **Working Components:**
1. **Session Management**: Previous session tokens remain valid
2. **User Profile**: Profile correctly shows updated phone number
3. **New Identifier Login**: Login with updated phone number working
4. **Webhook Integration**: OTP delivery and retrieval working

#### ❌ **Critical Security Issue Still Present:**
**Old Phone Number Still Works After Update**: After updating the phone number from `+84321339334` to `+84321339338`, the old phone number `+84321339334` still works for login challenges.

**Evidence**:
- User profile shows: `"phone": "+84321339338"`
- New phone login: ✅ Works
- Old phone login: ❌ **Still works** (should be blocked)

### Final Conclusion

**🚨 CRITICAL SECURITY ISSUE STILL NOT FIXED**

The latest retest confirms that the **critical security vulnerability** is still present:

1. **User Profile**: Correctly shows updated phone number `+84321339338`
2. **New Identifier**: Works for login as expected
3. **Old Identifier**: ❌ **Still works for login** - This is a security vulnerability

**Impact**: Users who have updated their phone numbers can still login with their old phone numbers, which defeats the purpose of the update and creates a security risk.

**Recommendation**: This issue needs to be fixed immediately before production deployment. The update identifier flow should properly remove/deactivate old identifiers when updating to new ones. 

## Third Retest Results - Bug Verification (2025-01-27 - Third Test)

### Test Summary
**STATUS**: ❌ **BUG STILL EXISTS** - The critical security issue where old phone numbers remain active after update has NOT been fixed in the third retest.

### Test Flow Executed

#### Step 1: Verify Existing Session ✅
```bash
# Check if previous session token is still valid
curl -X GET 'http://localhost:8080/api/v1/users/me' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Authorization: Bearer ory_st_AdsyQtlgdj4N13inNsA4OPDKe4JFwRst'

# Response: ✅ SUCCESS - Session still valid
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "global_user_id": "ba788764-8762-49dd-b9c6-a015d75e9608",
    "id": "ab1c352b-180c-4000-9fb9-96626de603b6",
    "phone": "+84321339338",
    "tenant": "genetica"
  }
}
```

#### Step 2: Test Old Phone Number ❌ **BUG CONFIRMED**
```bash
# Login with OLD phone number (should NOT work)
curl -X POST 'http://localhost:8080/api/v1/users/challenge-with-phone' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Content-Type: application/json' \
  -d '{"phone": "+84321339334"}'

# Response: ❌ BUG - Old phone still works!
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "flow_id": "a0ee8222-881f-4cf3-ab64-d0e0099ee17c",
    "receiver": "+84321339334",
    "challenge_at": 1753606222
  }
}
```

#### Step 3: Test New Phone Number ✅
```bash
# Login with new phone number
curl -X POST 'http://localhost:8080/api/v1/users/challenge-with-phone' \
  -H 'accept: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -H 'Content-Type: application/json' \
  -d '{"phone": "+84321339338"}'

# Response: ✅ SUCCESS
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "flow_id": "bf68e97c-17e6-495c-b8bb-33da7929eb51",
    "receiver": "+84321339338",
    "challenge_at": 1753606230
  }
}
```

### Third Retest Results Summary

#### ✅ **Working Components:**
1. **Session Management**: Previous session tokens remain valid
2. **User Profile**: Profile correctly shows updated phone number
3. **New Identifier Login**: Login with updated phone number working
4. **Webhook Integration**: OTP delivery and retrieval working

#### ❌ **Critical Security Issue Still Present:**
**Old Phone Number Still Works After Update**: After updating the phone number from `+84321339334` to `+84321339338`, the old phone number `+84321339334` still works for login challenges.

**Evidence**:
- User profile shows: `"phone": "+84321339338"`
- New phone login: ✅ Works
- Old phone login: ❌ **Still works** (should be blocked)

### Final Conclusion

**🚨 CRITICAL SECURITY ISSUE STILL NOT FIXED**

The third retest confirms that the **critical security vulnerability** is still present:

1. **User Profile**: Correctly shows updated phone number `+84321339338`
2. **New Identifier**: Works for login as expected
3. **Old Identifier**: ❌ **Still works for login** - This is a security vulnerability

**Impact**: Users who have updated their phone numbers can still login with their old phone numbers, which defeats the purpose of the update and creates a security risk.

**Recommendation**: This issue needs to be fixed immediately before production deployment. The update identifier flow should properly remove/deactivate old identifiers when updating to new ones. 