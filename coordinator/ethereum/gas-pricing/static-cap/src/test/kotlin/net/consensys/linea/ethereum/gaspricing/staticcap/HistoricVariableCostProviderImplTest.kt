package net.consensys.linea.ethereum.gaspricing.staticcap

import linea.OneKWei
import linea.domain.Block
import linea.domain.BlockParameter.Companion.toBlockParameter
import linea.kotlin.decodeHex
import linea.web3j.ExtendedWeb3J
import net.consensys.linea.ethereum.gaspricing.MinerExtraDataV1
import org.assertj.core.api.Assertions.assertThat
import org.assertj.core.api.Assertions.assertThatThrownBy
import org.junit.jupiter.api.Test
import org.mockito.kotlin.doReturn
import org.mockito.kotlin.eq
import org.mockito.kotlin.mock
import org.mockito.kotlin.times
import org.mockito.kotlin.verify
import tech.pegasys.teku.infrastructure.async.SafeFuture

class HistoricVariableCostProviderImplTest {
  val mockEthBlock = mock<Block> {
    on { extraData } doReturn MinerExtraDataV1(
      fixedCostInKWei = 1000U,
      variableCostInKWei = 10000U,
      ethGasPriceInKWei = 12000U,
    ).encode().decodeHex()
  }
  val targetBlockNumber = 100.toBigInteger()
  val mockWeb3jClient = mock<ExtendedWeb3J> {
    on { ethBlockNumber() } doReturn SafeFuture.completedFuture(targetBlockNumber)
    on { ethGetBlock(eq(targetBlockNumber.toBlockParameter())) } doReturn
      SafeFuture.completedFuture(mockEthBlock)
  }

  @Test
  fun test_getLatestVariableCost() {
    val historicVariableCostProvider = HistoricVariableCostProviderImpl(
      web3jClient = mockWeb3jClient,
    )

    val latestVariableCost = historicVariableCostProvider.getLatestVariableCost().get()

    val expectedVariableCost = 10000.0 * OneKWei
    assertThat(latestVariableCost).isEqualTo(expectedVariableCost)
  }

  @Test
  fun test_getLatestVariableCost_return_cached_variable_cost() {
    val historicVariableCostProvider = HistoricVariableCostProviderImpl(
      web3jClient = mockWeb3jClient,
    )

    // initialize the cache
    val expectedVariableCost = 10000.0 * OneKWei
    assertThat(historicVariableCostProvider.getLatestVariableCost().get()).isEqualTo(expectedVariableCost)

    // subsequent calls with the same block #100 should return the same value by the cache
    (0..4).forEach { _ ->
      assertThat(
        historicVariableCostProvider.getLatestVariableCost().get(),
      ).isEqualTo(expectedVariableCost)
    }

    // verified it only called ethGetBlock once
    verify(mockWeb3jClient, times(1))
      .ethGetBlock(eq(targetBlockNumber.toBlockParameter()))
  }

  @Test
  fun test_getLatestVariableCost_throws_error_if_ethGetBlock_return_null() {
    val mockWeb3jClient = mock<ExtendedWeb3J> {
      on { ethBlockNumber() } doReturn SafeFuture.completedFuture(targetBlockNumber)
      on { ethGetBlock(eq(targetBlockNumber.toBlockParameter())) } doReturn
        SafeFuture.completedFuture(null)
    }
    val historicVariableCostProvider = HistoricVariableCostProviderImpl(
      web3jClient = mockWeb3jClient,
    )

    assertThatThrownBy {
      historicVariableCostProvider.getLatestVariableCost().get()
    }.hasCauseInstanceOf(NullPointerException::class.java)
  }

  @Test
  fun test_getLatestVariableCost_throws_error_if_ethGetBlock_throws_error() {
    val expectedException = RuntimeException("Error from ethGetBlock")
    val mockWeb3jClient = mock<ExtendedWeb3J> {
      on { ethBlockNumber() } doReturn SafeFuture.completedFuture(targetBlockNumber)
      on { ethGetBlock(eq(targetBlockNumber.toBlockParameter())) } doReturn
        SafeFuture.failedFuture(expectedException)
    }
    val historicVariableCostProvider = HistoricVariableCostProviderImpl(
      web3jClient = mockWeb3jClient,
    )

    assertThatThrownBy {
      historicVariableCostProvider.getLatestVariableCost().get()
    }.hasCause(expectedException)
  }

  @Test
  fun test_getLatestVariableCost_returns_zero_if_MinerExtraData_decode_throws_error() {
    val mockEthBlock = mock<Block> {
      // extra data hex string with unsupported version 0xFF
      on { extraData } doReturn "0xff000003e80000271000002ee000000000000000000000000000000000000000".decodeHex()
    }
    val mockWeb3jClient = mock<ExtendedWeb3J> {
      on { ethBlockNumber() } doReturn SafeFuture.completedFuture(targetBlockNumber)
      on { ethGetBlock(eq(targetBlockNumber.toBlockParameter())) } doReturn
        SafeFuture.completedFuture(mockEthBlock)
    }
    val historicVariableCostProvider = HistoricVariableCostProviderImpl(
      web3jClient = mockWeb3jClient,
    )

    assertThat(
      historicVariableCostProvider.getLatestVariableCost().get(),
    ).isEqualTo(0.0)
  }
}
