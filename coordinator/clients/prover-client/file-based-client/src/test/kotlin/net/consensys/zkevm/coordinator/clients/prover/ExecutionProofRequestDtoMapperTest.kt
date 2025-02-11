package net.consensys.zkevm.coordinator.clients.prover

import build.linea.clients.GetZkEVMStateMerkleProofResponse
import com.fasterxml.jackson.databind.node.ArrayNode
import linea.domain.Block
import linea.domain.createBlock
import net.consensys.ByteArrayExt
import net.consensys.encodeHex
import net.consensys.zkevm.coordinator.clients.BatchExecutionProofRequestV1
import net.consensys.zkevm.coordinator.clients.GenerateTracesResponse
import net.consensys.zkevm.domain.CommonTestData
import net.consensys.zkevm.domain.RlpBridgeLogsData
import net.consensys.zkevm.encoding.BlockEncoder
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.kotlin.spy
import kotlin.random.Random

class ExecutionProofRequestDtoMapperTest {

  private lateinit var encoder: BlockEncoder
  private lateinit var requestDtoMapper: ExecutionProofRequestDtoMapper
  private val fakeEncoder: BlockEncoder = object : BlockEncoder {
    override fun encode(block: Block): ByteArray {
      return block.number.toString().toByteArray()
    }
  }

  @BeforeEach
  fun beforeEach() {
    encoder = spy(fakeEncoder)
    requestDtoMapper = ExecutionProofRequestDtoMapper(encoder)
  }

  @Test
  fun `should decorate data with bridge logs and parent stateRootHash`() {
    val block1 = createBlock(number = 747066UL)
    val block2 = createBlock(number = 747067UL)
    val type2StateResponse = GetZkEVMStateMerkleProofResponse(
      zkStateMerkleProof = ArrayNode(null),
      zkParentStateRootHash = ByteArrayExt.random32(),
      zkEndStateRootHash = ByteArrayExt.random32(),
      zkStateManagerVersion = "2.0.0"
    )
    val generateTracesResponse = GenerateTracesResponse(
      tracesFileName = "747066-747067-conflated-traces.json",
      tracesEngineVersion = "1.0.0"
    )
    val stateRoot = Random.nextBytes(32)
    val request = BatchExecutionProofRequestV1(
      blocks = listOf(block1, block2),
      bridgeLogs = CommonTestData.bridgeLogs,
      tracesResponse = generateTracesResponse,
      type2StateData = type2StateResponse,
      keccakParentStateRootHash = stateRoot
    )

    val requestDto = requestDtoMapper.invoke(request).get()

    assertThat(requestDto.keccakParentStateRootHash).isEqualTo(stateRoot.encodeHex())
    assertThat(requestDto.zkParentStateRootHash).isEqualTo(type2StateResponse.zkParentStateRootHash.encodeHex())
    assertThat(requestDto.conflatedExecutionTracesFile).isEqualTo("747066-747067-conflated-traces.json")
    assertThat(requestDto.tracesEngineVersion).isEqualTo("1.0.0")
    assertThat(requestDto.type2StateManagerVersion).isEqualTo("2.0.0")
    assertThat(requestDto.zkStateMerkleProof).isEqualTo(type2StateResponse.zkStateMerkleProof)
    assertThat(requestDto.blocksData).hasSize(2)
    assertThat(requestDto.blocksData[0]).isEqualTo(
      RlpBridgeLogsData(
        rlp = "747066".toByteArray().encodeHex(),
        bridgeLogs = CommonTestData.bridgeLogs
      )
    )
    assertThat(requestDto.blocksData[1]).isEqualTo(
      RlpBridgeLogsData(
        rlp = "747067".toByteArray().encodeHex(),
        bridgeLogs = emptyList()
      )
    )
  }
}
