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
import maru.consensus.ProposerSelector
import maru.consensus.ValidatorProvider
import maru.consensus.state.StateTransition
import maru.consensus.toConsensusRoundIdentifier
import maru.consensus.validation.BlockValidator.BlockValidationError
import maru.consensus.validation.BlockValidator.Companion.error
import maru.core.BeaconBlock
import maru.core.BeaconBlockHeader
import maru.core.HashUtil
import maru.core.Validator
import maru.database.BeaconChain
import maru.executionlayer.client.ExecutionLayerClient
import maru.executionlayer.extensions.hasValidExecutionPayload
import maru.serialization.rlp.bodyRoot
import maru.serialization.rlp.stateRoot
import org.hyperledger.besu.consensus.common.bft.BftHelpers
import tech.pegasys.teku.infrastructure.async.SafeFuture

interface BlockValidator {
  data class BlockValidationError(
    val message: String,
  )

  companion object {
    fun ok(): Result<Unit, BlockValidationError> = Ok(Unit)

    fun error(message: String): Result<Unit, BlockValidationError> = Err(BlockValidationError(message))

    fun require(
      condition: Boolean,
      errorMessageProvider: () -> String,
    ): Result<Unit, BlockValidationError> =
      if (condition) {
        ok()
      } else {
        error(errorMessageProvider())
      }
  }

  fun validateBlock(block: BeaconBlock): SafeFuture<Result<Unit, BlockValidationError>>
}

class CompositeBlockValidator(
  private val blockValidators: List<BlockValidator>,
) : BlockValidator {
  override fun validateBlock(block: BeaconBlock): SafeFuture<Result<Unit, BlockValidationError>> {
    val validationResultFutures =
      blockValidators
        .map { it.validateBlock(block) }
        .stream()
    return SafeFuture.collectAll(validationResultFutures).thenApply { validationResults ->
      val errors = validationResults.mapNotNull { it.component2() }
      BlockValidator.require(errors.isEmpty()) {
        errors.joinToString { it.message }
      }
    }
  }
}

class BlockNumberValidator(
  private val parentBlockHeader: BeaconBlockHeader,
) : BlockValidator {
  override fun validateBlock(block: BeaconBlock): SafeFuture<Result<Unit, BlockValidationError>> {
    val parentBlockNumber = parentBlockHeader.number
    return SafeFuture.completedFuture(
      BlockValidator.require(block.beaconBlockHeader.number == parentBlockNumber + 1u) {
        "Block number is not the next block number blockNumber=${block.beaconBlockHeader.number} " +
          "parentBlockNumber=$parentBlockNumber"
      },
    )
  }
}

class TimestampValidator(
  private val parentBlockHeader: BeaconBlockHeader,
) : BlockValidator {
  override fun validateBlock(block: BeaconBlock): SafeFuture<Result<Unit, BlockValidationError>> {
    val parentBlockTimeStamp = parentBlockHeader.timestamp
    return SafeFuture.completedFuture(
      BlockValidator.require(
        block.beaconBlockHeader.timestamp > parentBlockTimeStamp,
      ) {
        "Block timestamp is not greater than previous block timestamp " +
          "blockTimestamp=${block.beaconBlockHeader.timestamp} " +
          "parentBlockTimestamp=$parentBlockTimeStamp"
      },
    )
  }
}

class ProposerValidator(
  private val proposerSelector: ProposerSelector,
) : BlockValidator {
  override fun validateBlock(block: BeaconBlock): SafeFuture<Result<Unit, BlockValidationError>> =
    proposerSelector
      .getProposerForBlock(block.beaconBlockHeader.toConsensusRoundIdentifier())
      .thenApply { proposerForNewBlock ->
        BlockValidator.require(block.beaconBlockHeader.proposer == proposerForNewBlock) {
          "Proposer is not expected proposer proposer=${block.beaconBlockHeader.proposer} " +
            "expectedProposer=$proposerForNewBlock"
        }
      }
}

class ParentRootValidator(
  private val parentBlockHeader: BeaconBlockHeader,
) : BlockValidator {
  override fun validateBlock(block: BeaconBlock): SafeFuture<Result<Unit, BlockValidationError>> =
    SafeFuture.completedFuture(
      BlockValidator.require(
        block.beaconBlockHeader.parentRoot.contentEquals(
          parentBlockHeader
            .hash,
        ),
      ) {
        "Parent root does not match parent block root parentRoot=${block.beaconBlockHeader.parentRoot.encodeHex()} " +
          "expectedParentRoot=${parentBlockHeader.hash.encodeHex()}"
      },
    )
}

class StateRootValidator(
  private val stateTransition: StateTransition,
) : BlockValidator {
  override fun validateBlock(block: BeaconBlock): SafeFuture<Result<Unit, BlockValidationError>> =
    stateTransition
      .processBlock(block)
      .thenApply { postState ->
        val stateRootHeader =
          postState.latestBeaconBlockHeader.copy(
            stateRoot = BeaconBlockHeader.EMPTY_HASH,
          )
        val expectedStateRoot = HashUtil.stateRoot(postState.copy(latestBeaconBlockHeader = stateRootHeader))
        BlockValidator.require(block.beaconBlockHeader.stateRoot.contentEquals(expectedStateRoot)) {
          "State root in header does not match state root stateRoot=${block.beaconBlockHeader.stateRoot.encodeHex()} " +
            "expectedStateRoot=${expectedStateRoot.encodeHex()}"
        }
      }.exceptionally {
        error("State root validation failed: ${it.message}")
      }
}

class BodyRootValidator : BlockValidator {
  override fun validateBlock(block: BeaconBlock): SafeFuture<Result<Unit, BlockValidationError>> {
    val beaconBodyRoot = HashUtil.bodyRoot(block.beaconBlockBody)
    return SafeFuture.completedFuture(
      BlockValidator.require(
        block.beaconBlockHeader.bodyRoot.contentEquals
          (beaconBodyRoot),
      ) {
        "Body root in header does not match body root bodyRoot=${block.beaconBlockHeader.bodyRoot.encodeHex()} " +
          "expectedBodyRoot=${beaconBodyRoot.encodeHex()}"
      },
    )
  }
}

class PrevCommitSealValidator(
  private val sealVerifier: SealVerifier,
  private val beaconChain: BeaconChain,
  private val validatorProvider: ValidatorProvider,
  private val config: Config,
) : BlockValidator {
  data class Config(
    val prevBlockOffset: UInt,
  )

  override fun validateBlock(block: BeaconBlock): SafeFuture<Result<Unit, BlockValidationError>> {
    val prevBlockNumber = block.beaconBlockHeader.number - config.prevBlockOffset

    val prevBlock =
      beaconChain.getSealedBeaconBlock(prevBlockNumber)?.beaconBlock ?: return SafeFuture.completedFuture(
        error("Previous block not found, previousBlockNumber=$prevBlockNumber"),
      )

    return validatorProvider
      .getValidatorsForBlock(prevBlock.beaconBlockHeader.number)
      .thenApply { validatorsForPrevBlock -> verifySeals(block, prevBlock, validatorsForPrevBlock) }
  }

  private fun verifySeals(
    newBlock: BeaconBlock,
    prevBlock: BeaconBlock,
    validatorsForPrevBlock: Set<Validator>,
  ): Result<Unit, BlockValidationError> {
    val committers = mutableSetOf<Validator>()
    for (seal in newBlock.beaconBlockBody.prevCommitSeals) {
      when (val sealVerificationResult = sealVerifier.extractValidator(seal, prevBlock.beaconBlockHeader)) {
        is Ok -> {
          val sealValidator = sealVerificationResult.value
          if (sealValidator !in validatorsForPrevBlock) {
            return error(
              "Seal validator is not in the parent block's validator set seal=$seal sealValidator=$sealValidator " +
                "validatorsForParentBlock=$validatorsForPrevBlock",
            )
          }
          committers.add(sealVerificationResult.value)
        }

        is Err ->
          return error("Previous block seal verification failed. Reason: ${sealVerificationResult.error.message}")
      }
    }
    val quorumCount = BftHelpers.calculateRequiredValidatorQuorum(validatorsForPrevBlock.size)
    return BlockValidator.require(committers.size >= quorumCount) {
      "Quorum threshold not met. committers=${committers.size} validators=${validatorsForPrevBlock.size} " +
        "quorumCount=$quorumCount"
    }
  }
}

class ExecutionPayloadValidator(
  private val executionLayerClient: ExecutionLayerClient,
) : BlockValidator {
  override fun validateBlock(block: BeaconBlock): SafeFuture<Result<Unit, BlockValidationError>> =
    executionLayerClient.newPayload(block.beaconBlockBody.executionPayload).thenApply { newPayloadResponse ->
      BlockValidator.require(
        newPayloadResponse.isSuccess && newPayloadResponse.payload.hasValidExecutionPayload(),
      ) {
        "Execution payload validation failed: ${newPayloadResponse.errorMessage}"
      }
    }
}

object EmptyBlockValidator : BlockValidator {
  override fun validateBlock(block: BeaconBlock): SafeFuture<Result<Unit, BlockValidationError>> =
    SafeFuture.completedFuture(
      BlockValidator.require(
        block.beaconBlockBody.executionPayload.transactions
          .isNotEmpty(),
      ) {
        "Block number=${block.beaconBlockHeader.number} hash=${block.beaconBlockHeader.hash.encodeHex()} is empty!"
      },
    )
}
