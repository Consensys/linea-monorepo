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
package maru.consensus.blockimport

import kotlin.text.contains
import maru.core.ext.DataGenerators
import maru.p2p.SealedBeaconBlockHandler
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

class NewSealedBeaconBlockHandlerMultiplexerTest {
  @Test
  fun `should invoke all handlers for SealedBeaconBlock`() {
    val sealedBlock = DataGenerators.randomSealedBeaconBlock(1u)
    val handler1 =
      mock<SealedBeaconBlockHandler<Unit>> {
        on { handleSealedBlock(any()) } doReturn SafeFuture.completedFuture(Unit)
      }
    val handler2 =
      mock<SealedBeaconBlockHandler<Unit>> {
        on { handleSealedBlock(any()) } doReturn SafeFuture.completedFuture(Unit)
      }
    val multiplexer =
      NewSealedBeaconBlockHandlerMultiplexer<Unit>(
        handlersMap = mapOf("h1" to handler1, "h2" to handler2),
      )

    val future = multiplexer.handleSealedBlock(sealedBlock)
    future.join()

    verify(handler1).handleSealedBlock(sealedBlock)
    verify(handler2).handleSealedBlock(sealedBlock)
    assertThat(future.isDone).isTrue()
  }

  @Test
  fun `should log and throw error if sealed handler throws`() {
    val sealedBlock = DataGenerators.randomSealedBeaconBlock(1u)
    val handler =
      mock<SealedBeaconBlockHandler<Unit>> {
        on { handleSealedBlock(any()) } doThrow RuntimeException("fail")
      }
    val logger: Logger = mock()
    val multiplexer =
      NewSealedBeaconBlockHandlerMultiplexer<Unit>(
        handlersMap = mapOf(pair = "h" to handler),
        log = logger,
      )

    assertThrows<Throwable> {
      multiplexer.handleSealedBlock(sealedBlock).get()
    }
    verify(logger).error(
      argThat<String> {
        contains("New sealed block handler h failed processing") &&
          contains("number=${sealedBlock.beaconBlock.beaconBlockHeader.number}")
      },
      any<Throwable>(),
    )
  }
}
