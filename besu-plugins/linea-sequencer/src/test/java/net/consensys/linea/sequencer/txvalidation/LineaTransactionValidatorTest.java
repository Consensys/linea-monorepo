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

package net.consensys.linea.sequencer.txvalidation;

import java.util.ArrayList;
import java.util.Optional;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.junit.jupiter.api.Assertions;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

@Slf4j
@RequiredArgsConstructor
public class LineaTransactionValidatorTest {

  public static final Address DENIED =
      Address.fromHexString("0x0000000000000000000000000000000000001000");
  public static final Address NOT_DENIED =
      Address.fromHexString("0x0000000000000000000000000000000000001001");
  public static final Address PRECOMPILED = Address.precompiled(0xa);
  public static final int MAX_TX_GAS_LIMIT = 9000000;
  private LineaTransactionValidator lineaTransactionValidator;

  @BeforeEach
  public void initialize() {
    ArrayList<Address> denied = new ArrayList<>();
    denied.add(DENIED);
    lineaTransactionValidator =
        new LineaTransactionValidator(
            new LineaTransactionValidatorConfiguration("", MAX_TX_GAS_LIMIT), denied);
  }

  @Test
  public void validatedIfNoneOnList() {
    final org.hyperledger.besu.ethereum.core.Transaction.Builder builder =
        org.hyperledger.besu.ethereum.core.Transaction.builder();
    final org.hyperledger.besu.ethereum.core.Transaction transaction =
        builder.sender(NOT_DENIED).to(NOT_DENIED).gasPrice(Wei.ZERO).build();
    Assertions.assertEquals(
        lineaTransactionValidator.validateTransaction(transaction), Optional.empty());
  }

  @Test
  public void deniedIfFromAddressIsOnList() {
    final org.hyperledger.besu.ethereum.core.Transaction.Builder builder =
        org.hyperledger.besu.ethereum.core.Transaction.builder();
    final org.hyperledger.besu.ethereum.core.Transaction transaction =
        builder.sender(DENIED).to(NOT_DENIED).gasPrice(Wei.ZERO).build();
    Assertions.assertEquals(
        lineaTransactionValidator.validateTransaction(transaction).orElseThrow(),
        "sender 0x0000000000000000000000000000000000001000 is blocked as appearing on the SDN or other legally prohibited list");
  }

  @Test
  public void deniedIfToAddressIsOnList() {
    final org.hyperledger.besu.ethereum.core.Transaction.Builder builder =
        org.hyperledger.besu.ethereum.core.Transaction.builder();
    final org.hyperledger.besu.ethereum.core.Transaction transaction =
        builder.sender(NOT_DENIED).to(DENIED).gasPrice(Wei.ZERO).build();
    Assertions.assertEquals(
        lineaTransactionValidator.validateTransaction(transaction).orElseThrow(),
        "recipient 0x0000000000000000000000000000000000001000 is blocked as appearing on the SDN or other legally prohibited list");
  }

  @Test
  public void deniedIfToAddressIsPrecompiled() {
    final org.hyperledger.besu.ethereum.core.Transaction.Builder builder =
        org.hyperledger.besu.ethereum.core.Transaction.builder();
    final org.hyperledger.besu.ethereum.core.Transaction transaction =
        builder.sender(NOT_DENIED).to(PRECOMPILED).gasPrice(Wei.ZERO).build();
    Assertions.assertEquals(
        lineaTransactionValidator.validateTransaction(transaction).orElseThrow(),
        "destination address is a precompile address and cannot receive transactions");
  }

  @Test
  public void validatedWithValidGasLimit() {
    final org.hyperledger.besu.ethereum.core.Transaction.Builder builder =
        org.hyperledger.besu.ethereum.core.Transaction.builder();
    final org.hyperledger.besu.ethereum.core.Transaction transaction =
        builder
            .sender(NOT_DENIED)
            .to(NOT_DENIED)
            .gasLimit(MAX_TX_GAS_LIMIT)
            .gasPrice(Wei.ZERO)
            .build();
    Assertions.assertEquals(
        lineaTransactionValidator.validateTransaction(transaction), Optional.empty());
  }

  @Test
  public void rejectedWithMaxGasLimitPlusOne() {
    final org.hyperledger.besu.ethereum.core.Transaction.Builder builder =
        org.hyperledger.besu.ethereum.core.Transaction.builder();
    final org.hyperledger.besu.ethereum.core.Transaction transaction =
        builder
            .sender(NOT_DENIED)
            .to(NOT_DENIED)
            .gasLimit(MAX_TX_GAS_LIMIT + 1)
            .gasPrice(Wei.ZERO)
            .build();
    Assertions.assertEquals(
        lineaTransactionValidator.validateTransaction(transaction).orElseThrow(),
        "Gas limit of transaction is greater than the allowed max of " + MAX_TX_GAS_LIMIT);
  }
}
