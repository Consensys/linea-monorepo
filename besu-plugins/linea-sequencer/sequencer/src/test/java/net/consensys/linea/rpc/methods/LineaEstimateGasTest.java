/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */

package net.consensys.linea.rpc.methods;

import static net.consensys.linea.config.TransactionGasLimitCap.EIP_7825_MAX_TRANSACTION_GAS_LIMIT;
import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.junit.jupiter.api.Assertions.assertDoesNotThrow;
import static org.mockito.Mockito.mock;

import java.util.Map;
import java.util.Set;
import net.consensys.linea.bl.TransactionProfitabilityCalculator;
import net.consensys.linea.config.LineaProfitabilityConfiguration;
import net.consensys.linea.config.LineaRpcConfiguration;
import net.consensys.linea.config.LineaTracerConfiguration;
import net.consensys.linea.config.LineaTransactionPoolValidatorConfiguration;
import net.consensys.linea.plugins.config.LineaL1L2BridgeSharedConfiguration;
import org.hyperledger.besu.plugin.services.BesuConfiguration;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.RpcEndpointService;
import org.hyperledger.besu.plugin.services.TransactionSimulationService;
import org.hyperledger.besu.plugin.services.WorldStateService;
import org.hyperledger.besu.plugin.services.exception.PluginRpcEndpointException;
import org.junit.jupiter.api.Test;

class LineaEstimateGasTest {
  private static final String EIP_7825_LIMIT_EXCEEDED_MESSAGE =
      "Gas limit 16777217 exceeds EIP-7825 maximum transaction gas limit of 16777216";
  private static final String CONFIGURED_LIMIT_EXCEEDED_MESSAGE =
      "Gas limit 9000001 exceeds configured maximum transaction gas limit of 9000000 "
          + "(EIP-7825 maximum transaction gas limit is 16777216)";

  @Test
  void acceptsEstimateAtEip7825MaxWhenConfiguredLimitIsHigher() {
    final LineaEstimateGas method = createLineaEstimateGas(24_000_000);

    assertDoesNotThrow(() -> method.validateEstimatedGasLimit(EIP_7825_MAX_TRANSACTION_GAS_LIMIT));
  }

  @Test
  void rejectsEstimateAboveEip7825MaxWhenConfiguredLimitIsHigher() {
    final LineaEstimateGas method = createLineaEstimateGas(24_000_000);

    assertThatThrownBy(
            () -> method.validateEstimatedGasLimit(EIP_7825_MAX_TRANSACTION_GAS_LIMIT + 1L))
        .isInstanceOf(PluginRpcEndpointException.class)
        .extracting(exception -> ((PluginRpcEndpointException) exception).getRpcMethodError())
        .satisfies(
            error ->
                assertThat(error.getMessage()).isEqualTo(EIP_7825_LIMIT_EXCEEDED_MESSAGE));
  }

  @Test
  void rejectsEstimateAboveConfiguredMaxWhenConfiguredLimitIsLower() {
    final LineaEstimateGas method = createLineaEstimateGas(9_000_000);

    assertThatThrownBy(() -> method.validateEstimatedGasLimit(9_000_001L))
        .isInstanceOf(PluginRpcEndpointException.class)
        .extracting(exception -> ((PluginRpcEndpointException) exception).getRpcMethodError())
        .satisfies(
            error ->
                assertThat(error.getMessage()).isEqualTo(CONFIGURED_LIMIT_EXCEEDED_MESSAGE));
  }

  private static LineaEstimateGas createLineaEstimateGas(final int maxTxGasLimit) {
    final LineaEstimateGas method =
        new LineaEstimateGas(
            mock(BesuConfiguration.class),
            mock(TransactionSimulationService.class),
            mock(BlockchainService.class),
            mock(RpcEndpointService.class));

    method.init(
        mock(LineaRpcConfiguration.class),
        LineaTransactionPoolValidatorConfiguration.builder()
            .denyListPath("")
            .deniedAddresses(Set.of())
            .bundleOverridingDenyListPath("")
            .bundleDeniedAddresses(Set.of())
            .maxTxGasLimit(maxTxGasLimit)
            .txPoolSimulationCheckApiEnabled(false)
            .txPoolSimulationCheckP2pEnabled(false)
            .build(),
        mock(LineaProfitabilityConfiguration.class),
        mock(LineaL1L2BridgeSharedConfiguration.class),
        LineaTracerConfiguration.builder()
            .moduleLimitsFilePath("")
            .moduleLimitsMap(Map.of())
            .isLimitless(true)
            .build(),
        mock(WorldStateService.class),
        mock(TransactionProfitabilityCalculator.class));

    return method;
  }
}
