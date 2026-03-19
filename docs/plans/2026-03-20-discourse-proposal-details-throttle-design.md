# Discourse Proposal Details Throttle Design

**Date:** 2026-03-20

**Status:** Approved

**Scope:** `native-yield-operations/lido-governance-monitor`

## Problem

The Lido governance monitor is receiving `429` responses from the Discourse topic-details endpoint while processing proposal topics. The current `DiscourseFetcher` implementation already fetches topic details sequentially, but it issues the next request immediately after the previous one completes. That still creates a rapid burst of serial calls that can exceed Discourse rate limits.

## Decision

Add a configurable delay between individual Discourse topic-details requests inside `DiscourseFetcher`.

This delay will:
- apply only within a single monitor process
- apply only between topic-details requests
- default to `250ms`
- remain configurable through environment-based application config

This change will not introduce cross-process coordination, global rate limiting, or new retry behavior.

## Confirmed Current Behavior

- `ProposalFetcher` runs source fetchers concurrently, so the Discourse fetcher and on-chain fetcher may execute at the same time.
- Inside `DiscourseFetcher`, topic-details requests are already serial because `getLatestProposals()` iterates topics with a sequential `for ... of` loop and awaits `processTopic(topic.id)` on each iteration.
- The production issue is therefore not caused by concurrent `fetchProposalDetails()` calls within a single `DiscourseFetcher` run. It is caused by detail requests being sent too close together.

## Considered Approaches

### 1. Throttle in `DiscourseFetcher` between topic-details requests

Recommended.

This keeps the pacing logic at the call site that creates the request burst today. It is the smallest change, keeps `DiscourseClient` as a thin HTTP client, and makes the behavior easy to test at the fetcher level.

### 2. Throttle inside `DiscourseClient.fetchProposalDetails()`

Rejected.

This would make the client stateful and implicitly throttle every future caller of `fetchProposalDetails()`, even if that caller does not want the behavior.

### 3. Build a reusable rate-limited queue abstraction

Rejected.

This is more abstraction than the package currently needs. There is only one serial request path that requires spacing today.

## Architecture

Extend `discourse` application config with a new numeric field:

- `proposalDetailsDelayMs`

Read it from a new environment variable:

- `DISCOURSE_PROPOSAL_DETAILS_DELAY_MS`

Use `250` as the default value when the environment variable is not set.

Pass the configured value through bootstrap into `DiscourseFetcher`, and keep the existing one-request-at-a-time loop. After completing one topic's processing, wait for `proposalDetailsDelayMs` before starting the next topic-details request. Do not wait before the first topic and do not add a trailing wait after the last topic.

## Runtime Behavior

Polling flow remains:

1. Fetch the Discourse topic list.
2. Slice the list to `maxTopicsPerPoll`.
3. Process each topic sequentially.

Per-topic processing remains:

1. Fetch proposal details.
2. Normalize the proposal.
3. Upsert it to storage.

New behavior:

1. Finish processing the current topic.
2. If another topic remains, wait `proposalDetailsDelayMs`.
3. Start the next topic-details request.

The delay is only a throughput cap. It does not change retry semantics, add exponential backoff logic, or special-case `429` responses. Existing failure handling remains intact:

- if details fetch fails, log and continue
- if normalization fails, log and continue
- if persistence fails, log and continue

In all cases, the next topic still observes the configured inter-request delay.

## Testing Strategy

Add fetcher-level tests that prove:

- topic-details requests remain sequential
- the configured delay is applied between requests
- no delay is applied before the first topic
- the fetcher still continues when one topic fails

To keep tests deterministic, inject an internal sleep dependency into `DiscourseFetcher` with a default implementation used in production. Tests can then assert call order and delay invocations without waiting on real timers.

Add config tests that prove:

- `DISCOURSE_PROPOSAL_DETAILS_DELAY_MS` is parsed into `config.discourse.proposalDetailsDelayMs`
- the default value is `250`
- invalid non-positive values are rejected

Update package documentation and the tracked environment template so operators can discover the new setting.

## Files Expected To Change

- `native-yield-operations/lido-governance-monitor/src/services/fetchers/DiscourseFetcher.ts`
- `native-yield-operations/lido-governance-monitor/src/services/__tests__/fetchers/DiscourseFetcher.test.ts`
- `native-yield-operations/lido-governance-monitor/src/application/main/config/index.ts`
- `native-yield-operations/lido-governance-monitor/src/application/main/__tests__/config.test.ts`
- `native-yield-operations/lido-governance-monitor/src/application/main/LidoGovernanceMonitorBootstrap.ts`
- `native-yield-operations/lido-governance-monitor/.env.sample`
- `native-yield-operations/lido-governance-monitor/README.md`

## Non-Goals

- no distributed coordination across replicas or overlapping runs
- no new dependency for rate limiting
- no change to `DiscourseClient` retry behavior
- no change to on-chain fetch pacing
