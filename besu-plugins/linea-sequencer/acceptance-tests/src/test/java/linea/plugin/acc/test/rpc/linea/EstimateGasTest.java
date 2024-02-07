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
  private static final Wei MIN_GAS_PRICE = Wei.of(1_000_000_000);
  public static final int MAX_TRANSACTION_GAS_LIMIT = 30_000_000;

  @Override
  public List<String> getTestCliOptions() {
    return new TestCommandLineOptionsBuilder()
        .set("--plugin-linea-verification-gas-cost=", String.valueOf(VERIFICATION_GAS_COST))
        .set("--plugin-linea-verification-capacity=", String.valueOf(VERIFICATION_CAPACITY))
        .set("--plugin-linea-gas-price-ratio=", String.valueOf(GAS_PRICE_RATIO))
        .set("--plugin-linea-min-margin=", String.valueOf(MIN_MARGIN))
        .set("--plugin-linea-max-tx-gas-limit=", String.valueOf(MAX_TRANSACTION_GAS_LIMIT))
        .build();
  }

  @BeforeEach
  public void setMinGasPrice() {
    minerNode.getMiningParameters().setMinTransactionGasPrice(MIN_GAS_PRICE);
  }

  @Test
  public void estimateGas() {

    final Account sender = accounts.getSecondaryBenefactor();
    final Account recipient = accounts.createAccount("recipient");
    final org.web3j.protocol.core.methods.request.Transaction txTransfer =
        org.web3j.protocol.core.methods.request.Transaction.createEtherTransaction(
            sender.getAddress(),
            BigInteger.ZERO,
            BigInteger.valueOf(1_000_000_000),
            BigInteger.TEN,
            recipient.getAddress(),
            BigInteger.ONE);

    final var req = new LineaEstimateGasRequest(txTransfer);
    final var resp = req.execute(minerNode.nodeRequests());
    assertThat(resp.gasLimit()).isEqualTo("0x5208");
  }

  class LineaEstimateGasRequest implements Transaction<LineaEstimateGasRequest.Response> {
    private final org.web3j.protocol.core.methods.request.Transaction transaction;

    public LineaEstimateGasRequest(
        final org.web3j.protocol.core.methods.request.Transaction transaction) {
      this.transaction = transaction;
    }

    @Override
    public LineaEstimateGasRequest.Response execute(final NodeRequests nodeRequests) {
      try {
        return new Request<>(
                "linea_estimateGas",
                List.of(transaction),
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
}
