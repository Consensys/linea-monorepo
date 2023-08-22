package net.consensys.zkevm.coordinator.clients

import com.fasterxml.jackson.annotation.JsonProperty
import com.fasterxml.jackson.databind.annotation.JsonDeserialize
import com.github.michaelbull.result.Result
import net.consensys.linea.errors.ErrorResponse
import org.apache.tuweni.bytes.Bytes
import org.apache.tuweni.bytes.Bytes32
import tech.pegasys.teku.ethereum.executionclient.schema.ExecutionPayloadV1
import tech.pegasys.teku.ethereum.executionclient.serialization.Bytes32Deserializer
import tech.pegasys.teku.ethereum.executionclient.serialization.BytesDeserializer
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.time.Instant

enum class ProverErrorType {
  // to complete as we go
  ResponseNotFound,
  ParseError,
  ResponseTimeout,
  Unknown
}

data class GetProofResponse(
  @JsonDeserialize(using = BytesDeserializer::class) val proof: Bytes,
  val verifierIndex: Long, // TODO: make it work with ULong
  @JsonProperty("parentStateRootHash")
  @JsonDeserialize(using = Bytes32Deserializer::class)
  val zkParentStateRootHash: Bytes32,
  val blocksData: List<BlockData>,
  val proverVersion: String
) {
  data class BlockData(
    /** Type 2 State Manager RootHash */
    @JsonProperty("rootHash")
    @JsonDeserialize(using = Bytes32Deserializer::class)
    val zkRootHash: Bytes32,
    /** Unix timestamp in seconds */
    val timestamp: Instant,
    val rlpEncodedTransactions: List<String>,
    val batchReceptionIndices: List<UShort>,
    /** L2->L1 Message service Smart Contract Logs abi encoded */
    val l2ToL1MsgHashes: List<String>,
    @JsonDeserialize(using = BytesDeserializer::class)
    val fromAddresses: Bytes
  )
}

interface ProverClient {
  fun getZkProof(
    blocks: List<ExecutionPayloadV1>,
    tracesResponse: GenerateTracesResponse,
    type2StateData: GetZkEVMStateMerkleProofResponse
  ): SafeFuture<Result<GetProofResponse, ErrorResponse<ProverErrorType>>>
}
