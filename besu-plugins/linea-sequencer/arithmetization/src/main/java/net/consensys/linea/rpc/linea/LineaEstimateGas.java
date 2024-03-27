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

package net.consensys.linea.rpc.linea;

import static org.hyperledger.besu.ethereum.api.jsonrpc.internal.results.Quantity.create;

import java.math.BigDecimal;
import java.math.BigInteger;
import java.math.RoundingMode;
import java.util.concurrent.atomic.AtomicInteger;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.google.common.annotations.VisibleForTesting;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.bl.TransactionProfitabilityCalculator;
import net.consensys.linea.config.LineaProfitabilityConfiguration;
import net.consensys.linea.config.LineaRpcConfiguration;
import net.consensys.linea.config.LineaTransactionPoolValidatorConfiguration;
import org.apache.tuweni.bytes.Bytes;
import org.bouncycastle.asn1.sec.SECNamedCurves;
import org.bouncycastle.asn1.x9.X9ECParameters;
import org.bouncycastle.crypto.params.ECDomainParameters;
import org.hyperledger.besu.crypto.SECPSignature;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.api.jsonrpc.internal.exception.InvalidJsonRpcParameters;
import org.hyperledger.besu.ethereum.api.jsonrpc.internal.parameters.JsonCallParameter;
import org.hyperledger.besu.ethereum.api.jsonrpc.internal.parameters.JsonRpcParameter;
import org.hyperledger.besu.ethereum.api.jsonrpc.internal.response.RpcErrorType;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.evm.tracing.EstimateGasOperationTracer;
import org.hyperledger.besu.plugin.services.BesuConfiguration;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.TransactionSimulationService;
import org.hyperledger.besu.plugin.services.exception.PluginRpcEndpointException;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcRequest;
import org.hyperledger.besu.plugin.services.rpc.RpcMethodError;

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
  private LineaRpcConfiguration rpcConfiguration;
  private LineaTransactionPoolValidatorConfiguration txValidatorConf;
  private LineaProfitabilityConfiguration profitabilityConf;
  private TransactionProfitabilityCalculator txProfitabilityCalculator;

  public LineaEstimateGas(
      final BesuConfiguration besuConfiguration,
      final TransactionSimulationService transactionSimulationService,
      final BlockchainService blockchainService) {
    this.besuConfiguration = besuConfiguration;
    this.transactionSimulationService = transactionSimulationService;
    this.blockchainService = blockchainService;
  }

  public void init(
      LineaRpcConfiguration rpcConfiguration,
      final LineaTransactionPoolValidatorConfiguration transactionValidatorConfiguration,
      final LineaProfitabilityConfiguration profitabilityConf) {
    this.rpcConfiguration = rpcConfiguration;
    this.txValidatorConf = transactionValidatorConfiguration;
    this.profitabilityConf = profitabilityConf;
    this.txProfitabilityCalculator = new TransactionProfitabilityCalculator(profitabilityConf);
  }

  public String getNamespace() {
    return "linea";
  }

  public String getName() {
    return "estimateGas";
  }

  public LineaEstimateGas.Response execute(final PluginRpcRequest request) {
    if (log.isDebugEnabled()) {
      // no matter if it overflows, since it is only used to correlate logs for this request,
      // so we only print callParameters once at the beginning, and we can reference them using the
      // sequence.
      LOG_SEQUENCE.incrementAndGet();
    }
    final var callParameters = parseRequest(request.getParams());
    final var minGasPrice = besuConfiguration.getMinGasPrice();

    final var transaction =
        createTransactionForSimulation(
            callParameters, txValidatorConf.maxTxGasLimit(), minGasPrice);
    log.atDebug()
        .setMessage("[{}] Parsed call parameters: {}; Transaction: {}")
        .addArgument(LOG_SEQUENCE::get)
        .addArgument(callParameters)
        .addArgument(transaction::toTraceLog)
        .log();
    final var estimatedGasUsed = estimateGasUsed(callParameters, transaction, minGasPrice);

    final Wei baseFee =
        blockchainService
            .getNextBlockBaseFee()
            .orElseThrow(
                () ->
                    new PluginRpcEndpointException(
                        RpcErrorType.INVALID_REQUEST, "Not on a baseFee market"));

    final Wei estimatedPriorityFee =
        getEstimatedPriorityFee(transaction, baseFee, minGasPrice, estimatedGasUsed);

    final var response =
        new Response(create(estimatedGasUsed), create(baseFee), create(estimatedPriorityFee));
    log.atDebug()
        .setMessage("[{}] Response for call params {} is {}")
        .addArgument(LOG_SEQUENCE::get)
        .addArgument(callParameters)
        .addArgument(response)
        .log();

    return response;
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

    final Wei profitablePriorityFee =
        txProfitabilityCalculator.profitablePriorityFeePerGas(
            transaction, profitabilityConf.estimateGasMinMargin(), minGasPrice, estimatedGasUsed);

    if (profitablePriorityFee.greaterOrEqualThan(priorityFeeLowerBound)) {
      return profitablePriorityFee;
    }

    log.atDebug()
        .setMessage(
            "[{}] Estimated priority fee {} is lower that the lower bound {}, returning the latter")
        .addArgument(LOG_SEQUENCE::get)
        .addArgument(profitablePriorityFee::toHumanReadableString)
        .addArgument(priorityFeeLowerBound::toHumanReadableString)
        .log();
    return priorityFeeLowerBound;
  }

  private Long estimateGasUsed(
      final JsonCallParameter callParameters,
      final Transaction transaction,
      final Wei minGasPrice) {
    final var tracer = new EstimateGasOperationTracer();
    final var chainHeadHash = blockchainService.getChainHeadHash();
    final var maybeSimulationResults =
        transactionSimulationService.simulate(transaction, chainHeadHash, tracer, true);

    return maybeSimulationResults
        .map(
            r -> {
              // if the transaction is invalid or doesn't have enough gas with the max it never will
              if (r.isInvalid()) {
                log.atDebug()
                    .setMessage("[{}] Invalid transaction {}, reason {}")
                    .addArgument(LOG_SEQUENCE::get)
                    .addArgument(transaction::toTraceLog)
                    .addArgument(r.result())
                    .log();
                throw new PluginRpcEndpointException(
                    new TransactionSimulationError(r.result().getInvalidReason().orElse("")));
              }
              if (!r.isSuccessful()) {
                log.atDebug()
                    .setMessage("[{}] Failed transaction {}, reason {}")
                    .addArgument(LOG_SEQUENCE::get)
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
                    new TransactionSimulationError(
                        "Failed transaction"
                            + invalidReason.map(ir -> ", reason: " + ir).orElse("")));
              }

              final var lowGasEstimation = r.result().getEstimateGasUsedByTransaction();
              final var lowResult =
                  transactionSimulationService.simulate(
                      createTransactionForSimulation(callParameters, lowGasEstimation, minGasPrice),
                      chainHeadHash,
                      tracer,
                      true);

              return lowResult
                  .map(
                      lr -> {
                        // if with the low estimation gas is successful then return this estimation
                        if (lr.isSuccessful()) {
                          log.atTrace()
                              .setMessage("[{}] Low gas estimation {} successful")
                              .addArgument(LOG_SEQUENCE::get)
                              .addArgument(lowGasEstimation)
                              .log();
                          return lowGasEstimation;
                        } else {
                          log.atTrace()
                              .setMessage("[{}] Low gas estimation {} unsuccessful, result{}")
                              .addArgument(LOG_SEQUENCE::get)
                              .addArgument(lowGasEstimation)
                              .addArgument(lr::result)
                              .log();

                          // else do a binary search to find the right estimation
                          int iterations = 0;
                          var high = highGasEstimation(lr.getGasEstimate(), tracer);
                          var mid = high;
                          var low = lowGasEstimation;
                          while (low + 1 < high) {
                            ++iterations;
                            mid = (high + low) / 2;

                            final var binarySearchResult =
                                transactionSimulationService.simulate(
                                    createTransactionForSimulation(
                                        callParameters, mid, minGasPrice),
                                    chainHeadHash,
                                    tracer,
                                    true);

                            if (binarySearchResult.isEmpty()
                                || !binarySearchResult.get().isSuccessful()) {
                              log.atTrace()
                                  .setMessage(
                                      "[{}]-[{}] Binary gas estimation search low={},mid={},high={}, unsuccessful result {}")
                                  .addArgument(LOG_SEQUENCE::get)
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
                                      "[{}]-[{}} Binary gas estimation search low={},mid={},high={}, successful")
                                  .addArgument(LOG_SEQUENCE::get)
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
                              .addArgument(LOG_SEQUENCE::get)
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

  private JsonCallParameter parseRequest(final Object[] params) {
    final var callParameters = parameterParser.required(params, 0, JsonCallParameter.class);
    validateParameters(callParameters);
    return callParameters;
  }

  private void validateParameters(final JsonCallParameter callParameters) {
    if (callParameters.getGasPrice() != null
        && (callParameters.getMaxFeePerGas().isPresent()
            || callParameters.getMaxPriorityFeePerGas().isPresent()
            || callParameters.getMaxFeePerBlobGas().isPresent())) {
      throw new InvalidJsonRpcParameters(
          "gasPrice cannot be used with maxFeePerGas or maxPriorityFeePerGas or maxFeePerBlobGas");
    }

    if (callParameters.getGasLimit() > 0
        && callParameters.getGasLimit() > txValidatorConf.maxTxGasLimit()) {
      throw new InvalidJsonRpcParameters(
          "gasLimit above maximum of: " + txValidatorConf.maxTxGasLimit());
    }
  }

  /**
   * Estimate gas by adding minimum gas remaining for some operation and the necessary gas for sub
   * calls
   *
   * @param gasEstimation transaction gas estimation
   * @param operationTracer estimate gas operation tracer
   * @return estimate gas
   */
  private long highGasEstimation(
      final long gasEstimation, final EstimateGasOperationTracer operationTracer) {
    // no more than 63/64s of the remaining gas can be passed to the sub calls
    final double subCallMultiplier =
        Math.pow(SUB_CALL_REMAINING_GAS_RATIO, operationTracer.getMaxDepth());
    // and minimum gas remaining is necessary for some operation (additionalStipend)
    final long gasStipend = operationTracer.getStipendNeeded();
    return ((long) ((gasEstimation + gasStipend) * subCallMultiplier));
  }

  private Transaction createTransactionForSimulation(
      final JsonCallParameter callParameters, final long maxTxGasLimit, final Wei minGasPrice) {

    final var txBuilder =
        Transaction.builder()
            .sender(callParameters.getFrom())
            .to(callParameters.getTo())
            .gasLimit(maxTxGasLimit)
            .payload(
                callParameters.getPayload() == null ? Bytes.EMPTY : callParameters.getPayload())
            .gasPrice(
                callParameters.getGasPrice() == null ? minGasPrice : callParameters.getGasPrice())
            .value(callParameters.getValue() == null ? Wei.ZERO : callParameters.getValue())
            .signature(FAKE_SIGNATURE_FOR_SIZE_CALCULATION);

    callParameters.getMaxFeePerGas().ifPresent(txBuilder::maxFeePerGas);
    callParameters.getMaxPriorityFeePerGas().ifPresent(txBuilder::maxPriorityFeePerGas);
    callParameters.getAccessList().ifPresent(txBuilder::accessList);

    return txBuilder.build();
  }

  public record Response(
      @JsonProperty String gasLimit,
      @JsonProperty String baseFeePerGas,
      @JsonProperty String priorityFeePerGas) {}

  private record TransactionSimulationError(String errorReason) implements RpcMethodError {
    @Override
    public int getCode() {
      return -32000;
    }

    @Override
    public String getMessage() {
      return errorReason;
    }
  }
}
