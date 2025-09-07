# Testing & Error Assertion Best Practices

## Error Assertions with Custom Error Types in Go

When using custom error types (such as *DomainError), always use assert.Nil(t, err) for success cases in your tests, rather than assert.NoError(t, err). This is because a zero-value custom error struct is not nil in Go, and assert.NoError will fail even if the error is a zero-value struct.

**Pattern:**

```go
// For success cases:
assert.Nil(t, err) // Passes if err is nil or a zero-value custom error

// For failure cases:
assert.Error(t, err)
```

This approach ensures your tests are robust and do not fail due to Go’s handling of non-nil zero-value error structs.


