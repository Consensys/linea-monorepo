/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */

package org.hyperledger.besu.tests.acceptance.dsl;

import static java.nio.charset.StandardCharsets.UTF_8;
import static org.assertj.core.api.Assertions.assertThat;
import static org.awaitility.Awaitility.await;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.lang.ProcessBuilder.Redirect;
import java.math.BigInteger;
import java.time.Duration;
import java.util.List;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.stream.IntStream;
import lombok.extern.slf4j.Slf4j;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.api.util.DomainObjectDecodeUtils;
import org.hyperledger.besu.tests.acceptance.dsl.account.Account;
import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts;
import org.hyperledger.besu.tests.acceptance.dsl.blockchain.Blockchain;
import org.hyperledger.besu.tests.acceptance.dsl.condition.admin.AdminConditions;
import org.hyperledger.besu.tests.acceptance.dsl.condition.bft.BftConditions;
import org.hyperledger.besu.tests.acceptance.dsl.condition.clique.CliqueConditions;
import org.hyperledger.besu.tests.acceptance.dsl.condition.eth.EthConditions;
import org.hyperledger.besu.tests.acceptance.dsl.condition.login.LoginConditions;
import org.hyperledger.besu.tests.acceptance.dsl.condition.net.NetConditions;
import org.hyperledger.besu.tests.acceptance.dsl.condition.perm.PermissioningConditions;
import org.hyperledger.besu.tests.acceptance.dsl.condition.process.ExitedWithCode;
import org.hyperledger.besu.tests.acceptance.dsl.condition.txpool.TxPoolConditions;
import org.hyperledger.besu.tests.acceptance.dsl.condition.web3.Web3Conditions;
import org.hyperledger.besu.tests.acceptance.dsl.contract.ContractVerifier;
import org.hyperledger.besu.tests.acceptance.dsl.node.BesuNode;
import org.hyperledger.besu.tests.acceptance.dsl.node.Node;
import org.hyperledger.besu.tests.acceptance.dsl.node.cluster.Cluster;
import org.hyperledger.besu.tests.acceptance.dsl.node.configuration.BesuNodeFactory;
import org.hyperledger.besu.tests.acceptance.dsl.node.configuration.permissioning.PermissionedNodeBuilder;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.account.AccountTransactions;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.admin.AdminTransactions;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.bft.BftTransactions;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.clique.CliqueTransactions;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.contract.ContractTransactions;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.eth.EthTransactions;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.miner.MinerTransactions;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.net.NetTransactions;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.perm.PermissioningTransactions;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.txpool.TxPoolTransactions;
import org.hyperledger.besu.tests.acceptance.dsl.transaction.web3.Web3Transactions;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.extension.ExtendWith;
import org.web3j.protocol.core.DefaultBlockParameter;

/** Base class for acceptance tests. */
@ExtendWith(AcceptanceTestBaseTestWatcher.class)
@Tag("AcceptanceTest")
@Slf4j
public abstract class AcceptanceTestBase {
  protected final Accounts accounts;
  protected final AccountTransactions accountTransactions;
  protected final AdminConditions admin;
  protected final AdminTransactions adminTransactions;
  protected final Blockchain blockchain;
  protected final CliqueConditions clique;
  protected final CliqueTransactions cliqueTransactions;
  protected final Cluster cluster;
  protected final ContractVerifier contractVerifier;
  protected final ContractTransactions contractTransactions;
  protected final EthConditions eth;
  protected final EthTransactions ethTransactions;
  protected final BftTransactions bftTransactions;
  protected final BftConditions bft;
  protected final LoginConditions login;
  protected final NetConditions net;
  protected final BesuNodeFactory besu;
  protected final PermissioningConditions perm;
  protected final PermissionedNodeBuilder permissionedNodeBuilder;
  protected final PermissioningTransactions permissioningTransactions;
  protected final MinerTransactions minerTransactions;
  protected final Web3Conditions web3;
  protected final TxPoolConditions txPoolConditions;
  protected final TxPoolTransactions txPoolTransactions;
  protected final ExitedWithCode exitedSuccessfully;

  private final ExecutorService outputProcessorExecutor = Executors.newCachedThreadPool();
  protected BesuNode minerNode;

  protected AcceptanceTestBase() {
    ethTransactions = new EthTransactions();
    accounts = new Accounts(ethTransactions);
    adminTransactions = new AdminTransactions();
    cliqueTransactions = new CliqueTransactions();
    bftTransactions = new BftTransactions();
    accountTransactions = new AccountTransactions(accounts);
    permissioningTransactions = new PermissioningTransactions();
    contractTransactions = new ContractTransactions();
    minerTransactions = new MinerTransactions();

    blockchain = new Blockchain(ethTransactions);
    clique = new CliqueConditions(ethTransactions, cliqueTransactions);
    eth = new EthConditions(ethTransactions);
    bft = new BftConditions(bftTransactions);
    login = new LoginConditions();
    net = new NetConditions(new NetTransactions());
    cluster = new Cluster(net);
    perm = new PermissioningConditions(permissioningTransactions);
    admin = new AdminConditions(adminTransactions);
    web3 = new Web3Conditions(new Web3Transactions());
    besu = new BesuNodeFactory();
    txPoolTransactions = new TxPoolTransactions();
    txPoolConditions = new TxPoolConditions(txPoolTransactions);
    contractVerifier = new ContractVerifier(accounts.getPrimaryBenefactor());
    permissionedNodeBuilder = new PermissionedNodeBuilder();
    exitedSuccessfully = new ExitedWithCode(0);
  }

  @AfterEach
  public void tearDownAcceptanceTestBase() {
    reportMemory();
    cluster.close();
  }

  /** Report memory usage after test execution. */
  public void reportMemory() {
    String os = System.getProperty("os.name");
    String[] command = null;
    if (os.contains("Linux")) {
      command = new String[] {"/usr/bin/top", "-n", "1", "-o", "%MEM", "-b", "-c", "-w", "180"};
    }
    if (os.contains("Mac")) {
      command = new String[] {"/usr/bin/top", "-l", "1", "-o", "mem", "-n", "20"};
    }
    if (command != null) {
      log.info("Memory usage at end of test:");
      final ProcessBuilder processBuilder =
          new ProcessBuilder(command).redirectErrorStream(true).redirectInput(Redirect.INHERIT);
      try {
        final Process memInfoProcess = processBuilder.start();
        outputProcessorExecutor.execute(() -> printOutput(memInfoProcess));
        memInfoProcess.waitFor();
        log.debug("Memory info process exited with code {}", memInfoProcess.exitValue());
      } catch (final Exception e) {
        log.warn("Error running memory information process", e);
      }
    } else {
      log.info("Don't know how to report memory for OS {}", os);
    }
  }

  private void printOutput(final Process process) {
    try (final BufferedReader in =
        new BufferedReader(new InputStreamReader(process.getInputStream(), UTF_8))) {
      String line = in.readLine();
      while (line != null) {
        log.info(line);
        line = in.readLine();
      }
    } catch (final IOException e) {
      log.warn("Failed to read output from memory information process: ", e);
    }
  }

  protected void waitForBlockHeight(final Node node, final long blockchainHeight) {
    WaitUtils.waitFor(
        120,
        () ->
            assertThat(node.execute(ethTransactions.blockNumber()))
                .isGreaterThanOrEqualTo(BigInteger.valueOf(blockchainHeight)));
  }

  protected List<Account> createAccounts(int numAccounts, int initialBalanceEther) {
    return createAccountsBulk(numAccounts, initialBalanceEther);
  }

  private List<Account> createAccountsBulk(int numAccounts, int initialBalanceEther) {
    try {
      final var newAccounts =
          IntStream.rangeClosed(1, numAccounts)
              .mapToObj(i -> accounts.createAccount("Account" + i))
              .toList();

      final var chainId =
          minerNode.nodeRequests().eth().ethChainId().send().getChainId().longValue();
      final var senderAccount = accounts.getPrimaryBenefactor();
      final AtomicInteger fundingAccNonce =
          new AtomicInteger(
              minerNode
                  .nodeRequests()
                  .eth()
                  .ethGetTransactionCount(
                      senderAccount.getAddress(), DefaultBlockParameter.valueOf("LATEST"))
                  .send()
                  .getTransactionCount()
                  .intValue());
      final var founderBalance =
          ethTransactions.getBalance(senderAccount).execute(minerNode.nodeRequests());
      final var requiredBalance =
          Wei.fromEth((long) initialBalanceEther * numAccounts)
              .getAsBigInteger()
              .add(Wei.of(3_000_000_000L * numAccounts).getAsBigInteger());
      System.out.println("balance: " + Wei.of(founderBalance).getAsBigInteger());
      System.out.println("targetBalance: " + requiredBalance);
      assertThat(founderBalance).isGreaterThanOrEqualTo(requiredBalance);

      // if (fundingAccNonce.incrementAndGet() > 0) {
      //   System.exit(0);
      // }
      // fund the new accounts
      final var transfers =
          newAccounts.stream()
              .map(
                  account -> {
                    // final var tx =
                    // new TransferTransactionBuilder()
                    //     .sender(senderAccount)
                    //     .recipient(account)
                    //     .amount(Amount.ether(initialBalanceEther))
                    //     .transactionType(TransactionType.EIP1559)
                    //     .gasPrice(Amount.wei(BigInteger.valueOf(2_000_000_000L)))
                    //     .chainId(chainId)
                    //     .transactionType(TransactionType.FRONTIER)
                    //     .nonce(BigInteger.valueOf(fundingAccNonce.incrementAndGet()))
                    //     .build();

                    final var tx =
                        accountTransactions.createTransfer(
                            senderAccount,
                            account,
                            initialBalanceEther,
                            /*nonce*/ BigInteger.valueOf(fundingAccNonce.incrementAndGet()));

                    final var decodedTx =
                        DomainObjectDecodeUtils.decodeRawTransaction(tx.signedTransactionData());
                    System.out.println(
                        "----TRANSFER_TRANSACTION " + fundingAccNonce.get() + "----");
                    System.out.println("tx hash=" + tx.transactionHash());
                    System.out.println("Decoded tx hash=" + decodedTx.getHash().toHexString());
                    System.out.println("Sender: " + senderAccount.getAddress());
                    System.out.println("Recipient: " + account.getAddress());
                    System.out.println("Decoded Sender: " + decodedTx.getSender().toString());
                    System.out.println("Decoded Recipient: " + decodedTx.getTo().get().toString());
                    final var nodeTxHash = tx.execute(minerNode.nodeRequests());
                    System.out.println("EL Node tx hash: " + nodeTxHash.toHexString());
                    assertThat(decodedTx.getSender().toString())
                        .isEqualTo(senderAccount.getAddress());
                    return nodeTxHash;
                  })
              .toList();

      // transfers.forEach((Hash tx) -> {
      //   await()
      //     .atMost(Duration.ofMinutes(3))
      //     .pollInterval(Duration.ofMillis(1000))
      //     .untilAsserted(
      //       () -> {
      //         minerNode.verify(eth.expectSuccessfulTransactionReceipt(tx.toString()));
      //       });
      // });

      newAccounts.forEach(
          account -> {
            System.out.println("Waiting for account " + account.getAddress() + " to be funded");
            assertThatAddressHasBalance(
                account.getAddress(),
                Wei.fromEth(initialBalanceEther).getAsBigInteger(),
                Duration.ofMinutes(3));
          });

      return newAccounts;
    } catch (IOException e) {
      throw new RuntimeException(e);
    }
  }

  public void assertThatAddressHasBalance(
      String address, BigInteger expectedBalanceWei, Duration timeout) {
    await()
        .atMost(timeout)
        .pollInterval(Duration.ofMillis(1000))
        .untilAsserted(
            () -> {
              final var balance =
                  minerNode
                      .nodeRequests()
                      .eth()
                      .ethGetBalance(address, DefaultBlockParameter.valueOf("LATEST"))
                      .send()
                      .getBalance();
              assertThat(balance).isGreaterThan(expectedBalanceWei);
            });
  }
}
