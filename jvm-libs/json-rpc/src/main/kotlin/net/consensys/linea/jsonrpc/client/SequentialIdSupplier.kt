package net.consensys.linea.jsonrpc.client

import java.math.BigInteger
import java.util.function.Supplier
import javax.annotation.concurrent.ThreadSafe

@ThreadSafe
class SequentialIdSupplier : Supplier<Any> {
  private var id = BigInteger.ZERO

  @Synchronized
  override fun get(): Any {
    val nexId = id.add(BigInteger.ONE)
    id = nexId
    return nexId
  }

  companion object {
    val singleton = SequentialIdSupplier()
  }
}
