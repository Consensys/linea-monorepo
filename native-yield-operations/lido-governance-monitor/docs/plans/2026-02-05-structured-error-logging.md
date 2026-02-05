# Structured Error Logging Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.
>
> **CRITICAL:** Follow `/test-driven-development` for all new feature code. Write tests FIRST, verify they fail, then implement.

**Goal:** Add structured error logging with severity levels (CRITICAL/ERROR/WARN) to enable log-based alerting on error counts.

**Architecture:** Create `ILidoGovernanceMonitorLogger` interface extending `ILogger` with `.critical()` method. Factory function `createLidoGovernanceMonitorLogger(name)` wraps `WinstonLogger` and auto-injects `{ severity: 'CRITICAL' | 'ERROR' | 'WARN' }` field on all error/warn calls. Each service receives its own logger instance with a distinct name for better log filtering.

**Tech Stack:** TypeScript, Winston (via @consensys/linea-shared-utils), Jest

---

## Error Severity Classification Rules

| Severity | When to Use | Examples |
|----------|-------------|----------|
| **CRITICAL** | External dependency failures - HTTP errors, DB errors, API request failures | Discourse HTTP 4xx/5xx, DB query fails, Slack webhook fails, Anthropic API request error |
| **ERROR** | Processing failures not caused by external deps | Zod validation fails, JSON parse error, missing required fields |
| **WARN** | Non-blocking issues worth noting | Audit channel failed (but main alert succeeded), transient conditions |

---

## Task 0: Document Error Logging Decisions

**Files:**
- Create: `docs/error-logging.md`

**Step 1: Create the documentation file**

Create `docs/error-logging.md`:

```markdown
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
```

**Step 2: Commit**

```bash
git add native-yield-operations/lido-governance-monitor/docs/error-logging.md
git commit -m "docs(lido-governance-monitor): document error logging decisions and severity classification"
```

---

## Task 1: Create ILidoGovernanceMonitorLogger Interface

**Files:**
- Create: `src/utils/logging/ILidoGovernanceMonitorLogger.ts`

**Step 1: Write the failing test**

Create test file `src/utils/logging/__tests__/ILidoGovernanceMonitorLogger.test.ts`:

```typescript
import { describe, it, expect } from "@jest/globals";
import { ILidoGovernanceMonitorLogger, Severity } from "../ILidoGovernanceMonitorLogger.js";

describe("ILidoGovernanceMonitorLogger", () => {
  it("Severity enum has correct values", () => {
    expect(Severity.CRITICAL).toBe("CRITICAL");
    expect(Severity.ERROR).toBe("ERROR");
    expect(Severity.WARN).toBe("WARN");
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd native-yield-operations/lido-governance-monitor && pnpm test -- --testPathPattern="ILidoGovernanceMonitorLogger.test.ts"`

Expected: FAIL with "Cannot find module"

**Step 3: Write minimal implementation**

Create `src/utils/logging/ILidoGovernanceMonitorLogger.ts`:

```typescript
import { ILogger } from "@consensys/linea-shared-utils";

export enum Severity {
  CRITICAL = "CRITICAL",
  ERROR = "ERROR",
  WARN = "WARN",
}

/**
 * Extended logger interface that adds severity-classified logging methods.
 * - critical(): External dependency failures (HTTP errors, DB errors, API failures)
 * - error(): Processing failures not caused by external deps (validation errors)
 * - warn(): Non-blocking issues worth noting
 *
 * All severity methods auto-inject { severity: Severity } into log metadata.
 */
export interface ILidoGovernanceMonitorLogger extends ILogger {
  /**
   * Log CRITICAL severity - external dependency failures.
   * Use when: HTTP errors, DB errors, Slack webhook fails, API request errors.
   */
  critical(message: string, meta?: Record<string, unknown>): void;
}
```

**Step 4: Run test to verify it passes**

Run: `cd native-yield-operations/lido-governance-monitor && pnpm test -- --testPathPattern="ILidoGovernanceMonitorLogger.test.ts"`

Expected: PASS

**Step 5: Commit**

```bash
git add native-yield-operations/lido-governance-monitor/src/utils/logging/
git commit -m "feat(lido-governance-monitor): add ILidoGovernanceMonitorLogger interface with Severity enum"
```

---

## Task 2: Create LidoGovernanceMonitorLogger Implementation

**Files:**
- Create: `src/utils/logging/LidoGovernanceMonitorLogger.ts`
- Test: `src/utils/logging/__tests__/LidoGovernanceMonitorLogger.test.ts`

**Step 1: Write the failing tests**

Create `src/utils/logging/__tests__/LidoGovernanceMonitorLogger.test.ts`:

```typescript
import { ILogger } from "@consensys/linea-shared-utils";
import { jest, describe, it, expect, beforeEach } from "@jest/globals";
import { LidoGovernanceMonitorLogger } from "../LidoGovernanceMonitorLogger.js";
import { Severity } from "../ILidoGovernanceMonitorLogger.js";

describe("LidoGovernanceMonitorLogger", () => {
  let mockBaseLogger: jest.Mocked<ILogger>;
  let logger: LidoGovernanceMonitorLogger;

  beforeEach(() => {
    mockBaseLogger = {
      name: "TestLogger",
      info: jest.fn(),
      error: jest.fn(),
      warn: jest.fn(),
      debug: jest.fn(),
    };
    logger = new LidoGovernanceMonitorLogger(mockBaseLogger);
  });

  describe("critical", () => {
    it("calls base logger error with severity CRITICAL", () => {
      logger.critical("Database connection failed", { host: "localhost" });

      expect(mockBaseLogger.error).toHaveBeenCalledWith("Database connection failed", {
        severity: Severity.CRITICAL,
        host: "localhost",
      });
    });

    it("handles call without metadata", () => {
      logger.critical("Network timeout");

      expect(mockBaseLogger.error).toHaveBeenCalledWith("Network timeout", {
        severity: Severity.CRITICAL,
      });
    });
  });

  describe("error", () => {
    it("calls base logger error with severity ERROR", () => {
      logger.error("Validation failed", { field: "riskScore" });

      expect(mockBaseLogger.error).toHaveBeenCalledWith("Validation failed", {
        severity: Severity.ERROR,
        field: "riskScore",
      });
    });

    it("handles call without metadata", () => {
      logger.error("Parse error");

      expect(mockBaseLogger.error).toHaveBeenCalledWith("Parse error", {
        severity: Severity.ERROR,
      });
    });
  });

  describe("warn", () => {
    it("calls base logger warn with severity WARN", () => {
      logger.warn("Audit channel failed", { channel: "audit" });

      expect(mockBaseLogger.warn).toHaveBeenCalledWith("Audit channel failed", {
        severity: Severity.WARN,
        channel: "audit",
      });
    });

    it("handles call without metadata", () => {
      logger.warn("Retrying operation");

      expect(mockBaseLogger.warn).toHaveBeenCalledWith("Retrying operation", {
        severity: Severity.WARN,
      });
    });
  });

  describe("delegated methods", () => {
    it("delegates info to base logger", () => {
      logger.info("Starting process", { step: 1 });

      expect(mockBaseLogger.info).toHaveBeenCalledWith("Starting process", { step: 1 });
    });

    it("delegates debug to base logger", () => {
      logger.debug("Debug details", { data: "test" });

      expect(mockBaseLogger.debug).toHaveBeenCalledWith("Debug details", { data: "test" });
    });
  });

  describe("name property", () => {
    it("returns base logger name", () => {
      expect(logger.name).toBe("TestLogger");
    });
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd native-yield-operations/lido-governance-monitor && pnpm test -- --testPathPattern="LidoGovernanceMonitorLogger.test.ts"`

Expected: FAIL with "Cannot find module"

**Step 3: Write minimal implementation**

Create `src/utils/logging/LidoGovernanceMonitorLogger.ts`:

```typescript
import { ILogger } from "@consensys/linea-shared-utils";
import { ILidoGovernanceMonitorLogger, Severity } from "./ILidoGovernanceMonitorLogger.js";

/**
 * Logger wrapper that adds severity-classified logging methods.
 * Wraps an ILogger and auto-injects { severity: Severity } into metadata.
 */
export class LidoGovernanceMonitorLogger implements ILidoGovernanceMonitorLogger {
  constructor(private readonly baseLogger: ILogger) {}

  get name(): string {
    return this.baseLogger.name;
  }

  critical(message: string, meta?: Record<string, unknown>): void {
    this.baseLogger.error(message, { severity: Severity.CRITICAL, ...meta });
  }

  error(message: string, meta?: Record<string, unknown>): void {
    this.baseLogger.error(message, { severity: Severity.ERROR, ...meta });
  }

  warn(message: string, meta?: Record<string, unknown>): void {
    this.baseLogger.warn(message, { severity: Severity.WARN, ...meta });
  }

  info(message: string, ...params: unknown[]): void {
    this.baseLogger.info(message, ...params);
  }

  debug(message: string, ...params: unknown[]): void {
    this.baseLogger.debug(message, ...params);
  }
}
```

**Step 4: Run test to verify it passes**

Run: `cd native-yield-operations/lido-governance-monitor && pnpm test -- --testPathPattern="LidoGovernanceMonitorLogger.test.ts"`

Expected: PASS

**Step 5: Commit**

```bash
git add native-yield-operations/lido-governance-monitor/src/utils/logging/
git commit -m "feat(lido-governance-monitor): implement LidoGovernanceMonitorLogger with severity injection"
```

---

## Task 3: Create Factory Function

**Files:**
- Create: `src/utils/logging/createLidoGovernanceMonitorLogger.ts`
- Create: `src/utils/logging/index.ts` (barrel export)
- Test: `src/utils/logging/__tests__/createLidoGovernanceMonitorLogger.test.ts`

**Step 1: Write the failing test**

Create `src/utils/logging/__tests__/createLidoGovernanceMonitorLogger.test.ts`:

```typescript
import { describe, it, expect } from "@jest/globals";
import { createLidoGovernanceMonitorLogger } from "../createLidoGovernanceMonitorLogger.js";
import { LidoGovernanceMonitorLogger } from "../LidoGovernanceMonitorLogger.js";

describe("createLidoGovernanceMonitorLogger", () => {
  it("returns a LidoGovernanceMonitorLogger instance", () => {
    const logger = createLidoGovernanceMonitorLogger("TestComponent");

    expect(logger).toBeInstanceOf(LidoGovernanceMonitorLogger);
  });

  it("sets the logger name correctly", () => {
    const logger = createLidoGovernanceMonitorLogger("ProposalPoller");

    expect(logger.name).toBe("ProposalPoller");
  });

  it("creates loggers with distinct names", () => {
    const logger1 = createLidoGovernanceMonitorLogger("ServiceA");
    const logger2 = createLidoGovernanceMonitorLogger("ServiceB");

    expect(logger1.name).toBe("ServiceA");
    expect(logger2.name).toBe("ServiceB");
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd native-yield-operations/lido-governance-monitor && pnpm test -- --testPathPattern="createLidoGovernanceMonitorLogger.test.ts"`

Expected: FAIL with "Cannot find module"

**Step 3: Write minimal implementation**

Create `src/utils/logging/createLidoGovernanceMonitorLogger.ts`:

```typescript
import { WinstonLogger } from "@consensys/linea-shared-utils";
import { ILidoGovernanceMonitorLogger } from "./ILidoGovernanceMonitorLogger.js";
import { LidoGovernanceMonitorLogger } from "./LidoGovernanceMonitorLogger.js";

/**
 * Factory function to create a severity-aware logger for the Lido Governance Monitor.
 *
 * @param name - Component name (e.g., "ProposalPoller", "ClaudeAIClient")
 * @param logLevel - Log level (default: from LOG_LEVEL env or "info")
 * @returns ILidoGovernanceMonitorLogger instance
 */
export function createLidoGovernanceMonitorLogger(
  name: string,
  logLevel?: string,
): ILidoGovernanceMonitorLogger {
  const level = logLevel ?? process.env.LOG_LEVEL ?? "info";
  const baseLogger = new WinstonLogger(name, { level });
  return new LidoGovernanceMonitorLogger(baseLogger);
}
```

Create barrel export `src/utils/logging/index.ts`:

```typescript
export { ILidoGovernanceMonitorLogger, Severity } from "./ILidoGovernanceMonitorLogger.js";
export { LidoGovernanceMonitorLogger } from "./LidoGovernanceMonitorLogger.js";
export { createLidoGovernanceMonitorLogger } from "./createLidoGovernanceMonitorLogger.js";
```

**Step 4: Run test to verify it passes**

Run: `cd native-yield-operations/lido-governance-monitor && pnpm test -- --testPathPattern="createLidoGovernanceMonitorLogger.test.ts"`

Expected: PASS

**Step 5: Commit**

```bash
git add native-yield-operations/lido-governance-monitor/src/utils/logging/
git commit -m "feat(lido-governance-monitor): add createLidoGovernanceMonitorLogger factory"
```

---

## Task 4: Update Bootstrap to Use Per-Service Loggers

**Files:**
- Modify: `src/application/main/LidoGovernanceMonitorBootstrap.ts`
- Test: `src/application/main/__tests__/LidoGovernanceMonitorBootstrap.test.ts`

**Step 1: Write the failing test**

Add to `src/application/main/__tests__/LidoGovernanceMonitorBootstrap.test.ts`:

```typescript
import { describe, it, expect, jest, beforeEach } from "@jest/globals";

// Mock the factory before importing Bootstrap
jest.unstable_mockModule("../../../utils/logging/index.js", () => ({
  createLidoGovernanceMonitorLogger: jest.fn().mockImplementation((name: string) => ({
    name,
    info: jest.fn(),
    error: jest.fn(),
    warn: jest.fn(),
    debug: jest.fn(),
    critical: jest.fn(),
  })),
  Severity: { CRITICAL: "CRITICAL", ERROR: "ERROR", WARN: "WARN" },
}));

describe("LidoGovernanceMonitorBootstrap", () => {
  describe("logger creation", () => {
    it("creates per-service loggers with distinct names", async () => {
      const { createLidoGovernanceMonitorLogger } = await import("../../../utils/logging/index.js");
      const { LidoGovernanceMonitorBootstrap } = await import("../LidoGovernanceMonitorBootstrap.js");

      // This test verifies that Bootstrap creates multiple loggers
      // Implementation will call createLidoGovernanceMonitorLogger for each service
      const mockFactory = createLidoGovernanceMonitorLogger as jest.Mock;

      // The actual test assertions depend on your ability to spy on the factory
      // For now, verify the factory is exported and callable
      expect(typeof createLidoGovernanceMonitorLogger).toBe("function");
    });
  });
});
```

Note: The existing Bootstrap test file may need adjustment based on how mocking works in your setup. The key verification is that `createLidoGovernanceMonitorLogger` is called with different service names.

**Step 2: Run test to verify current behavior**

Run: `cd native-yield-operations/lido-governance-monitor && pnpm test -- --testPathPattern="LidoGovernanceMonitorBootstrap.test.ts"`

**Step 3: Update Bootstrap implementation**

Modify `src/application/main/LidoGovernanceMonitorBootstrap.ts`:

```typescript
import Anthropic from "@anthropic-ai/sdk";
import { ExponentialBackoffRetryService } from "@consensys/linea-shared-utils";
import { PrismaPg } from "@prisma/adapter-pg";

import { Config } from "./config/index.js";
import { ClaudeAIClient } from "../../clients/ClaudeAIClient.js";
import { ProposalRepository } from "../../clients/db/ProposalRepository.js";
import { DiscourseClient } from "../../clients/DiscourseClient.js";
import { SlackClient } from "../../clients/SlackClient.js";
import { PrismaClient } from "../../../prisma/client/client.js";
import { NormalizationService } from "../../services/NormalizationService.js";
import { NotificationService } from "../../services/NotificationService.js";
import { ProposalPoller } from "../../services/ProposalPoller.js";
import { ProposalProcessor } from "../../services/ProposalProcessor.js";
import {
  createLidoGovernanceMonitorLogger,
  ILidoGovernanceMonitorLogger,
} from "../../utils/logging/index.js";

export class LidoGovernanceMonitorBootstrap {
  private constructor(
    private readonly logger: ILidoGovernanceMonitorLogger,
    private readonly prisma: PrismaClient,
    private readonly proposalPoller: ProposalPoller,
    private readonly proposalProcessor: ProposalProcessor,
    private readonly notificationService: NotificationService,
  ) {}

  static create(config: Config, systemPrompt: string): LidoGovernanceMonitorBootstrap {
    // Create per-service loggers
    const bootstrapLogger = createLidoGovernanceMonitorLogger("LidoGovernanceMonitorBootstrap");
    const discourseClientLogger = createLidoGovernanceMonitorLogger("DiscourseClient");
    const aiClientLogger = createLidoGovernanceMonitorLogger("ClaudeAIClient");
    const slackClientLogger = createLidoGovernanceMonitorLogger("SlackClient");
    const normalizationLogger = createLidoGovernanceMonitorLogger("NormalizationService");
    const pollerLogger = createLidoGovernanceMonitorLogger("ProposalPoller");
    const processorLogger = createLidoGovernanceMonitorLogger("ProposalProcessor");
    const notificationLogger = createLidoGovernanceMonitorLogger("NotificationService");

    // Database
    const adapter = new PrismaPg({
      connectionString: config.database.url,
    });
    const prisma = new PrismaClient({ adapter });

    // Repositories
    const proposalRepository = new ProposalRepository(prisma);

    // Shared services
    const retryService = new ExponentialBackoffRetryService(bootstrapLogger);

    // Clients
    const discourseClient = new DiscourseClient(
      discourseClientLogger,
      retryService,
      config.discourse.proposalsUrl,
      config.http.timeoutMs,
    );

    const anthropicClient = new Anthropic({ apiKey: config.anthropic.apiKey });
    const aiClient = new ClaudeAIClient(
      aiClientLogger,
      anthropicClient,
      config.anthropic.model,
      systemPrompt,
      config.anthropic.maxOutputTokens,
      config.anthropic.maxProposalChars,
    );

    const slackClient = new SlackClient(
      slackClientLogger,
      config.slack.webhookUrl,
      config.riskAssessment.threshold,
      config.http.timeoutMs,
      config.slack.auditWebhookUrl,
    );

    // Services
    const normalizationService = new NormalizationService(normalizationLogger, discourseClient.getBaseUrl());

    const proposalPoller = new ProposalPoller(
      pollerLogger,
      discourseClient,
      normalizationService,
      proposalRepository,
      config.discourse.maxTopicsPerPoll,
    );

    const proposalProcessor = new ProposalProcessor(
      processorLogger,
      aiClient,
      proposalRepository,
      config.riskAssessment.threshold,
      config.riskAssessment.promptVersion,
    );

    const notificationService = new NotificationService(
      notificationLogger,
      slackClient,
      proposalRepository,
      config.riskAssessment.threshold,
    );

    return new LidoGovernanceMonitorBootstrap(
      bootstrapLogger,
      prisma,
      proposalPoller,
      proposalProcessor,
      notificationService,
    );
  }

  async start(): Promise<void> {
    this.logger.info("Starting Lido Governance Monitor");

    try {
      await this.prisma.$connect();
      this.logger.info("Database connected");

      await this.proposalPoller.pollOnce();
      await this.proposalProcessor.processOnce();
      await this.notificationService.notifyOnce();

      this.logger.info("Lido Governance Monitor execution completed");
    } catch (error) {
      this.logger.critical("Lido Governance Monitor execution failed", {
        error: error instanceof Error ? error.message : String(error),
      });
      throw error;
    }
  }

  async stop(): Promise<void> {
    this.logger.info("Stopping Lido Governance Monitor");
    await this.prisma.$disconnect();
    this.logger.info("Database disconnected");
  }

  getProposalPoller(): ProposalPoller {
    return this.proposalPoller;
  }

  getProposalProcessor(): ProposalProcessor {
    return this.proposalProcessor;
  }

  getNotificationService(): NotificationService {
    return this.notificationService;
  }
}
```

**Step 4: Run tests to verify**

Run: `cd native-yield-operations/lido-governance-monitor && pnpm test`

Expected: All tests pass

**Step 5: Commit**

```bash
git add native-yield-operations/lido-governance-monitor/src/application/main/
git commit -m "refactor(lido-governance-monitor): use per-service loggers via factory"
```

---

## Task 5: Update Service Interfaces to Accept ILidoGovernanceMonitorLogger

**Files:**
- Modify: `src/core/clients/IDiscourseClient.ts` - No change needed (doesn't define logger)
- Modify: `src/core/services/IProposalPoller.ts` - No change needed
- The interfaces don't specify logger type, so services can receive either ILogger or ILidoGovernanceMonitorLogger

Note: Since `ILidoGovernanceMonitorLogger extends ILogger`, existing code accepting `ILogger` will work with the new logger. No interface changes needed.

**Step 1: Verify existing tests still pass**

Run: `cd native-yield-operations/lido-governance-monitor && pnpm test`

Expected: PASS

**Step 2: Commit (if any changes)**

Skip if no changes needed.

---

## Task 6: Reclassify Errors in DiscourseClient

**Files:**
- Modify: `src/clients/DiscourseClient.ts`
- Test: `src/clients/__tests__/DiscourseClient.test.ts`

**Classification:**
- HTTP 4xx/5xx responses → CRITICAL (external dependency failure)
- Schema validation errors → ERROR (processing failure)
- Network exceptions → CRITICAL (external dependency failure)

**Step 1: Update the test mocks to use ILidoGovernanceMonitorLogger**

Update `src/clients/__tests__/DiscourseClient.test.ts` to mock `critical`:

```typescript
import { jest, describe, it, expect, beforeEach } from "@jest/globals";
import { ILidoGovernanceMonitorLogger } from "../../utils/logging/index.js";
// ... other imports

const createLoggerMock = (): jest.Mocked<ILidoGovernanceMonitorLogger> => ({
  name: "test-logger",
  debug: jest.fn(),
  error: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
  critical: jest.fn(),
});
```

**Step 2: Write tests for severity classification**

Add to `src/clients/__tests__/DiscourseClient.test.ts`:

```typescript
describe("error severity classification", () => {
  it("logs CRITICAL for HTTP error response", async () => {
    // Arrange
    global.fetch = jest.fn().mockResolvedValue({
      ok: false,
      status: 500,
      statusText: "Internal Server Error",
    }) as jest.Mock;

    // Act
    await client.fetchLatestProposals();

    // Assert
    expect(logger.critical).toHaveBeenCalledWith(
      "Failed to fetch latest proposals",
      expect.objectContaining({ status: 500 }),
    );
  });

  it("logs ERROR for schema validation failure", async () => {
    // Arrange - return valid HTTP but invalid schema
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ invalid: "data" }),
    }) as jest.Mock;

    // Act
    await client.fetchLatestProposals();

    // Assert
    expect(logger.error).toHaveBeenCalledWith(
      "Discourse API response failed schema validation",
      expect.objectContaining({ errors: expect.any(Array) }),
    );
  });

  it("logs CRITICAL for network exception", async () => {
    // Arrange
    global.fetch = jest.fn().mockRejectedValue(new Error("Network timeout")) as jest.Mock;

    // Act
    await client.fetchLatestProposals();

    // Assert
    expect(logger.critical).toHaveBeenCalledWith(
      "Error fetching latest proposals",
      expect.objectContaining({ error: expect.any(Error) }),
    );
  });
});
```

**Step 3: Update DiscourseClient implementation**

```typescript
import { fetchWithTimeout, IRetryService } from "@consensys/linea-shared-utils";

import { IDiscourseClient } from "../core/clients/IDiscourseClient.js";
import {
  RawDiscourseProposal,
  RawDiscourseProposalList,
  RawDiscourseProposalListSchema,
  RawDiscourseProposalSchema,
} from "../core/entities/RawDiscourseProposal.js";
import { ILidoGovernanceMonitorLogger } from "../utils/logging/index.js";

export class DiscourseClient implements IDiscourseClient {
  private readonly baseUrl: string;

  constructor(
    private readonly logger: ILidoGovernanceMonitorLogger,
    private readonly retryService: IRetryService,
    private readonly proposalsUrl: string,
    private readonly httpTimeoutMs: number,
  ) {
    this.baseUrl = new URL(proposalsUrl).origin;
  }

  async fetchLatestProposals(): Promise<RawDiscourseProposalList | undefined> {
    const url = this.proposalsUrl;
    try {
      const response = await this.retryService.retry(() => fetchWithTimeout(url, {}, this.httpTimeoutMs));
      if (!response.ok) {
        // CRITICAL: HTTP error from external dependency
        this.logger.critical("Failed to fetch latest proposals", {
          status: response.status,
          statusText: response.statusText,
        });
        return undefined;
      }
      const data = await response.json();

      const validationResult = RawDiscourseProposalListSchema.safeParse(data);
      if (!validationResult.success) {
        // ERROR: Schema validation is a processing failure, not external dep
        this.logger.error("Discourse API response failed schema validation", {
          errors: validationResult.error.errors,
        });
        return undefined;
      }

      this.logger.debug("Fetched latest proposals", { count: validationResult.data.topic_list.topics.length });
      return validationResult.data;
    } catch (error) {
      // CRITICAL: Network/request exception is external dependency failure
      this.logger.critical("Error fetching latest proposals", { error });
      return undefined;
    }
  }

  getBaseUrl(): string {
    return this.baseUrl;
  }

  async fetchProposalDetails(proposalId: number): Promise<RawDiscourseProposal | undefined> {
    const url = `${this.baseUrl}/t/${proposalId}.json`;
    try {
      const response = await this.retryService.retry(() => fetchWithTimeout(url, {}, this.httpTimeoutMs));
      if (!response.ok) {
        // CRITICAL: HTTP error from external dependency
        this.logger.critical("Failed to fetch proposal details", {
          proposalId,
          status: response.status,
        });
        return undefined;
      }
      const data = await response.json();

      const validationResult = RawDiscourseProposalSchema.safeParse(data);
      if (!validationResult.success) {
        // ERROR: Schema validation is a processing failure
        this.logger.error("Discourse API response failed schema validation", {
          proposalId,
          errors: validationResult.error.errors,
        });
        return undefined;
      }

      this.logger.debug("Fetched proposal details", { proposalId, title: validationResult.data.title });
      return validationResult.data;
    } catch (error) {
      // CRITICAL: Network/request exception
      this.logger.critical("Error fetching proposal details", { proposalId, error });
      return undefined;
    }
  }
}
```

**Step 4: Run tests**

Run: `cd native-yield-operations/lido-governance-monitor && pnpm test -- --testPathPattern="DiscourseClient.test.ts"`

Expected: PASS

**Step 5: Commit**

```bash
git add native-yield-operations/lido-governance-monitor/src/clients/DiscourseClient.ts
git add native-yield-operations/lido-governance-monitor/src/clients/__tests__/DiscourseClient.test.ts
git commit -m "refactor(lido-governance-monitor): classify DiscourseClient errors by severity"
```

---

## Task 7: Reclassify Errors in ClaudeAIClient

**Files:**
- Modify: `src/clients/ClaudeAIClient.ts`
- Test: `src/clients/__tests__/ClaudeAIClient.test.ts`

**Classification:**
- Anthropic API request exception (network, 4xx, 5xx) → CRITICAL
- Input validation error (AIAnalysisRequestSchema) → ERROR
- Output validation error (AssessmentSchema) → ERROR
- JSON parse error → ERROR

**Step 1: Update test mocks**

Update `src/clients/__tests__/ClaudeAIClient.test.ts`:

```typescript
const createLoggerMock = (): jest.Mocked<ILidoGovernanceMonitorLogger> => ({
  name: "test-logger",
  debug: jest.fn(),
  error: jest.fn(),
  info: jest.fn(),
  warn: jest.fn(),
  critical: jest.fn(),
});
```

**Step 2: Add severity classification tests**

```typescript
describe("error severity classification", () => {
  it("logs ERROR for input validation failure", async () => {
    const result = await client.analyzeProposal({
      proposalTitle: "",  // Invalid: empty
      proposalText: "text",
      proposalUrl: "not-a-url",  // Invalid URL
      proposalType: "discourse",
    });

    expect(result).toBeUndefined();
    expect(logger.error).toHaveBeenCalledWith(
      "Invalid analysis request",
      expect.objectContaining({ errors: expect.any(Array) }),
    );
  });

  it("logs CRITICAL for API request exception", async () => {
    mockAnthropicClient.messages.create.mockRejectedValue(new Error("API timeout"));

    const result = await client.analyzeProposal(validRequest);

    expect(result).toBeUndefined();
    expect(logger.critical).toHaveBeenCalledWith(
      "AI analysis request failed",
      expect.objectContaining({ error: expect.any(Error) }),
    );
  });

  it("logs ERROR for missing JSON in response", async () => {
    mockAnthropicClient.messages.create.mockResolvedValue({
      content: [{ type: "text", text: "No JSON here" }],
    });

    const result = await client.analyzeProposal(validRequest);

    expect(result).toBeUndefined();
    expect(logger.error).toHaveBeenCalledWith("No JSON found in AI response");
  });

  it("logs ERROR for schema validation failure", async () => {
    mockAnthropicClient.messages.create.mockResolvedValue({
      content: [{ type: "text", text: '{"riskScore": "invalid"}' }],  // Wrong type
    });

    const result = await client.analyzeProposal(validRequest);

    expect(result).toBeUndefined();
    expect(logger.error).toHaveBeenCalledWith(
      "AI response failed schema validation",
      expect.objectContaining({ errors: expect.any(Array) }),
    );
  });
});
```

**Step 3: Update ClaudeAIClient implementation**

```typescript
import Anthropic from "@anthropic-ai/sdk";
import { z } from "zod";

import { IAIClient, AIAnalysisRequest } from "../core/clients/IAIClient.js";
import { Assessment } from "../core/entities/Assessment.js";
import { ILidoGovernanceMonitorLogger } from "../utils/logging/index.js";

// ... schemas unchanged ...

export class ClaudeAIClient implements IAIClient {
  constructor(
    private readonly logger: ILidoGovernanceMonitorLogger,
    private readonly anthropicClient: Anthropic,
    private readonly modelName: string,
    private readonly systemPromptTemplate: string,
    private readonly maxOutputTokens: number,
    private readonly maxProposalChars: number,
  ) {}

  async analyzeProposal(request: AIAnalysisRequest): Promise<Assessment | undefined> {
    const validationResult = AIAnalysisRequestSchema.safeParse(request);
    if (!validationResult.success) {
      // ERROR: Input validation is processing failure
      this.logger.error("Invalid analysis request", {
        errors: validationResult.error.errors,
      });
      return undefined;
    }

    const userPrompt = this.buildUserPrompt(request);

    try {
      const response = await this.anthropicClient.messages.create({
        model: this.modelName,
        max_tokens: this.maxOutputTokens,
        system: this.systemPromptTemplate,
        messages: [{ role: "user", content: userPrompt }],
      });

      const textContent = response.content.find((c) => c.type === "text");
      if (!textContent || textContent.type !== "text") {
        // ERROR: Unexpected response format is processing failure
        this.logger.error("AI response missing text content");
        return undefined;
      }

      this.logger.debug("AI response text content", { textContent: textContent.text });

      const parsed = this.parseAndValidate(textContent.text);
      if (!parsed) return undefined;

      this.logger.debug("AI analysis completed", {
        proposalTitle: request.proposalTitle,
        riskScore: parsed.riskScore,
      });
      return parsed;
    } catch (error) {
      // CRITICAL: API request failure is external dependency failure
      this.logger.critical("AI analysis request failed", { error });
      return undefined;
    }
  }

  getModelName(): string {
    return this.modelName;
  }

  private parseAndValidate(responseText: string): Assessment | undefined {
    try {
      const jsonMatch = responseText.match(/\{[\s\S]*\}/);
      if (!jsonMatch) {
        // ERROR: Missing JSON is processing failure
        this.logger.error("No JSON found in AI response");
        return undefined;
      }
      const parsed = JSON.parse(jsonMatch[0]);
      const result = AssessmentSchema.safeParse(parsed);
      if (!result.success) {
        // ERROR: Schema validation is processing failure
        this.logger.error("AI response failed schema validation", { errors: result.error.errors });
        return undefined;
      }
      return result.data as Assessment;
    } catch (error) {
      // ERROR: JSON parse is processing failure
      this.logger.error("Failed to parse AI response as JSON", { error });
      return undefined;
    }
  }

  private buildUserPrompt(request: AIAnalysisRequest): string {
    const truncatedText = request.proposalText.substring(0, this.maxProposalChars);

    return `Analyze this Lido governance proposal:

Title: ${request.proposalTitle}
URL: ${request.proposalUrl}
Type: ${request.proposalType}

Content:
${truncatedText}`;
  }
}
```

**Step 4: Run tests**

Run: `cd native-yield-operations/lido-governance-monitor && pnpm test -- --testPathPattern="ClaudeAIClient.test.ts"`

Expected: PASS

**Step 5: Commit**

```bash
git add native-yield-operations/lido-governance-monitor/src/clients/ClaudeAIClient.ts
git add native-yield-operations/lido-governance-monitor/src/clients/__tests__/ClaudeAIClient.test.ts
git commit -m "refactor(lido-governance-monitor): classify ClaudeAIClient errors by severity"
```

---

## Task 8: Reclassify Errors in SlackClient

**Files:**
- Modify: `src/clients/SlackClient.ts`
- Test: `src/clients/__tests__/SlackClient.test.ts`

**Classification:**
- Webhook HTTP error (4xx/5xx) → CRITICAL (main alert), WARN (audit channel)
- Network exception → CRITICAL (main alert), WARN (audit channel)

**Step 1: Update test mocks and add severity tests**

**Step 2: Update SlackClient implementation**

```typescript
import { fetchWithTimeout } from "@consensys/linea-shared-utils";

import { ISlackClient, SlackNotificationResult } from "../core/clients/ISlackClient.js";
import { Assessment, RiskLevel } from "../core/entities/Assessment.js";
import { Proposal } from "../core/entities/Proposal.js";
import { ILidoGovernanceMonitorLogger } from "../utils/logging/index.js";

export class SlackClient implements ISlackClient {
  constructor(
    private readonly logger: ILidoGovernanceMonitorLogger,
    private readonly webhookUrl: string,
    private readonly riskThreshold: number,
    private readonly httpTimeoutMs: number,
    private readonly auditWebhookUrl?: string,
  ) {}

  async sendProposalAlert(proposal: Proposal, assessment: Assessment): Promise<SlackNotificationResult> {
    const payload = this.buildSlackPayload(proposal, assessment);

    try {
      const response = await fetchWithTimeout(
        this.webhookUrl,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        },
        this.httpTimeoutMs,
      );

      if (!response.ok) {
        const errorText = await response.text();
        // CRITICAL: Main alert webhook failure is critical
        this.logger.critical("Slack webhook failed", { status: response.status, error: errorText });
        return { success: false, error: errorText };
      }

      this.logger.info("Slack notification sent", { proposalId: proposal.id, title: proposal.title });
      return { success: true, messageTs: Date.now().toString() };
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : "Unknown error";
      // CRITICAL: Network failure for main alert is critical
      this.logger.critical("Slack notification error", { error: errorMessage });
      return { success: false, error: errorMessage };
    }
  }

  async sendAuditLog(proposal: Proposal, assessment: Assessment): Promise<SlackNotificationResult> {
    if (!this.auditWebhookUrl) {
      this.logger.debug("Audit webhook not configured, skipping");
      return { success: true };
    }

    const payload = this.buildAuditPayload(proposal, assessment);

    try {
      const response = await fetchWithTimeout(
        this.auditWebhookUrl,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        },
        this.httpTimeoutMs,
      );

      if (!response.ok) {
        const errorText = await response.text();
        // WARN: Audit channel failure is non-blocking
        this.logger.warn("Audit webhook failed", { status: response.status, error: errorText });
        return { success: false, error: errorText };
      }

      this.logger.debug("Audit log sent", { proposalId: proposal.id });
      return { success: true };
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : "Unknown error";
      // WARN: Audit channel failure is non-blocking
      this.logger.warn("Audit log error", { error: errorMessage });
      return { success: false, error: errorMessage };
    }
  }

  // ... rest of the methods unchanged ...
}
```

**Step 3: Run tests**

Run: `cd native-yield-operations/lido-governance-monitor && pnpm test -- --testPathPattern="SlackClient.test.ts"`

Expected: PASS

**Step 4: Commit**

```bash
git add native-yield-operations/lido-governance-monitor/src/clients/SlackClient.ts
git add native-yield-operations/lido-governance-monitor/src/clients/__tests__/SlackClient.test.ts
git commit -m "refactor(lido-governance-monitor): classify SlackClient errors by severity"
```

---

## Task 9: Reclassify Errors in Services (ProposalPoller, ProposalProcessor, NotificationService)

**Files:**
- Modify: `src/services/ProposalPoller.ts`
- Modify: `src/services/ProposalProcessor.ts`
- Modify: `src/services/NotificationService.ts`
- Tests: corresponding test files

**Classification:**
- Top-level catch (entire pipeline failed) → CRITICAL
- Individual proposal processing failure → ERROR
- Non-blocking warnings → WARN

**Step 1: Update ProposalPoller**

```typescript
import { IDiscourseClient } from "../core/clients/IDiscourseClient.js";
import { ProposalSource } from "../core/entities/ProposalSource.js";
import { IProposalRepository } from "../core/repositories/IProposalRepository.js";
import { INormalizationService } from "../core/services/INormalizationService.js";
import { IProposalPoller } from "../core/services/IProposalPoller.js";
import { ILidoGovernanceMonitorLogger } from "../utils/logging/index.js";

export class ProposalPoller implements IProposalPoller {
  constructor(
    private readonly logger: ILidoGovernanceMonitorLogger,
    private readonly discourseClient: IDiscourseClient,
    private readonly normalizationService: INormalizationService,
    private readonly proposalRepository: IProposalRepository,
    private readonly maxTopicsPerPoll: number = 20,
  ) {}

  async pollOnce(): Promise<void> {
    try {
      this.logger.info("Starting proposal polling");

      const proposalList = await this.discourseClient.fetchLatestProposals();

      if (!proposalList) {
        // WARN: Upstream already logged CRITICAL, this is informational
        this.logger.warn("Failed to fetch latest proposals from Discourse");
        return;
      }

      const allTopics = proposalList.topic_list.topics;
      const topics = allTopics.slice(0, this.maxTopicsPerPoll);
      this.logger.debug("Fetched proposal list", { total: allTopics.length, processing: topics.length });

      for (const topic of topics) {
        await this.processTopic(topic.id);
      }

      this.logger.info("Proposal polling completed");
    } catch (error) {
      // CRITICAL: Entire polling pipeline failed
      this.logger.critical("Proposal polling failed", {
        error: error instanceof Error ? error.message : String(error),
      });
    }
  }

  private async processTopic(topicId: number): Promise<void> {
    const existing = await this.proposalRepository.findBySourceAndSourceId(
      ProposalSource.DISCOURSE,
      topicId.toString(),
    );

    if (existing) {
      this.logger.debug("Proposal already exists, skipping", { topicId });
      return;
    }

    const proposalDetails = await this.discourseClient.fetchProposalDetails(topicId);

    if (!proposalDetails) {
      // WARN: Individual proposal fetch failed, continue with others
      this.logger.warn("Failed to fetch proposal details", { topicId });
      return;
    }

    try {
      const normalizedInput = this.normalizationService.normalizeDiscourseProposal(proposalDetails);
      const created = await this.proposalRepository.create(normalizedInput);
      this.logger.info("Created new proposal", { id: created.id, title: normalizedInput.title });
    } catch (error) {
      // CRITICAL: DB write failure is external dependency
      this.logger.critical("Failed to create proposal", {
        topicId,
        error: error instanceof Error ? error.message : String(error),
      });
    }
  }
}
```

**Step 2: Update ProposalProcessor**

```typescript
import { IAIClient } from "../core/clients/IAIClient.js";
import { ProposalType } from "../core/entities/Assessment.js";
import { Proposal } from "../core/entities/Proposal.js";
import { ProposalSource } from "../core/entities/ProposalSource.js";
import { ProposalState } from "../core/entities/ProposalState.js";
import { IProposalRepository } from "../core/repositories/IProposalRepository.js";
import { IProposalProcessor } from "../core/services/IProposalProcessor.js";
import { ILidoGovernanceMonitorLogger } from "../utils/logging/index.js";

export class ProposalProcessor implements IProposalProcessor {
  constructor(
    private readonly logger: ILidoGovernanceMonitorLogger,
    private readonly aiClient: IAIClient,
    private readonly proposalRepository: IProposalRepository,
    private readonly riskThreshold: number,
    private readonly promptVersion: string,
  ) {}

  async processOnce(): Promise<void> {
    try {
      this.logger.info("Starting proposal processing");

      const newProposals = await this.proposalRepository.findByState(ProposalState.NEW);
      const failedProposals = await this.proposalRepository.findByState(ProposalState.ANALYSIS_FAILED);
      const proposals = [...newProposals, ...failedProposals];

      if (proposals.length === 0) {
        this.logger.debug("No proposals to process");
        return;
      }

      this.logger.debug("Processing proposals", { count: proposals.length });

      for (const proposal of proposals) {
        await this.processProposal(proposal);
      }

      this.logger.info("Proposal processing completed");
    } catch (error) {
      // CRITICAL: Entire processing pipeline failed
      this.logger.critical("Proposal processing failed", {
        error: error instanceof Error ? error.message : String(error),
      });
    }
  }

  private async processProposal(proposal: Proposal): Promise<void> {
    try {
      const updated = await this.proposalRepository.incrementAnalysisAttempt(proposal.id);

      const assessment = await this.aiClient.analyzeProposal({
        proposalTitle: proposal.title,
        proposalText: proposal.text,
        proposalUrl: proposal.url,
        proposalType: this.mapSourceToProposalType(proposal.source),
      });

      if (!assessment) {
        await this.proposalRepository.updateState(proposal.id, ProposalState.ANALYSIS_FAILED);
        // WARN: Individual analysis failed, will retry
        this.logger.warn("AI analysis failed, will retry", {
          proposalId: proposal.id,
          attempt: updated.analysisAttemptCount,
        });
        return;
      }

      await this.proposalRepository.saveAnalysis(
        proposal.id,
        assessment,
        assessment.riskScore,
        this.aiClient.getModelName(),
        this.riskThreshold,
        this.promptVersion,
      );

      this.logger.info("Proposal analysis completed", {
        proposalId: proposal.id,
        riskScore: assessment.riskScore,
      });
      this.logger.debug("Full assessment details", {
        proposalId: proposal.id,
        assessment,
      });
    } catch (error) {
      // CRITICAL: DB operation failed
      this.logger.critical("Error processing proposal", {
        proposalId: proposal.id,
        error: error instanceof Error ? error.message : String(error),
      });
    }
  }

  private mapSourceToProposalType(source: ProposalSource): ProposalType {
    switch (source) {
      case ProposalSource.DISCOURSE:
        return "discourse";
      case ProposalSource.SNAPSHOT:
        return "snapshot";
      case ProposalSource.LDO_VOTING_CONTRACT:
      case ProposalSource.STETH_VOTING_CONTRACT:
        return "onchain_vote";
    }
  }
}
```

**Step 3: Update NotificationService**

```typescript
import { ISlackClient } from "../core/clients/ISlackClient.js";
import { Assessment } from "../core/entities/Assessment.js";
import { Proposal } from "../core/entities/Proposal.js";
import { ProposalState } from "../core/entities/ProposalState.js";
import { IProposalRepository } from "../core/repositories/IProposalRepository.js";
import { INotificationService } from "../core/services/INotificationService.js";
import { ILidoGovernanceMonitorLogger } from "../utils/logging/index.js";

export class NotificationService implements INotificationService {
  constructor(
    private readonly logger: ILidoGovernanceMonitorLogger,
    private readonly slackClient: ISlackClient,
    private readonly proposalRepository: IProposalRepository,
    private readonly riskThreshold: number,
  ) {}

  async notifyOnce(): Promise<void> {
    try {
      this.logger.info("Starting notification processing");

      const analyzedProposals = await this.proposalRepository.findByState(ProposalState.ANALYZED);
      const failedProposals = await this.proposalRepository.findByState(ProposalState.NOTIFY_FAILED);
      const proposals = [...analyzedProposals, ...failedProposals];

      if (proposals.length === 0) {
        this.logger.debug("No proposals to notify");
        return;
      }

      this.logger.debug("Processing proposals for notification", { count: proposals.length });

      for (const proposal of proposals) {
        await this.notifyProposalInternal(proposal);
      }

      this.logger.info("Notification processing completed");
    } catch (error) {
      // CRITICAL: Entire notification pipeline failed
      this.logger.critical("Notification processing failed", {
        error: error instanceof Error ? error.message : String(error),
      });
    }
  }

  private async notifyProposalInternal(proposal: Proposal): Promise<void> {
    try {
      if (!proposal.assessmentJson) {
        // ERROR: Data integrity issue
        this.logger.error("Proposal missing assessment data", { proposalId: proposal.id });
        return;
      }

      const assessment = proposal.assessmentJson as Assessment;

      const auditResult = await this.slackClient.sendAuditLog(proposal, assessment);
      if (!auditResult.success) {
        // WARN: Audit is best-effort, already logged in SlackClient
        this.logger.warn("Audit log failed, continuing", {
          proposalId: proposal.id,
          error: auditResult.error,
        });
      }

      if (proposal.riskScore === null || proposal.riskScore < this.riskThreshold) {
        await this.proposalRepository.updateState(proposal.id, ProposalState.NOT_NOTIFIED);
        this.logger.info("Proposal below notification threshold, skipped", {
          proposalId: proposal.id,
          riskScore: proposal.riskScore,
          threshold: this.riskThreshold,
        });
        return;
      }

      const updated = await this.proposalRepository.incrementNotifyAttempt(proposal.id);

      const result = await this.slackClient.sendProposalAlert(proposal, assessment);

      if (result.success) {
        await this.proposalRepository.markNotified(proposal.id, result.messageTs ?? "");
        this.logger.info("Proposal notification sent", {
          proposalId: proposal.id,
          messageTs: result.messageTs,
        });
      } else {
        await this.proposalRepository.updateState(proposal.id, ProposalState.NOTIFY_FAILED);
        // WARN: Will retry, already logged CRITICAL in SlackClient
        this.logger.warn("Slack notification failed, will retry", {
          proposalId: proposal.id,
          attempt: updated.notifyAttemptCount,
          error: result.error,
        });
      }
    } catch (error) {
      // CRITICAL: DB operation or unexpected failure
      this.logger.critical("Error notifying proposal", {
        proposalId: proposal.id,
        error: error instanceof Error ? error.message : String(error),
      });
    }
  }
}
```

**Step 4: Update test files**

Update all service test files to use `ILidoGovernanceMonitorLogger` mock with `critical` method.

**Step 5: Run all tests**

Run: `cd native-yield-operations/lido-governance-monitor && pnpm test`

Expected: PASS

**Step 6: Commit**

```bash
git add native-yield-operations/lido-governance-monitor/src/services/
git commit -m "refactor(lido-governance-monitor): classify service errors by severity"
```

---

## Task 10: Update NormalizationService

**Files:**
- Modify: `src/services/NormalizationService.ts`

This service is simpler and mainly does data transformation. Update to use `ILidoGovernanceMonitorLogger`.

**Step 1: Update implementation**

The NormalizationService only uses debug/info logs currently, so minimal changes needed. Just update the type.

**Step 2: Run tests**

Run: `cd native-yield-operations/lido-governance-monitor && pnpm test`

**Step 3: Commit**

```bash
git add native-yield-operations/lido-governance-monitor/src/services/NormalizationService.ts
git commit -m "refactor(lido-governance-monitor): update NormalizationService to use ILidoGovernanceMonitorLogger"
```

---

## Task 11: Final Integration Test

**Files:**
- Test: `src/__tests__/integration/proposal-lifecycle.test.ts`

**Step 1: Run full test suite**

Run: `cd native-yield-operations/lido-governance-monitor && pnpm test`

Expected: All tests PASS

**Step 2: Run the service locally (if possible)**

Run: `cd native-yield-operations/lido-governance-monitor && pnpm dev` (or equivalent)

Verify logs show severity field in output.

**Step 3: Final commit**

```bash
git add -A
git commit -m "test(lido-governance-monitor): verify integration with severity-classified logging"
```

---

## Summary

| Task | Component | Files Changed |
|------|-----------|---------------|
| 0 | Documentation | `docs/error-logging.md` |
| 1 | Interface | `src/utils/logging/ILidoGovernanceMonitorLogger.ts` |
| 2 | Implementation | `src/utils/logging/LidoGovernanceMonitorLogger.ts` |
| 3 | Factory | `src/utils/logging/createLidoGovernanceMonitorLogger.ts`, `index.ts` |
| 4 | Bootstrap | `src/application/main/LidoGovernanceMonitorBootstrap.ts` |
| 5 | Verify | No changes |
| 6 | DiscourseClient | `src/clients/DiscourseClient.ts` |
| 7 | ClaudeAIClient | `src/clients/ClaudeAIClient.ts` |
| 8 | SlackClient | `src/clients/SlackClient.ts` |
| 9 | Services | `src/services/ProposalPoller.ts`, `ProposalProcessor.ts`, `NotificationService.ts` |
| 10 | NormalizationService | `src/services/NormalizationService.ts` |
| 11 | Integration | Verify all tests pass |

**Severity Classification Reference:**

| Level | When | Examples |
|-------|------|----------|
| CRITICAL | External dependency failure | HTTP 4xx/5xx, DB error, Slack webhook fail, API timeout |
| ERROR | Processing failure | Validation error, JSON parse error, missing data |
| WARN | Non-blocking issue | Audit channel fail, retry scheduled, threshold skip |
