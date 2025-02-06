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

package net.consensys.linea.testing;

import static com.google.common.base.Preconditions.*;

import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.function.Consumer;

import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.module.constants.GlobalConstants;
import net.consensys.linea.zktracer.module.hub.Hub;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;

/**
 * A BytecodeRunner takes bytecode, then run it in a single transaction in a single block, and
 * ensures that it executed correctly.
 */
@Accessors(fluent = true)
public final class BytecodeRunner {
  public static final long DEFAULT_GAS_LIMIT = 61_000_000L;
  private final Bytes byteCode;
  ToyExecutionEnvironmentV2 toyExecutionEnvironmentV2;

  /**
   * @param byteCode the byte code to test
   */
  public BytecodeRunner(Bytes byteCode) {
    this.byteCode = byteCode;
  }

  public static BytecodeRunner of(BytecodeCompiler program) {
    return new BytecodeRunner(program.compile());
  }

  public static BytecodeRunner of(Bytes byteCode) {
    return new BytecodeRunner(byteCode);
  }

  @Setter private Consumer<ZkTracer> zkTracerValidator = zkTracer -> {};

  // Default run method
  public void run() {
    this.run(Wei.fromEth(1), (long) GlobalConstants.LINEA_BLOCK_GAS_LIMIT, List.of());
  }

  // Ad-hoc senderBalance
  public void run(Wei senderBalance) {
    this.run(senderBalance, (long) GlobalConstants.LINEA_BLOCK_GAS_LIMIT, List.of());
  }

  // Ad-hoc gasLimit
  public void run(Long gasLimit) {
    this.run(Wei.fromEth(1), gasLimit, List.of());
  }

  // Ad-hoc senderBalance and gasLimit
  public void run(Wei senderBalance, Long gasLimit) {
    this.run(senderBalance, gasLimit, List.of());
  }

  // Ad-hoc accounts
  public void run(List<ToyAccount> additionalAccounts) {
    this.run(Wei.fromEth(1), (long) GlobalConstants.LINEA_BLOCK_GAS_LIMIT, additionalAccounts);
  }

  // Ad-hoc gasLimit and accounts
  public void run(Long gasLimit, List<ToyAccount> additionalAccounts) {
    this.run(Wei.fromEth(1), gasLimit, additionalAccounts);
  }

  // Ad-hoc senderBalance, gasLimit and accounts
  public void run(Wei senderBalance, Long gasLimit, List<ToyAccount> additionalAccounts) {
    checkArgument(byteCode != null, "byteCode cannot be empty");

    KeyPair keyPair = new SECP256K1().generateKeyPair();
    Address senderAddress = Address.extract(Hash.hash(keyPair.getPublicKey().getEncodedBytes()));

    final ToyAccount senderAccount =
        ToyAccount.builder().balance(senderBalance).nonce(5).address(senderAddress).build();

    final Long selectedGasLimit = Optional.of(gasLimit).orElse(DEFAULT_GAS_LIMIT);

    final ToyAccount receiverAccount =
        ToyAccount.builder()
            .balance(Wei.fromEth(1))
            .nonce(6)
            .address(Address.fromHexString("0x1111111111111111111111111111111111111111"))
            .code(byteCode)
            .build();

    final Transaction tx =
        ToyTransaction.builder()
            .sender(senderAccount)
            .to(receiverAccount)
            .value(Wei.of(69))
            .keyPair(keyPair)
            .gasLimit(selectedGasLimit)
            .gasPrice(Wei.of(8))
            .build();

    List<ToyAccount> accounts = new ArrayList<>();
    accounts.add(senderAccount);
    accounts.add(receiverAccount);
    accounts.addAll(additionalAccounts);

    toyExecutionEnvironmentV2 =
        ToyExecutionEnvironmentV2.builder()
            .transactionProcessingResultValidator(
                TransactionProcessingResultValidator.EMPTY_VALIDATOR)
            .accounts(accounts)
            .zkTracerValidator(zkTracerValidator)
            .transaction(tx)
            .build();
    toyExecutionEnvironmentV2.run();
  }

  public Hub getHub() {
    return toyExecutionEnvironmentV2.getHub();
  }
}
