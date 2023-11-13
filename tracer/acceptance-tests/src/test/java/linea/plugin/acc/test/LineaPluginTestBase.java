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

import static org.assertj.core.api.Assertions.*;

import java.io.IOException;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Objects;
import java.util.Set;
import java.util.stream.Collectors;

import linea.plugin.acc.test.tests.web3j.generated.SimpleStorage;
import org.apache.commons.lang3.RandomStringUtils;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.tests.acceptance.dsl.AcceptanceTestBase;
import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts;
import org.hyperledger.besu.tests.acceptance.dsl.node.BesuNode;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.web3j.crypto.Credentials;
import org.web3j.protocol.Web3j;
import org.web3j.protocol.core.RemoteCall;
import org.web3j.protocol.core.methods.response.TransactionReceipt;
import org.web3j.protocol.exceptions.TransactionException;
import org.web3j.tx.RawTransactionManager;
import org.web3j.tx.TransactionManager;
import org.web3j.tx.gas.DefaultGasProvider;
import org.web3j.tx.response.PollingTransactionReceiptProcessor;
import org.web3j.tx.response.TransactionReceiptProcessor;

/** Base class for plugin tests. */
public class LineaPluginTestBase extends AcceptanceTestBase {
  public static final int MAX_CALLDATA_SIZE = 1188; // contract has a call data size of 1160
  public static final int MAX_TX_GAS_LIMIT = DefaultGasProvider.GAS_LIMIT.intValue();
  public static final long CHAIN_ID = 1337L;
  protected BesuNode minerNode;

  @BeforeEach
  public void setup() throws Exception {
    minerNode = besu.createMinerNodeWithExtraCliOptions("miner1", getTestCliOptions());
    cluster.start(minerNode);
  }

  public List<String> getTestCliOptions() {
    return new TestCommandLineOptionsBuilder().build();
  }

  @AfterEach
  public void stop() {
    cluster.stop();
    cluster.close();
  }

  protected void sendTransactionsWithGivenLengthPayload(
      final SimpleStorage simpleStorage,
      final List<String> accounts,
      final Web3j web3j,
      final int num) {
    final String contractAddress = simpleStorage.getContractAddress();
    final String txData =
        simpleStorage.set(RandomStringUtils.randomAlphabetic(num)).encodeFunctionCall();
    final List<String> hashes = new ArrayList<>();
    accounts.forEach(
        a -> {
          final Credentials credentials = Credentials.create(a);
          TransactionManager txManager = new RawTransactionManager(web3j, credentials, CHAIN_ID);
          for (int i = 0; i < 5; i++) {
            try {
              hashes.add(
                  txManager
                      .sendTransaction(
                          DefaultGasProvider.GAS_PRICE,
                          DefaultGasProvider.GAS_LIMIT,
                          contractAddress,
                          txData,
                          BigInteger.ZERO)
                      .getTransactionHash());
            } catch (IOException e) {
              throw new RuntimeException(e);
            }
          }
        });

    assertTransactionsInCorrectBlocks(web3j, hashes, num);
  }

  private void assertTransactionsInCorrectBlocks(Web3j web3j, List<String> hashes, int num) {
    final HashMap<Long, Integer> txMap = new HashMap<>();
    TransactionReceiptProcessor receiptProcessor = createReceiptProcessor(web3j);

    // CallData for the transaction for empty String is 68 and grows in steps of 32 with (String
    // size / 32)
    final int maxTxs = MAX_CALLDATA_SIZE / (68 + ((num + 31) / 32) * 32);

    // Wait for transaction to be mined and check that there are no more than maxTxs per block
    hashes.forEach(
        h -> {
          final TransactionReceipt transactionReceipt;
          try {
            transactionReceipt = receiptProcessor.waitForTransactionReceipt(h);
          } catch (IOException | TransactionException e) {
            throw new RuntimeException(e);
          }

          final long blockNumber = transactionReceipt.getBlockNumber().longValue();
          txMap.compute(blockNumber, (b, n) -> n == null ? 1 : n + 1);

          // make sure that no block contained more than maxTxs
          assertThat(txMap.get(blockNumber)).isLessThanOrEqualTo(maxTxs);
        });
    // make sure that at least one block has maxTxs
    assertThat(txMap).containsValue(maxTxs);
  }

  protected SimpleStorage deploySimpleStorage() throws Exception {
    final Web3j web3j = minerNode.nodeRequests().eth();
    final Credentials credentials = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY);
    TransactionManager txManager = new RawTransactionManager(web3j, credentials, CHAIN_ID);

    final RemoteCall<SimpleStorage> deploy =
        SimpleStorage.deploy(web3j, txManager, new DefaultGasProvider());
    return deploy.send();
  }

  public static String getResourcePath(String resource) {
    return Objects.requireNonNull(LineaPluginTestBase.class.getResource(resource)).getPath();
  }

  protected void assertTransactionsMinedInSeparateBlocks(Web3j web3j, List<String> hashes)
      throws Exception {
    TransactionReceiptProcessor receiptProcessor = createReceiptProcessor(web3j);

    final HashSet<Long> blockNumbers = new HashSet<>();
    for (String hash : hashes) {
      TransactionReceipt receipt = receiptProcessor.waitForTransactionReceipt(hash);
      assertThat(receipt).isNotNull();
      boolean isAdded = blockNumbers.add(receipt.getBlockNumber().longValue());
      assertThat(isAdded).isEqualTo(true);
    }
  }

  protected void assertTransactionsMinedInSameBlock(Web3j web3j, List<String> hashes) {
    TransactionReceiptProcessor receiptProcessor = createReceiptProcessor(web3j);
    Set<Long> blockNumbers =
        hashes.stream()
            .map(
                hash -> {
                  try {
                    TransactionReceipt receipt = receiptProcessor.waitForTransactionReceipt(hash);
                    assertThat(receipt).isNotNull();
                    return receipt.getBlockNumber().longValue();
                  } catch (IOException | TransactionException e) {
                    throw new RuntimeException(e);
                  }
                })
            .collect(Collectors.toSet());

    assertThat(blockNumbers.size()).isEqualTo(1);
  }

  private TransactionReceiptProcessor createReceiptProcessor(Web3j web3j) {
    return new PollingTransactionReceiptProcessor(
        web3j,
        TransactionManager.DEFAULT_POLLING_FREQUENCY,
        TransactionManager.DEFAULT_POLLING_ATTEMPTS_PER_TX_HASH);
  }

  protected String sendTransactionWithGivenLengthPayload(
      final String account, final Web3j web3j, final int num) throws IOException {
    String to = Address.fromHexString("fe3b557e8fb62b89f4916b721be55ceb828dbd73").toString();
    TransactionManager txManager = new RawTransactionManager(web3j, Credentials.create(account));

    return txManager
        .sendTransaction(
            DefaultGasProvider.GAS_PRICE,
            BigInteger.valueOf(MAX_TX_GAS_LIMIT),
            to,
            RandomStringUtils.randomAlphabetic(num),
            BigInteger.ZERO)
        .getTransactionHash();
  }
}
