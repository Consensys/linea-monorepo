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

import static org.assertj.core.api.Assertions.assertThat;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;

import linea.plugin.acc.test.tests.web3j.generated.SimpleStorage;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.tests.acceptance.dsl.account.Account;
import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts;
import org.junit.jupiter.api.Test;
import org.web3j.crypto.Credentials;
import org.web3j.protocol.Web3j;
import org.web3j.tx.RawTransactionManager;
import org.web3j.tx.TransactionManager;
import org.web3j.tx.gas.DefaultGasProvider;

public class TransactionGasLimitTest extends LineaPluginTestBase {

  public static final int MAX_TX_GAS_LIMIT = DefaultGasProvider.GAS_LIMIT.intValue();
  private static final BigInteger GAS_PRICE = DefaultGasProvider.GAS_PRICE;
  private static final BigInteger VALUE = BigInteger.ZERO;

  @Override
  public List<String> getTestCliOptions() {
    return new TestCommandLineOptionsBuilder()
        .set("--plugin-linea-max-tx-gas-limit=", String.valueOf(MAX_TX_GAS_LIMIT))
        .build();
  }

  @Test
  public void transactionIsMinedWhenGasLimitIsNotExceeded() throws Exception {
    final SimpleStorage simpleStorage = deploySimpleStorage();

    final Web3j web3j = minerNode.nodeRequests().eth();
    final String contractAddress = simpleStorage.getContractAddress();
    final Credentials credentials = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY);
    TransactionManager txManager = new RawTransactionManager(web3j, credentials, CHAIN_ID);

    final String txData = simpleStorage.set("hello").encodeFunctionCall();

    final String hashGood =
        txManager
            .sendTransaction(
                GAS_PRICE, BigInteger.valueOf(MAX_TX_GAS_LIMIT), contractAddress, txData, VALUE)
            .getTransactionHash();

    // make sure that a transaction that is not too big was mined
    minerNode.verify(eth.expectSuccessfulTransactionReceipt(hashGood));
  }

  @Test
  public void transactionIsNotMinedWhenGasLimitIsExceeded() throws Exception {
    final SimpleStorage simpleStorage = deploySimpleStorage();

    final Web3j web3j = minerNode.nodeRequests().eth();
    final String contractAddress = simpleStorage.getContractAddress();
    final Credentials credentials = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY);
    TransactionManager txManager = new RawTransactionManager(web3j, credentials, CHAIN_ID);

    final String txData = simpleStorage.set("hello").encodeFunctionCall();

    final var txTooBigResp =
        txManager.sendTransaction(
            GAS_PRICE, BigInteger.valueOf(MAX_TX_GAS_LIMIT + 1), contractAddress, txData, VALUE);

    assertThat(txTooBigResp.hasError()).isTrue();
    assertThat(txTooBigResp.getError().getMessage())
        .isEqualTo("Gas limit of transaction is greater than the allowed max of 9000000");
  }

  /**
   * if we have a list of transactions [t_small, t_tooBig, t_small, ..., t_small] where t_tooBig is
   * too big to fit in a block, we have blocks created that contain all t_small transactions.
   *
   * @throws Exception if send transaction fails
   */
  @Test
  public void multipleSmallTxsMinedWhileTxTooBigNot() throws Exception {
    final SimpleStorage simpleStorage = deploySimpleStorage();

    final Web3j web3j = minerNode.nodeRequests().eth();
    final String contractAddress = simpleStorage.getContractAddress();
    final Credentials credentials = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY);
    TransactionManager txManager = new RawTransactionManager(web3j, credentials, CHAIN_ID);

    final Account lowGasSender = accounts.getSecondaryBenefactor();
    final Account recipient = accounts.createAccount("recipient");

    final List<Hash> expectedConfirmedTxs = new ArrayList<>(4);

    expectedConfirmedTxs.add(
        minerNode.execute(accountTransactions.createTransfer(lowGasSender, recipient, 1)));

    final String txData = simpleStorage.set("too BIG").encodeFunctionCall();

    final var txTooBigResp =
        txManager.sendTransaction(
            GAS_PRICE, BigInteger.valueOf(MAX_TX_GAS_LIMIT + 1), contractAddress, txData, VALUE);

    expectedConfirmedTxs.addAll(
        minerNode.execute(
            accountTransactions.createIncrementalTransfers(lowGasSender, recipient, 3)));

    assertThat(txTooBigResp.hasError()).isTrue();
    assertThat(txTooBigResp.getError().getMessage())
        .isEqualTo("Gas limit of transaction is greater than the allowed max of 9000000");

    expectedConfirmedTxs.stream()
        .map(Hash::toHexString)
        .forEach(hash -> minerNode.verify(eth.expectSuccessfulTransactionReceipt(hash)));
  }
}
