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

  // TODO - Consider test scenario of block import as follows:
  // 1. Start an additional minimal node with Prague config
  // 2. Ensure additional node is peered to minerNode
  // 3. Send blob tx to additional node
  // 4. Construct block on additional node
  // 5. Send 'debug_getBadBlocks' RPC request to minerNode, confirm that block is rejected from
  // import
  //
  // If the above is not possible, we can try the following:
  // 1. Create a premade block containing a blob tx
  // 2. Import the premade block using 'engine_newPayloadV4' Engine API call
}
