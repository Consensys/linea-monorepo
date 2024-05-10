package net.consensys.linea.contract

import net.consensys.toBigInteger
import net.consensys.zkevm.coordinator.clients.DataSubmittedEvent
import org.apache.tuweni.bytes.Bytes32
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.BeforeEach
import org.junit.jupiter.api.Test
import org.mockito.Mockito
import org.mockito.kotlin.any
import org.mockito.kotlin.eq
import org.mockito.kotlin.mock
import org.mockito.kotlin.whenever
import org.web3j.abi.EventEncoder
import org.web3j.abi.FunctionEncoder
import org.web3j.abi.TypeEncoder
import org.web3j.abi.datatypes.generated.Uint256
import org.web3j.protocol.core.methods.response.EthLog.LogObject
import tech.pegasys.teku.infrastructure.async.SafeFuture
import java.util.concurrent.TimeUnit

class RollupSmartContractClientWeb3JImplTest {
  private lateinit var web3jLogsClient: Web3JLogsClient
  private lateinit var lineaRollupAsyncFriendly: LineaRollupAsyncFriendly
  private lateinit var smartContractClient: RollupSmartContractClientWeb3JImpl

  @BeforeEach
  fun beforeEach() {
    web3jLogsClient = mock()
    lineaRollupAsyncFriendly = mock(defaultAnswer = Mockito.RETURNS_DEEP_STUBS)
    smartContractClient = RollupSmartContractClientWeb3JImpl(
      web3jLogsClient = web3jLogsClient,
      lineaRollup = lineaRollupAsyncFriendly
    )
  }

  @Test
  fun `findLatestDataSubmittedEmittedEventSince when there are no events should return null`() {
    whenever(web3jLogsClient.getLogs(any()))
      .thenReturn(SafeFuture.completedFuture(emptyList()))

    assertThat(
      smartContractClient
        .findLatestDataSubmittedEmittedEvent(0, 100).get()
    )
      .isNull()
  }

  @Test
  fun `findLatestDataSubmittedEmittedEventSince when there is an event shoult return the last`() {
    val dataHash = Bytes32.random()
    val expectedLastEvent = DataSubmittedEvent(
      dataHash = dataHash.toArray(),
      startBlock = 301u,
      endBlock = 400u
    )
    val ethLogsEvents: List<LogObject> = listOf(
      createDataSubmitEvent(blockNumber = 1u, blobStartBlockNumber = 100u, blobEndBlockNumber = 200u),
      createDataSubmitEvent(blockNumber = 2u, blobStartBlockNumber = 201u, blobEndBlockNumber = 300u),
      createDataSubmitEvent(
        blockNumber = 3u,
        blobStartBlockNumber = 301u,
        blobEndBlockNumber = 400u,
        dataHashValue = dataHash
      )
    )

    whenever(web3jLogsClient.getLogs(any()))
      .thenReturn(SafeFuture.completedFuture(ethLogsEvents))

    assertThat(
      smartContractClient
        .findLatestDataSubmittedEmittedEvent(0, 100).get()
    )
      .isEqualTo(expectedLastEvent)
  }

  private fun createDataSubmitEvent(
    blockNumber: ULong,
    blobStartBlockNumber: ULong,
    blobEndBlockNumber: ULong,
    dataHashValue: Bytes32 = Bytes32.random(),
    testFee: Uint256 = Uint256(12345),
    testValue: Uint256 = Uint256(123456789),
    testNonce: Uint256 = Uint256(1),
    callData: org.web3j.abi.datatypes.DynamicBytes = org.web3j.abi.datatypes.DynamicBytes("".encodeToByteArray())
  ): LogObject {
    val log = LogObject()
    val eventSignature: String = EventEncoder.encode(LineaRollup.DATASUBMITTED_EVENT)
    val dataHash = org.web3j.abi.datatypes.generated.Bytes32(dataHashValue.toArray())

    log.topics =
      listOf(
        eventSignature,
        TypeEncoder.encode(dataHash),
        TypeEncoder.encode(org.web3j.abi.datatypes.Int(blobStartBlockNumber.toBigInteger())),
        TypeEncoder.encode(org.web3j.abi.datatypes.Int(blobEndBlockNumber.toBigInteger()))
      )

    log.data =
      FunctionEncoder.encodeConstructor(
        listOf(
          testFee,
          testValue,
          testNonce,
          callData
        )
      )

    log.setBlockNumber(blockNumber.toString())
    log.setLogIndex("0")

    return log
  }

  @Test
  fun `getMessageRollingHash_return hash from given messageNumber`() {
    val messageNumber = 101L
    val expectedRollingHash = Bytes32.random().toArray()
    whenever(lineaRollupAsyncFriendly.rollingHashes(eq(messageNumber.toBigInteger())).sendAsync())
      .thenAnswer { SafeFuture.completedFuture(expectedRollingHash) }

    assertThat(
      smartContractClient.getMessageRollingHash(messageNumber)
        .get(1, TimeUnit.SECONDS)
    )
      .isEqualTo(expectedRollingHash)
  }
}
