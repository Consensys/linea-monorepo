---
description: Security-Sensitive Information Logging Prevention
globs: **/*.ts, **/*.js, **/*.kt, **/*.java, **/*.go, **/*.sol, **/*.py
alwaysApply: true
---

# Security-Sensitive Information Logging Prevention

## Core Principle

**NEVER log, print, or expose security-sensitive information in code, logs, comments, or error messages.**

This rule prevents accidental exposure of credentials, tokens, API keys, and other sensitive data that could lead to security breaches.

## Protected Information Categories

### 1. Authentication & Authorization

**NEVER log:**
- Passwords (plain text or hashed)
- API keys and secrets
- Authentication tokens (JWT, Bearer, OAuth)
- Session IDs
- Private keys (cryptographic, SSH, PGP)
- Access tokens and refresh tokens
- Security signatures
- Biometric data

**Examples of VIOLATIONS:**

```typescript
// WRONG - Logging API key
console.log(`API Key: ${apiKey}`);
logger.info(`Using key: ${process.env.SECRET_KEY}`);

// WRONG - Logging password
console.error(`Login failed for ${username} with password ${password}`);
logger.debug(`Auth attempt: ${JSON.stringify({ user, password })}`);

// WRONG - Logging tokens
console.log(`Bearer token: ${authToken}`);
logger.info(`Session ID: ${sessionId}`);
```

**CORRECT approaches:**

```typescript
// CORRECT - Log without sensitive data
console.log('API key configured successfully');
logger.info('Authentication credentials loaded');

// CORRECT - Redact sensitive parts
logger.debug(`Auth attempt for user: ${username} [credentials redacted]`);
logger.info(`Token type: ${tokenType} [value redacted]`);

// CORRECT - Use partial identifiers
logger.info(`Using key ending in: ...${apiKey.slice(-4)}`);
```

### 2. URLs with Sensitive Parameters

**NEVER log URLs containing:**
- API keys or tokens in query parameters
- Authentication credentials
- Session identifiers
- Temporary access codes

**Examples of VIOLATIONS:**

```typescript
// WRONG - URL with token in query param
console.log(`Calling API: https://api.example.com/data?token=${token}`);
logger.info(`Request URL: ${fullUrl}`); // if fullUrl contains secrets

// WRONG - URL with API key
fetch(`https://api.service.com/endpoint?apiKey=${key}`)
  .catch(err => console.error(`Failed to fetch: https://api.service.com/endpoint?apiKey=${key}`));
```

**CORRECT approaches:**

```typescript
// CORRECT - Sanitize URLs before logging
function sanitizeUrl(url: string): string {
  const urlObj = new URL(url);
  const sensitiveParams = ['token', 'apiKey', 'api_key', 'secret', 'password', 'auth'];
  
  sensitiveParams.forEach(param => {
    if (urlObj.searchParams.has(param)) {
      urlObj.searchParams.set(param, '[REDACTED]');
    }
  });
  
  return urlObj.toString();
}

console.log(`Calling API: ${sanitizeUrl(fullUrl)}`);

// CORRECT - Log only the endpoint path
const endpoint = new URL(fullUrl).pathname;
logger.info(`Request to endpoint: ${endpoint}`);

// CORRECT - Log without query parameters
logger.info(`API call to: ${urlBase} [query params redacted]`);
```

### 3. Environment Variables & Configuration

**NEVER log:**
- Environment variables containing secrets (`*_KEY`, `*_SECRET`, `*_TOKEN`, `*_PASSWORD`)
- Configuration values for credentials
- Connection strings with passwords
- Database credentials

**Examples of VIOLATIONS:**

```typescript
// WRONG - Logging environment variables
console.log('Environment:', process.env);
logger.debug(`Config: ${JSON.stringify(config)}`); // if config has secrets

// WRONG - Database connection string
console.log(`DB connection: ${dbConnectionString}`);
logger.info(`Connecting to: postgres://user:password@host/db`);
```

**CORRECT approaches:**

```typescript
// CORRECT - Log only non-sensitive env vars
const safeEnvVars = Object.keys(process.env)
  .filter(key => !/(KEY|SECRET|TOKEN|PASSWORD|PRIVATE)/i.test(key))
  .reduce((obj, key) => ({ ...obj, [key]: process.env[key] }), {});
console.log('Safe environment:', safeEnvVars);

// CORRECT - Redact credentials in connection strings
function sanitizeConnectionString(connStr: string): string {
  return connStr.replace(/(:\/\/[^:]+:)[^@]+(@)/, '$1[REDACTED]$2');
}
logger.info(`Database: ${sanitizeConnectionString(dbConnectionString)}`);
```

### 4. Common Secret Patterns

**Detect and prevent logging of:**

- AWS keys: `AKIA[0-9A-Z]{16}`, `aws_secret_access_key`
- GitHub tokens: `ghp_[a-zA-Z0-9]{36}`, `github_token`
- Private keys: `-----BEGIN PRIVATE KEY-----`, `-----BEGIN RSA PRIVATE KEY-----`
- JWT tokens: `eyJ[a-zA-Z0-9_-]*\.eyJ[a-zA-Z0-9_-]*\.[a-zA-Z0-9_-]*`
- Generic secrets: Variables containing `secret`, `password`, `token`, `key`, `credential`

### 5. User Privacy Data

**Be cautious with:**
- Email addresses (in some contexts)
- Phone numbers
- IP addresses (may be PII in some jurisdictions)
- Physical addresses
- Payment information (credit card numbers, account numbers)

**CORRECT approach:**

```typescript
// CORRECT - Hash or mask PII
logger.info(`User action: ${hashUserId(userId)}`);
logger.debug(`Email domain: ${email.split('@')[1]}`);

// CORRECT - Use partial data
logger.info(`Payment method ending in: ${cardNumber.slice(-4)}`);
```

## Detection Patterns (Regex)

Use these patterns to detect potential secrets in logs:

```regex
# API Keys & Tokens
(api[_-]?key|apikey|access[_-]?token|auth[_-]?token|bearer[_-]?token)["\s:=]+[a-zA-Z0-9_\-]{20,}

# AWS Credentials
AKIA[0-9A-Z]{16}
aws_secret_access_key["\s:=]+[a-zA-Z0-9/+=]{40}

# GitHub Tokens
gh[pousr]_[a-zA-Z0-9]{36}

# Generic Secrets
(password|passwd|pwd|secret)["\s:=]+\S+

# Private Keys
-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----

# JWT Tokens
eyJ[a-zA-Z0-9_-]+\.eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+

# URLs with sensitive params
(https?:\/\/[^\s]*[?&](token|key|secret|password|auth)=[^\s&]+)
```

## Logging Best Practices

### DO:

✅ Log user actions without credentials
✅ Log endpoint paths without query parameters
✅ Log error types and categories
✅ Use structured logging with redaction
✅ Log partial identifiers (last 4 characters)
✅ Log configuration keys (not values)
✅ Use log levels appropriately

### DON'T:

❌ Log request/response bodies containing auth headers
❌ Log full stack traces with credential values
❌ Log debugging info with secrets in development
❌ Include secrets in error messages
❌ Log entire objects without sanitization
❌ Use `console.log` for debugging in production

## Code Review Checklist

Before committing code, verify:

- [ ] No passwords, API keys, or tokens in logs
- [ ] URLs are sanitized before logging
- [ ] Environment variables are not logged wholesale
- [ ] Configuration dumps exclude sensitive fields
- [ ] Error messages don't contain credentials
- [ ] Stack traces don't expose secrets
- [ ] Test fixtures don't use real credentials
- [ ] Comments don't contain hardcoded secrets

## Testing with Sensitive Data

```typescript
// WRONG - Real credentials in tests
const testApiKey = 'sk-real-api-key-12345';

// CORRECT - Fake/mock credentials
const testApiKey = 'test-mock-key-' + randomString();
const mockToken = 'mock-jwt-token-for-testing';

// CORRECT - Use environment with test values
process.env.API_KEY = 'test-key-not-real';
```

## Exception Handling

When logging errors:

```typescript
// WRONG - May expose secrets in error details
catch (error) {
  console.error('Request failed:', error, requestConfig);
}

// CORRECT - Sanitize error details
catch (error) {
  const safeError = {
    message: error.message,
    code: error.code,
    endpoint: sanitizeUrl(error.config?.url)
  };
  logger.error('Request failed:', safeError);
}
```

## Automated Detection

### Pre-commit Hooks

Consider adding checks for common secret patterns:

```bash
# Detect potential secrets before commit
git diff --cached | grep -E "(password|api_key|secret|token)['\"]?\s*[:=]"
```

### Code Review

Reviewers should:
1. Search for logging statements (`console.`, `logger.`, `log.`, `print`, etc.)
2. Check logged variables for sensitive names
3. Verify URL logging is sanitized
4. Ensure error handlers redact credentials

## Domain-Specific Examples

### Kotlin/Java (Coordinator Services)

```kotlin
// WRONG - Logging sensitive config
log.info("Starting coordinator with config: $config")
log.debug("API key: ${apiKey}")

// CORRECT - Safe logging
log.info("Starting coordinator service")
log.info("Configuration loaded")
```

```kotlin
// WRONG - Config class exposes secrets via toString()
data class CoordinatorConfig(val apiKey: String, val secretKey: String)

// CORRECT - Override toString() to redact
data class CoordinatorConfig(val apiKey: String, val secretKey: String) {
    override fun toString(): String {
        return "CoordinatorConfig(apiKey=[REDACTED], secretKey=[REDACTED])"
    }
}
```

```kotlin
// WRONG - Logging DB connection string
log.info("Database connection: $dbConnectionString")

// CORRECT - Redact credentials
fun sanitizeConnectionString(connStr: String): String {
    return connStr.replace(Regex("://([^:]+):([^@]+)@"), "://$1:[REDACTED]@")
}
log.info("Database connection: ${sanitizeConnectionString(dbConnectionString)}")
```

```kotlin
// WRONG - Logging request URL with tokens
log.info("Making request to: $fullUrl")
log.debug("Request headers: $headers") // May contain Authorization

// CORRECT - Sanitize URL, omit headers
fun sanitizeUrl(url: String): String {
    return url.replace(Regex("[?&](token|apiKey|api_key|secret|password|auth)=[^&]*"), "")
}
log.info("Making request to: ${sanitizeUrl(fullUrl)}")
```

### Blockchain / Web3

```typescript
// WRONG - Logging private key during signing
console.log('Signing transaction with key:', privateKey);
console.log('Full wallet:', wallet);

// CORRECT - Log only addresses and tx hashes
console.log(`From address: ${wallet.address}`);
console.log(`Transaction hash: ${txHash}`);
```

```typescript
// WRONG - Deployment script exposes secrets
console.log(`Deploying with private key: ${process.env.DEPLOYER_PRIVATE_KEY}`);
console.log(`Using API key: ${apiKey}`);

// CORRECT - Confirm presence, not value
console.log('Deployer account configured');
console.log(`Deploying from address: ${deployerAddress}`);
```

```typescript
// WRONG - RPC URL with embedded project token
console.log(`Connecting to: ${rpcUrlWithToken}`);
console.log(`API endpoint: https://api.infura.io/v3/${projectId}`);

// CORRECT - Strip embedded tokens
function sanitizeRpcUrl(url: string): string {
  return url.replace(/\/v3\/[a-zA-Z0-9]+/, '/v3/[REDACTED]');
}
console.log(`Connecting to: ${sanitizeRpcUrl(rpcUrl)}`);
```

```typescript
// WRONG - Provider config with secrets
console.log('Initializing provider:', providerConfig);

// CORRECT - Safe config dump
const safeConfig = {
  network: config.network,
  chainId: config.chainId,
  rpcUrl: config.rpcUrl ? '[CONFIGURED]' : '[NOT SET]',
  apiKey: config.apiKey ? '[CONFIGURED]' : '[NOT SET]',
};
console.log('Deployment config:', JSON.stringify(safeConfig));
```

### API Testing / Postman

```json
// WRONG - Real credentials in Postman environment file
{
  "API_KEY": "sk-real-api-key-abc123",
  "AUTH_TOKEN": "real-bearer-token-xyz"
}

// CORRECT - Placeholder values
{
  "API_KEY": "{{YOUR_API_KEY}}",
  "AUTH_TOKEN": "{{YOUR_AUTH_TOKEN}}"
}
```

```typescript
// WRONG - Logging full request/response in tests
console.log('Making request:', { url: fullUrl, headers: headers });
console.log('Response:', response);

// CORRECT - Log only safe fields
console.log('Making request to:', sanitizeUrl(fullUrl));
console.log('Response status:', response.status);
```

```typescript
// WRONG - Real credentials in E2E tests
const PRIVATE_KEY = '0x1234...real-private-key';
const API_ENDPOINT = 'https://api.linea.build?token=real-token-abc123';

// CORRECT - Well-known test keys and env vars
const TEST_PRIVATE_KEY = '0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80';
const API_ENDPOINT = process.env.TEST_API_ENDPOINT || 'http://localhost:8545';
```

## Tools & Libraries

**Recommended sanitization libraries:**
- TypeScript/JavaScript: `redact-object`, `fast-redact`
- Java/Kotlin: `logback-mask`, custom filters
- Go: `zap` with custom encoders
- Python: `logging` with custom formatters