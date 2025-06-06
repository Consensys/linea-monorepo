/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
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
import maru.p2p.ValidationResult
import org.apache.logging.log4j.LogManager
import org.apache.logging.log4j.Logger
import tech.pegasys.teku.infrastructure.async.SafeFuture

// This is basically Chain of Responsibility design pattern, except it doesn't allow multiple children
// Multiplexer class was created to address that
fun interface SealedBeaconBlockImporter<T> {
  fun importBlock(sealedBeaconBlock: SealedBeaconBlock): SafeFuture<T>
}

class NewSealedBeaconBlockHandlerMultiplexer<T>(
  handlersMap: Map<String, SealedBeaconBlockHandler<*>>,
  log: Logger = LogManager.getLogger(CallAndForgetFutureMultiplexer<*>::javaClass)!!,
) : CallAndForgetFutureMultiplexer<SealedBeaconBlock>(
    handlersMap = sealedBlockHandlersToGenericHandlers(handlersMap),
    log = log,
  ),
  SealedBeaconBlockHandler<Unit> {
  companion object {
    fun sealedBlockHandlersToGenericHandlers(
      handlersMap: Map<String, SealedBeaconBlockHandler<*>>,
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

  override fun handleSealedBlock(sealedBeaconBlock: SealedBeaconBlock): SafeFuture<Unit> = handle(sealedBeaconBlock)
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
) : SealedBeaconBlockImporter<ValidationResult> {
  private val log: Logger = LogManager.getLogger(this::javaClass)

  override fun importBlock(sealedBeaconBlock: SealedBeaconBlock): SafeFuture<ValidationResult> {
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
          ValidationResult.Companion.Valid as ValidationResult
        }.exceptionally { ex ->
          updater.rollback()
          ValidationResult.Companion.Invalid(ex.message!!, ex.cause)
        }.whenComplete { _, _ ->
          updater.close()
        }
    } catch (e: Exception) {
      log.error("Block import state transition failed!: ${e.message}", e)
      return SafeFuture.failedFuture(e)
    }
  }
}

/**
 * Verifies the seal and delegates to another beaconBlockImporter
 */
class ValidatingSealedBeaconBlockImporter(
  private val sealsVerifier: SealsVerifier,
  private val beaconBlockImporter: SealedBeaconBlockImporter<ValidationResult>,
  private val beaconBlockValidatorFactory: BeaconBlockValidatorFactory,
) : SealedBeaconBlockImporter<ValidationResult> {
  companion object {
    fun Result<Unit, String>.toDomain(): ValidationResult =
      when (this) {
        is Ok -> ValidationResult.Companion.Valid
        is Err -> ValidationResult.Companion.Invalid(this.error, null)
      }
  }

  private val log = LogManager.getLogger(this.javaClass)

  override fun importBlock(sealedBeaconBlock: SealedBeaconBlock): SafeFuture<ValidationResult> {
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
              beaconBlockImporter.importBlock(sealedBeaconBlock).thenApply { it }
            }

            is Err -> {
              log.error(
                "Validation failed for blockNumber=${sealedBeaconBlock.beaconBlock.beaconBlockHeader.number}, " +
                  "hash=${sealedBeaconBlock.beaconBlock.beaconBlockHeader.hash.encodeHex()}! " +
                  "error=${combinedValidationResult.error}",
              )
              SafeFuture.completedFuture(combinedValidationResult.toDomain())
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
