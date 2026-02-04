# Error Logging & Severity Classification

This document describes the error logging architecture and severity classification rules for the Lido Governance Monitor service.

## Overview

The service uses structured logging with severity levels to enable log-based alerting. Logs are aggregated externally (Loki/CloudWatch), and alerts are configured to trigger on counts of specific severity levels.

## Design Decisions

### Why Log-Based Metrics?

The Lido Governance Monitor runs as a **Kubernetes CronJob** (every ~24 hours). Traditional Prometheus counters/gauges don't work well because:

1. **Staleness**: Prometheus marks metrics as stale after ~5 minutes of no scrapes. Since the pod terminates after each run, metrics disappear.
2. **No Pushgateway**: We chose not to add Pushgateway infrastructure complexity.
3. **Existing Infrastructure**: We already have log aggregation (Loki/CloudWatch) with alerting capabilities.

### Severity via Metadata Field

We use Winston's standard log levels (`error`, `warn`, `info`, `debug`) with an additional `severity` metadata field:

```
level=ERROR severity=CRITICAL message="Database connection failed"
level=ERROR severity=ERROR message="Validation failed"
level=WARN severity=WARN message="Audit channel failed"
```

**Why not a custom Winston level?**
- Winston's `colorize()` controls ANSI colors in ArgoCD logs
- Adding custom levels requires modifying shared library
- Metadata field approach is simpler and equally queryable

### Per-Service Loggers

Each service receives its own logger instance with a distinct name:
- `ProposalPoller`
- `ProposalProcessor`
- `NotificationService`
- `DiscourseClient`
- `ClaudeAIClient`
- `SlackClient`

This enables filtering logs by component in queries.

## Severity Levels

### CRITICAL

**Definition**: External dependency failures that require immediate investigation.

**When to use**:
- HTTP 4xx/5xx responses from external APIs
- Database connection or query failures
- Slack webhook failures (main alert channel)
- Network timeouts or connection errors
- Any failure in communication with external systems

**Examples**:
- Discourse API returns 500 Internal Server Error
- PostgreSQL connection refused
- Anthropic API rate limited (429)
- Slack webhook returns 403 Forbidden

**Alert action**: Page on-call, investigate external dependency.

### ERROR

**Definition**: Processing failures not caused by external dependencies.

**When to use**:
- Input validation failures (Zod schema)
- Output validation failures (AI response schema)
- JSON parse errors
- Missing required data (data integrity issues)
- Unexpected response formats

**Examples**:
- AI response missing `riskScore` field
- Invalid proposal URL format
- JSON.parse() throws on malformed response
- Proposal missing assessment data

**Alert action**: Review logs, may indicate prompt/schema drift.

### WARN

**Definition**: Non-blocking issues worth noting but not actionable immediately.

**When to use**:
- Audit channel failures (best-effort, main alert succeeded)
- Scheduled retries (will be attempted again)
- Threshold skips (expected behavior)
- Transient conditions that self-resolve

**Examples**:
- Audit webhook failed but main alert sent
- AI analysis failed, will retry next run
- Proposal below notification threshold
- Individual proposal fetch failed, continuing with others

**Alert action**: Monitor trends, investigate if sustained.

## Component-Specific Classification

### DiscourseClient
| Scenario | Severity |
|----------|----------|
| HTTP 4xx/5xx response | CRITICAL |
| Schema validation failure | ERROR |
| Network exception | CRITICAL |

### ClaudeAIClient
| Scenario | Severity |
|----------|----------|
| API request exception (timeout, 4xx, 5xx) | CRITICAL |
| Input validation error | ERROR |
| Output schema validation error | ERROR |
| JSON parse error | ERROR |
| Missing text content in response | ERROR |

### SlackClient
| Scenario | Severity |
|----------|----------|
| Main alert webhook HTTP error | CRITICAL |
| Main alert network exception | CRITICAL |
| Audit webhook HTTP error | WARN |
| Audit webhook network exception | WARN |

### Services (Poller, Processor, Notification)
| Scenario | Severity |
|----------|----------|
| Top-level catch (entire pipeline failed) | CRITICAL |
| DB operation failure | CRITICAL |
| Individual proposal processing failure | WARN |
| Data integrity issue (missing assessment) | ERROR |

## Alerting Configuration

Configure alerts in your log aggregation system:

```yaml
# Example Loki alerting rule
groups:
  - name: lido-governance-monitor
    rules:
      - alert: LidoGovernanceMonitorCritical
        expr: |
          count_over_time({app="lido-governance-monitor"} |= "severity=CRITICAL" [1h]) > 0
        for: 0m
        labels:
          severity: critical
        annotations:
          summary: "Critical error in Lido Governance Monitor"
          description: "External dependency failure detected"

      - alert: LidoGovernanceMonitorErrorRate
        expr: |
          count_over_time({app="lido-governance-monitor"} |= "severity=ERROR" [24h]) > 5
        for: 0m
        labels:
          severity: warning
        annotations:
          summary: "Elevated error rate in Lido Governance Monitor"
          description: "Multiple processing failures in last 24h"
```

## Logger Interface

```typescript
interface ILidoGovernanceMonitorLogger extends ILogger {
  /**
   * Log CRITICAL severity - external dependency failures.
   * Internally calls logger.error() with { severity: 'CRITICAL' }
   */
  critical(message: string, meta?: Record<string, unknown>): void;

  /**
   * Log ERROR severity - processing failures.
   * Internally calls logger.error() with { severity: 'ERROR' }
   */
  error(message: string, meta?: Record<string, unknown>): void;

  /**
   * Log WARN severity - non-blocking issues.
   * Internally calls logger.warn() with { severity: 'WARN' }
   */
  warn(message: string, meta?: Record<string, unknown>): void;
}
```

## Usage Example

```typescript
// Factory creates per-service logger
const logger = createLidoGovernanceMonitorLogger("DiscourseClient");

// External dependency failure
try {
  const response = await fetch(url);
  if (!response.ok) {
    logger.critical("Failed to fetch proposals", {
      status: response.status,
      statusText: response.statusText,
    });
    return undefined;
  }
} catch (error) {
  logger.critical("Network error fetching proposals", { error });
  return undefined;
}

// Processing failure
const result = schema.safeParse(data);
if (!result.success) {
  logger.error("Schema validation failed", {
    errors: result.error.errors,
  });
  return undefined;
}

// Non-blocking warning
if (!auditResult.success) {
  logger.warn("Audit log failed, continuing", {
    error: auditResult.error,
  });
}
```
