# BUGBOT Rules

## Core Mission

Automated security and code quality enforcement across the Linea monorepo.

## Execution Protocol

### 1. Security Logging Guidelines

- **ALWAYS** load and reference [security-logging-guidelines](../.cursor/rules/security-logging-guidelines/RULE.md)
- Applies to all code files: `.ts`, `.js`, `.kt`, `.java`, `.go`, `.sol`, `.py`
- **CRITICAL**: This rule applies to ALL code reviews and changes

### 2. Sensitive Information Detection

**NEVER allow logging of:**
- Passwords, API keys, authentication tokens
- URLs with sensitive query parameters (token, apiKey, secret, password, auth)
- Private keys (cryptographic, SSH, PGP)
- Environment variables containing secrets (`*_KEY`, `*_SECRET`, `*_TOKEN`, `*_PASSWORD`)
- Session IDs, Bearer tokens, OAuth tokens
- Database connection strings with credentials
- AWS keys, GitHub tokens, JWT tokens

### 3. Code Review Checks

When reviewing code changes, ALWAYS verify:

#### Logging Statements
- [ ] Check all `console.log`, `console.error`, `console.debug` statements
- [ ] Check all `logger.*` calls (info, debug, warn, error)
- [ ] Check all `print`, `println`, `log.*` statements
- [ ] Verify logged URLs are sanitized (no tokens in query params)
- [ ] Verify logged objects don't contain sensitive fields

#### Common Patterns to Flag

```typescript
// VIOLATIONS - Flag these patterns:
console.log(`API Key: ${apiKey}`);
logger.info(`Token: ${token}`);
console.error(`Failed with credentials: ${creds}`);
logger.debug(`Full config: ${JSON.stringify(config)}`);
console.log(`URL: ${urlWithToken}`);
logger.info(process.env); // Logging all env vars
```

```typescript
// ACCEPTABLE - These are safe:
console.log('API key configured successfully');
logger.info('Authentication completed');
logger.debug(`User: ${username} [credentials redacted]`);
logger.info(`Endpoint: ${sanitizeUrl(url)}`);
```

#### Error Handling
- [ ] Verify error logs don't expose credentials
- [ ] Check stack traces are sanitized
- [ ] Ensure caught errors don't log sensitive request/response data

#### Detection Regex Patterns

Flag any logging that matches these patterns:
```regex
(api[_-]?key|apikey|access[_-]?token|auth[_-]?token|bearer[_-]?token)
(password|passwd|pwd|secret)
AKIA[0-9A-Z]{16}
gh[pousr]_[a-zA-Z0-9]{36}
-----BEGIN\s+(RSA\s+)?PRIVATE\s+KEY-----
eyJ[a-zA-Z0-9_-]+\.eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+
```

### 4. Configuration & Environment Files

- [ ] Verify `.env` files are in `.gitignore`
- [ ] Check that example configs use placeholder values
- [ ] Ensure no hardcoded secrets in configuration files
- [ ] Verify test files use mock credentials, not real ones

### 5. URL Sanitization

When URLs are logged, they MUST be sanitized:
- Remove or redact query parameters: `token`, `apiKey`, `api_key`, `secret`, `password`, `auth`
- Only log the endpoint path when possible
- Use utility functions to sanitize URLs before logging

### 6. Acceptable Logging

**DO allow:**
- Log levels and categories
- User identifiers (hashed or partial)
- Endpoint paths (without query params)
- Error types and codes
- Timestamp and request IDs
- Configuration keys (not values)
- Last 4 characters of identifiers

### 7. Enforcement Actions

When sensitive information is detected in logs:
1. **BLOCK** the change with clear explanation
2. Suggest safe alternatives (sanitization, redaction)
3. Reference the [security-logging-guidelines](../.cursor/rules/security-logging-guidelines/RULE.md)
4. Provide code examples of correct approach

## Quick Reference

### Safe Logging Pattern

```typescript
// Import or define sanitization utilities
function sanitizeUrl(url: string): string {
  const urlObj = new URL(url);
  ['token', 'apiKey', 'api_key', 'secret', 'password', 'auth'].forEach(param => {
    if (urlObj.searchParams.has(param)) {
      urlObj.searchParams.set(param, '[REDACTED]');
    }
  });
  return urlObj.toString();
}

function redactSensitiveFields(obj: any): any {
  const sensitiveKeys = /(password|secret|token|key|credential|auth)/i;
  // ... implement redaction logic
}

// Use in logging
logger.info(`Request to: ${sanitizeUrl(fullUrl)}`);
logger.debug('Config loaded', redactSensitiveFields(config));
```

Use the rules in [security-logging-guidelines](../.cursor/rules/security-logging-guidelines/RULE.md) to enforce security standards and prevent credential exposure.
