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

import org.hyperledger.besu.tests.acceptance.dsl.account.Accounts;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.web3j.crypto.Credentials;
import org.web3j.protocol.Web3j;
import org.web3j.tx.RawTransactionManager;
import org.web3j.tx.TransactionManager;
import org.web3j.tx.gas.DefaultGasProvider;

/**
 * Example test using Besu node configured for Prague. Note that block building must be triggered
 * explicitly through `this.buildNewBlock()`
 */
public class ExamplePragueTest extends LineaPluginTestBasePrague {
  private static final BigInteger GAS_PRICE = DefaultGasProvider.GAS_PRICE;
  private static final BigInteger GAS_LIMIT = DefaultGasProvider.GAS_LIMIT;
  private static final BigInteger VALUE = BigInteger.ZERO;
  private static final String DATA = "0x";

  private Web3j web3j;
  private Credentials credentials;
  private TransactionManager txManager;
  private String recipient;

  @Override
  @BeforeEach
  public void setup() throws Exception {
    super.setup();
    web3j = minerNode.nodeRequests().eth();
    credentials = Credentials.create(Accounts.GENESIS_ACCOUNT_ONE_PRIVATE_KEY);
    txManager = new RawTransactionManager(web3j, credentials, CHAIN_ID);
    recipient = accounts.getSecondaryBenefactor().getAddress();
  }

  @Test
  public void legacyTransactionsAreAccepted() throws Exception {
    // Act - Send a legacy transaction
    String txHash =
        txManager
            .sendTransaction(GAS_PRICE, GAS_LIMIT, recipient, DATA, VALUE)
            .getTransactionHash();

    this.buildNewBlock();

    // Assert
    minerNode.verify(eth.expectSuccessfulTransactionReceipt(txHash));
  }
}
