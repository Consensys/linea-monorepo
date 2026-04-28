package net.consensys.zkevm.persistence.db

import tech.pegasys.teku.infrastructure.async.SafeFuture

abstract class RetryingDaoBase<D>(
  protected val delegate: D,
  private val persistenceRetryer: PersistenceRetryer,
) {
  protected fun <T> retrying(query: () -> SafeFuture<T>): SafeFuture<T> =
    persistenceRetryer.retryQuery(query)
}
