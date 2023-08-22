package net.consensys.linea.forkchoicestate

import com.fasterxml.jackson.databind.annotation.JsonDeserialize
import com.fasterxml.jackson.databind.annotation.JsonSerialize
import org.apache.tuweni.bytes.Bytes32
import org.web3j.protocol.Web3j
import org.web3j.protocol.core.DefaultBlockParameterName
import org.web3j.protocol.core.methods.response.EthBlock
import tech.pegasys.teku.ethereum.executionclient.serialization.Bytes32Deserializer
import tech.pegasys.teku.ethereum.executionclient.serialization.BytesSerializer
import java.math.BigInteger

// TODO: Latter Wrap this into a class with a message signature

data class ForkChoiceState(
  @JsonSerialize(using = BytesSerializer::class)
  @JsonDeserialize(using = Bytes32Deserializer::class)
  val headBlockHash: Bytes32,
  @JsonSerialize(using = BytesSerializer::class)
  @JsonDeserialize(using = Bytes32Deserializer::class)
  val safeBlockHash: Bytes32,
  @JsonSerialize(using = BytesSerializer::class)
  @JsonDeserialize(using = Bytes32Deserializer::class)
  val finalizedBlockHash: Bytes32
)

data class ForkChoiceStateInfoV0(
  val headBlockNumber: Long,
  val headBlockTimestamp: Long,
  val forkChoiceState: ForkChoiceState
) {
  fun copyWithUpdatedHeadAndSafe(
    headBlockNumber: Long,
    headBlockTimestamp: Long,
    headBlockHash: Bytes32
  ): ForkChoiceStateInfoV0 {
    return ForkChoiceStateInfoV0(
      headBlockNumber,
      headBlockTimestamp,
      forkChoiceState.copy(headBlockHash = headBlockHash, safeBlockHash = headBlockHash)
    )
  }
}

data class ForkChoiceStateBlocks(
  val headBlock: EthBlock.Block,
  val safeBlock: EthBlock.Block,
  val finalizedBlock: EthBlock.Block
) {
  fun toForkChoiceStateInfoV0(): ForkChoiceStateInfoV0 {
    return ForkChoiceStateInfoV0(
      headBlock.number.toLong(),
      headBlock.timestamp.toLong(),
      ForkChoiceState(
        Bytes32.fromHexString(headBlock.hash),
        Bytes32.fromHexString(safeBlock.hash),
        Bytes32.fromHexString(finalizedBlock.hash)
      )
    )
  }
}

fun forkChoiceStateBlocks(web3jClient: Web3j): ForkChoiceStateBlocks {
  val headBlock =
    web3jClient.ethGetBlockByNumber(DefaultBlockParameterName.LATEST, false).send().block

  val safeBlock =
    if (headBlock.number.equals(BigInteger.ZERO)) {
      headBlock
    } else {
      web3jClient.ethGetBlockByNumber(DefaultBlockParameterName.SAFE, false).send().block
    }

  val finalizedBlock =
    if (headBlock.number.equals(BigInteger.ZERO)) {
      headBlock
    } else {
      web3jClient.ethGetBlockByNumber(DefaultBlockParameterName.FINALIZED, false).send().block
    }

  return ForkChoiceStateBlocks(headBlock, safeBlock, finalizedBlock)
}
