# Lido Governance Monitor

## Overview

The Lido Governance Monitor is an off-chain service that continuously monitors Lido governance proposals and alerts the Security Council when high-risk proposals may impact Linea's Native Yield system.

The service implements a polling-based architecture with three concurrent processing pipelines:

1. **ProposalPoller** - Fetches new proposals from Lido's Discourse forum at configurable intervals
2. **ProposalProcessor** - Performs AI-assisted risk analysis using Claude API, scoring proposals 0-100
3. **NotificationService** - Sends Slack alerts for proposals exceeding the risk threshold

Proposals flow through a state machine: `NEW` → `ANALYZED` → `NOTIFIED` or `NOT_NOTIFIED`. Failed operations are automatically retried on subsequent cycles.

## Codebase Architecture

The codebase follows a **Layered Architecture with Dependency Inversion**, incorporating concepts from Hexagonal Architecture (Ports and Adapters) and Domain-Driven Design:

- **`core/`** - Domain layer containing interfaces (ports), entities, and enums. This layer has no dependencies on other internal layers.
- **`services/`** - Application layer containing business logic that orchestrates operations using interfaces from `core/`.
- **`clients/`** - Infrastructure layer containing adapter implementations of interfaces defined in `core/`.
- **`application/`** - Composition layer that wires dependencies and bootstraps the service.

Dependencies flow inward: `application` → `services/clients` → `core`. This ensures business logic remains independent of infrastructure concerns, making the codebase testable and maintainable.

## Folder Structure

```
lido-governance-monitor/
├── prisma/
│   └── schema.prisma         # Database schema for proposal state tracking
├── src/
│   ├── application/          # Application bootstrap and configuration
│   │   └── main/
│   │       ├── config/       # Zod-validated configuration schema
│   │       └── LidoGovernanceMonitorBootstrap.ts
│   ├── clients/              # External service clients
│   │   ├── db/               # Prisma repository implementation
│   │   ├── ClaudeAIClient.ts # Anthropic Claude API client
│   │   ├── DiscourseClient.ts # Lido forum API client
│   │   └── SlackClient.ts    # Slack webhook client
│   ├── core/                 # Interfaces and domain entities
│   │   ├── clients/          # Client interfaces (IAIClient, IDiscourseClient, ISlackClient)
│   │   ├── entities/         # Domain entities (Proposal, Assessment, ProposalState)
│   │   ├── repositories/     # Repository interfaces
│   │   └── services/         # Service interfaces
│   ├── prompts/              # AI system prompts for risk assessment
│   │   └── risk-assessment-system.md
│   ├── services/             # Business logic services
│   │   ├── NormalizationService.ts  # Converts raw Discourse data to domain entities
│   │   ├── NotificationService.ts   # Sends Slack alerts for high-risk proposals
│   │   ├── ProposalPoller.ts        # Fetches proposals from Discourse
│   │   └── ProposalProcessor.ts     # AI-assisted risk analysis
│   └── __tests__/
│       └── integration/      # Integration tests for proposal lifecycle
└── run.ts                    # Service entry point
```

## Proposal State Machine

```
                    ┌──────────────────┐
                    │       NEW        │
                    └────────┬─────────┘
                             │ AI Analysis
                    ┌────────▼─────────┐
                    │     ANALYZED     │
                    └────────┬─────────┘
                             │
              ┌──────────────┴──────────────┐
              │                             │
     riskScore >= threshold        riskScore < threshold
              │                             │
     ┌────────▼─────────┐          ┌───────▼────────┐
     │     NOTIFIED     │          │  NOT_NOTIFIED  │
     └──────────────────┘          └────────────────┘

Failed states (ANALYSIS_FAILED, NOTIFY_FAILED) are automatically
retried on subsequent processing cycles.
```

## Configuration

See the [configuration schema file](./src/application/main/config/index.ts)

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Required |
| `DISCOURSE_PROPOSALS_URL` | Full URL to Lido proposals feed | Required |
| `DISCOURSE_POLLING_INTERVAL_MS` | How often to poll for new proposals | `3600000` (1 hour) |
| `ANTHROPIC_API_KEY` | Claude API key | Required |
| `CLAUDE_MODEL` | Claude model to use | `claude-sonnet-4-20250514` |
| `SLACK_WEBHOOK_URL` | Slack incoming webhook URL | Required |
| `RISK_THRESHOLD` | Score (0-100) above which to notify | `60` |
| `PROMPT_VERSION` | Version identifier for risk prompt | `v1.0` |
| `PROCESSING_INTERVAL_MS` | How often to process proposals | `60000` (1 minute) |

## Development

### Prerequisites

- Node.js 18+
- PostgreSQL database
- Anthropic API key
- Slack incoming webhook

### Setup

1. Install dependencies:
   ```bash
   pnpm install
   ```

2. Generate Prisma client:
   ```bash
   pnpm db:generate
   ```

3. Create database tables:
   ```bash
   pnpm db:push
   ```

4. Create `.env` file with required configuration (see Configuration section)

### Running

```bash
pnpm --filter @consensys/lido-governance-monitor exec tsx run.ts
```

### Build

```bash
pnpm --filter @consensys/lido-governance-monitor build
```

### Unit Tests

```bash
pnpm --filter @consensys/lido-governance-monitor test
```

### Lint

```bash
pnpm --filter @consensys/lido-governance-monitor lint
```

## License

This package is licensed under the [Apache 2.0](../../LICENSE-APACHE) and the [MIT](../../LICENSE-MIT) licenses.
