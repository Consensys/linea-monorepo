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
package maru.consensus.delegated

import java.util.concurrent.TimeUnit
import maru.consensus.ConsensusConfig
import maru.consensus.ForkSpec
import maru.consensus.NewBlockHandler
import maru.consensus.ProtocolFactory
import maru.core.Protocol
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import tech.pegasys.teku.infrastructure.async.SafeFuture

class ElDelegatedConsensusFactory(
  private val ethereumJsonRpcClient: Web3j,
  private val newBlockHandler: NewBlockHandler,
) : ProtocolFactory {
  override fun create(forkSpec: ForkSpec): ElDelegatedConsensus =
    ElDelegatedConsensus(
      ethereumJsonRpcClient = ethereumJsonRpcClient,
      onNewBlock = newBlockHandler,
      blockTimeSeconds = forkSpec.blockTimeSeconds,
    )
}

class ElDelegatedConsensus(
  private val ethereumJsonRpcClient: Web3j,
  private val onNewBlock: NewBlockHandler,
  private val blockTimeSeconds: Int,
) : Protocol {
  // Only for comparisons in the tests to set common ground
  object ElDelegatedConfig : ConsensusConfig

  private val log: Logger = LogManager.getLogger(this::class.java)

  @Volatile
  private var currentTask: SafeFuture<Unit>? = null

  override fun start() {
    if (currentTask == null) {
      poll()
    } else {
      throw IllegalStateException("Timer has already been started!")
    }
  }

  override fun stop() {
    if (currentTask != null) {
      currentTask!!.cancel(false)
      currentTask = null
    } else {
      throw IllegalStateException("EventProducer hasn't been started to stop it!")
    }
  }

  private fun poll(): SafeFuture<Unit> {
    log.debug("Polling EL for new blocks")
    if (currentTask != null) {
      if (!currentTask!!.isDone) {
        log.warn("Current task isn't done. Scheduling the next one, but results may be unexpected!")
      }
      stop()
    }

    return SafeFuture
      .of(
        ethereumJsonRpcClient.ethGetBlockByNumber(DefaultBlockParameter.valueOf("latest"), false).sendAsync(),
      ).thenApply {
        onNewBlock.handleNewBlock(Mapper.mapWeb3jBlockToBeaconBlock(it.block))
      }.handleException {
        log.error(it.message, it)
      }.thenApply {
        SafeFuture
          .runAsync(
            {
              currentTask = poll()
            },
            SafeFuture.delayedExecutor(blockTimeSeconds.toLong(), TimeUnit.SECONDS),
          )
      }
  }
}
