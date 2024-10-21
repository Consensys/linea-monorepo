package net.consensys.linea.jsonrpc.client

import java.util.concurrent.atomic.AtomicLong
import java.util.function.Supplier
import javax.annotation.concurrent.ThreadSafe

@ThreadSafe
class SequentialIdSupplier : Supplier<Any> {
  private var id = AtomicLong(0)

  // if application makes 1_000 requests per second, it will take 292,277,026,596 years of uptime to overflow
  override fun get(): Any {
    return id.incrementAndGet()
  }

  companion object {
    val singleton = SequentialIdSupplier()
  }
}
