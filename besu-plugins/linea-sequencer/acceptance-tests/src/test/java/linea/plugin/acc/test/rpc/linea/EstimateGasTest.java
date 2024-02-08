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
package linea.plugin.acc.test.rpc.linea;

import static org.assertj.core.api.Assertions.assertThat;

import java.io.IOException;
import java.math.BigInteger;
import java.util.List;

import linea.plugin.acc.test.LineaPluginTestBase;
import linea.plugin.acc.test.TestCommandLineOptionsBuilder;
import net.consensys.linea.bl.TransactionProfitabilityCalculator;
import net.consensys.linea.config.LineaTransactionSelectorCliOptions;
import net.consensys.linea.config.LineaTransactionSelectorConfiguration;
import net.consensys.linea.rpc.linea.LineaEstimateGas;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.units.bigints.UInt64;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.tests.acceptance.dsl.account.Account;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.NodeRequests;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.Transaction;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.web3j.protocol.core.Request;

public class EstimateGasTest extends LineaPluginTestBase {
  private static final int VERIFICATION_GAS_COST = 1_200_000;
  private static final int VERIFICATION_CAPACITY = 90_000;
  private static final int GAS_PRICE_RATIO = 15;
  private static final double MIN_MARGIN = 1.0;
  private static final double ESTIMATE_GAS_MIN_MARGIN = 1.0;
  private static final Wei MIN_GAS_PRICE = Wei.of(1_000_000_000);
  public static final int MAX_TRANSACTION_GAS_LIMIT = 30_000_000;
  private LineaTransactionSelectorConfiguration txSelectorConf;

  @Override
  public List<String> getTestCliOptions() {
    return new TestCommandLineOptionsBuilder()
        .set("--plugin-linea-verification-gas-cost=", String.valueOf(VERIFICATION_GAS_COST))
        .set("--plugin-linea-verification-capacity=", String.valueOf(VERIFICATION_CAPACITY))
        .set("--plugin-linea-gas-price-ratio=", String.valueOf(GAS_PRICE_RATIO))
        .set("--plugin-linea-min-margin=", String.valueOf(MIN_MARGIN))
        .set("--plugin-linea-estimate-gas-min-margin=", String.valueOf(ESTIMATE_GAS_MIN_MARGIN))
        .set("--plugin-linea-max-tx-gas-limit=", String.valueOf(MAX_TRANSACTION_GAS_LIMIT))
        .build();
  }

  @BeforeEach
  public void setMinGasPrice() {
    minerNode.getMiningParameters().setMinTransactionGasPrice(MIN_GAS_PRICE);
  }

  @BeforeEach
  public void createDefaultConfigurations() {
    txSelectorConf =
        LineaTransactionSelectorCliOptions.create().toDomainObject().toBuilder()
            .verificationCapacity(VERIFICATION_CAPACITY)
            .verificationGasCost(VERIFICATION_GAS_COST)
            .gasPriceRatio(GAS_PRICE_RATIO)
            .minMargin(MIN_MARGIN)
            .estimateGasMinMargin(ESTIMATE_GAS_MIN_MARGIN)
            .build();
  }

  @Test
  public void lineaEstimateGasMatchesEthEstimateGas() {

    final Account sender = accounts.getSecondaryBenefactor();

    final CallParams callParams = new CallParams(sender.getAddress(), null);

    final var reqEth = new RawEstimateGasRequest(callParams);
    final var reqLinea = new LineaEstimateGasRequest(callParams);
    final var respEth = reqEth.execute(minerNode.nodeRequests());
    final var respLinea = reqLinea.execute(minerNode.nodeRequests());
    assertThat(respEth).isEqualTo(respLinea.gasLimit());
  }

  @Test
  public void lineaEstimateGasIsProfitable() {

    final Account sender = accounts.getSecondaryBenefactor();

    final CallParams callParams = new CallParams(sender.getAddress(), null);

    final var reqLinea = new LineaEstimateGasRequest(callParams);
    final var respLinea = reqLinea.execute(minerNode.nodeRequests());

    final var gasLimit = UInt64.fromHexString(respLinea.gasLimit()).toLong();
    final var baseFee = Wei.fromHexString(respLinea.baseFeePerGas());
    final var priorityFee = Wei.fromHexString(respLinea.priorityFeePerGas());
    final var maxGasPrice = baseFee.add(priorityFee);

    final var tx =
        org.hyperledger.besu.ethereum.core.Transaction.builder()
            .sender(Address.fromHexString(sender.getAddress()))
            .gasLimit(gasLimit)
            .maxFeePerGas(maxGasPrice)
            .maxPriorityFeePerGas(priorityFee)
            .chainId(BigInteger.valueOf(CHAIN_ID))
            .value(Wei.ZERO)
            .payload(Bytes.EMPTY)
            .signature(LineaEstimateGas.FAKE_SIGNATURE_FOR_SIZE_CALCULATION)
            .build();

    final var profitabilityCalculator = new TransactionProfitabilityCalculator(txSelectorConf);
    assertThat(
            profitabilityCalculator.isProfitable(
                "Test",
                tx,
                minerNode
                    .getMiningParameters()
                    .getMinTransactionGasPrice()
                    .getAsBigInteger()
                    .doubleValue(),
                maxGasPrice.getAsBigInteger().doubleValue(),
                gasLimit))
        .isTrue();
  }

  class LineaEstimateGasRequest implements Transaction<LineaEstimateGasRequest.Response> {
    private final CallParams callParams;

    public LineaEstimateGasRequest(final CallParams callParams) {
      this.callParams = callParams;
    }

    @Override
    public LineaEstimateGasRequest.Response execute(final NodeRequests nodeRequests) {
      try {
        return new Request<>(
                "linea_estimateGas",
                List.of(callParams),
                nodeRequests.getWeb3jService(),
                LineaEstimateGasResponse.class)
            .send()
            .getResult();
      } catch (IOException e) {
        throw new RuntimeException(e);
      }
    }

    static class LineaEstimateGasResponse extends org.web3j.protocol.core.Response<Response> {}

    record Response(String gasLimit, String baseFeePerGas, String priorityFeePerGas) {}
  }

  class RawEstimateGasRequest implements Transaction<String> {
    private final CallParams callParams;

    public RawEstimateGasRequest(final CallParams callParams) {
      this.callParams = callParams;
    }

    @Override
    public String execute(final NodeRequests nodeRequests) {
      try {
        return new Request<>(
                "eth_estimateGas",
                List.of(callParams),
                nodeRequests.getWeb3jService(),
                RawEstimateGasResponse.class)
            .send()
            .getResult();
      } catch (IOException e) {
        throw new RuntimeException(e);
      }
    }

    static class RawEstimateGasResponse extends org.web3j.protocol.core.Response<String> {}
  }

  record CallParams(String from, String value) {}
}
