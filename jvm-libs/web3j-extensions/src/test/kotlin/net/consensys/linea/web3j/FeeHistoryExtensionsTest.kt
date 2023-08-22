package net.consensys.linea.web3j

import net.consensys.linea.FeeHistory
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.web3j.protocol.core.methods.response.EthFeeHistory
import java.math.BigInteger

class FeeHistoryExtensionsTest {
  @Test
  fun mapSingleEntry() {
    val feeHistory = EthFeeHistory.FeeHistory().apply {
      this.setOldestBlock("0xf")
      this.setBaseFeePerGas(listOf("0xa1", "0xa2"))
      this.setReward(listOf(listOf("0xba1", "0xba2")))
      this.gasUsedRatio = listOf(0.25)
    }

    assertThat(feeHistory.toLineaDomain()).isEqualTo(
      FeeHistory(
        oldestBlock = BigInteger("f", 16),
        baseFeePerGas = listOf(BigInteger("a1", 16), BigInteger("a2", 16)),
        reward = listOf(listOf(BigInteger("ba1", 16), BigInteger("ba2", 16))),
        gasUsedRatio = listOf(0.25.toBigDecimal())
      )
    )
  }

  @Test
  fun mapMultipleEntries() {
    val feeHistory = EthFeeHistory.FeeHistory().apply {
      this.setOldestBlock("0xf")
      this.setBaseFeePerGas(listOf("0xa1", "0xa2", "0xa3"))
      this.setReward(
        listOf(
          listOf("0xba1", "0xba2"),
          listOf("0xbb1", "0xbb2")
        )
      )
      this.gasUsedRatio = listOf(0.25, 0.75)
    }

    assertThat(feeHistory.toLineaDomain()).isEqualTo(
      FeeHistory(
        oldestBlock = BigInteger("f", 16),
        baseFeePerGas = listOf(BigInteger("a1", 16), BigInteger("a2", 16), BigInteger("a3", 16)),
        reward = listOf(
          listOf(BigInteger("ba1", 16), BigInteger("ba2", 16)),
          listOf(BigInteger("bb1", 16), BigInteger("bb2", 16))
        ),
        gasUsedRatio = listOf(0.25.toBigDecimal(), 0.75.toBigDecimal())
      )
    )
  }
}
