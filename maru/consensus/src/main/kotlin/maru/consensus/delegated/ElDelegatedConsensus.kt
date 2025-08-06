/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package maru.consensus.delegated

import java.util.concurrent.Executors
import java.util.concurrent.ScheduledExecutorService
import java.util.concurrent.TimeUnit
import maru.consensus.ForkSpec
import maru.consensus.NewBlockHandler
import maru.consensus.ProtocolFactory
import maru.core.BeaconBlock
import maru.core.BeaconBlockBody
import maru.core.BeaconBlockHeader
import maru.core.EMPTY_HASH
import maru.core.ExecutionPayload
import maru.core.Protocol
import maru.core.Validator
import maru.mappers.Mappers.toDomain
import maru.serialization.rlp.RLPSerializers
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameter
import tech.pegasys.teku.infrastructure.async.SafeFuture

class ElDelegatedConsensusFactory(
  private val ethereumJsonRpcClient: Web3j,
  private val newBlockHandler: NewBlockHandler<*>,
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
  private val onNewBlock: NewBlockHandler<*>,
  private val blockTimeSeconds: Int,
  private val executor: ScheduledExecutorService =
    Executors.newSingleThreadScheduledExecutor(
      Thread.ofVirtual().factory(),
    ),
) : Protocol {
  private val log: Logger = LogManager.getLogger(this.javaClass)

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
    if (currentTask != null && !currentTask!!.isDone) {
      log.warn("Current task isn't done. Cancelling it before scheduling the next one.")
      currentTask!!.cancel(false)
    }

    val future =
      SafeFuture
        .of(
          ethereumJsonRpcClient.ethGetBlockByNumber(DefaultBlockParameter.valueOf("latest"), true).sendAsync(),
        ).thenApply {
          onNewBlock.handleNewBlock(wrapIntoDummyBeaconBlock(it.block.toDomain()))
        }.handleException {
          log.error(it.message, it)
        }.thenCompose {
          SafeFuture
            .runAsync(
              {
                poll()
              },
              SafeFuture.delayedExecutor(blockTimeSeconds.toLong(), TimeUnit.SECONDS, executor),
            ).thenApply { }
        }

    currentTask = future
    return future
  }

  private fun wrapIntoDummyBeaconBlock(executionPayload: ExecutionPayload): BeaconBlock {
    val beaconBlockBody = BeaconBlockBody(prevCommitSeals = emptySet(), executionPayload = executionPayload)

    val beaconBlockHeader =
      BeaconBlockHeader(
        number = 0u,
        round = 0u,
        timestamp = executionPayload.timestamp,
        proposer = Validator(executionPayload.feeRecipient),
        parentRoot = EMPTY_HASH,
        stateRoot = EMPTY_HASH,
        bodyRoot = EMPTY_HASH,
        headerHashFunction = RLPSerializers.DefaultHeaderHashFunction,
      )
    return BeaconBlock(beaconBlockHeader, beaconBlockBody)
  }
}
