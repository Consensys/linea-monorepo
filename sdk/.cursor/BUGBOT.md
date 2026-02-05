# BUGBOT Rules - SDK

## Core Mission

Security enforcement for SDK client libraries and utilities.

## Execution Protocol

### 1. Security Logging Guidelines

- **ALWAYS** load and reference [security-logging-guidelines](../../.cursor/rules/security-logging-guidelines/RULE.md)
- Applies to all SDK code in `sdk-core/`, `sdk-ethers/`, `sdk-viem/`

### 2. SDK-Specific Security Checks

#### Client Libraries
- [ ] Client initialization must not log RPC URLs with embedded tokens
- [ ] Provider configuration should redact sensitive connection details
- [ ] Error messages from clients must not expose credentials

```typescript
// WRONG - Logging provider config with secrets
console.log('Initializing provider:', providerConfig);
console.log(`RPC URL: ${rpcUrl}`); // May contain tokens

// CORRECT - Safe provider initialization
console.log('Provider initialized');
console.log(`Network: ${network}`);
```

#### Transaction Handling
- [ ] Signed transactions should not log private keys or raw signatures
- [ ] Transaction parameters can be logged (to, value, data)
- [ ] Wallet operations must not expose private keys

```typescript
// WRONG
console.log('Signing transaction with key:', privateKey);
console.log('Full wallet:', wallet);

// CORRECT
console.log('Transaction signed');
console.log(`From address: ${wallet.address}`);
console.log(`Transaction hash: ${txHash}`);
```

#### Message Signing & Verification
- [ ] Message signing should not log private keys
- [ ] Signatures can be logged (they're public after broadcast)
- [ ] Verification processes should not expose secrets

```typescript
// CORRECT - Safe message signing
export async function signMessage(
  message: string,
  signer: Signer
): Promise<string> {
  console.log('Signing message');
  const signature = await signer.signMessage(message);
  console.log('Message signed successfully');
  return signature;
  // Don't log the signer's private key
}
```

#### API Integration
- [ ] API client creation must not log API keys
- [ ] Request logging should sanitize URLs and headers
- [ ] Response logging should avoid sensitive fields

```typescript
// WRONG
console.log('API client config:', {
  url: apiUrl,
  apiKey: apiKey,
  headers: headers
});

// CORRECT
console.log('API client configured');
console.log(`Base URL: ${new URL(apiUrl).origin}`);
```

### 3. Error Handling in SDK

- [ ] SDK errors must not expose internal credentials
- [ ] Provider errors should be sanitized before propagation
- [ ] Connection failures should not log authentication details

```typescript
// WRONG
catch (error) {
  throw new Error(`Provider error: ${JSON.stringify(providerConfig)}`);
}

// CORRECT
catch (error) {
  throw new Error(`Provider error: ${error.message}`);
}
```

### 4. Testing & Examples

- [ ] Test files should use mock/test credentials
- [ ] Example code must show placeholder values
- [ ] Test wallets should use well-known test private keys

```typescript
// WRONG - Real credentials in examples
const provider = new ethers.JsonRpcProvider(
  'https://mainnet.infura.io/v3/abc123realtoken'
);

// CORRECT - Placeholder in examples
const provider = new ethers.JsonRpcProvider(
  process.env.RPC_URL || 'https://mainnet.infura.io/v3/YOUR_INFURA_KEY'
);
```

### 5. Cache & Storage

- [ ] Cached data must not include sensitive credentials
- [ ] Storage utilities should not log stored values if sensitive
- [ ] Cache keys can be logged, but not values containing secrets

```typescript
// WRONG
console.log('Caching:', cacheKey, cacheValue); // value might be sensitive

// CORRECT
console.log(`Cached item with key: ${cacheKey}`);
```

### 6. Utility Functions

When creating utilities:
- [ ] URL utilities should include sanitization helpers
- [ ] Serialization should offer redaction options
- [ ] Logging utilities should have built-in secret detection

```typescript
// GOOD - Utility with built-in sanitization
export function sanitizeUrl(url: string): string {
  const urlObj = new URL(url);
  const sensitiveParams = ['token', 'apiKey', 'api_key', 'secret', 'password', 'auth'];
  
  sensitiveParams.forEach(param => {
    if (urlObj.searchParams.has(param)) {
      urlObj.searchParams.set(param, '[REDACTED]');
    }
  });
  
  return urlObj.toString();
}
```

Use the rules in [security-logging-guidelines](../../.cursor/rules/security-logging-guidelines/RULE.md) to maintain security standards in SDK code.
