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

import com.google.common.io.Resources;
import java.io.File;
import java.io.IOException;
import java.math.BigInteger;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.util.List;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.web3j.crypto.Blob;
import org.web3j.crypto.BlobUtils;
import org.web3j.crypto.Credentials;
import org.web3j.crypto.RawTransaction;
import org.web3j.crypto.TransactionEncoder;
import org.web3j.protocol.Web3j;
import org.web3j.protocol.core.DefaultBlockParameterName;
import org.web3j.protocol.core.methods.response.EthSendTransaction;
import org.web3j.tx.gas.DefaultGasProvider;
import org.web3j.utils.Numeric;

/**
 * Tests that verify the LineaTransactionValidationPlugin correctly rejects BLOB transactions while
 * allowing other transaction types.
 */
public class BlobTransactionDenialTest extends LineaPluginTestBasePrague {
  @Override
  protected String getGenesisFileTemplatePath() {
    // We cannot use clique-prague-zero-blobs because `config.blobSchedule.prague.max = 0` will
    // block all blob txs
    return "/clique/clique-prague-one-blob.json.tpl";
  }

  @Override
  public List<String> getTestCliOptions() {
    return new TestCommandLineOptionsBuilder()
        .set("--plugin-linea-blob-tx-enabled=", "false")
        .build();
  }

  private static final BigInteger GAS_PRICE = DefaultGasProvider.GAS_PRICE;
  private static final BigInteger GAS_LIMIT = DefaultGasProvider.GAS_LIMIT;
  private static final BigInteger VALUE = BigInteger.ZERO;
  private static final String DATA = "0x";

  private Web3j web3j;
  private Credentials credentials;
  private String recipient;

  @Override
  @BeforeEach
  public void setup() throws Exception {
    super.setup();
    web3j = minerNode.nodeRequests().eth();
    credentials = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY);
    recipient = accounts.getSecondaryBenefactor().getAddress();
  }

  @Test
  public void blobTransactionsIsRejectedFromTransactionPool() throws Exception {
    // Act - Send a blob transaction to transaction pool
    EthSendTransaction response = sendRawBlobTransaction();
    this.buildNewBlock();

    // Assert
    assertThat(response.hasError()).isTrue();
    assertThat(response.getError().getMessage())
        .contains("Plugin has marked the transaction as invalid");
  }

  // TODO - Test that block import from one node to another fails for blob tx
  // TODO - Create EngineApiHelper method to import a premade block with blob tx

  private EthSendTransaction sendRawBlobTransaction() throws IOException {
    BigInteger nonce =
        web3j
            .ethGetTransactionCount(credentials.getAddress(), DefaultBlockParameterName.PENDING)
            .send()
            .getTransactionCount();

    URL blobUrl = new File(getResourcePath("/blob.txt")).toURI().toURL();
    final var blobHexString = Resources.toString(blobUrl, StandardCharsets.UTF_8);
    final Blob blob = new Blob(Numeric.hexStringToByteArray(blobHexString));
    final Bytes kzgCommitment = BlobUtils.getCommitment(blob);
    final Bytes kzgProof = BlobUtils.getProof(blob, kzgCommitment);
    final Bytes versionedHash = BlobUtils.kzgToVersionedHash(kzgCommitment);
    final RawTransaction rawTransaction =
        RawTransaction.createTransaction(
            List.of(blob),
            List.of(kzgCommitment),
            List.of(kzgProof),
            CHAIN_ID,
            nonce,
            GAS_PRICE,
            GAS_PRICE,
            GAS_LIMIT,
            recipient,
            VALUE,
            DATA,
            BigInteger.ONE,
            List.of(versionedHash));
    byte[] signedMessage = TransactionEncoder.signMessage(rawTransaction, credentials);
    String hexValue = Numeric.toHexString(signedMessage);
    return web3j.ethSendRawTransaction(hexValue).send();
  }
}
