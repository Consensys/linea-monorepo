# BUGBOT Rules - Coordinator

## Core Mission

Security enforcement for coordinator service code (Kotlin/Java).

## Execution Protocol

### 1. Security Logging Guidelines

- **ALWAYS** load and reference [security-logging-guidelines](../../.cursor/rules/security-logging-guidelines/RULE.md)
- Applies to all Kotlin/Java code in the coordinator service

### 2. Coordinator-Specific Security Checks

#### Logging Configuration
- [ ] Log4j/SLF4J configurations must not output sensitive data
- [ ] Logger statements should not include credentials
- [ ] Structured logging should redact sensitive fields

```kotlin
// WRONG - Logging sensitive data
log.info("Starting coordinator with config: $config") // May contain secrets
log.debug("API key: ${apiKey}")
log.error("Authentication failed with credentials: $credentials")

// CORRECT - Safe logging
log.info("Starting coordinator service")
log.debug("Configuration loaded successfully")
log.error("Authentication failed for user: ${username}")
```

#### Configuration Management
- [ ] Application properties/YAML must not contain hardcoded secrets
- [ ] Configuration classes should not log sensitive properties
- [ ] Environment variable loading should not expose values

```kotlin
// WRONG
data class CoordinatorConfig(
    val apiKey: String,
    val secretKey: String
) {
    init {
        log.info("Config: apiKey=$apiKey, secretKey=$secretKey") // VIOLATION
    }
}

// CORRECT
data class CoordinatorConfig(
    val apiKey: String,
    val secretKey: String
) {
    init {
        log.info("Configuration loaded")
        require(apiKey.isNotBlank()) { "API key must be configured" }
        require(secretKey.isNotBlank()) { "Secret key must be configured" }
    }
    
    override fun toString(): String {
        return "CoordinatorConfig(apiKey=[REDACTED], secretKey=[REDACTED])"
    }
}
```

#### HTTP Client Operations
- [ ] Request URLs must be sanitized before logging
- [ ] Request/response interceptors should redact auth headers
- [ ] API client errors should not expose credentials

```kotlin
// WRONG
log.info("Making request to: $fullUrl") // May contain tokens
log.debug("Request headers: $headers") // May contain Authorization

// CORRECT
fun sanitizeUrl(url: String): String {
    return url.replace(Regex("[?&](token|apiKey|api_key|secret|password|auth)=[^&]*"), "")
}

log.info("Making request to: ${sanitizeUrl(fullUrl)}")
log.debug("Request configured with authentication")
```

#### Database Operations
- [ ] Connection strings must redact passwords before logging
- [ ] Database configurations should not log credentials
- [ ] Migration scripts should not contain real data

```kotlin
// WRONG
log.info("Database connection: $dbConnectionString") // Contains password

// CORRECT
fun sanitizeConnectionString(connStr: String): String {
    return connStr.replace(Regex("://([^:]+):([^@]+)@"), "://$1:[REDACTED]@")
}

log.info("Database connection: ${sanitizeConnectionString(dbConnectionString)}")
```

### 3. Service Communication

- [ ] gRPC/REST service calls should not log request bodies with secrets
- [ ] Inter-service authentication should not expose tokens
- [ ] Service mesh configurations should redact sensitive values

```kotlin
// CORRECT - Safe service call logging
suspend fun callExternalService(request: ServiceRequest) {
    log.info("Calling external service: ${request.endpoint}")
    try {
        val response = client.call(request)
        log.info("Service call successful, status: ${response.status}")
    } catch (e: Exception) {
        log.error("Service call failed: ${e.message}")
        // Don't log the full request which may contain auth tokens
    }
}
```

### 4. Error Handling

- [ ] Exception messages must not include credentials
- [ ] Stack traces should be sanitized in production logs
- [ ] Error responses should not leak internal secrets

```kotlin
// WRONG
catch (e: Exception) {
    log.error("Operation failed with config: $config", e)
    throw RuntimeException("Failed: $config", e)
}

// CORRECT
catch (e: Exception) {
    log.error("Operation failed: ${e.message}", e)
    throw RuntimeException("Operation failed", e)
}
```

### 5. Testing

- [ ] Test fixtures should use mock credentials
- [ ] Test configurations should not use production values
- [ ] Test logs should not expose real secrets

```kotlin
// CORRECT - Test configuration
class CoordinatorServiceTest {
    companion object {
        private const val TEST_API_KEY = "test-api-key-12345"
        private const val TEST_SECRET = "test-secret-not-real"
        private const val MOCK_RPC_URL = "http://localhost:8545"
    }
    
    @Test
    fun testServiceInitialization() {
        val config = CoordinatorConfig(
            apiKey = TEST_API_KEY,
            secretKey = TEST_SECRET
        )
        // Test implementation
    }
}
```

### 6. Observability & Metrics

- [ ] Metrics should not include sensitive labels
- [ ] Trace data should not contain credentials
- [ ] Health check endpoints should not expose secrets

Use the rules in [security-logging-guidelines](../../.cursor/rules/security-logging-guidelines/RULE.md) to maintain security in coordinator code.
