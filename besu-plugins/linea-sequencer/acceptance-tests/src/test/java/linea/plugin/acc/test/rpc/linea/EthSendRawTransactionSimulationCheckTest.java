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
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;

import linea.plugin.acc.test.LineaPluginTestBase;
import linea.plugin.acc.test.TestCommandLineOptionsBuilder;
import linea.plugin.acc.test.tests.web3j.generated.RevertExample;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.tests.acceptance.dsl.account.Account;
import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts;
import org.junit.jupiter.api.Test;
import org.web3j.crypto.Credentials;
import org.web3j.crypto.RawTransaction;
import org.web3j.crypto.TransactionEncoder;
import org.web3j.protocol.Web3j;
import org.web3j.protocol.core.methods.response.EthSendTransaction;
import org.web3j.tx.gas.DefaultGasProvider;
import org.web3j.utils.Numeric;

public class EthSendRawTransactionSimulationCheckTest extends LineaPluginTestBase {

  private static final BigInteger GAS_LIMIT = DefaultGasProvider.GAS_LIMIT;
  private static final BigInteger VALUE = BigInteger.ZERO;
  private static final BigInteger GAS_PRICE = BigInteger.TEN.pow(9);

  @Override
  public List<String> getTestCliOptions() {
    return new TestCommandLineOptionsBuilder()
        .set(
            "--plugin-linea-module-limit-file-path=",
            getResourcePath("/moduleLimits_sendRawTx.toml"))
        .set("--plugin-linea-tx-pool-simulation-check-api-enabled=", "true")
        .build();
  }

  @Test
  public void transactionOverModuleLineCountNotAccepted() throws Exception {
    // this tx will not be accepted since it goes above the line count limit
    assertThatThrownBy(() -> deploySimpleStorage())
        .hasMessage(
            "JsonRpcError thrown with code -32000. Message: Transaction 0x63ea5c9ec0f0fae53683c1b5f1998fe806f1ad5b655a1d03a9771f1976be45a9 line count for module ROM=2369 is above the limit 2300");

    assertThat(getTxPoolContent()).isEmpty();

    // these are under the line count limit and should be accepted and selected
    final Account fewLinesSender = accounts.getSecondaryBenefactor();
    final Account recipient = accounts.createAccount("recipient");
    final List<Hash> expectedConfirmedTxs = new ArrayList<>(4);

    expectedConfirmedTxs.addAll(
        minerNode.execute(
            accountTransactions.createIncrementalTransfers(fewLinesSender, recipient, 4)));

    final var txPoolContentByHash = getTxPoolContent().stream().map(e -> e.get("hash")).toList();
    assertThat(txPoolContentByHash)
        .containsExactlyInAnyOrderElementsOf(
            expectedConfirmedTxs.stream().map(Hash::toHexString).toList());

    expectedConfirmedTxs.stream()
        .map(Hash::toHexString)
        .forEach(hash -> minerNode.verify(eth.expectSuccessfulTransactionReceipt(hash)));
  }

  @Test
  public void transactionsThatRevertAreAccepted() throws Exception {
    final RevertExample revertExample = deployRevertExample();
    final Web3j web3j = minerNode.nodeRequests().eth();
    final String contractAddress = revertExample.getContractAddress();
    final String txData = revertExample.setValue(BigInteger.ZERO).encodeFunctionCall();

    // this tx reverts but nevertheless it is accepted in the pool
    final RawTransaction txThatReverts =
        RawTransaction.createTransaction(
            CHAIN_ID,
            BigInteger.ZERO,
            GAS_LIMIT.divide(BigInteger.TEN),
            contractAddress,
            VALUE,
            txData,
            GAS_PRICE,
            GAS_PRICE.multiply(BigInteger.TEN).add(BigInteger.ONE));
    final byte[] signedTxContractInteraction =
        TransactionEncoder.signMessage(
            txThatReverts, Credentials.create(Accounts.GENESIS_ACCOUNT_TWO_PRIVATE_KEY));

    final EthSendTransaction signedTxContractInteractionResp =
        web3j.ethSendRawTransaction(Numeric.toHexString(signedTxContractInteraction)).send();

    assertThat(signedTxContractInteractionResp.hasError()).isFalse();

    final var expectedConfirmedTxHash = signedTxContractInteractionResp.getTransactionHash();

    minerNode.verify(eth.expectSuccessfulTransactionReceipt(expectedConfirmedTxHash));
  }
}
