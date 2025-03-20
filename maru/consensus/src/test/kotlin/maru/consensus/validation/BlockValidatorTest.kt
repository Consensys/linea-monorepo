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
import maru.core.BeaconBlock
import maru.core.BeaconBlockHeader
import maru.core.HashUtil
import maru.core.Seal
import maru.core.Validator
import maru.core.ext.DataGenerators
import maru.executionlayer.client.ExecutionLayerClient
import maru.serialization.rlp.bodyRoot
import org.assertj.core.api.Assertions.assertThat
import org.hyperledger.besu.consensus.common.bft.BftHelpers
import org.junit.jupiter.api.Test
import org.mockito.kotlin.any
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.mock
import org.mockito.kotlin.spy
import tech.pegasys.teku.ethereum.executionclient.schema.PayloadStatusV1
import tech.pegasys.teku.ethereum.executionclient.schema.Response
import tech.pegasys.teku.infrastructure.async.SafeFuture
import tech.pegasys.teku.spec.executionlayer.ExecutionPayloadStatus

class BlockValidatorTest {
  private val validators = (1..3).map { DataGenerators.randomValidator() }
  private val nonValidatorNode = DataGenerators.randomValidator()

  private val currBlockNumber = 9uL
  private val newBlockNumber = 10uL

  private val validCurrBlockBody = DataGenerators.randomBeaconBlockBody(numSeals = validators.size)
  private val validCurrBlockHeader =
    DataGenerators.randomBeaconBlockHeader(currBlockNumber).copy(
      proposer = validators[0],
      bodyRoot = HashUtil.bodyRoot(validCurrBlockBody),
    )

  private val validNewBlockBody = DataGenerators.randomBeaconBlockBody(numSeals = validators.size)
  private val validNewBlockHeader =
    DataGenerators.randomBeaconBlockHeader(newBlockNumber).copy(
      proposer = validators[1],
      parentRoot = validCurrBlockHeader.hash,
      timestamp = validCurrBlockHeader.timestamp + 1u,
      bodyRoot = HashUtil.bodyRoot(validNewBlockBody),
    )
  private val validNewBlock = BeaconBlock(validNewBlockHeader, validNewBlockBody)

  @Test
  fun `test valid block`() {
    val sealVerifier =
      object : SealVerifier {
        override fun extractValidator(
          seal: Seal,
          beaconBlockHeader: BeaconBlockHeader,
        ): Result<Validator, SealVerifier.SealValidationError> {
          assertThat(beaconBlockHeader).isEqualTo(validCurrBlockHeader)
          return when (seal) {
            validNewBlockBody.prevCommitSeals[0] -> return Ok(validators[0])
            validNewBlockBody.prevCommitSeals[1] -> return Ok(validators[1])
            validNewBlockBody.prevCommitSeals[2] -> return Ok(validators[2])
            else -> Err(SealVerifier.SealValidationError("Invalid seal"))
          }
        }
      }

    val executionLayerClient =
      mock<ExecutionLayerClient> {
        on { newPayload(any()) }.thenReturn(
          SafeFuture.completedFuture(
            Response.fromPayloadReceivedAsSsz(
              PayloadStatusV1(ExecutionPayloadStatus.VALID, null, null),
            ),
          ),
        )
      }

    val compositeBlockValidator =
      CompositeBlockValidator(
        blockValidators =
          listOf(
            BlockValidators.BlockNumberValidator,
            BlockValidators.TimestampValidator,
            BlockValidators.ProposerValidator,
            BlockValidators.ParentRootValidator,
            BlockValidators.BodyRootValidator,
//            PrevCommitSealValidator(sealVerifier = sealVerifier),
            ExecutionPayloadValidator(executionLayerClient),
          ),
      )
    val result =
      compositeBlockValidator
        .validateBlock(
          newBlock = validNewBlock,
          proposerForNewBlock = validNewBlock.beaconBlockHeader.proposer,
          parentBlockHeader = validCurrBlockHeader,
        ).get()
    assertThat(result is Ok).isTrue()
  }

  @Test
  fun `test invalid previous block number`() {
    val prevBlockHeader = validCurrBlockHeader.copy(number = 8u)
    val result =
      BlockValidators.BlockNumberValidator
        .validateBlock(
          newBlock = validNewBlock,
          proposerForNewBlock = validNewBlock.beaconBlockHeader.proposer,
          parentBlockHeader = prevBlockHeader,
        ).get()
    val expectedResult =
      error(
        "Block number is not the next block number " +
          "blockNumber=$newBlockNumber " +
          "nextBlockNumber=${prevBlockHeader.number + 1u}",
      )
    assertThat(result).isEqualTo(expectedResult)
  }

  @Test
  fun `test invalid block timestamp, equal to previous block timestamp`() {
    val blockHeader = validNewBlockHeader.copy(timestamp = validCurrBlockHeader.timestamp)
    val result =
      BlockValidators.TimestampValidator
        .validateBlock(
          newBlock = validNewBlock.copy(beaconBlockHeader = blockHeader),
          proposerForNewBlock = validNewBlock.beaconBlockHeader.proposer,
          parentBlockHeader = validCurrBlockHeader,
        ).get()
    val expectedResult =
      error(
        "Block timestamp is not greater than previous block timestamp " +
          "blockTimestamp=${blockHeader.timestamp} " +
          "parentBlockTimestamp=${validCurrBlockHeader.timestamp}",
      )
    assertThat(result).isEqualTo(expectedResult)
  }

  @Test
  fun `test invalid block timestamp, less than previous block timestamp`() {
    val blockHeader = validNewBlockHeader.copy(timestamp = validCurrBlockHeader.timestamp - 1u)
    val result =
      BlockValidators.TimestampValidator
        .validateBlock(
          newBlock = validNewBlock.copy(beaconBlockHeader = blockHeader),
          proposerForNewBlock = validNewBlock.beaconBlockHeader.proposer,
          parentBlockHeader = validCurrBlockHeader,
        ).get()
    val expectedResult =
      error(
        "Block timestamp is not greater than previous block timestamp " +
          "blockTimestamp=${blockHeader.timestamp} " +
          "parentBlockTimestamp=${validCurrBlockHeader.timestamp}",
      )

    assertThat(result).isEqualTo(expectedResult)
  }

  @Test
  fun `test invalid block proposer`() {
    val result =
      BlockValidators.ProposerValidator
        .validateBlock(
          newBlock = validNewBlock,
          proposerForNewBlock = validators.last(),
          parentBlockHeader = validCurrBlockHeader,
        ).get()
    val expectedResult =
      error(
        "Proposer is not expected proposer " +
          "proposer=${validNewBlockHeader.proposer} " +
          "expectedProposer=${validators.last()}",
      )
    assertThat(result).isEqualTo(expectedResult)
  }

  @Test
  fun `test invalid parent root`() {
    val blockHeader = validNewBlockHeader.copy(parentRoot = validCurrBlockHeader.hash.reversedArray())
    val result =
      BlockValidators.ParentRootValidator
        .validateBlock(
          newBlock = validNewBlock.copy(beaconBlockHeader = blockHeader),
          proposerForNewBlock = validNewBlockHeader.proposer,
          parentBlockHeader = validCurrBlockHeader,
        ).get()
    val expectedResult =
      error(
        "Parent root does not match parent block root " +
          "parentRoot=${blockHeader.parentRoot.encodeHex()} " +
          "expectedParentRoot=${validCurrBlockHeader.hash.encodeHex()}",
      )
    assertThat(result).isEqualTo(expectedResult)
  }

  @Test
  fun `test invalid body root`() {
    val blockHeader = validNewBlockHeader.copy(bodyRoot = validNewBlockHeader.bodyRoot.reversedArray())
    val result =
      BlockValidators.BodyRootValidator
        .validateBlock(
          newBlock = validNewBlock.copy(beaconBlockHeader = blockHeader),
          proposerForNewBlock = validCurrBlockHeader.proposer,
          parentBlockHeader = validCurrBlockHeader,
        ).get()
    val expectedResult =
      error(
        "Body root in header does not match body root " +
          "bodyRoot=${blockHeader.bodyRoot.encodeHex()} " +
          "expectedBodyRoot=${validNewBlockHeader.bodyRoot.encodeHex()}",
      )
    assertThat(result).isEqualTo(expectedResult)
  }

  @Test
  fun `test not enough commit seals`() {
    assertThat(BftHelpers.calculateRequiredValidatorQuorum(3)).isEqualTo(2)

    val blockBodyWithEnoughSeals =
      validNewBlockBody.copy(
        prevCommitSeals =
          listOf(
            validNewBlockBody.prevCommitSeals[0],
            validNewBlockBody.prevCommitSeals[1],
          ),
      )

    val blockBodyWithLessSeals =
      validNewBlockBody.copy(
        prevCommitSeals = listOf(validNewBlockBody.prevCommitSeals[0]),
      )

    val sealVerifier =
      object : SealVerifier {
        override fun extractValidator(
          seal: Seal,
          beaconBlockHeader: BeaconBlockHeader,
        ): Result<Validator, SealVerifier.SealValidationError> =
          when (seal) {
            validNewBlockBody.prevCommitSeals[0] -> Ok(validators[0])
            validNewBlockBody.prevCommitSeals[1] -> Ok(validators[1])
            else -> Err(SealVerifier.SealValidationError("Invalid seal"))
          }
      }

    val prevCommitSealValidator = spy(PrevCommitSealValidator(sealVerifier = sealVerifier))
    doReturn(
      SafeFuture.completedFuture(validators.toSet()),
    ).`when`(prevCommitSealValidator).getValidatorsForAncestorBlock(
      any(),
    )
    val validResult =
      prevCommitSealValidator
        .validateBlock(
          newBlock = validNewBlock.copy(beaconBlockBody = blockBodyWithEnoughSeals),
          proposerForNewBlock = validNewBlockHeader.proposer,
          parentBlockHeader = validCurrBlockHeader,
        ).get()
    assertThat(validResult is Ok).isTrue()

    val inValidResult =
      prevCommitSealValidator
        .validateBlock(
          newBlock = validNewBlock.copy(beaconBlockBody = blockBodyWithLessSeals),
          proposerForNewBlock = validNewBlockHeader.proposer,
          parentBlockHeader = validCurrBlockHeader,
        ).get()
    val expectedResult =
      error(
        "Quorum threshold not met. " +
          "committers=1 " +
          "validators=3 " +
          "quorumCount=2",
      )
    assertThat(inValidResult).isEqualTo(expectedResult)
  }

  @Test
  fun `test commit seals not from validator`() {
    val sealVerifier =
      object : SealVerifier {
        override fun extractValidator(
          seal: Seal,
          beaconBlockHeader: BeaconBlockHeader,
        ): Result<Validator, SealVerifier.SealValidationError> = Ok(nonValidatorNode)
      }

    val prevCommitSealValidator = spy(PrevCommitSealValidator(sealVerifier = sealVerifier))
    doReturn(
      SafeFuture.completedFuture(validators.toSet()),
    ).`when`(prevCommitSealValidator).getValidatorsForAncestorBlock(
      any(),
    )

    val result =
      prevCommitSealValidator
        .validateBlock(
          newBlock = validNewBlock,
          proposerForNewBlock = validNewBlockHeader.proposer,
          parentBlockHeader = validCurrBlockHeader,
        ).get()
    val expectedResult =
      error(
        "Seal validator is not in the parent block's validator set " +
          "seal=${validNewBlockBody.prevCommitSeals[0]} " +
          "sealValidator=$nonValidatorNode " +
          "validatorsForParentBlock=$validators",
      )
    assertThat(result).isEqualTo(expectedResult)
  }

  @Test
  fun `test invalid commit seals`() {
    val sealVerifier =
      object : SealVerifier {
        override fun extractValidator(
          seal: Seal,
          beaconBlockHeader: BeaconBlockHeader,
        ): Result<Validator, SealVerifier.SealValidationError> = Err(SealVerifier.SealValidationError("Invalid seal"))
      }
    val prevCommitSealValidator = spy(PrevCommitSealValidator(sealVerifier = sealVerifier))
    doReturn(
      SafeFuture.completedFuture(validators.toSet()),
    ).`when`(prevCommitSealValidator).getValidatorsForAncestorBlock(
      any(),
    )

    val result =
      prevCommitSealValidator
        .validateBlock(
          newBlock = validNewBlock,
          proposerForNewBlock = validNewBlockHeader.proposer,
          parentBlockHeader = validCurrBlockHeader,
        ).get()
    val expectedResult = error("Previous block seal verification failed. Reason: Invalid seal")
    assertThat(result).isEqualTo(expectedResult)
  }

  @Test
  fun `test invalid execution payload`() {
    val blockBody =
      validNewBlockBody.copy(
        executionPayload =
          validNewBlockBody.executionPayload.copy(
            timestamp = validNewBlockBody.executionPayload.timestamp + 1u,
          ),
      )
    val invalidExecutionClient =
      mock<ExecutionLayerClient> {
        on { newPayload(any()) }.thenReturn(
          SafeFuture.completedFuture(Response.fromErrorMessage<PayloadStatusV1>("Invalid execution payload")),
        )
      }
    val result =
      ExecutionPayloadValidator(executionLayerClient = invalidExecutionClient)
        .validateBlock(
          newBlock = validNewBlock.copy(beaconBlockBody = blockBody),
          proposerForNewBlock = validNewBlockHeader.proposer,
          parentBlockHeader = validCurrBlockHeader,
        ).get()
    val expectedResult =
      error(
        "Execution payload validation failed: Invalid execution payload",
      )
    assertThat(result).isEqualTo(expectedResult)
  }
}
