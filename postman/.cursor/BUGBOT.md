# BUGBOT Rules - Postman/API Testing

## Core Mission

Security enforcement for API testing scripts and collections.

## Execution Protocol

### 1. Security Logging Guidelines

- **ALWAYS** load and reference [security-logging-guidelines](../../.cursor/rules/security-logging-guidelines/RULE.md)
- Applies to all Postman scripts and API testing code

### 2. API Testing-Specific Checks

#### Request Logging
- [ ] **NEVER** log full request URLs with tokens in query parameters
- [ ] Request headers containing auth tokens must be redacted
- [ ] API keys in requests should not be logged

```typescript
// WRONG - Logging request with sensitive data
console.log('Making request:', {
  url: fullUrl, // May contain ?token=...
  headers: headers // May contain Authorization header
});

// CORRECT - Safe request logging
console.log('Making request to:', sanitizeUrl(fullUrl));
console.log('Headers configured');
```

#### Response Handling
- [ ] Response logging should avoid auth tokens in response bodies
- [ ] Verify response headers don't leak session tokens
- [ ] Sensitive fields in JSON responses should be redacted

```typescript
// WRONG
console.log('Response:', response);
console.log('Response headers:', response.headers);

// CORRECT
console.log('Response status:', response.status);
console.log('Response received successfully');
// Only log specific safe fields if needed
```

#### Environment Variables
- [ ] Postman environment files must not contain real credentials
- [ ] Variables should use placeholders: `{{API_KEY}}`, `{{TOKEN}}`
- [ ] Test environments should be clearly marked as non-production

```json
// WRONG - Real credentials in environment file
{
  "API_KEY": "sk-real-api-key-abc123",
  "AUTH_TOKEN": "real-bearer-token-xyz"
}

// CORRECT - Placeholder values
{
  "API_KEY": "{{YOUR_API_KEY}}",
  "AUTH_TOKEN": "{{YOUR_AUTH_TOKEN}}",
  "NOTE": "Replace placeholders with your actual test credentials"
}
```

#### Test Scripts
- [ ] Pre-request scripts should not log sensitive variables
- [ ] Test assertions can verify status codes, not credential values
- [ ] Collection variables must use safe test values

```typescript
// WRONG - In Postman test script
pm.test("API key works", function() {
  console.log("Testing with API key:", pm.environment.get("API_KEY"));
});

// CORRECT
pm.test("API key works", function() {
  pm.response.to.have.status(200);
  console.log("Authentication successful");
});
```

### 3. Collection Configuration

- [ ] Collection exports should not include sensitive auth
- [ ] Shared collections must use environment variables, not hardcoded secrets
- [ ] Documentation should reference placeholder values

### 4. Mock Servers & Stubs

- [ ] Mock responses should use fake tokens and credentials
- [ ] Test data should be obviously non-production
- [ ] Mock configurations should not mirror real infrastructure

```typescript
// CORRECT - Mock data
const mockResponse = {
  token: "mock-jwt-token-for-testing-" + randomString(),
  apiKey: "test-api-key-12345",
  userId: "test-user-001"
};
```

Use the rules in [security-logging-guidelines](../../.cursor/rules/security-logging-guidelines/RULE.md) to maintain security in API testing.
