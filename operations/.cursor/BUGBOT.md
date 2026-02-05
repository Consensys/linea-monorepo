# BUGBOT Rules - Operations

## Core Mission

Security enforcement for operational scripts and deployment tools.

## Execution Protocol

### 1. Security Logging Guidelines

- **ALWAYS** load and reference [security-logging-guidelines](../../.cursor/rules/security-logging-guidelines/RULE.md)
- Applies to all operational scripts and deployment code

### 2. Operations-Specific Security Checks

#### Deployment Scripts
- [ ] **NEVER** log deployment private keys or API keys
- [ ] Verify environment variable usage is secure
- [ ] Check that credential loading doesn't expose values
- [ ] Ensure error messages don't leak secrets

```typescript
// WRONG - Logging credentials in deployment
console.log(`Deploying with private key: ${process.env.DEPLOYER_PRIVATE_KEY}`);
console.log(`Using API key: ${apiKey}`);

// CORRECT - Safe deployment logging
console.log('Deployer account configured');
console.log(`Deploying from address: ${deployerAddress}`);
console.log('API credentials loaded successfully');
```

#### Configuration Management
- [ ] Configuration dumps must redact sensitive fields
- [ ] Verify network configs don't include raw RPC URLs with tokens
- [ ] Check that credential validation doesn't log values

```typescript
// WRONG
console.log('Deployment config:', JSON.stringify(config));

// CORRECT
const safeConfig = {
  network: config.network,
  chainId: config.chainId,
  // Redact sensitive fields
  rpcUrl: config.rpcUrl ? '[CONFIGURED]' : '[NOT SET]',
  apiKey: config.apiKey ? '[CONFIGURED]' : '[NOT SET]'
};
console.log('Deployment config:', JSON.stringify(safeConfig));
```

#### RPC & API Endpoints
- [ ] RPC URLs with embedded tokens must be sanitized before logging
- [ ] API endpoint logs should not include auth parameters
- [ ] Connection strings must redact credentials

```typescript
// WRONG
console.log(`Connecting to: ${rpcUrlWithToken}`);
console.log(`API endpoint: https://api.infura.io/v3/${projectId}`);

// CORRECT
function sanitizeRpcUrl(url: string): string {
  return url.replace(/\/v3\/[a-zA-Z0-9]+/, '/v3/[REDACTED]');
}
console.log(`Connecting to: ${sanitizeRpcUrl(rpcUrl)}`);
console.log('API endpoint configured');
```

### 3. Transaction & Contract Operations

- [ ] Transaction signing should never log private keys
- [ ] Deployment transactions can log addresses and tx hashes
- [ ] Contract verification should not expose API keys

```typescript
// CORRECT - Safe transaction logging
console.log('Transaction signed');
console.log(`Transaction hash: ${txHash}`);
console.log(`Contract deployed at: ${contractAddress}`);

// Don't log the signer's private key or the raw signature data
```

### 4. Secret Management

- [ ] Secrets should be loaded from environment or secure vaults
- [ ] Never hardcode secrets in operation scripts
- [ ] Log only that secrets are configured, not their values

```typescript
// WRONG
const apiKey = 'sk-1234567890abcdef'; // Hardcoded
console.log(`API Key: ${process.env.API_KEY}`);

// CORRECT
const apiKey = process.env.API_KEY;
if (!apiKey) {
  throw new Error('API_KEY environment variable not set');
}
console.log('API key loaded from environment');
```

### 5. Error Handling in Operations

- [ ] Operational errors must not expose credentials
- [ ] Stack traces should be sanitized
- [ ] Failed transaction logs should not include signing details

```typescript
// WRONG
catch (error) {
  console.error('Deployment failed:', error, deploymentConfig);
}

// CORRECT
catch (error) {
  console.error('Deployment failed:', {
    message: error.message,
    network: deploymentConfig.network,
    // Don't include privateKey, apiKey, etc.
  });
}
```

### 6. Documentation & Comments

- [ ] README files should reference `.env.template`, not real values
- [ ] Code comments should not contain example credentials
- [ ] Documentation should show placeholder values only

```markdown
<!-- WRONG -->
Set your API key: `export INFURA_KEY=abc123def456real`

<!-- CORRECT -->
Set your API key: `export INFURA_KEY=your_infura_project_id_here`
```

Use the rules in [security-logging-guidelines](../../.cursor/rules/security-logging-guidelines/RULE.md) to maintain security in operational code.
