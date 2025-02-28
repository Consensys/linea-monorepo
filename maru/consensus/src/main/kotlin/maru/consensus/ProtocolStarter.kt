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

import java.util.concurrent.atomic.AtomicReference
import maru.core.Protocol
import maru.executionlayer.client.ExecutionLayerClient
import maru.executionlayer.manager.BlockMetadata
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.hyperledger.besu.ethereum.core.Block

class MetadataOnlyHandlerAdapter(
  private val protocolStarter: ProtocolStarter,
) : NewBlockHandler {
  override fun handleNewBlock(block: Block) {
    val blockHeader = block.header
    val blockMetadata =
      BlockMetadata(
        blockHeader.number.toULong(),
        blockHeader.blockHash.toArray(),
        blockHeader.timestamp,
      )
    protocolStarter.handleNewBlock(blockMetadata)
  }
}

class ProtocolStarter(
  private val forksSchedule: ForksSchedule,
  private val protocolFactory: ProtocolFactory,
  private val executionLayerClient: ExecutionLayerClient,
) : Protocol {
  data class ProtocolWithConfig(
    val protocol: Protocol,
    val config: ConsensusConfig,
  )

  private val log: Logger = LogManager.getLogger(this::class.java)

  internal val currentProtocolWithConfig: AtomicReference<ProtocolWithConfig> = AtomicReference()

  @Synchronized
  fun handleNewBlock(block: BlockMetadata) {
    log.debug("New block {} received", { block.blockNumber })
    val latestBlockNumber = block.blockNumber
    val nextForkSpec = forksSchedule.getForkByNumber(latestBlockNumber + 1UL)

    val currentProtocol = currentProtocolWithConfig.get()
    if (currentProtocol?.config != nextForkSpec) {
      val newProtocol: Protocol = protocolFactory.create(nextForkSpec)

      val newProtocolWithConfig =
        ProtocolWithConfig(
          newProtocol,
          nextForkSpec,
        )
      log.debug("Switching from {} to protocol {}", currentProtocol, newProtocolWithConfig)
      currentProtocolWithConfig.set(
        newProtocolWithConfig,
      )
      currentProtocol?.protocol?.stop()
      newProtocol.start()
    } else {
      log.trace("Block {} was produced, but the fork switch isn't required", { block.blockNumber })
    }
  }

  override fun start() {
    val latestBlock = executionLayerClient.getLatestBlockMetadata().get()
    handleNewBlock(latestBlock)
  }

  override fun stop() {
    currentProtocolWithConfig.get().protocol.stop()
    currentProtocolWithConfig.set(null)
  }
}
