# Permission Handler Testing

Summary of Permission Handler Testing:

Last Test Run: [Current Timestamp]

1. Test Environment:
- Base URL: [REDACTED_IP]
- Two test tenants:
  - Tenant 1 (Genetica): c7928076-2cfc-49c3-b7ea-d7519ad52929
    * Name: genetica
    * Public URL: [REDACTED_URL]
    * Admin URL: [REDACTED_URL]
  - Tenant 2 (Life AI): 671284f6-dcac-4d03-86c4-5d19279f6f77
    * Name: life_ai
    * Public URL: [REDACTED_URL]
    * Admin URL: [REDACTED_URL]
- Default namespace: "user_profile" (Used consistently across all test cases)

3. Test Cases:

a. Basic Permission Creation:
```
Request:
curl -X 'POST' \
  '[REDACTED_IP]/api/v1/permissions/relation-tuples' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -d '{
  "namespace": "user_profile",
  "object": "genetica:doc1",
  "relation": "viewer",
  "subject_id": "user1"
}'

Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": "Relation tuple created successfully"
}
```

b. Different Object and Subject:
```
Request:
curl -X 'POST' \
  '[REDACTED_IP]/api/v1/permissions/relation-tuples' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -d '{
  "namespace": "user_profile",
  "object": "genetica:profile1",
  "relation": "viewer",
  "subject_id": "user2"
}'

Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": "Relation tuple created successfully"
}
```

c. Different Relation Type:
```
Request:
curl -X 'POST' \
  '[REDACTED_IP]/api/v1/permissions/relation-tuples' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -d '{
  "namespace": "user_profile",
  "object": "genetica:doc1",
  "relation": "owner",
  "subject_id": "user1"
}'

Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": "Relation tuple created successfully"
}
```

d. Subject Set Permission:

Step 1: Create Group Membership
```
Request:
curl -X 'POST' \
  '[REDACTED_IP]/api/v1/permissions/relation-tuples' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -d '{
  "namespace": "user_profile",
  "object": "genetica:group1",
  "relation": "member",
  "subject_id": "user1"
}'

Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": "Relation tuple created successfully"
}
```

Step 2: Grant Permission to Group Members
```
Request:
curl -X 'POST' \
  '[REDACTED_IP]/api/v1/permissions/relation-tuples' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -H 'X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929' \
  -d '{
  "namespace": "user_profile",
  "object": "genetica:document123",
  "relation": "viewer",
  "subject_set": {
    "namespace": "user_profile",
    "object": "genetica:group1",
    "relation": "member"
  }
}'

Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": "Relation tuple created successfully"
}
```

Subject Set Permission Details:
1. Group Creation:
   - Namespace: Must be "user_profile"
   - Object: Group identifier (e.g., "genetica:group1")
   - Relation: "member" (defines group membership)
   - Subject: Individual user ID (e.g., "user1")

2. Permission Assignment:
   - Namespace: Must be "user_profile"
   - Object: Resource identifier (e.g., "genetica:document123")
   - Relation: Permission type (e.g., "viewer")
   - Subject Set:
     * Namespace: Must match main namespace ("user_profile")
     * Object: Group identifier used in step 1
     * Relation: Must be "member" to reference group membership

Result:
- User "user1" is a member of "genetica:group1"
- All members of "genetica:group1" get "viewer" access to "genetica:document123"
- Permissions are transitive: new members added to the group automatically get the permissions
```

4. Working Patterns:

a. Namespace:
✅ Use "user_profile" consistently for both main permission and subject sets
✅ Same namespace must be used across all operations

b. Object Naming:
✅ Format: "{tenant}:{resource}" (e.g., "genetica:doc1")
✅ Consistent prefix for all objects (e.g., "genetica:")
✅ Works with different resource types (doc1, profile1, group1)

c. Relations:
✅ Supports multiple relation types (viewer, owner)
✅ Consistent relation naming across permissions

d. Subject Types:
✅ Direct subject_id (e.g., "user1", "user2")
✅ Subject sets with matching namespace
✅ Group membership via subject sets

5. Key Requirements:
- Always use "user_profile" as the namespace
- Always include tenant prefix in object names
- Use consistent relation names
- For subject sets, use the same namespace as the main permission

4. Validation Tests:

a. Input Validation:
✅ Required fields are properly validated
✅ Either subject_id or subject_set must be provided
✅ Namespace, object, and relation are required
✅ Tenant header is required and validated

b. Object Naming:
✅ Basic names accepted (e.g., "test")
✅ Tenant-prefixed names accepted (e.g., "genetica:doc1")
✅ Special characters properly handled
❌ No validation for tenant prefix matching

c. Subject Set Format:
✅ Proper structure validation
✅ Required fields checked
✅ Cross-namespace references allowed
❌ No validation for circular references

6. Integration Status:

a. Keto Service Connection:
✅ Direct Keto write API working properly
✅ All permission creation attempts successful
✅ Basic permissions working
✅ Subject set permissions working

b. Expected vs Actual:
- Expected: 200 Success
- Actual: 200 Success with confirmation message
- Note: All test cases working as expected
- Latest Test Results: All test cases succeeded with proper response format

7. Updated Recommendations:

a. Current Focus:
1. Documentation and Guidelines:
   - Document working permission patterns
   - Create examples for common use cases
   - Add validation rules to API documentation
   - Create user guide for permission management

2. Testing Coverage:
   - Add automated tests for all working patterns
   - Include edge cases in test suite
   - Document test scenarios

b. Code Improvements:
1. Input Validation:
   - Add specific validation for subject set format
   - Validate namespace consistency
   - Add proper error messages for invalid inputs

2. Documentation Updates:
   - Document successful permission patterns
   - Add examples of working configurations
   - Include troubleshooting guide

c. Next Steps:
1. Immediate Actions:
   - Create comprehensive test suite
   - Add more example use cases
   - Document best practices

2. Follow-up Tasks:
   - Implement automated testing
   - Add monitoring for permission operations
   - Create user documentation 