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

package net.consensys.linea.zktracer.testing;

import java.util.List;
import java.util.function.Consumer;

import com.google.common.base.Preconditions;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ZkTracer;
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
  private final Bytes byteCode;

  /**
   * @param byteCode the byte code to test
   */
  public BytecodeRunner(Bytes byteCode) {
    this.byteCode = byteCode;
  }

  public static BytecodeRunner of(Bytes byteCode) {
    return new BytecodeRunner(byteCode);
  }

  @Setter private Consumer<ZkTracer> zkTracerValidator = zkTracer -> {};

  public void run() {
    Preconditions.checkArgument(byteCode != null, "byteCode cannot be empty");

    KeyPair keyPair = new SECP256K1().generateKeyPair();
    Address senderAddress = Address.extract(Hash.hash(keyPair.getPublicKey().getEncodedBytes()));

    final ToyAccount senderAccount =
        ToyAccount.builder().balance(Wei.fromEth(1)).nonce(5).address(senderAddress).build();

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
            .keyPair(keyPair)
            .gasLimit(25000000L)
            .build();

    final ToyWorld toyWorld =
        ToyWorld.builder().accounts(List.of(senderAccount, receiverAccount)).build();

    ToyExecutionEnvironment.builder()
        .testValidator(x -> {})
        .toyWorld(toyWorld)
        .zkTracerValidator(zkTracerValidator)
        .transaction(tx)
        .build()
        .run();
  }
}
