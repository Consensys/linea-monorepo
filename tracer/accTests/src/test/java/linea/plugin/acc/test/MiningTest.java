/*
 * Copyright ConsenSys AG.
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

import org.apache.commons.lang3.RandomStringUtils;
import org.assertj.core.api.Assertions;
import org.hyperledger.besu.tests.acceptance.dsl.AcceptanceTestBase;
import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts;
import org.hyperledger.besu.tests.acceptance.dsl.node.BesuNode;
import org.hyperledger.besu.tests.web3j.generated.SimpleStorage;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.web3j.crypto.Credentials;
import org.web3j.protocol.Web3j;
import org.web3j.protocol.core.methods.response.EthGetTransactionReceipt;
import org.web3j.protocol.core.methods.response.TransactionReceipt;
import org.web3j.protocol.exceptions.TransactionException;
import org.web3j.tx.RawTransactionManager;
import org.web3j.tx.TransactionManager;
import org.web3j.tx.gas.DefaultGasProvider;
import org.web3j.tx.response.PollingTransactionReceiptProcessor;
import org.web3j.tx.response.TransactionReceiptProcessor;

import java.io.IOException;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;

public class MiningTest extends AcceptanceTestBase {

    public static final int MAX_CALLDATA_SIZE = 1092; // contract has a call data size of 979
    private BesuNode minerNode;

    @BeforeEach
    public void setUp() throws Exception {
        final List<String> cliOptions =
                List.of(
                        "--plugin-linea-max-tx-calldata-size=" + MAX_CALLDATA_SIZE,
                        "--plugin-linea-max-block-calldata-size=" + MAX_CALLDATA_SIZE);
        minerNode = besu.createMinerNodeWithExtraCliOptions("miner1", cliOptions);
        cluster.start(minerNode);
    }

    @AfterEach
    public void stop() throws Exception {
        cluster.stop();
        cluster.close();
    }

    @Test
    public void shouldMineTransactions() {

        final SimpleStorage simpleStorage =
                minerNode.execute(contractTransactions.createSmartContract(SimpleStorage.class));
        List<String> accounts =
                List.of(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY, Accounts.GENESIS_ACCOUNT_TWO_PRIVATE_KEY);

        final Web3j web3j = minerNode.nodeRequests().eth();
        final List<Integer> numCaractersInStringList = List.of(150, 200, 400);

        numCaractersInStringList.forEach(num -> sendTransactionsWithGivenLengthPayload(simpleStorage, accounts, web3j, num));
    }

    @Test
    public void transactionIsNotMinedWhenTooBig() throws IOException, TransactionException {

        final SimpleStorage simpleStorage =
                minerNode.execute(contractTransactions.createSmartContract(SimpleStorage.class));
        final Web3j web3j = minerNode.nodeRequests().eth();
        final String contractAddress = simpleStorage.getContractAddress();
        final Credentials credentials = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY);
        TransactionManager txManager = new RawTransactionManager(web3j, credentials);

        final String txDataGood =
                simpleStorage.set(RandomStringUtils.randomAlphabetic(MAX_CALLDATA_SIZE-68)).encodeFunctionCall();
        final String hashGood = txManager.sendTransaction(
                DefaultGasProvider.GAS_PRICE,
                DefaultGasProvider.GAS_LIMIT,
                contractAddress,
                txDataGood,
                BigInteger.ZERO).getTransactionHash();

        final String txDataTooBig =
                simpleStorage.set(RandomStringUtils.randomAlphabetic(MAX_CALLDATA_SIZE-67)).encodeFunctionCall();
        final String hashTooBig = txManager.sendTransaction(
                DefaultGasProvider.GAS_PRICE,
                DefaultGasProvider.GAS_LIMIT,
                contractAddress,
                txDataTooBig,
                BigInteger.ZERO).getTransactionHash();

        TransactionReceiptProcessor receiptProcessor = new PollingTransactionReceiptProcessor(
                web3j,
                TransactionManager.DEFAULT_POLLING_FREQUENCY,
                TransactionManager.DEFAULT_POLLING_ATTEMPTS_PER_TX_HASH);

        // make sure that a transaction that is not too big was mined
        final TransactionReceipt transactionReceipt = receiptProcessor.waitForTransactionReceipt(hashGood);
        Assertions.assertThat(transactionReceipt).isNotNull();

        final EthGetTransactionReceipt receipt = web3j.ethGetTransactionReceipt(hashTooBig).send();
        Assertions.assertThat(receipt.getTransactionReceipt()).isEmpty();
    }

    private void sendTransactionsWithGivenLengthPayload(
            final SimpleStorage simpleStorage, final List<String> accounts, final Web3j web3j, final int num) {
        final String contractAddress = simpleStorage.getContractAddress();
        final String txData = simpleStorage.set(RandomStringUtils.randomAlphabetic(num)).encodeFunctionCall();
        final List<String> hashes = new ArrayList<>();
        accounts.forEach(a -> {
            final Credentials credentials = Credentials.create(a);
            TransactionManager txManager = new RawTransactionManager(web3j, credentials);
            for (int i = 0; i < 5; i++) {
                try {
                    hashes.add(txManager.sendTransaction(
                            DefaultGasProvider.GAS_PRICE,
                            DefaultGasProvider.GAS_LIMIT,
                            contractAddress,
                            txData,
                            BigInteger.ZERO).getTransactionHash());
                } catch (IOException e) {
                    throw new RuntimeException(e);
                }
            }
        });

        final HashMap<Long, Integer> txMap = new HashMap<>();
        TransactionReceiptProcessor receiptProcessor = new PollingTransactionReceiptProcessor(
                web3j,
                TransactionManager.DEFAULT_POLLING_FREQUENCY,
                TransactionManager.DEFAULT_POLLING_ATTEMPTS_PER_TX_HASH);
        // CallData for the transaction for empty String is 68 and grows in steps of 32 with (String size / 32)
        final int maxTxs = MAX_CALLDATA_SIZE / (68 + ((num + 31) / 32) * 32);

        // Wait for transaction to be mined and check that there are no more than maxTxs per block
        hashes.forEach( h -> {
            final TransactionReceipt transactionReceipt;

            try {
                transactionReceipt = receiptProcessor.waitForTransactionReceipt(h);
            } catch (IOException | TransactionException e) {
                throw new RuntimeException(e);
            }

            final long blockNumber = transactionReceipt.getBlockNumber().longValue();

            txMap.compute(blockNumber, (b, n) -> {
                if (n == null) {
                    return 1;
                }
                return n + 1;
            });

            // make sure that no block contained more than maxTxs
            Assertions.assertThat(txMap.get(blockNumber)).isLessThanOrEqualTo(maxTxs);
        });
        // make sure that at least one block has maxTxs
        Assertions.assertThat(txMap).containsValue(maxTxs);
    }
}
