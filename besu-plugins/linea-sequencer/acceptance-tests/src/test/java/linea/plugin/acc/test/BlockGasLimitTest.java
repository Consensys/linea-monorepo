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

import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.web3j.crypto.Credentials;
import org.web3j.protocol.Web3j;
import org.web3j.tx.RawTransactionManager;
import org.web3j.tx.TransactionManager;

/** This class tests the block gas limit functionality of the plugin. */
public class BlockGasLimitTest extends LineaPluginTestBase {

  private static final BigInteger GAS_PRICE = BigInteger.TEN.pow(9);
  private static final BigInteger VALUE = BigInteger.TWO;

  @Override
  public List<String> getTestCliOptions() {
    return new TestCommandLineOptionsBuilder()
        .set("--plugin-linea-max-block-gas=", "300000")
        .build();
  }

  @Override
  @BeforeEach
  public void setup() throws Exception {
    minerNode = besu.createMinerNodeWithExtraCliOptions("miner1", getTestCliOptions());
    cluster.start(minerNode);
    minerNode.execute(minerTransactions.minerStop());
  }

  /**
   * if we have a list of transactions [t_0.3, t_0.3, t_0.66, t_0.4], just two blocks are created,
   * where t_x fills X% of a limit.
   */
  @Test
  public void multipleBlocksFilledRespectingUserBlockGasLimit() throws Exception {
    final Web3j web3j = minerNode.nodeRequests().eth();
    final Credentials credentials1 = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY);
    TransactionManager txManager1 = new RawTransactionManager(web3j, credentials1, CHAIN_ID);
    final Credentials credentials2 = Credentials.create(Accounts.GENESIS_ACCOUNT_TWO_PRIVATE_KEY);
    TransactionManager txManager2 = new RawTransactionManager(web3j, credentials2, CHAIN_ID);

    final var tx100kGas1 =
        txManager1.sendTransaction(
            GAS_PRICE,
            BigInteger.valueOf(MAX_TX_GAS_LIMIT).divide(BigInteger.TEN),
            accounts.getSecondaryBenefactor().getAddress(),
            "a".repeat(10000),
            VALUE);

    final var tx100kGas2 =
        txManager1.sendTransaction(
            GAS_PRICE.multiply(BigInteger.TWO),
            BigInteger.valueOf(MAX_TX_GAS_LIMIT).divide(BigInteger.TEN),
            accounts.getSecondaryBenefactor().getAddress(),
            "b".repeat(10000),
            VALUE);

    final var tx200kGas =
        txManager2.sendTransaction(
            GAS_PRICE.multiply(BigInteger.TEN),
            BigInteger.valueOf(MAX_TX_GAS_LIMIT).divide(BigInteger.TEN),
            accounts.getPrimaryBenefactor().getAddress(),
            "c".repeat(20000),
            VALUE);

    final var tx125kGas =
        txManager1.sendTransaction(
            GAS_PRICE.multiply(BigInteger.TWO),
            BigInteger.valueOf(MAX_TX_GAS_LIMIT).divide(BigInteger.TEN),
            accounts.getSecondaryBenefactor().getAddress(),
            "d".repeat(12500),
            VALUE);

    startMining();

    assertTransactionsMinedInSameBlock(
        web3j, List.of(tx100kGas1.getTransactionHash(), tx200kGas.getTransactionHash()));
    assertTransactionsMinedInSameBlock(
        web3j, List.of(tx100kGas2.getTransactionHash(), tx125kGas.getTransactionHash()));
  }

  private void startMining() {
    minerNode.execute(minerTransactions.minerStart());
  }
}
