/*
 * Copyright Consensys Software Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
 * the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package net.consensys.linea.rpc.methods;

import static net.consensys.linea.sequencer.modulelimit.ModuleLineCountValidator.ModuleLineCountResult.MODULE_NOT_DEFINED;
import static net.consensys.linea.sequencer.modulelimit.ModuleLineCountValidator.ModuleLineCountResult.TX_MODULE_LINE_COUNT_OVERFLOW;
import static net.consensys.linea.zktracer.Fork.LONDON;
import static org.hyperledger.besu.ethereum.api.jsonrpc.internal.results.Quantity.create;

import java.math.BigDecimal;
import java.math.BigInteger;
import java.math.RoundingMode;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.atomic.AtomicInteger;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.google.common.annotations.VisibleForTesting;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.bl.TransactionProfitabilityCalculator;
import net.consensys.linea.config.LineaProfitabilityConfiguration;
import net.consensys.linea.config.LineaRpcConfiguration;
import net.consensys.linea.config.LineaTransactionPoolValidatorConfiguration;
import net.consensys.linea.plugins.config.LineaL1L2BridgeSharedConfiguration;
import net.consensys.linea.sequencer.TracerAggregator;
import net.consensys.linea.sequencer.modulelimit.ModuleLimitsValidationResult;
import net.consensys.linea.sequencer.modulelimit.ModuleLineCountValidator;
import net.consensys.linea.zktracer.ZkTracer;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.units.bigints.UInt256;
import org.bouncycastle.asn1.sec.SECNamedCurves;
import org.bouncycastle.asn1.x9.X9ECParameters;
import org.bouncycastle.crypto.params.ECDomainParameters;
import org.hyperledger.besu.crypto.SECPSignature;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.StateOverrideMap;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.api.jsonrpc.internal.exception.InvalidJsonRpcParameters;
import org.hyperledger.besu.ethereum.api.jsonrpc.internal.exception.InvalidJsonRpcRequestException;
import org.hyperledger.besu.ethereum.api.jsonrpc.internal.parameters.JsonRpcParameter;
import org.hyperledger.besu.ethereum.api.jsonrpc.internal.response.RpcErrorType;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.ethereum.transaction.CallParameter;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.operation.Operation;
import org.hyperledger.besu.evm.operation.SStoreOperation;
import org.hyperledger.besu.evm.tracing.OperationTracer;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;
import org.hyperledger.besu.plugin.services.BesuConfiguration;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.RpcEndpointService;
import org.hyperledger.besu.plugin.services.TransactionSimulationService;
import org.hyperledger.besu.plugin.services.exception.PluginRpcEndpointException;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcRequest;
import org.hyperledger.besu.plugin.services.rpc.RpcMethodError;
import org.hyperledger.besu.plugin.services.rpc.RpcResponseType;

@Slf4j
public class LineaEstimateGas {
  @VisibleForTesting public static final SECPSignature FAKE_SIGNATURE_FOR_SIZE_CALCULATION;

  private static final double SUB_CALL_REMAINING_GAS_RATIO = 65D / 64D;
  private static final AtomicInteger LOG_SEQUENCE = new AtomicInteger();

  static {
    final X9ECParameters params = SECNamedCurves.getByName("secp256k1");
    final ECDomainParameters curve =
        new ECDomainParameters(params.getCurve(), params.getG(), params.getN(), params.getH());
    FAKE_SIGNATURE_FOR_SIZE_CALCULATION =
        SECPSignature.create(
            new BigInteger(
                "66397251408932042429874251838229702988618145381408295790259650671563847073199"),
            new BigInteger(
                "24729624138373455972486746091821238755870276413282629437244319694880507882088"),
            (byte) 0,
            curve.getN());
  }

  private final JsonRpcParameter parameterParser = new JsonRpcParameter();
  private final BesuConfiguration besuConfiguration;
  private final TransactionSimulationService transactionSimulationService;
  private final BlockchainService blockchainService;
  private final RpcEndpointService rpcEndpointService;
  private LineaRpcConfiguration rpcConfiguration;
  private LineaTransactionPoolValidatorConfiguration txValidatorConf;
  private LineaProfitabilityConfiguration profitabilityConf;
  private TransactionProfitabilityCalculator txProfitabilityCalculator;
  private LineaL1L2BridgeSharedConfiguration l1L2BridgeConfiguration;
  private ModuleLineCountValidator moduleLineCountValidator;
  private UInt256 maxTxGasLimit;

  public LineaEstimateGas(
      final BesuConfiguration besuConfiguration,
      final TransactionSimulationService transactionSimulationService,
      final BlockchainService blockchainService,
      final RpcEndpointService rpcEndpointService) {
    this.besuConfiguration = besuConfiguration;
    this.transactionSimulationService = transactionSimulationService;
    this.blockchainService = blockchainService;
    this.rpcEndpointService = rpcEndpointService;
  }

  public void init(
      final LineaRpcConfiguration rpcConfiguration,
      final LineaTransactionPoolValidatorConfiguration transactionValidatorConfiguration,
      final LineaProfitabilityConfiguration profitabilityConf,
      final Map<String, Integer> limitsMap,
      final LineaL1L2BridgeSharedConfiguration l1L2BridgeConfiguration) {
    this.rpcConfiguration = rpcConfiguration;
    this.txValidatorConf = transactionValidatorConfiguration;
    this.profitabilityConf = profitabilityConf;
    this.txProfitabilityCalculator = new TransactionProfitabilityCalculator(profitabilityConf);
    this.l1L2BridgeConfiguration = l1L2BridgeConfiguration;
    this.moduleLineCountValidator = new ModuleLineCountValidator(limitsMap);
    this.maxTxGasLimit = UInt256.valueOf(txValidatorConf.maxTxGasLimit());
  }

  public String getNamespace() {
    return "linea";
  }

  public String getName() {
    return "estimateGas";
  }

  public LineaEstimateGas.Response execute(final PluginRpcRequest request) {
    try {
      final long logId;
      if (log.isDebugEnabled()) {
        // no matter if it overflows, since it is only used to correlate logs for this request,
        // so we only print callParameters once at the beginning, and we can reference them using
        // the logId.
        logId = LOG_SEQUENCE.incrementAndGet();
      } else {
        logId = 0;
      }

      final var callParameters = parseCallParameters(request.getParams());
      final var maybeStateOverrides = getStateOverrideMap(request.getParams());
      final var minGasPrice = besuConfiguration.getMinGasPrice();
      final var gasLimitUpperBound = calculateGasLimitUpperBound(callParameters, logId);
      final Wei baseFee =
          blockchainService
              .getNextBlockBaseFee()
              .orElseThrow(
                  () ->
                      new PluginRpcEndpointException(
                          RpcErrorType.INVALID_REQUEST, "Not on a baseFee market"));
      final var transaction =
          createTransactionForSimulation(callParameters, gasLimitUpperBound, baseFee, logId);
      log.atDebug()
          .setMessage("[{}] Parsed call parameters: {}; Transaction: {}; Gas limit upper bound {}")
          .addArgument(logId)
          .addArgument(callParameters)
          .addArgument(transaction::toTraceLog)
          .addArgument(gasLimitUpperBound)
          .log();
      final var estimatedGasUsed =
          estimateGasUsed(callParameters, maybeStateOverrides, transaction, baseFee, logId);

      final Wei estimatedPriorityFee =
          getEstimatedPriorityFee(transaction, baseFee, minGasPrice, estimatedGasUsed);

      final var response =
          new Response(create(estimatedGasUsed), create(baseFee), create(estimatedPriorityFee));
      log.atDebug()
          .setMessage("[{}] Response for call params {} is {}")
          .addArgument(logId)
          .addArgument(callParameters)
          .addArgument(response)
          .log();

      return response;
    } catch (PluginRpcEndpointException | InvalidJsonRpcRequestException e) {
      throw e;
    } catch (Exception e) {
      throw new PluginRpcEndpointException(new InternalError(e.getMessage()), null, e);
    }
  }

  private long calculateGasLimitUpperBound(final CallParameter callParameters, final long logId) {
    if (callParameters.getSender().isPresent()) {
      final var sender = callParameters.getSender().get();
      final var maxGasPrice = calculateTxMaxGasPrice(callParameters);
      log.atTrace()
          .setMessage("[{}] Calculated max gas price {}")
          .addArgument(logId)
          .addArgument(maxGasPrice)
          .log();
      if (!maxGasPrice.equals(Wei.ZERO)) {
        final Wei balance = getSenderBalance(sender, logId);
        if (balance.greaterThan(Wei.ZERO)) {
          final var value = callParameters.getValue().orElse(Wei.ZERO);
          final var balanceForGas = balance.subtract(value);
          final var gasLimitForBalance = balanceForGas.divide(maxGasPrice).toUInt256();
          if (gasLimitForBalance.lessThan(maxTxGasLimit)) {
            final var gasLimitUpperBound = gasLimitForBalance.toLong();
            log.atTrace()
                .setMessage(
                    "[{}] Calculated gasLimitUpperBound {}; gasLimitForBalance {}, balance {}, value {}, balanceForGas {}, maxGasPrice {}")
                .addArgument(logId)
                .addArgument(gasLimitUpperBound)
                .addArgument(gasLimitForBalance::toDecimalString)
                .addArgument(balance::toHumanReadableString)
                .addArgument(value::toHumanReadableString)
                .addArgument(balanceForGas::toHumanReadableString)
                .addArgument(maxGasPrice::toHumanReadableString)
                .log();
            return gasLimitUpperBound;
          }
        }
      }
    }

    return txValidatorConf.maxTxGasLimit();
  }

  private Wei getSenderBalance(final Address sender, final long logId) {
    final var resp =
        rpcEndpointService.call("eth_getBalance", new Object[] {sender.toHexString(), "latest"});
    if (!resp.getType().equals(RpcResponseType.SUCCESS)) {
      throw new PluginRpcEndpointException(new InternalError("Unable to query sender balance"));
    }
    final Wei balance = Wei.fromHexString((String) resp.getResult());
    log.atTrace()
        .setMessage("[{}] eth_getBalance response for {} is {}, balance {}")
        .addArgument(logId)
        .addArgument(sender)
        .addArgument(resp::getResult)
        .addArgument(balance::toHumanReadableString)
        .log();
    return balance;
  }

  private Wei calculateTxMaxGasPrice(final CallParameter callParameters) {
    return callParameters
        .getMaxFeePerGas()
        .orElseGet(() -> callParameters.getGasPrice().orElse(Wei.ZERO));
  }

  private Wei getEstimatedPriorityFee(
      final Transaction transaction,
      final Wei baseFee,
      final Wei minGasPrice,
      final long estimatedGasUsed) {
    final Wei priorityFeeLowerBound = minGasPrice.subtract(baseFee);

    if (rpcConfiguration.estimateGasCompatibilityModeEnabled()) {
      return Wei.of(
          rpcConfiguration
              .estimateGasCompatibilityMultiplier()
              .multiply(new BigDecimal(priorityFeeLowerBound.getAsBigInteger()))
              .setScale(0, RoundingMode.CEILING)
              .toBigInteger());
    }

    return txProfitabilityCalculator.profitablePriorityFeePerGas(
        transaction, profitabilityConf.estimateGasMinMargin(), estimatedGasUsed, minGasPrice);
  }

  private Long estimateGasUsed(
      final CallParameter callParameters,
      final Optional<StateOverrideMap> maybeStateOverrides,
      final Transaction transaction,
      final Wei baseFee,
      final long logId) {

    final var estimateGasTracer = new EstimateGasOperationTracer();
    final var pendingBlockHeader = transactionSimulationService.simulatePendingBlockHeader();
    final var zkTracer = createZkTracer(pendingBlockHeader, blockchainService.getChainId().get());
    final TracerAggregator zkAndGasTracer = TracerAggregator.create(estimateGasTracer, zkTracer);

    final var maybeSimulationResults =
        transactionSimulationService.simulate(
            transaction, maybeStateOverrides, pendingBlockHeader, zkAndGasTracer, false, true);

    ModuleLimitsValidationResult moduleLimit =
        moduleLineCountValidator.validate(zkTracer.getModulesLineCount());

    if (moduleLimit.getResult() != ModuleLineCountValidator.ModuleLineCountResult.VALID) {
      handleModuleOverLimit(moduleLimit);
    }

    return maybeSimulationResults
        .map(
            r -> {
              // if the transaction is invalid or doesn't have enough gas with the max it never will
              if (r.isInvalid()) {
                log.atDebug()
                    .setMessage("[{}] Invalid transaction {}, reason {}")
                    .addArgument(logId)
                    .addArgument(transaction::toTraceLog)
                    .addArgument(r.result())
                    .log();
                throw new PluginRpcEndpointException(
                    new InternalError(r.result().getInvalidReason().orElse("")));
              }
              if (!r.isSuccessful()) {
                log.atDebug()
                    .setMessage("[{}] Failed transaction {}, reason {}")
                    .addArgument(logId)
                    .addArgument(transaction::toTraceLog)
                    .addArgument(r.result())
                    .log();
                r.getRevertReason()
                    .ifPresent(
                        rr -> {
                          throw new PluginRpcEndpointException(
                              RpcErrorType.REVERT_ERROR, rr.toHexString());
                        });
                final var invalidReason = r.result().getInvalidReason();
                throw new PluginRpcEndpointException(
                    new InternalError(
                        "Failed transaction"
                            + invalidReason.map(ir -> ", reason: " + ir).orElse("")));
              }

              final var lowGasEstimation = r.result().getEstimateGasUsedByTransaction();
              final var lowResult =
                  transactionSimulationService.simulate(
                      createTransactionForSimulation(
                          callParameters, lowGasEstimation, baseFee, logId),
                      maybeStateOverrides,
                      pendingBlockHeader,
                      estimateGasTracer,
                      false,
                      true);

              return lowResult
                  .map(
                      lr -> {
                        // if with the low estimation gas is successful then return this estimation
                        if (lr.isSuccessful()) {
                          log.atTrace()
                              .setMessage("[{}] Low gas estimation {} successful")
                              .addArgument(logId)
                              .addArgument(lowGasEstimation)
                              .log();
                          return lowGasEstimation;
                        } else {
                          log.atTrace()
                              .setMessage("[{}] Low gas estimation {} unsuccessful, result{}")
                              .addArgument(logId)
                              .addArgument(lowGasEstimation)
                              .addArgument(lr::result)
                              .log();

                          // else do a binary search to find the right estimation
                          int iterations = 0;
                          var high = highGasEstimation(lr.getGasEstimate(), estimateGasTracer);
                          var mid = high;
                          var low = lowGasEstimation;
                          while (low + 1 < high) {
                            ++iterations;
                            mid = (high + low) / 2;

                            final var binarySearchResult =
                                transactionSimulationService.simulate(
                                    createTransactionForSimulation(
                                        callParameters, mid, baseFee, logId),
                                    maybeStateOverrides,
                                    pendingBlockHeader,
                                    estimateGasTracer,
                                    false,
                                    true);

                            if (binarySearchResult.isEmpty()
                                || !binarySearchResult.get().isSuccessful()) {
                              log.atTrace()
                                  .setMessage(
                                      "[{}]-[{}] Binary gas estimation search low={},mid={},high={}, unsuccessful result {}")
                                  .addArgument(logId)
                                  .addArgument(iterations)
                                  .addArgument(low)
                                  .addArgument(mid)
                                  .addArgument(high)
                                  .addArgument(
                                      () ->
                                          binarySearchResult
                                              .map(result -> result.result().toString())
                                              .orElse("empty"))
                                  .log();
                              low = mid;
                            } else {
                              log.atTrace()
                                  .setMessage(
                                      "[{}]-[{}] Binary gas estimation search low={},mid={},high={}, successful")
                                  .addArgument(logId)
                                  .addArgument(iterations)
                                  .addArgument(low)
                                  .addArgument(mid)
                                  .addArgument(high)
                                  .log();
                              high = mid;
                            }
                          }
                          log.atDebug()
                              .setMessage(
                                  "[{}] Binary gas estimation search={} after {} iterations")
                              .addArgument(logId)
                              .addArgument(high)
                              .addArgument(iterations)
                              .log();
                          return high;
                        }
                      })
                  .orElseThrow(
                      () ->
                          new PluginRpcEndpointException(
                              RpcErrorType.PLUGIN_INTERNAL_ERROR, "Empty result from simulation"));
            })
        .orElseThrow(
            () ->
                new PluginRpcEndpointException(
                    RpcErrorType.PLUGIN_INTERNAL_ERROR, "Empty result from simulation"));
  }

  private CallParameter parseCallParameters(final Object[] params) {
    final CallParameter callParameters;
    try {
      callParameters = parameterParser.required(params, 0, CallParameter.class);
    } catch (JsonRpcParameter.JsonRpcParameterException e) {
      throw new InvalidJsonRpcParameters(
          "Invalid call parameters (index 0)", RpcErrorType.INVALID_CALL_PARAMS);
    }
    validateCallParameters(callParameters);
    return callParameters;
  }

  private void validateCallParameters(final CallParameter callParameters) {
    if (callParameters.getGasPrice().isPresent() && isBaseFeeTransaction(callParameters)) {
      throw new InvalidJsonRpcParameters(
          "gasPrice cannot be used with maxFeePerGas or maxPriorityFeePerGas or maxFeePerBlobGas");
    }

    final var gasLimit = callParameters.getGas().orElse(0L);
    if (gasLimit > txValidatorConf.maxTxGasLimit()) {
      throw new InvalidJsonRpcParameters(
          "gasLimit above maximum of: " + txValidatorConf.maxTxGasLimit());
    }
  }

  protected Optional<StateOverrideMap> getStateOverrideMap(final Object[] params) {
    try {
      return parameterParser.optional(params, 1, StateOverrideMap.class);
    } catch (JsonRpcParameter.JsonRpcParameterException e) {
      throw new InvalidJsonRpcRequestException(
          "Invalid account overrides parameter (index 1)", RpcErrorType.INVALID_CALL_PARAMS, e);
    }
  }

  private boolean isBaseFeeTransaction(final CallParameter callParameters) {
    return (callParameters.getMaxFeePerGas().isPresent()
        || callParameters.getMaxPriorityFeePerGas().isPresent()
        || callParameters.getMaxFeePerBlobGas().isPresent());
  }

  /**
   * Estimate gas by adding minimum gas remaining for some operation and the necessary gas for sub
   * calls
   *
   * @param gasEstimation transaction gas estimation
   * @param estimateGasTracer estimate gas operation tracer
   * @return estimate gas
   */
  private long highGasEstimation(
      final long gasEstimation, final EstimateGasOperationTracer estimateGasTracer) {

    // no more than 63/64s of the remaining gas can be passed to the sub calls
    final double subCallMultiplier =
        Math.pow(SUB_CALL_REMAINING_GAS_RATIO, estimateGasTracer.getMaxDepth());
    // and minimum gas remaining is necessary for some operation (additionalStipend)
    final long gasStipend = estimateGasTracer.getStipendNeeded();
    return ((long) ((gasEstimation + gasStipend) * subCallMultiplier));
  }

  private Transaction createTransactionForSimulation(
      final CallParameter callParameters,
      final long maxTxGasLimit,
      final Wei baseFee,
      final long logId) {

    final var txBuilder =
        Transaction.builder()
            .sender(callParameters.getSender().orElse(Address.ZERO))
            .nonce(callParameters.getNonce().orElseGet(() -> getSenderNonce(callParameters, logId)))
            .gasLimit(maxTxGasLimit)
            .payload(callParameters.getPayload().orElse(Bytes.EMPTY))
            .value(callParameters.getValue().orElse(Wei.ZERO))
            .signature(FAKE_SIGNATURE_FOR_SIZE_CALCULATION);

    callParameters.getTo().ifPresent(txBuilder::to);

    if (isBaseFeeTransaction(callParameters)) {
      txBuilder.maxFeePerGas(callParameters.getMaxFeePerGas().orElse(Wei.ZERO));
      txBuilder.maxPriorityFeePerGas(callParameters.getMaxPriorityFeePerGas().orElse(Wei.ZERO));
    } else {
      txBuilder.gasPrice(callParameters.getGasPrice().orElse(baseFee));
    }

    callParameters.getAccessList().ifPresent(txBuilder::accessList);

    final var txType = txBuilder.guessType().getTransactionType();

    if (txType.supportsBlob()) {
      txBuilder.maxFeePerBlobGas(callParameters.getMaxFeePerBlobGas().orElse(Wei.ZERO));
    }

    callParameters
        .getChainId()
        .ifPresentOrElse(
            txBuilder::chainId,
            () -> {
              if (txType.requiresChainId()) {
                blockchainService.getChainId().ifPresent(txBuilder::chainId);
              }
            });

    return txBuilder.build();
  }

  private long getSenderNonce(final CallParameter callParameters, final long logId) {
    return callParameters
        .getSender()
        .map(
            sender -> {
              final var resp =
                  rpcEndpointService.call(
                      "eth_getTransactionCount", new Object[] {sender.toHexString(), "latest"});
              if (!resp.getType().equals(RpcResponseType.SUCCESS)) {
                throw new PluginRpcEndpointException(
                    new InternalError("Unable to query sender nonce"));
              }
              final Long nonce = Long.decode((String) resp.getResult());
              log.atTrace()
                  .setMessage("[{}] eth_getTransactionCount response for {} is {}, nonce {}")
                  .addArgument(logId)
                  .addArgument(sender)
                  .addArgument(resp::getResult)
                  .addArgument(nonce)
                  .log();
              return nonce;
            })
        .orElse(0L);
  }

  private ZkTracer createZkTracer(
      final ProcessableBlockHeader pendingBlockHeader, final BigInteger chainId) {
    var zkTracer = new ZkTracer(LONDON, l1L2BridgeConfiguration, chainId);
    zkTracer.traceStartConflation(1L);
    zkTracer.traceStartBlock(pendingBlockHeader, pendingBlockHeader.getCoinbase());
    return zkTracer;
  }

  private void handleModuleOverLimit(ModuleLimitsValidationResult moduleLimitResult) {
    // Throw specific exceptions based on the type of limit exceeded
    if (moduleLimitResult.getResult() == MODULE_NOT_DEFINED) {
      String moduleNotDefinedMsg =
          String.format(
              "Module %s does not exist in the limits file.", moduleLimitResult.getModuleName());
      log.error(moduleNotDefinedMsg);
      throw new PluginRpcEndpointException(new InternalError(moduleNotDefinedMsg));
    }
    if (moduleLimitResult.getResult() == TX_MODULE_LINE_COUNT_OVERFLOW) {
      String txOverflowMsg =
          String.format(
              "Transaction line count for module %s=%s is above the limit %s",
              moduleLimitResult.getModuleName(),
              moduleLimitResult.getModuleLineCount(),
              moduleLimitResult.getModuleLineLimit());
      log.warn(txOverflowMsg);
      throw new PluginRpcEndpointException(new InternalError(txOverflowMsg));
    }

    final String internalErrorMsg =
        String.format("Do not know what to do with result %s", moduleLimitResult.getResult());
    log.error(internalErrorMsg);
    throw new PluginRpcEndpointException(RpcErrorType.PLUGIN_INTERNAL_ERROR, internalErrorMsg);
  }

  public record Response(
      @JsonProperty String gasLimit,
      @JsonProperty String baseFeePerGas,
      @JsonProperty String priorityFeePerGas) {}

  private record InternalError(String errorReason) implements RpcMethodError {
    @Override
    public int getCode() {
      return -32000;
    }

    @Override
    public String getMessage() {
      return errorReason;
    }
  }

  private static class EstimateGasOperationTracer implements OperationTracer {

    private int maxDepth = 0;

    private long sStoreStipendNeeded = 0L;

    /** Default constructor. */
    public EstimateGasOperationTracer() {}

    @Override
    public void tracePostExecution(
        final MessageFrame frame, final Operation.OperationResult operationResult) {
      if (frame.getCurrentOperation() instanceof SStoreOperation sStoreOperation
          && sStoreStipendNeeded == 0L) {
        sStoreStipendNeeded = sStoreOperation.getMinimumGasRemaining();
      }
      if (maxDepth < frame.getDepth()) {
        maxDepth = frame.getDepth();
      }
    }

    /**
     * Gets max depth.
     *
     * @return the max depth
     */
    public int getMaxDepth() {
      return maxDepth;
    }

    /**
     * Gets stipend needed.
     *
     * @return the stipend needed
     */
    public long getStipendNeeded() {
      return sStoreStipendNeeded;
    }
  }
}
