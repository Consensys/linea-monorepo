# BUGBOT Rules - E2E Tests

## Core Mission

Security-focused code review for E2E test suite to prevent credential exposure.

## Execution Protocol

### 1. Security Logging Guidelines

- **ALWAYS** load and reference [security-logging-guidelines](../../.cursor/rules/security-logging-guidelines/RULE.md)
- Applies to all test files in `e2e/`

### 2. E2E Test-Specific Checks

#### Test Credentials
- [ ] **NEVER** use real API keys, private keys, or tokens in tests
- [ ] Verify all credentials are mock/test values
- [ ] Check that `.env.template` uses placeholders, not real values
- [ ] Ensure test accounts use non-production credentials

```typescript
// WRONG - Real credentials in tests
const PRIVATE_KEY = '0x1234...real-private-key';
const API_ENDPOINT = 'https://api.linea.build?token=real-token-abc123';

// CORRECT - Mock/test credentials
const TEST_PRIVATE_KEY = '0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80'; // Well-known test key
const API_ENDPOINT = process.env.TEST_API_ENDPOINT || 'http://localhost:8545';
```

#### Logging in Tests
- [ ] Verify test logs don't expose environment variables
- [ ] Check console outputs don't print sensitive test data
- [ ] Ensure error logs in tests are sanitized

```typescript
// WRONG
console.log('Test config:', process.env);
console.log(`Testing with account: ${account.privateKey}`);

// CORRECT
console.log('Test config loaded');
console.log(`Testing with account: ${account.address}`);
```

#### API Calls
- [ ] URLs in test logs must be sanitized
- [ ] Request/response logging should redact auth headers
- [ ] Test fixtures should not contain real tokens

```typescript
// WRONG
console.log('Calling API:', fullUrlWithToken);
console.log('Response:', response); // May contain sensitive headers

// CORRECT
console.log('Calling API:', sanitizeUrl(fullUrlWithToken));
console.log('Response status:', response.status);
```

### 3. Environment Configuration

- [ ] `.env.template` must only contain placeholder values
- [ ] No real credentials in committed `.env` files
- [ ] Test setup should validate env vars are set, not log their values

```typescript
// WRONG
console.log('Environment variables:', process.env);

// CORRECT
const requiredVars = ['RPC_URL', 'PRIVATE_KEY'];
requiredVars.forEach(v => {
  if (!process.env[v]) {
    throw new Error(`Missing required env var: ${v}`);
  }
});
console.log('All required environment variables configured');
```

### 4. Account Management

When using test accounts:
- [ ] Log addresses, not private keys
- [ ] Use well-known test private keys for local testing
- [ ] Never commit real account credentials

```typescript
// CORRECT - From environment-based-account-manager.ts pattern
export class EnvironmentBasedAccountManager {
  // Good: Loads from env, doesn't log the private key
  private loadAccount(): Account {
    const privateKey = process.env.PRIVATE_KEY;
    if (!privateKey) throw new Error('PRIVATE_KEY not set');
    // Don't log privateKey here!
    return new Account(privateKey);
  }
  
  // Good: Logs only address
  logAccountInfo(account: Account) {
    console.log(`Using account: ${account.address}`);
  }
}
```

### 5. Test Data & Fixtures

- [ ] Hardcoded test data should use fake values
- [ ] Token amounts and addresses should be obvious test values
- [ ] Comments should not contain real credentials

Use the rules in [security-logging-guidelines](../../.cursor/rules/security-logging-guidelines/RULE.md) to maintain security standards in E2E tests.
