package io.vertx.ext.web.healthchecks

import io.vertx.core.Future
import io.vertx.core.Handler
import io.vertx.core.Promise
import io.vertx.core.Vertx
import io.vertx.ext.auth.authentication.AuthenticationProvider
import io.vertx.ext.healthchecks.CheckResult
import io.vertx.ext.healthchecks.HealthChecks
import io.vertx.ext.healthchecks.Status
import io.vertx.ext.web.RoutingContext
import java.util.function.Function
import io.vertx.ext.healthchecks.HealthCheckHandler as Vertx4HealthCheckHandler

/**
 * Compatibility adapter for the Vert.x health-check package move.
 *
 * The monorepo ObservabilityServer is compiled against Vert.x 5, where this interface lives under
 * io.vertx.ext.web.healthchecks. Besu 26.5.0 test utilities still require Vert.x 4 at runtime, where the
 * same API lives under io.vertx.ext.healthchecks. Keep this adapter in Maru test-utils so only test-support
 * classpaths see it.
 */
interface HealthCheckHandler : Handler<RoutingContext> {
  companion object {
    @JvmStatic
    fun create(
      vertx: Vertx,
      authenticationProvider: AuthenticationProvider?,
    ): HealthCheckHandler =
      DelegatingHealthCheckHandler(Vertx4HealthCheckHandler.create(vertx, authenticationProvider))

    @JvmStatic
    fun create(vertx: Vertx): HealthCheckHandler =
      DelegatingHealthCheckHandler(Vertx4HealthCheckHandler.create(vertx))

    @JvmStatic
    fun createWithHealthChecks(
      healthChecks: HealthChecks,
      authenticationProvider: AuthenticationProvider?,
    ): HealthCheckHandler =
      DelegatingHealthCheckHandler(
        Vertx4HealthCheckHandler.createWithHealthChecks(healthChecks, authenticationProvider),
      )

    @JvmStatic
    fun createWithHealthChecks(healthChecks: HealthChecks): HealthCheckHandler =
      DelegatingHealthCheckHandler(Vertx4HealthCheckHandler.createWithHealthChecks(healthChecks))
  }

  fun register(
    name: String,
    procedure: Handler<Promise<Status>>,
  ): HealthCheckHandler

  fun register(
    name: String,
    timeout: Long,
    procedure: Handler<Promise<Status>>,
  ): HealthCheckHandler

  fun unregister(name: String): HealthCheckHandler

  fun resultMapper(mapper: Function<CheckResult, Future<CheckResult>>): HealthCheckHandler
}

private class DelegatingHealthCheckHandler(
  private val delegate: Vertx4HealthCheckHandler,
) : HealthCheckHandler {
  override fun handle(event: RoutingContext) {
    delegate.handle(event)
  }

  override fun register(
    name: String,
    procedure: Handler<Promise<Status>>,
  ): HealthCheckHandler {
    delegate.register(name, procedure)
    return this
  }

  override fun register(
    name: String,
    timeout: Long,
    procedure: Handler<Promise<Status>>,
  ): HealthCheckHandler {
    delegate.register(name, timeout, procedure)
    return this
  }

  override fun unregister(name: String): HealthCheckHandler {
    delegate.unregister(name)
    return this
  }

  override fun resultMapper(mapper: Function<CheckResult, Future<CheckResult>>): HealthCheckHandler {
    delegate.resultMapper(mapper)
    return this
  }
}
