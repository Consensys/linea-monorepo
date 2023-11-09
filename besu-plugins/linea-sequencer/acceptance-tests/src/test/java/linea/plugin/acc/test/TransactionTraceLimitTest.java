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
import java.util.ArrayList;
import java.util.List;

import linea.plugin.acc.test.tests.web3j.generated.SimpleStorage;
import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts;
import org.junit.jupiter.api.Test;
import org.web3j.crypto.Credentials;
import org.web3j.crypto.RawTransaction;
import org.web3j.crypto.TransactionEncoder;
import org.web3j.protocol.Web3j;
import org.web3j.protocol.core.methods.response.EthSendTransaction;
import org.web3j.tx.gas.DefaultGasProvider;
import org.web3j.utils.Numeric;

public class TransactionTraceLimitTest extends LineaPluginTestBase {

  private static final BigInteger GAS_LIMIT = DefaultGasProvider.GAS_LIMIT;
  private static final BigInteger VALUE = BigInteger.ZERO;
  private static final BigInteger GAS_PRICE = BigInteger.ONE;
  private static final BigInteger NONCE = BigInteger.ONE;

  @Override
  public List<String> getTestCliOptions() {
    return new TestCommandLineOptionsBuilder()
        .set("--plugin-linea-module-limit-file-path=", getResourcePath("/moduleLimits.json"))
        .build();
  }

  @Test
  public void sendTransactions() throws Exception {
    final SimpleStorage simpleStorage = deploySimpleStorage();
    final Web3j web3j = minerNode.nodeRequests().eth();
    final String contractAddress = simpleStorage.getContractAddress();
    final Credentials credentials = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY);
    final String txData = simpleStorage.add(BigInteger.valueOf(100)).encodeFunctionCall();

    final ArrayList<String> hashes = new ArrayList<>();
    for (int i = 0; i < 5; i++) {
      final RawTransaction transaction =
          RawTransaction.createTransaction(
              CHAIN_ID,
              BigInteger.valueOf(i + 1),
              GAS_LIMIT,
              contractAddress,
              VALUE,
              txData,
              GAS_PRICE,
              NONCE);
      final byte[] signedTransaction = TransactionEncoder.signMessage(transaction, credentials);
      final EthSendTransaction response =
          web3j.ethSendRawTransaction(Numeric.toHexString(signedTransaction)).send();
      hashes.add(response.getTransactionHash());
    }

    // make sure that there are no more than one transaction per block, because the limit for the
    // add module only allows for one of these transactions.
    assertTransactionsInSeparateBlocks(web3j, hashes);
  }
}
