/*
   Copyright 2025 Consensys Software Inc.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */
package maru.consensus

import kotlin.text.contains
import maru.core.ext.DataGenerators
import org.apache.logging.log4j.Logger
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import org.mockito.kotlin.any
import org.mockito.kotlin.argThat
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.doThrow
import org.mockito.kotlin.mock
import org.mockito.kotlin.verify
import tech.pegasys.teku.infrastructure.async.SafeFuture

class NewBlockHandlerMultiplexerTest {
  @Test
  fun `should invoke all handlers for BeaconBlock`() {
    val block = DataGenerators.randomBeaconBlock(1u)
    val handler1 =
      mock<NewBlockHandler<Unit>> {
        on { handleNewBlock(any()) } doReturn SafeFuture.completedFuture(Unit)
      }
    val handler2 =
      mock<NewBlockHandler<Unit>> {
        on { handleNewBlock(any()) } doReturn SafeFuture.completedFuture(Unit)
      }
    val multiplexer =
      NewBlockHandlerMultiplexer(
        mapOf("h1" to handler1, "h2" to handler2),
      )

    val future = multiplexer.handleNewBlock(block)
    future.join()

    verify(handler1).handleNewBlock(block)
    verify(handler2).handleNewBlock(block)
    assertThat(future.isDone).isTrue()
  }

  @Test
  fun `should log and throw error if handler throws`() {
    val block = DataGenerators.randomBeaconBlock(1u)
    val handler =
      mock<NewBlockHandler<Unit>> {
        on { handleNewBlock(any()) } doThrow RuntimeException("fail")
      }
    val logger: Logger = mock<Logger>()
    val multiplexer = NewBlockHandlerMultiplexer(handlersMap = mapOf(pair = "h" to handler), log = logger)

    assertThrows<Throwable> {
      multiplexer.handleNewBlock(block).get()
    }
    verify(logger).error(
      argThat<String> {
        contains("New block handler h failed processing") &&
          contains("number=${block.beaconBlockHeader.number}")
      },
      any<Throwable>(),
    )
  }
}
