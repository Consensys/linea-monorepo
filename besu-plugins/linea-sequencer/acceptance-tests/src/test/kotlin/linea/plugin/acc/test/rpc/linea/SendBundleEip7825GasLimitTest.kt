/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package linea.plugin.acc.test.rpc.linea

import linea.plugin.acc.test.TestCommandLineOptionsBuilder
import linea.plugin.acc.test.rpc.SendBundleRequest
import net.consensys.linea.config.TransactionGasLimitCap.EIP_7825_MAX_TRANSACTION_GAS_LIMIT
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import java.math.BigInteger

class SendBundleEip7825GasLimitTest : AbstractSendBundleTest() {

  override fun getTestCliOptions(): List<String> {
    return TestCommandLineOptionsBuilder()
      .set("--plugin-linea-max-tx-gas-limit=", "24000000")
      .set(
        "--plugin-linea-module-limit-file-path=",
        getResourcePath("/moduleLimitsLimitless.toml"),
      )
      .set("--plugin-linea-limitless-enabled=", "true")
      .build()
  }

  @Test
  fun bundleTxAboveEip7825CapIsRejectedWhenConfiguredLimitIsHigher() {
    val mulmodExecutor = deployMulmodExecutor()
    val mulmodTx = mulmodOperation(
      mulmodExecutor,
      accounts.primaryBenefactor,
      1,
      1_000,
      BigInteger.valueOf(EIP_7825_MAX_TRANSACTION_GAS_LIMIT + 1L),
    )

    val sendBundleRequest =
      SendBundleRequest(BundleParams(arrayOf(mulmodTx.rawTx), Integer.toHexString(2)))
    val sendBundleResponse = sendBundleRequest.execute(minerNode.nodeRequests())

    assertThat(sendBundleResponse.hasError()).isTrue()
    assertThat(sendBundleResponse.error.message)
      .contains("Gas limit 16777217 exceeds EIP-7825 maximum transaction gas limit of 16777216")
  }
}
