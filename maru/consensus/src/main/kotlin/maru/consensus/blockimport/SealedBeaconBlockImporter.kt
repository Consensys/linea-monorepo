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

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import com.github.michaelbull.result.flatMap
import com.github.michaelbull.result.mapError
import maru.consensus.AsyncFunction
import maru.consensus.CallAndForgetFutureMultiplexer
import maru.consensus.state.StateTransition
import maru.consensus.validation.BeaconBlockValidatorFactory
import maru.consensus.validation.SealsVerifier
import maru.core.SealedBeaconBlock
import maru.database.BeaconChain
import maru.extensions.encodeHex
import maru.p2p.SealedBeaconBlockHandler
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture

// This is basically Chain of Responsibility design pattern, except it doesn't allow multiple children
// Multiplexer class was created to address that
fun interface SealedBeaconBlockImporter {
  fun importBlock(sealedBeaconBlock: SealedBeaconBlock): SafeFuture<*>
}

class NewSealedBeaconBeaconBlockHandlerMultiplexer(
  handlersMap: Map<String, SealedBeaconBlockHandler>,
  log: Logger = LogManager.getLogger(CallAndForgetFutureMultiplexer<*, *>::javaClass)!!,
) : CallAndForgetFutureMultiplexer<SealedBeaconBlock, Unit>(
    handlersMap = sealedBlockHandlersToGenericHandlers(handlersMap),
    log = log,
  ),
  SealedBeaconBlockHandler {
  companion object {
    fun sealedBlockHandlersToGenericHandlers(
      handlersMap: Map<String, SealedBeaconBlockHandler>,
    ): Map<String, AsyncFunction<SealedBeaconBlock, Unit>> =
      handlersMap.mapValues { newSealedBlockHandler ->
        {
          newSealedBlockHandler.value.handleSealedBlock(it).thenApply { }
        }
      }
  }

  override fun Logger.logError(
    handlerName: String,
    input: SealedBeaconBlock,
    ex: Exception,
  ) {
    this.error(
      "New sealed block handler $handlerName failed processing" +
        "blockHash=${input.beaconBlock.beaconBlockHeader.hash}, number=${input.beaconBlock.beaconBlockHeader.number} " +
        "executionPayloadBlockNumber=${input.beaconBlock.beaconBlockBody.executionPayload.blockNumber}!",
      ex,
    )
  }

  override fun handleSealedBlock(sealedBeaconBlock: SealedBeaconBlock): SafeFuture<*> = handle(sealedBeaconBlock)
}

/**
 * Responsible for: transactional  and El node
 * 1. state transition of node's BeaconChain
 * 2. new block import into an EL node
 * The import is transactional, I.e. all or nothing approach
 */
class TransactionalSealedBeaconBlockImporter(
  private val beaconChain: BeaconChain,
  private val stateTransition: StateTransition,
  private val beaconBlockImporter: BeaconBlockImporter,
) : SealedBeaconBlockImporter {
  private val log: Logger = LogManager.getLogger(this::javaClass)

  override fun importBlock(sealedBeaconBlock: SealedBeaconBlock): SafeFuture<*> {
    val updater = beaconChain.newUpdater()
    try {
      return stateTransition
        .processBlock(sealedBeaconBlock.beaconBlock)
        .thenCompose { resultingState ->
          updater
            .putBeaconState(resultingState)
            .putSealedBeaconBlock(sealedBeaconBlock)
          beaconBlockImporter
            .importBlock(resultingState, sealedBeaconBlock.beaconBlock)
        }.thenApply {
          updater.commit()
          it
        }.whenException {
          updater.rollback()
        }.whenComplete { _, _ ->
          updater.close()
        }
    } catch (e: Exception) {
      log.error("Block import state transition failed!: ${e.message}", e)
      return SafeFuture.failedFuture<Unit>(e)
    }
  }
}

/**
 * Verifies the seal and delegates to another beaconBlockImporter
 */
class ValidatingSealedBeaconBlockImporter(
  private val sealsVerifier: SealsVerifier,
  private val beaconBlockImporter: SealedBeaconBlockImporter,
  private val beaconBlockValidatorFactory: BeaconBlockValidatorFactory,
) : SealedBeaconBlockImporter {
  private val log = LogManager.getLogger(this.javaClass)

  override fun importBlock(sealedBeaconBlock: SealedBeaconBlock): SafeFuture<Result<*, String>> {
    try {
      val beaconBlock = sealedBeaconBlock.beaconBlock
      val beaconBlockHeader = beaconBlock.beaconBlockHeader
      log.debug("Received beacon block blockNumber={} hash={}", beaconBlockHeader.number, beaconBlockHeader.hash)
      val blockValidators =
        beaconBlockValidatorFactory
          .createValidatorForBlock(beaconBlockHeader)
      return sealsVerifier
        .verifySeals(sealedBeaconBlock.commitSeals, beaconBlockHeader)
        .thenComposeCombined(
          blockValidators.validateBlock(beaconBlock),
        ) { sealsVerificationResult, blockValidationResult ->
          val combinedValidationResult =
            sealsVerificationResult.flatMap { blockValidationResult.mapError { it.message } }
          when (combinedValidationResult) {
            is Ok -> {
              log.debug("Block is validated blockNumber={} hash={}", beaconBlockHeader.number, beaconBlockHeader.hash)
              beaconBlockImporter.importBlock(sealedBeaconBlock).thenApply { Ok(it) }
            }

            is Err -> {
              log.error(
                "Validation failed for blockNumber=${sealedBeaconBlock.beaconBlock.beaconBlockHeader.number}, " +
                  "hash=${sealedBeaconBlock.beaconBlock.beaconBlockHeader.hash.encodeHex()}! " +
                  "error=${combinedValidationResult.error}",
              )
              SafeFuture.completedFuture(combinedValidationResult)
            }
          }
        }.whenException {
          log.error("Exception during sealed block import!", it)
        }
    } catch (ex: Throwable) {
      log.error("Exception during sealed block import!", ex)
      throw ex
    }
  }
}
