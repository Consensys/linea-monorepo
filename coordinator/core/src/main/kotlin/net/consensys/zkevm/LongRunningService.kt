package net.consensys.zkevm

import java.util.concurrent.CompletableFuture

interface LongRunningService {
  fun start(): CompletableFuture<Unit>
  fun stop(): CompletableFuture<Unit>
}
