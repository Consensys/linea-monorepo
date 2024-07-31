package net.consensys.linea.async

import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.CompletableFuture

fun <T> CompletableFuture<T>.toSafeFuture(): SafeFuture<T> = SafeFuture.of(this)
