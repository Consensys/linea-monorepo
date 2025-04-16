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
import org.apache.logging.log4j.LogManager
import tech.pegasys.teku.infrastructure.async.SafeFuture

fun interface NewBlockHandler {
  fun handleNewBlock(block: BeaconBlock)
}

class NewBlockHandlerMultiplexer(
  handlersMap: Map<String, NewBlockHandler>,
) : NewBlockHandler {
  private val handlersMap = ConcurrentHashMap(handlersMap)
  private val log = LogManager.getLogger(NewBlockHandlerMultiplexer::class.java)!!

  fun addHandler(
    name: String,
    handler: NewBlockHandler,
  ) {
    handlersMap[name] = handler
  }

  override fun handleNewBlock(block: BeaconBlock) {
    val handlerFutures: List<CompletableFuture<Void>> =
      handlersMap.map {
        val (handlerName, handler) = it
        SafeFuture.runAsync {
          try {
            log.debug("Handling $handlerName")
            handler.handleNewBlock(block)
            log.debug("$handlerName handling completed successfully")
          } catch (ex: Exception) {
            log.error(
              "New block handler $handlerName failed processing" +
                " block hash=${block.beaconBlockHeader.hash}, number=${block.beaconBlockHeader.number} " +
                "executionPayloadBlockNumber=${block.beaconBlockBody.executionPayload.blockNumber}!",
              ex,
            )
          }
        }
      }
    SafeFuture.allOf(*handlerFutures.toTypedArray()).get()
  }
}
