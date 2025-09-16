# Permission Handler Testing

Summary of Permission Handler Testing:

Last Test Run: 2025-07-22T18:30:00Z

1. Test Environment:
- Base URL: http://localhost:8080
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
- Test User: tdtuan1702@gmail.com
- Test Phone: +84331339331

2. Test Cases:

a. Basic Permission Creation:
```
Request:
curl -X POST http://localhost:8080/api/v1/permissions/relation-tuples \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -H "Authorization: Bearer ory_st_hdwHaRkFUdXf1NMWhMT6etllviNoRmsp" \
  -d '{
    "namespace": "user_profile",
    "object": "genetica:document123",
    "relation": "viewer",
    "identifier": "tdtuan1702@gmail.com"
  }'

Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": "Relation tuple created successfully"
}
```

b. Permission Self-Check:
```
Request:
curl -X POST http://localhost:8080/api/v1/permissions/self-check \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -H "Authorization: Bearer ory_st_hdwHaRkFUdXf1NMWhMT6etllviNoRmsp" \
  -d '{
    "namespace": "user_profile",
    "object": "genetica:testdoc",
    "relation": "viewer"
  }'

Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "allowed": true
  }
}
```

c. Permission Check (Cross-User):
```
Request:
curl -X POST http://localhost:8080/api/v1/permissions/check \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -d '{
    "namespace": "user_profile",
    "object": "user_profile:testdoc",
    "relation": "viewer",
    "tenant_member": {
      "tenant_id": "c7928076-2cfc-49c3-b7ea-d7519ad52929",
      "identifier": "+84331339331"
    }
  }'

Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": {
    "allowed": true
  }
}
```

d. Delegation Access:
```
Request:
curl -X POST http://localhost:8080/api/v1/permissions/delegate \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -H "Authorization: Bearer ory_st_hdwHaRkFUdXf1NMWhMT6etllviNoRmsp" \
  -d '{
    "resource_type": "user_profile",
    "resource_id": "testdoc",
    "permission": "viewer",
    "tenant_id": "c7928076-2cfc-49c3-b7ea-d7519ad52929",
    "identifier": "+84331339331"
  }'

Response:
{
  "status": 200,
  "code": "MSG_SUCCESS",
  "message": "Success",
  "data": true
}
```

e. Comprehensive Test Flow:
```
Step 1: Create Relation Tuple
curl -X POST http://localhost:8080/api/v1/permissions/relation-tuples \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -H "Authorization: Bearer ory_st_hdwHaRkFUdXf1NMWhMT6etllviNoRmsp" \
  -d '{
    "namespace": "user_profile",
    "object": "genetica:testdoc",
    "relation": "viewer",
    "identifier": "tdtuan1702@gmail.com"
  }'

Step 2: Check Permission (Self)
curl -X POST http://localhost:8080/api/v1/permissions/self-check \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -H "Authorization: Bearer ory_st_hdwHaRkFUdXf1NMWhMT6etllviNoRmsp" \
  -d '{
    "namespace": "user_profile",
    "object": "genetica:testdoc",
    "relation": "viewer"
  }'

Step 3: Create Delegate Permission
curl -X POST http://localhost:8080/api/v1/permissions/relation-tuples \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -H "Authorization: Bearer ory_st_hdwHaRkFUdXf1NMWhMT6etllviNoRmsp" \
  -d '{
    "namespace": "user_profile",
    "object": "user_profile:testdoc",
    "relation": "delegate",
    "identifier": "tdtuan1702@gmail.com"
  }'

Step 4: Delegate Access
curl -X POST http://localhost:8080/api/v1/permissions/delegate \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -H "Authorization: Bearer ory_st_hdwHaRkFUdXf1NMWhMT6etllviNoRmsp" \
  -d '{
    "resource_type": "user_profile",
    "resource_id": "testdoc",
    "permission": "viewer",
    "tenant_id": "c7928076-2cfc-49c3-b7ea-d7519ad52929",
    "identifier": "+84331339331"
  }'

Step 5: Check Delegated Permission
curl -X POST http://localhost:8080/api/v1/permissions/check \
  -H "Content-Type: application/json" \
  -H "X-Tenant-Id: c7928076-2cfc-49c3-b7ea-d7519ad52929" \
  -d '{
    "namespace": "user_profile",
    "object": "user_profile:testdoc",
    "relation": "viewer",
    "tenant_member": {
      "tenant_id": "c7928076-2cfc-49c3-b7ea-d7519ad52929",
      "identifier": "+84331339331"
    }
  }'
```

3. Working Patterns:

a. Namespace:
✅ Use "user_profile" consistently for both main permission and subject sets
✅ Same namespace must be used across all operations

b. Object Naming:
✅ Format: "{tenant}:{resource}" (e.g., "genetica:doc1")
✅ For delegation checks: "user_profile:{resource}" format
✅ Consistent prefix for all objects (e.g., "genetica:")
✅ Works with different resource types (doc1, profile1, group1)

c. Relations:
✅ Supports multiple relation types (viewer, owner, editor, delegate)
✅ Consistent relation naming across permissions

d. Subject Types:
✅ Direct identifier (e.g., "tdtuan1702@gmail.com", "+84331339331")
✅ Email and phone number identifiers supported
✅ Cross-user permission checks working

e. Delegation:
✅ Requires "delegate" permission on resource
✅ Supports different permission types (viewer, editor, owner)
✅ Target user must exist in system
✅ Creates proper relation tuples for delegated permissions

4. Key Requirements:
- Always use "user_profile" as the namespace
- Always include tenant prefix in object names
- Use consistent relation names
- For delegation, ensure delegate permission exists first
- Target users must be registered in the system

5. Validation Tests:

a. Input Validation:
✅ Required fields are properly validated
✅ Namespace, object, and relation are required
✅ Tenant header is required and validated
✅ Authorization header required for authenticated endpoints

b. Object Naming:
✅ Basic names accepted (e.g., "test")
✅ Tenant-prefixed names accepted (e.g., "genetica:doc1")
✅ Special characters properly handled
✅ Delegation object format: "user_profile:{resource}"

c. Permission Checks:
✅ Self-check working for authenticated users
✅ Cross-user permission checks working
✅ Delegation permission validation working
✅ Proper error responses for invalid permissions

6. Integration Status:

a. Keto Service Connection:
✅ Direct Keto write API working properly
✅ All permission creation attempts successful
✅ Basic permissions working
✅ Delegation permissions working
✅ Permission checks working

b. Expected vs Actual:
- Expected: 200 Success for all operations
- Actual: 200 Success with proper response format
- Note: All test cases working as expected
- Latest Test Results: Comprehensive delegation flow working

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
   - Add specific validation for delegation format
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