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
package maru.consensus.validation

import com.github.michaelbull.result.Err
import com.github.michaelbull.result.Ok
import com.github.michaelbull.result.Result
import encodeHex
import maru.consensus.validation.BlockValidator.Companion.error
import maru.consensus.validation.BlockValidator.Companion.ok
import maru.core.BeaconBlock
import maru.core.BeaconBlockHeader
import maru.core.HashUtil
import maru.core.Validator
import maru.executionlayer.client.ExecutionLayerClient
import maru.executionlayer.extensions.hasValidExecutionPayload
import maru.serialization.rlp.bodyRoot
import org.hyperledger.besu.consensus.common.bft.BftHelpers
import tech.pegasys.teku.infrastructure.async.SafeFuture

fun interface BlockValidator {
  data class BlockValidationError(
    val message: String,
  )

  companion object {
    fun ok(): Result<Unit, BlockValidationError> = Ok(Unit)

    fun error(message: String): Result<Unit, BlockValidationError> = Err(BlockValidationError(message))
  }

  fun validateBlock(
    newBlock: BeaconBlock,
    proposerForNewBlock: Validator,
    parentBlockHeader: BeaconBlockHeader,
  ): SafeFuture<Result<Unit, BlockValidationError>>
}

class CompositeBlockValidator(
  private val blockValidators: List<BlockValidator>,
) : BlockValidator {
  override fun validateBlock(
    newBlock: BeaconBlock,
    proposerForNewBlock: Validator,
    parentBlockHeader: BeaconBlockHeader,
  ): SafeFuture<Result<Unit, BlockValidator.BlockValidationError>> {
    val validationResultFutures =
      blockValidators
        .map { it.validateBlock(newBlock, proposerForNewBlock, parentBlockHeader) }
        .stream()
    return SafeFuture.collectAll(validationResultFutures).thenApply { validationResults ->
      val errors = validationResults.mapNotNull { it.component2() }
      if (errors.isEmpty()) {
        ok()
      } else {
        error(errors.joinToString { it.message })
      }
    }
  }
}

object BlockValidators {
  val BlockNumberValidator =
    BlockValidator { block, _, parentBlockHeader ->
      if (block.beaconBlockHeader.number != parentBlockHeader.number + 1u) {
        SafeFuture.completedFuture(
          error(
            "Block number is not the next block number " +
              "blockNumber=${block.beaconBlockHeader.number} " +
              "nextBlockNumber=${parentBlockHeader.number + 1u}",
          ),
        )
      } else {
        SafeFuture.completedFuture(ok())
      }
    }

  val TimestampValidator =
    BlockValidator { block, _, parentBlockHeader ->
      if (block.beaconBlockHeader.timestamp <= parentBlockHeader.timestamp) {
        SafeFuture.completedFuture(
          error(
            "Block timestamp is not greater than previous block timestamp " +
              "blockTimestamp=${block.beaconBlockHeader.timestamp} " +
              "parentBlockTimestamp=${parentBlockHeader.timestamp}",
          ),
        )
      } else {
        SafeFuture.completedFuture(ok())
      }
    }

  val ProposerValidator =
    BlockValidator { block, proposerForBlock, _ ->
      if (block.beaconBlockHeader.proposer != proposerForBlock) {
        SafeFuture.completedFuture(
          Err(
            BlockValidator.BlockValidationError(
              "Proposer is not expected proposer " +
                "proposer=${block.beaconBlockHeader.proposer} " +
                "expectedProposer=$proposerForBlock",
            ),
          ),
        )
      } else {
        SafeFuture.completedFuture(ok())
      }
    }

  val ParentRootValidator =
    BlockValidator { block, _, parentBlockHeader ->
      if (!block.beaconBlockHeader.parentRoot.contentEquals(parentBlockHeader.hash)) {
        SafeFuture.completedFuture(
          error(
            "Parent root does not match parent block root " +
              "parentRoot=${block.beaconBlockHeader.parentRoot.encodeHex()} " +
              "expectedParentRoot=${parentBlockHeader.hash.encodeHex()}",
          ),
        )
      } else {
        SafeFuture.completedFuture(ok())
      }
    }

  val BodyRootValidator =
    BlockValidator { block, _, _ ->
      val beaconBodyRoot = HashUtil.bodyRoot(block.beaconBlockBody)
      if (!block.beaconBlockHeader.bodyRoot.contentEquals(beaconBodyRoot)) {
        SafeFuture.completedFuture(
          error(
            "Body root in header does not match body root " +
              "bodyRoot=${block.beaconBlockHeader.bodyRoot.encodeHex()} " +
              "expectedBodyRoot=${beaconBodyRoot.encodeHex()}",
          ),
        )
      } else {
        SafeFuture.completedFuture(ok())
      }
    }
}

class PrevCommitSealValidator(
  private val sealVerifier: SealVerifier,
) : BlockValidator {
  // Public for tests
  fun getValidatorsForAncestorBlock(newBlock: BeaconBlock): SafeFuture<Set<Validator>> {
    TODO("Figure out N, and retrive block newBlock.blockHeader.number - N and its validators")
  }

  override fun validateBlock(
    newBlock: BeaconBlock,
    proposerForNewBlock: Validator,
    parentBlockHeader: BeaconBlockHeader,
  ): SafeFuture<Result<Unit, BlockValidator.BlockValidationError>> {
    val validatorsForPrevBlockFuture = getValidatorsForAncestorBlock(newBlock)
    return validatorsForPrevBlockFuture.thenApply { validatorsForPrevBlock ->
      val committers = mutableSetOf<Validator>()
      for (seal in newBlock.beaconBlockBody.prevCommitSeals) {
        when (val sealVerificationResult = sealVerifier.extractValidator(seal, parentBlockHeader)) {
          is Ok -> {
            val sealValidator = sealVerificationResult.value
            if (sealValidator !in validatorsForPrevBlock) {
              return@thenApply (
                error(
                  "Seal validator is not in the parent block's validator set " +
                    "seal=$seal " +
                    "sealValidator=$sealValidator " +
                    "validatorsForParentBlock=$validatorsForPrevBlock",
                )
              )
            }
            committers.add(sealVerificationResult.value)
          }

          is Err ->
            return@thenApply (
              error("Previous block seal verification failed. Reason: ${sealVerificationResult.error.message}")
            )
        }
      }
      val quorumCount = BftHelpers.calculateRequiredValidatorQuorum(validatorsForPrevBlock.size)
      if (committers.size < quorumCount) {
        return@thenApply(
          error(
            "Quorum threshold not met. " +
              "committers=${committers.size} " +
              "validators=${validatorsForPrevBlock.size} " +
              "quorumCount=$quorumCount",
          )
        )
      }
      ok()
    }
  }
}

class ExecutionPayloadValidator(
  private val executionLayerClient: ExecutionLayerClient,
) : BlockValidator {
  override fun validateBlock(
    newBlock: BeaconBlock,
    proposerForNewBlock: Validator,
    parentBlockHeader: BeaconBlockHeader,
  ): SafeFuture<Result<Unit, BlockValidator.BlockValidationError>> =
    executionLayerClient.newPayload(newBlock.beaconBlockBody.executionPayload).thenApply { newPayloadResponse ->
      if (newPayloadResponse.isSuccess && newPayloadResponse.payload.hasValidExecutionPayload()) {
        ok()
      } else {
        error(
          "Execution payload validation failed: ${newPayloadResponse.errorMessage}",
        )
      }
    }
}
