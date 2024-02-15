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

import java.math.BigInteger;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.google.common.annotations.VisibleForTesting;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.bl.TransactionProfitabilityCalculator;
import net.consensys.linea.config.LineaTransactionSelectorConfiguration;
import net.consensys.linea.config.LineaTransactionValidatorConfiguration;
import org.apache.tuweni.bytes.Bytes;
import org.bouncycastle.asn1.sec.SECNamedCurves;
import org.bouncycastle.asn1.x9.X9ECParameters;
import org.bouncycastle.crypto.params.ECDomainParameters;
import org.hyperledger.besu.crypto.SECPSignature;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.api.jsonrpc.internal.exception.InvalidJsonRpcParameters;
import org.hyperledger.besu.ethereum.api.jsonrpc.internal.parameters.JsonCallParameter;
import org.hyperledger.besu.ethereum.api.jsonrpc.internal.parameters.JsonRpcParameter;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.evm.tracing.EstimateGasOperationTracer;
import org.hyperledger.besu.plugin.services.BesuConfiguration;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.TransactionSimulationService;
import org.hyperledger.besu.plugin.services.rpc.PluginRpcRequest;

@Slf4j
public class LineaEstimateGas {
  @VisibleForTesting public static final SECPSignature FAKE_SIGNATURE_FOR_SIZE_CALCULATION;

  private static final double SUB_CALL_REMAINING_GAS_RATIO = 65D / 64D;

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
  private LineaTransactionValidatorConfiguration txValidatorConf;
  private LineaTransactionSelectorConfiguration txSelectorConf;
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
      final LineaTransactionValidatorConfiguration transactionValidatorConfiguration,
      final LineaTransactionSelectorConfiguration transactionSelectorConfiguration) {
    this.txValidatorConf = transactionValidatorConfiguration;
    this.txSelectorConf = transactionSelectorConfiguration;
    this.txProfitabilityCalculator = new TransactionProfitabilityCalculator(txSelectorConf);
  }

  public String getNamespace() {
    return "linea";
  }

  public String getName() {
    return "estimateGas";
  }

  public LineaEstimateGas.Response execute(final PluginRpcRequest request) {
    final var callParameters = parseRequest(request.getParams());
    final var minGasPrice = besuConfiguration.getMinGasPrice();

    final var transaction =
        createTransactionForSimulation(
            callParameters, txValidatorConf.maxTxGasLimit(), minGasPrice);
    log.atTrace()
        .setMessage("Parsed call parameters: {}; Transaction: {}")
        .addArgument(callParameters)
        .addArgument(transaction::toTraceLog)
        .log();
    final var estimatedGasUsed = estimateGasUsed(callParameters, transaction, minGasPrice);

    final Wei estimatedPriorityFee =
        txProfitabilityCalculator.profitablePriorityFeePerGas(
            transaction, minGasPrice, estimatedGasUsed);

    final Wei baseFee =
        blockchainService
            .getNextBlockBaseFee()
            .orElseThrow(() -> new IllegalStateException("Not on a baseFee market"));

    final Wei priorityFeeLowerBound = minGasPrice.subtract(baseFee);
    final Wei boundedEstimatedPriorityFee;
    if (estimatedPriorityFee.lessThan(priorityFeeLowerBound)) {
      boundedEstimatedPriorityFee = priorityFeeLowerBound;
      log.atDebug()
          .setMessage(
              "Estimated priority fee {} is lower that the lower bound {}, returning the latter")
          .addArgument(estimatedPriorityFee::toHumanReadableString)
          .addArgument(boundedEstimatedPriorityFee::toHumanReadableString)
          .log();
    } else {
      boundedEstimatedPriorityFee = estimatedPriorityFee;
    }

    final var response =
        new Response(
            create(estimatedGasUsed), create(baseFee), create(boundedEstimatedPriorityFee));
    log.debug("Response for call params {} is {}", callParameters, response);

    return response;
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

              // if the transaction is invalid or doesn't have enough gas with the max it never
              // will!
              if (r.isInvalid() || !r.isSuccessful()) {
                log.atDebug()
                    .setMessage("Invalid or unsuccessful transaction {}, reason {}")
                    .addArgument(transaction::toTraceLog)
                    .addArgument(r.result())
                    .log();
                final var invalidReason = r.result().getInvalidReason();
                throw new RuntimeException(
                    "Invalid or unsuccessful transaction"
                        + invalidReason.map(ir -> ", reason: " + ir).orElse(""));
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
                        // if with the low estimation gas is successful the return this
                        // estimation
                        if (lr.isSuccessful()) {
                          log.trace(
                              "Low gas estimation {} successful, call params {}",
                              lowGasEstimation,
                              callParameters);
                          return lowGasEstimation;
                        } else {
                          log.trace(
                              "Low gas estimation {} unsuccessful, result{}, call params {}",
                              lowGasEstimation,
                              lr.result(),
                              callParameters);

                          // else do a binary search to find the right estimation
                          var high = highGasEstimation(lr.getGasEstimate(), tracer);
                          var mid = high;
                          var low = lowGasEstimation;
                          while (low + 1 < high) {
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
                              low = mid;
                              log.atTrace()
                                  .setMessage(
                                      "Binary gas estimation search low={},med={},high={}, unsuccessful result {}, call params {}")
                                  .addArgument(lowGasEstimation)
                                  .addArgument(
                                      () ->
                                          binarySearchResult
                                              .map(result -> result.result().toString())
                                              .orElse("empty"))
                                  .addArgument(callParameters)
                                  .log();

                            } else {
                              high = mid;
                              log.trace(
                                  "Binary gas estimation search low={},med={},high={}, successful, call params {}",
                                  lowGasEstimation,
                                  callParameters);
                            }
                          }
                          return high;
                        }
                      })
                  .orElseThrow();
            })
        .orElseThrow();
  }

  private JsonCallParameter parseRequest(final Object[] params) {
    final var callParameters = parameterParser.required(params, 0, JsonCallParameter.class);
    validateParameters(callParameters);
    return callParameters;
  }

  private void validateParameters(final JsonCallParameter callParameters) {
    if (callParameters.getGasPrice() != null
        && (callParameters.getMaxFeePerGas().isPresent()
            || callParameters.getMaxPriorityFeePerGas().isPresent())) {
      throw new InvalidJsonRpcParameters(
          "gasPrice cannot be used with maxFeePerGas or maxPriorityFeePerGas");
    }

    if (callParameters.getGasLimit() > 0
        && callParameters.getGasLimit() > txValidatorConf.maxTxGasLimit()) {
      throw new InvalidJsonRpcParameters("gasLimit above maximum");
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
}
