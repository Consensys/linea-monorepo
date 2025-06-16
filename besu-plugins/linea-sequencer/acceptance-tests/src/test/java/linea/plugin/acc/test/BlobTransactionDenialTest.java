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

import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.web3j.crypto.Credentials;
import org.web3j.protocol.Web3j;
import org.web3j.protocol.core.methods.response.EthSendTransaction;

/**
 * Tests that verify the LineaTransactionValidationPlugin correctly rejects BLOB transactions from
 * being executed
 */
public class BlobTransactionDenialTest extends LineaPluginTestBasePrague {
  private Web3j web3j;
  private Credentials credentials;
  private String recipient;

  @Override
  protected String getGenesisFileTemplatePath() {
    // We cannot use clique-prague-zero-blobs because `config.blobSchedule.prague.max = 0` will
    // block all blob txs
    return "/clique/clique-prague-one-blob.json.tpl";
  }

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
    EthSendTransaction response = sendRawBlobTransaction(web3j, credentials, recipient);
    this.buildNewBlock();

    // Assert
    assertThat(response.hasError()).isTrue();
    assertThat(response.getError().getMessage())
        .contains("Plugin has marked the transaction as invalid");
  }

  // @Test
  // public void blobTransactionsIsRejectedFromNodeImport() throws Exception {
  //   // Arrange
  //   final ObjectMapper mapper = new ObjectMapper();
  //   final ObjectNode executionPayload =
  //       mapper
  //           .createObjectNode()
  //           .put("parentHash",
  // "0xb41404ed3346dbfa0261faee076251d1746ca73ff2eaa9b0aea4d3e49956ea0d")
  //           .put("feeRecipient", "0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b")
  //           .put("stateRoot",
  // "0xdb034dcfd5b2a16f1691772acf8107f19d505a50351dcbeec01af71a49fc3ff1")
  //           .put(
  //               "logsBloom",
  //
  // "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
  //           .put("prevRandao",
  // "0x0000000000000000000000000000000000000000000000000000000000000000")
  //           .put("gasLimit", "0x1ca35ef")
  //           .put("gasUsed", "0x5208")
  //           .put("timestamp", "0x5")
  //           .put("extraData", "0x626573752032352e362e302d6c696e656131")
  //           .put("baseFeePerGas", "0x7")
  //           .put("excessBlobGas", "0x0")
  //           .put("blobGasUsed", "0x20000")
  //           .put("blockHash",
  // "0xa4e05a57d6070c4c88f3b124efed691fd52f7a219ecdfaf1ba6c7e5c97c26944")
  //           .put(
  //               "receiptsRoot",
  //               "0xeaa8c40899a61ae59615cf9985f5e2194f8fd2b57d273be63bde6733e89b12ab")
  //           .put("blockNumber", "0x1");
  //   final ArrayNode transactions = mapper.createArrayNode();
  //   transactions.add(
  //
  // "0x03f8908205398084f461090084f46109008389544094627306090abab3a6e1400e9345bc60c78a8bef578080c001e1a0010657f37554c781402a22917dee2f75def7ab966d7b770905398eba3c44401480a029e016f013bc577f372d54f84c371ae40e34fc05bb6c9968072ccca75a1f4cc4a0527d03cc62de8137e565254fd1169afe38133fb6ac7dac7134479a99f9a2d309");
  //   executionPayload.set("transactions", transactions);
  //   final ArrayNode withdrawals = mapper.createArrayNode();
  //   executionPayload.set("withdrawals", withdrawals);

  //   final String parentBeaconBlockRoot =
  //       "0x0000000000000000000000000000000000000000000000000000000000000000";

  //   final ArrayNode executionRequests = mapper.createArrayNode();
  //   //
  // executionRequests.add("0x01a4664c40aacebd82a2db79f0ea36c06bc6a19adbb10a4a15bf67b328c9b101d09e5c6ee6672978fdad9ef0d9e2ceffaee99223555d8601f0cb3bcc4ce1af9864779a416e0000000000000000");

  //   this.importPremadeBlock(executionPayload, parentBeaconBlockRoot, executionRequests);
  //   assertThat(false).isTrue();

  //   // Act
  //   // TODO: Use these values to create and send engine_newPayloadV4 request

  //   // Assert
  //   // TODO: Verify the response indicates rejection of blob transactions
  // }
}
