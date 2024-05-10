package net.consensys.zkevm.ethereum.coordination.common

import net.consensys.zkevm.ethereum.coordination.SimpleCompositeSafeFutureHandler
import org.junit.jupiter.api.Test
import org.mockito.ArgumentMatchers.eq
import org.mockito.Mockito.spy
import org.mockito.Mockito.times
import org.mockito.kotlin.verify
import tech.pegasys.teku.infrastructure.async.SafeFuture

class SimpleCompositeSafeFutureHandlerTest {
  @Test
  fun `SimpleCompositeSafeFutureHandler propagates argument to all the handlers exactly once`() {
    val testHandler = { _: Long -> SafeFuture.completedFuture(Unit) }
    val handlers = listOf(
      spy(testHandler),
      spy(testHandler)
    )

    val expectedArgument = 13L
    SimpleCompositeSafeFutureHandler(handlers).invoke(expectedArgument)

    handlers.forEach {
      verify(it, times(1)).invoke(eq(expectedArgument))
    }
  }
}
