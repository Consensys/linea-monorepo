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
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyDouble;
import static org.mockito.ArgumentMatchers.anyInt;
import static org.mockito.ArgumentMatchers.anyLong;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.EnumSet;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import net.consensys.linea.bl.TransactionProfitabilityCalculator;
import net.consensys.linea.config.LineaProfitabilityConfiguration;
import net.consensys.linea.config.LineaRpcConfiguration;
import net.consensys.linea.config.LineaTracerConfiguration;
import net.consensys.linea.config.LineaTransactionPoolValidatorConfiguration;
import net.consensys.linea.plugins.config.LineaL1L2BridgeSharedConfiguration;
import org.hyperledger.besu.datatypes.StateOverrideMap;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.api.jsonrpc.internal.exception.InvalidJsonRpcParameters;
import org.hyperledger.besu.ethereum.transaction.ImmutableCallParameter;
import org.hyperledger.besu.evm.tracing.OperationTracer;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;
import org.hyperledger.besu.plugin.services.BesuConfiguration;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.RpcEndpointService;
import org.hyperledger.besu.plugin.services.TransactionSimulationService;
import org.hyperledger.besu.plugin.services.TransactionSimulationService.SimulationParameters;
import org.hyperledger.besu.plugin.services.WorldStateService;
import org.hyperledger.besu.plugin.services.exception.PluginRpcEndpointException;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcRequest;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcResponse;
import org.hyperledger.besu.plugin.services.rpc.RpcResponseType;
import org.junit.jupiter.api.Test;

class LineaEstimateGasTest {
  private static final String EIP_7825_LIMIT_EXCEEDED_MESSAGE =
      "Gas limit 16777217 exceeds EIP-7825 maximum transaction gas limit of 16777216";
  private static final String CONFIGURED_LIMIT_EXCEEDED_MESSAGE =
      "Gas limit 9000001 exceeds configured maximum transaction gas limit of 9000000 "
          + "(EIP-7825 maximum transaction gas limit is 16777216)";

  @Test
  void acceptsEstimateAtEip7825MaxWhenConfiguredLimitIsHigher() {
    final TestContext context = createLineaEstimateGas(24_000_000);

    assertDoesNotThrow(
        () -> context.method().validateEstimatedGasLimit(EIP_7825_MAX_TRANSACTION_GAS_LIMIT));
  }

  @Test
  void rejectsEstimateAboveEip7825MaxWhenConfiguredLimitIsHigher() {
    final TestContext context = createLineaEstimateGas(24_000_000);

    assertThatThrownBy(
            () ->
                context
                    .method()
                    .validateEstimatedGasLimit(EIP_7825_MAX_TRANSACTION_GAS_LIMIT + 1L))
        .isInstanceOf(PluginRpcEndpointException.class)
        .extracting(exception -> ((PluginRpcEndpointException) exception).getRpcMethodError())
        .satisfies(
            error ->
                assertThat(error.getMessage()).isEqualTo(EIP_7825_LIMIT_EXCEEDED_MESSAGE));
  }

  @Test
  void rejectsEstimateAboveConfiguredMaxWhenConfiguredLimitIsLower() {
    final TestContext context = createLineaEstimateGas(9_000_000);

    assertThatThrownBy(() -> context.method().validateEstimatedGasLimit(9_000_001L))
        .isInstanceOf(PluginRpcEndpointException.class)
        .extracting(exception -> ((PluginRpcEndpointException) exception).getRpcMethodError())
        .satisfies(
            error ->
                assertThat(error.getMessage()).isEqualTo(CONFIGURED_LIMIT_EXCEEDED_MESSAGE));
  }

  @Test
  void executeRejectsEstimateAboveEip7825MaxBeforeSimulationAndProfitabilityCalculation() {
    final TestContext context = createLineaEstimateGas(24_000_000);
    final PluginRpcRequest request = request(callParameter());
    when(context.besuConfiguration().getMinGasPrice()).thenReturn(Wei.ZERO);
    when(context.blockchainService().getNextBlockBaseFee()).thenReturn(Optional.of(Wei.ZERO));
    when(context.rpcEndpointService().call(eq("eth_estimateGas"), any(Object[].class)))
        .thenReturn(successResponse("0x1000001"));

    assertThatThrownBy(() -> context.method().execute(request))
        .isInstanceOf(PluginRpcEndpointException.class)
        .extracting(exception -> ((PluginRpcEndpointException) exception).getRpcMethodError())
        .satisfies(
            error ->
                assertThat(error.getMessage()).isEqualTo(EIP_7825_LIMIT_EXCEEDED_MESSAGE));
    verify(context.transactionSimulationService(), never())
        .simulate(
            any(org.hyperledger.besu.datatypes.CallParameter.class),
            anyStateOverrideMap(),
            any(ProcessableBlockHeader.class),
            any(OperationTracer.class),
            anySimulationParameters());
    verify(context.transactionProfitabilityCalculator(), never()).getCompressedTxSize(any());
    verify(context.transactionProfitabilityCalculator(), never())
        .profitablePriorityFeePerGas(any(), anyDouble(), anyLong(), any(), anyInt());
  }

  @Test
  void executeRejectsCallerProvidedGasAboveEip7825MaxWhenConfiguredLimitIsHigher() {
    final TestContext context = createLineaEstimateGas(24_000_000);
    final PluginRpcRequest request = request(callParameter(EIP_7825_MAX_TRANSACTION_GAS_LIMIT + 1L));

    assertThatThrownBy(() -> context.method().execute(request))
        .isInstanceOf(InvalidJsonRpcParameters.class)
        .hasMessage(EIP_7825_LIMIT_EXCEEDED_MESSAGE);
  }

  @Test
  void executeRejectsCallerProvidedGasAboveConfiguredMaxWhenConfiguredLimitIsLower() {
    final TestContext context = createLineaEstimateGas(9_000_000);
    final PluginRpcRequest request = request(callParameter(9_000_001L));

    assertThatThrownBy(() -> context.method().execute(request))
        .isInstanceOf(InvalidJsonRpcParameters.class)
        .hasMessage(CONFIGURED_LIMIT_EXCEEDED_MESSAGE);
  }

  private static TestContext createLineaEstimateGas(final int maxTxGasLimit) {
    final BesuConfiguration besuConfiguration = mock(BesuConfiguration.class);
    final TransactionSimulationService transactionSimulationService =
        mock(TransactionSimulationService.class);
    final BlockchainService blockchainService = mock(BlockchainService.class);
    final RpcEndpointService rpcEndpointService = mock(RpcEndpointService.class);
    final TransactionProfitabilityCalculator transactionProfitabilityCalculator =
        mock(TransactionProfitabilityCalculator.class);
    final LineaEstimateGas method =
        new LineaEstimateGas(
            besuConfiguration, transactionSimulationService, blockchainService, rpcEndpointService);

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
        transactionProfitabilityCalculator);

    return new TestContext(
        method,
        besuConfiguration,
        transactionSimulationService,
        blockchainService,
        rpcEndpointService,
        transactionProfitabilityCalculator);
  }

  private static PluginRpcRequest request(final Object... params) {
    final PluginRpcRequest request = mock(PluginRpcRequest.class);
    when(request.getParams()).thenReturn(params);
    return request;
  }

  private static ImmutableCallParameter callParameter() {
    return ImmutableCallParameter.builder().build();
  }

  private static ImmutableCallParameter callParameter(final long gasLimit) {
    return ImmutableCallParameter.builder().gas(gasLimit).build();
  }

  private static Optional<StateOverrideMap> anyStateOverrideMap() {
    return any();
  }

  private static EnumSet<SimulationParameters> anySimulationParameters() {
    return any();
  }

  private static PluginRpcResponse successResponse(final Object result) {
    return new PluginRpcResponse() {
      @Override
      public Object getResult() {
        return result;
      }

      @Override
      public RpcResponseType getType() {
        return RpcResponseType.SUCCESS;
      }
    };
  }

  private record TestContext(
      LineaEstimateGas method,
      BesuConfiguration besuConfiguration,
      TransactionSimulationService transactionSimulationService,
      BlockchainService blockchainService,
      RpcEndpointService rpcEndpointService,
      TransactionProfitabilityCalculator transactionProfitabilityCalculator) {}
}
