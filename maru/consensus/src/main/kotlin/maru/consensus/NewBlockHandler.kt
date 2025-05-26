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

import java.util.concurrent.CompletableFuture
import java.util.concurrent.ConcurrentHashMap
import maru.core.BeaconBlock
import maru.core.SealedBeaconBlock
import maru.p2p.SealedBeaconBlockHandler
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture

class SealedBeaconBlockHandlerAdapter(
  val adaptee: NewBlockHandler,
) : SealedBeaconBlockHandler {
  override fun handleSealedBlock(sealedBeaconBlock: SealedBeaconBlock): SafeFuture<*> =
    adaptee.handleNewBlock(sealedBeaconBlock.beaconBlock)
}

fun interface NewBlockHandler {
  fun handleNewBlock(beaconBlock: BeaconBlock): SafeFuture<*>
}

typealias AsyncFunction<I, O> = (I) -> SafeFuture<O>

abstract class CallAndForgetFutureMultiplexer<I, O>(
  handlersMap: Map<String, AsyncFunction<I, O>>,
  protected val log: Logger = LogManager.getLogger(CallAndForgetFutureMultiplexer<*, *>::javaClass)!!,
) {
  private val handlersMap = ConcurrentHashMap(handlersMap)

  protected abstract fun Logger.logError(
    handlerName: String,
    input: I,
    ex: Exception,
  )

  fun addHandler(
    name: String,
    handler: AsyncFunction<I, O>,
  ) {
    handlersMap[name] = handler
  }

  fun handle(input: I): SafeFuture<*> {
    val handlerFutures: List<CompletableFuture<Void>> =
      handlersMap.map {
        val (handlerName, handler) = it
        SafeFuture.runAsync {
          try {
            log.debug("Handling $handlerName")
            handler(input)
            log.debug("$handlerName handling completed successfully")
          } catch (ex: Exception) {
            log.logError(handlerName, input, ex)
          }
        }
      }
    val completableFuture =
      SafeFuture
        .allOf(*handlerFutures.toTypedArray())
        .thenApply { }
    return SafeFuture.of(completableFuture)
  }
}

class NewBlockHandlerMultiplexer(
  handlersMap: Map<String, NewBlockHandler>,
  log: Logger = LogManager.getLogger(CallAndForgetFutureMultiplexer<*, *>::javaClass)!!,
) : CallAndForgetFutureMultiplexer<BeaconBlock, Unit>(
    handlersMap = blockHandlersToGenericHandlers(handlersMap),
    log = log,
  ),
  NewBlockHandler {
  companion object {
    fun blockHandlersToGenericHandlers(
      handlersMap: Map<String, NewBlockHandler>,
    ): Map<String, AsyncFunction<BeaconBlock, Unit>> =
      handlersMap.mapValues { newSealedBlockHandler ->
        {
          newSealedBlockHandler.value.handleNewBlock(it).thenApply { }
        }
      }
  }

  override fun Logger.logError(
    handlerName: String,
    input: BeaconBlock,
    ex: Exception,
  ) {
    this.error(
      "New block handler $handlerName failed processing" +
        " block hash=${input.beaconBlockHeader.hash}, number=${input.beaconBlockHeader.number} " +
        "executionPayloadBlockNumber=${input.beaconBlockBody.executionPayload.blockNumber}!",
      ex,
    )
  }

  override fun handleNewBlock(beaconBlock: BeaconBlock): SafeFuture<*> = handle(beaconBlock)
}
