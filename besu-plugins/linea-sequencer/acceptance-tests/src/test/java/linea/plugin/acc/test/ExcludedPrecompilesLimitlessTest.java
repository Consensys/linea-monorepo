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

import java.io.IOException;
import java.math.BigInteger;
import java.nio.charset.StandardCharsets;
import java.util.Arrays;
import java.util.List;
import linea.plugin.acc.test.tests.web3j.generated.ExcludedPrecompiles;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.tests.acceptance.dsl.account.Account;
import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts;
import org.junit.jupiter.api.Test;
import org.web3j.abi.datatypes.generated.Bytes8;
import org.web3j.crypto.Credentials;
import org.web3j.crypto.Hash;
import org.web3j.crypto.RawTransaction;
import org.web3j.crypto.TransactionEncoder;
import org.web3j.protocol.Web3j;
import org.web3j.protocol.core.methods.response.EthSendTransaction;
import org.web3j.tx.gas.DefaultGasProvider;
import org.web3j.utils.Numeric;

public class ExcludedPrecompilesLimitlessTest extends LineaPluginTestBase {
  private static final BigInteger GAS_LIMIT = DefaultGasProvider.GAS_LIMIT;
  private static final BigInteger GAS_PRICE = DefaultGasProvider.GAS_PRICE;

  @Override
  public List<String> getTestCliOptions() {
    return new TestCommandLineOptionsBuilder()
        // disable line count validation to accept excluded precompile txs in the txpool
        .set("--plugin-linea-tx-pool-simulation-check-api-enabled=", "false")
        // set the module limits file
        .set(
            "--plugin-linea-module-limit-file-path=",
            getResourcePath("/moduleLimitsLimitless.toml"))
        // enabled the ZkCounter
        .set("--plugin-linea-limitless-enabled=", "true")
        .build();
  }

  @Test
  public void transactionsWithExcludedPrecompilesAreNotAccepted() throws Exception {
    final ExcludedPrecompiles excludedPrecompiles = deployExcludedPrecompiles();
    final Web3j web3j = minerNode.nodeRequests().eth();
    final String contractAddress = excludedPrecompiles.getContractAddress();

    // fund a new account
    final var recipient = accounts.createAccount("recipient");
    final var txHashFundRecipient =
        accountTransactions
            .createTransfer(accounts.getPrimaryBenefactor(), recipient, 10, BigInteger.valueOf(1))
            .execute(minerNode.nodeRequests());
    minerNode.verify(eth.expectSuccessfulTransactionReceipt(txHashFundRecipient.toHexString()));

    record InvalidCall(
        String senderPrivateKey, int nonce, String encodedContractCall, String expectedTraceLog) {}

    final InvalidCall[] invalidCalls = {
      new InvalidCall(
          Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY,
          2,
          excludedPrecompiles
              .callRIPEMD160("I am not allowed here".getBytes(StandardCharsets.UTF_8))
              .encodeFunctionCall(),
          "Tx 0xe4648fd59d4289e59b112bf60931336440d306c85c2aac5a8b0c64ab35bc55b7 line count per module: [RIP=2147483647/2147483647/1,BLAKE=0/0/1,BLOCK_L1_SIZE=235/365/1000000,MODEXP=0/0/1,BLOCK_L2_L1_LOGS=0/0/16"),
      new InvalidCall(
          Accounts.GENESIS_ACCOUNT_TWO_PRIVATE_KEY,
          0,
          encodedCallBlake2F(excludedPrecompiles),
          "Tx 0x9f457b1b5244b03c54234f7f9e8225d4253135dd3c99a46dc527d115e7ea5dac line count per module: [RIP=0/0/1,BLAKE=2147483647/2147483647/1,BLOCK_L1_SIZE=462/592/1000000,MODEXP=0/0/1,BLOCK_L2_L1_LOGS=0/0/16]")
    };

    final var invalidTxHashes =
        Arrays.stream(invalidCalls)
            .map(
                invalidCall -> {
                  // this tx must not be accepted but not mined
                  final RawTransaction txInvalid =
                      RawTransaction.createTransaction(
                          CHAIN_ID,
                          BigInteger.valueOf(invalidCall.nonce),
                          GAS_LIMIT.divide(BigInteger.TEN),
                          contractAddress,
                          BigInteger.ZERO,
                          invalidCall.encodedContractCall,
                          GAS_PRICE,
                          GAS_PRICE);

                  final byte[] signedTxInvalid =
                      TransactionEncoder.signMessage(
                          txInvalid, Credentials.create(invalidCall.senderPrivateKey));

                  final EthSendTransaction signedTxInvalidResp;
                  try {
                    signedTxInvalidResp =
                        web3j.ethSendRawTransaction(Numeric.toHexString(signedTxInvalid)).send();
                  } catch (IOException e) {
                    throw new RuntimeException(e);
                  }

                  assertThat(signedTxInvalidResp.hasError()).isFalse();
                  return signedTxInvalidResp.getTransactionHash();
                })
            .toList();

    assertThat(getTxPoolContent()).hasSize(invalidTxHashes.size());

    // transfer used as sentry to ensure a new block is mined without the invalid txs
    final var transferTxHash1 =
        accountTransactions
            .createTransfer(recipient, accounts.getSecondaryBenefactor(), 1)
            .execute(minerNode.nodeRequests());

    // first sentry is mined and no tx of the bundle is mined
    minerNode.verify(eth.expectSuccessfulTransactionReceipt(transferTxHash1.toHexString()));
    Arrays.stream(invalidCalls)
        .forEach(
            invalidCall ->
                minerNode.verify(
                    eth.expectNoTransactionReceipt(Hash.sha3(invalidCall.encodedContractCall))));

    final String log = getLog();
    // verify trace log contains the exclusion cause
    Arrays.stream(invalidCalls)
        .forEach(invalidCall -> assertThat(log).contains(invalidCall.expectedTraceLog));
  }

  @Test
  public void invalidModExpCallsAreNotMined() throws Exception {
    final var modExp = deployModExp();

    final var modExpSenders = new Account[3];
    final var foundTxHashes = new String[3];
    for (int i = 0; i < 3; i++) {
      modExpSenders[i] = accounts.createAccount("sender" + i);
      foundTxHashes[i] =
          accountTransactions
              .createTransfer(
                  accounts.getSecondaryBenefactor(), modExpSenders[i], 1, BigInteger.valueOf(i))
              .execute(minerNode.nodeRequests())
              .toHexString();
    }
    Arrays.stream(foundTxHashes)
        .forEach(
            fundTxHash -> minerNode.verify(eth.expectSuccessfulTransactionReceipt(fundTxHash)));

    final Bytes[][] invalidInputs = {
      {Bytes.fromHexString("0000000000000000000000000000000000000000000000000000000000000201")},
      {
        Bytes.fromHexString("00000000000000000000000000000000000000000000000000000000000003"),
        Bytes.fromHexString("ff")
      },
      {
        Bytes.fromHexString("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"),
        Bytes.fromHexString("00000000000000000000000000000000000000000000000000000000000003"),
        Bytes.fromHexString("ff")
      }
    };

    for (int i = 0; i < invalidInputs.length; i++) {
      final var invalidCallTxHashes = new String[invalidInputs[i].length];
      for (int j = 0; j < invalidInputs[i].length; j++) {

        // use always the same nonce since we expect this tx not to be mined
        final var mulmodOverflow =
            encodedCallModExp(modExp, modExpSenders[j], 0, invalidInputs[i][j]);

        final Web3j web3j = minerNode.nodeRequests().eth();
        final EthSendTransaction resp =
            web3j.ethSendRawTransaction(Numeric.toHexString(mulmodOverflow)).send();
        invalidCallTxHashes[j] = resp.getTransactionHash();
      }

      // transfer used as sentry to ensure a new block is mined without the invalid modexp call
      final var transferTxHash =
          accountTransactions
              .createTransfer(
                  accounts.getPrimaryBenefactor(),
                  accounts.getSecondaryBenefactor(),
                  1,
                  BigInteger.valueOf(i + 1))
              .execute(minerNode.nodeRequests());

      // sentry is mined and the invalid modexp txs are not
      minerNode.verify(eth.expectSuccessfulTransactionReceipt(transferTxHash.toHexString()));
      final var blockLog = getAndResetLog();
      Arrays.stream(invalidCallTxHashes)
          .forEach(
              invalidCallTxHash -> {
                minerNode.verify(eth.expectNoTransactionReceipt(invalidCallTxHash));
                assertThat(blockLog)
                    .contains(
                        "Tx "
                            + invalidCallTxHash
                            + " line count for module MODEXP=2147483647 is above the limit 1, removing from the txpool");
              });
    }
  }

  private String encodedCallBlake2F(final ExcludedPrecompiles excludedPrecompiles) {
    return excludedPrecompiles
        .callBlake2f(
            BigInteger.valueOf(12),
            List.of(
                Bytes32.fromHexString(
                        "0x48c9bdf267e6096a3ba7ca8485ae67bb2bf894fe72f36e3cf1361d5f3af54fa5")
                    .toArrayUnsafe(),
                Bytes32.fromHexString(
                        "0xd182e6ad7f520e511f6c3e2b8c68059b6bbd41fbabd9831f79217e1319cde05b")
                    .toArrayUnsafe()),
            List.of(
                Bytes32.fromHexString(
                        "0x6162630000000000000000000000000000000000000000000000000000000000")
                    .toArrayUnsafe(),
                Bytes32.ZERO.toArrayUnsafe(),
                Bytes32.ZERO.toArrayUnsafe(),
                Bytes32.ZERO.toArrayUnsafe()),
            List.of(Bytes8.DEFAULT.getValue(), Bytes8.DEFAULT.getValue()),
            true)
        .encodeFunctionCall();
  }
}
