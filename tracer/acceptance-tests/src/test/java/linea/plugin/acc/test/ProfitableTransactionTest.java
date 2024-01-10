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
package linea.plugin.acc.test;

import java.math.BigInteger;
import java.util.List;

import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.tests.acceptance.dsl.account.Account;
import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.account.TransferTransaction;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.web3j.crypto.Credentials;
import org.web3j.protocol.Web3j;
import org.web3j.tx.RawTransactionManager;
import org.web3j.tx.TransactionManager;

public class ProfitableTransactionTest extends LineaPluginTestBase {
  private static final int VERIFICATION_GAS_COST = 1_200_000;
  private static final int VERIFICATION_CAPACITY = 90_000;
  private static final int GAS_PRICE_RATIO = 15;
  private static final double MIN_MARGIN = 1.0;
  private static final Wei MIN_GAS_PRICE = Wei.of(1_000_000_000);

  @Override
  public List<String> getTestCliOptions() {
    return new TestCommandLineOptionsBuilder()
        .set("--plugin-linea-verification-gas-cost=", String.valueOf(VERIFICATION_GAS_COST))
        .set("--plugin-linea-verification-capacity=", String.valueOf(VERIFICATION_CAPACITY))
        .set("--plugin-linea-gas-price-ratio=", String.valueOf(GAS_PRICE_RATIO))
        .set("--plugin-linea-min-margin=", String.valueOf(MIN_MARGIN))
        .build();
  }

  @BeforeEach
  public void setMinGasPrice() {
    minerNode.getMiningParameters().setMinTransactionGasPrice(MIN_GAS_PRICE);
  }

  @Test
  public void transactionIsNotMinedWhenUnprofitable() throws Exception {

    final Web3j web3j = minerNode.nodeRequests().eth();
    final Credentials credentials = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY);
    TransactionManager txManager = new RawTransactionManager(web3j, credentials, CHAIN_ID);

    final String txData = "not profitable transaction".repeat(10);

    final var txUnprofitable =
        txManager.sendTransaction(
            MIN_GAS_PRICE.getAsBigInteger(),
            BigInteger.valueOf(MAX_TX_GAS_LIMIT / 2),
            credentials.getAddress(),
            txData,
            BigInteger.ZERO);

    final Account sender = accounts.getSecondaryBenefactor();
    final Account recipient = accounts.createAccount("recipient");
    final TransferTransaction transferTx = accountTransactions.createTransfer(sender, recipient, 1);
    final var txHash = minerNode.execute(transferTx);

    minerNode.verify(eth.expectSuccessfulTransactionReceipt(txHash.toHexString()));

    // assert that tx below margin is not confirmed
    minerNode.verify(eth.expectNoTransactionReceipt(txUnprofitable.getTransactionHash()));
  }

  /**
   * if we have a list of transactions [t_small, t_tooBig, t_small, ..., t_small] where t_tooBig is
   * too big to fit in a block, we have blocks created that contain all t_small transactions.
   *
   * @throws Exception if send transaction fails
   */
  @Test
  public void transactionIsMinedWhenProfitable() {
    minerNode.getMiningParameters().setMinTransactionGasPrice(MIN_GAS_PRICE);
    final Account sender = accounts.getSecondaryBenefactor();
    final Account recipient = accounts.createAccount("recipient");

    final TransferTransaction transferTx = accountTransactions.createTransfer(sender, recipient, 1);
    final var txHash = minerNode.execute(transferTx);

    minerNode.verify(eth.expectSuccessfulTransactionReceipt(txHash.toHexString()));
  }
}
