/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package linea.plugin.acc.test;

import static org.assertj.core.api.Assertions.assertThat;
import static org.junit.jupiter.api.Assertions.assertThrows;

import java.io.IOException;
import java.math.BigInteger;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.StandardOpenOption;
import java.util.List;
import linea.plugin.acc.test.tests.web3j.generated.LogEmitter;
import net.consensys.linea.sequencer.txselection.selectors.TransactionEventFilter;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.awaitility.core.ConditionTimeoutException;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.log.LogTopic;
import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.NodeRequests;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.Transaction;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;
import org.web3j.crypto.Credentials;
import org.web3j.protocol.Web3j;
import org.web3j.protocol.core.Request;
import org.web3j.tx.RawTransactionManager;
import org.web3j.tx.TransactionManager;

public class TransactionEventDenialTest extends LineaPluginTestBase {

  private static final BigInteger GAS_PRICE = BigInteger.TEN.pow(9);
  private static final BigInteger GAS_LIMIT = BigInteger.valueOf(210_000);
  private static final BigInteger VALUE = BigInteger.ZERO;

  @TempDir static Path tempDir;
  static Path denyEventListPath;

  @Override
  public List<String> getTestCliOptions() {
    denyEventListPath = tempDir.resolve("denyEventList.txt");
    try {
      if (!Files.exists(denyEventListPath)) {
        Files.createFile(denyEventListPath);
      }
    } catch (IOException e) {
      throw new RuntimeException("Failed to create deny list file", e);
    }

    return new TestCommandLineOptionsBuilder()
        .set("--plugin-linea-events-deny-list-path=", denyEventListPath.toString())
        .set("--plugin-linea-events-bundle-deny-list-path=", denyEventListPath.toString())
        .build();
  }

  /** Test that a transaction emitting a denied topic is rejected. */
  @Test
  public void transactionWithDeniedTopicIsRejected() throws Exception {
    LogEmitter logEmitter = deployLogEmitter();
    Web3j web3j = minerNode.nodeRequests().eth();
    TransactionManager txManager = createTransactionManager(web3j);

    Bytes32 blockedTopic = Bytes32.fromHexStringLenient("0xaa");
    byte[] payload = "data".getBytes(StandardCharsets.UTF_8);

    addDenyListFilterAndReload(logEmitter.getContractAddress(), blockedTopic.toArray());

    String txData = logEmitter.log1(blockedTopic.toArray(), payload).encodeFunctionCall();
    var transaction =
        txManager.sendTransaction(
            GAS_PRICE, GAS_LIMIT, logEmitter.getContractAddress(), txData, VALUE);

    // Transaction should not be mined
    assertThrows(
        ConditionTimeoutException.class,
        () ->
            minerNode.verify(
                eth.expectSuccessfulTransactionReceipt(transaction.getTransactionHash())));

    // Because the transaction is denied based on its emitted event, it should not even be in the
    // pool
    assertTransactionNotInThePool(transaction.getTransactionHash());
  }

  private TransactionManager createTransactionManager(Web3j web3j) {
    Credentials credentials = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY);
    return new RawTransactionManager(web3j, credentials, CHAIN_ID);
  }

  // Helper method to add a filter to the deny list and reload configuration
  private void addDenyListFilterAndReload(String address, byte[] blockedTopic) throws IOException {
    // Add filter to deny list
    addFilterToDenyList(
        new TransactionEventFilter(
            Address.fromHexString(address),
            LogTopic.wrap(Bytes.wrap(blockedTopic)),
            null,
            null,
            null));
    reloadPluginConfiguration();
  }

  private void addFilterToDenyList(TransactionEventFilter filter) throws IOException {
    String entry =
        String.format(
            "%s,%s,%s,%s,%s",
            filter.contractAddress(),
            filter.topic0() != null ? filter.topic0() : "",
            filter.topic1() != null ? filter.topic1() : "",
            filter.topic2() != null ? filter.topic2() : "",
            filter.topic3() != null ? filter.topic3() : "");
    Files.writeString(denyEventListPath, entry + "\n", StandardOpenOption.APPEND);
  }

  private void reloadPluginConfiguration() {
    ReloadPluginConfigRequest request = new ReloadPluginConfigRequest();
    String result = request.execute(minerNode.nodeRequests());
    assertThat(result).isEqualTo("Success");
  }

  static class ReloadPluginConfigRequest implements Transaction<String> {
    @Override
    public String execute(NodeRequests nodeRequests) {
      try {
        return new Request<>(
                "plugins_reloadPluginConfig",
                List.of("net.consensys.linea.sequencer.txselection.LineaTransactionSelectorPlugin"),
                nodeRequests.getWeb3jService(),
                ReloadPluginConfigResponse.class)
            .send()
            .getResult();
      } catch (IOException e) {
        throw new RuntimeException("Failed to reload plugin configuration", e);
      }
    }
  }

  static class ReloadPluginConfigResponse extends org.web3j.protocol.core.Response<String> {}
}
