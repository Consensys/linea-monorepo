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

package net.consensys.linea.zktracer.module.rlp_txrcpt;

import static org.assertj.core.api.Assertions.assertThat;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.Random;

import net.consensys.linea.zktracer.ZkTraceBuilder;
import net.consensys.linea.zktracer.corset.CorsetValidator;
import net.consensys.linea.zktracer.opcode.OpCodes;
import net.consensys.linea.zktracer.testing.ToyAccount;
import net.consensys.linea.zktracer.testing.ToyTransaction;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.datatypes.AccessListEntry;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.log.LogTopic;
import org.junit.jupiter.api.Test;

public class RandomTxrcptTests {
  @Test
  public void testRandomTxrcpt() {
    RlpTxrcpt rlpTxrcpt = new RlpTxrcpt();
    OpCodes.load();

    // SET UP THE WORLD
    KeyPair keyPair = new SECP256K1().generateKeyPair();
    Address senderAddress = Address.extract(Hash.hash(keyPair.getPublicKey().getEncodedBytes()));

    ToyAccount senderAccount =
        ToyAccount.builder().balance(Wei.of(5)).nonce(32).address(senderAddress).build();

    // Create few tx
    for (int i = 0; i < 200; i++) {
      final Transaction tx = randTransaction(senderAccount, keyPair);

      // Create a mock test receipt

      final Bytes output = Bytes.random(20);
      final boolean status = rnd.nextBoolean();
      final List<Log> logs = randomListLog(rnd.nextInt(10));
      final long gasUsed = rnd.nextLong(21000, 0xfffffffffffffffL);

      // Call the module
      rlpTxrcpt.traceEndTx(null, tx, status, output, logs, gasUsed);
    }

    //
    // Check the trace
    //
    assertThat(CorsetValidator.isValid(new ZkTraceBuilder().addTrace(rlpTxrcpt).build().toJson()))
        .isTrue();
  }

  private final Random rnd = new Random(666);

  private Log randomLog(int dataType, int nbTopic) {
    Bytes data =
        switch (dataType) {
          case 0 -> Bytes.EMPTY;
          case 1 -> Bytes.of(0x00);
          case 2 -> Bytes.minimalBytes(rnd.nextInt(1, 128));
          case 3 -> Bytes.minimalBytes(rnd.nextInt(128, 256));
          case 4 -> Bytes.random(rnd.nextInt(2, 56));
          case 5 -> Bytes.random(rnd.nextInt(56, 6666));
          default -> null;
        };

    List<LogTopic> topics = new ArrayList<>();
    for (int i = 0; i < nbTopic; i++) {
      topics.add(LogTopic.of(Bytes.random(32)));
    }
    return new Log(Address.wrap(Bytes.random(20)), data, topics);
  }

  private List<Log> randomListLog(int nLog) {
    List<Log> logs = new java.util.ArrayList<>(List.of());
    for (int i = 0; i < nLog; i++) {
      logs.add(randomLog(rnd.nextInt(0, 6), rnd.nextInt(0, 5)));
    }
    return logs;
  }

  private Transaction randTransaction(ToyAccount senderAccount, KeyPair keyPair) {
    int txType = rnd.nextInt(0, 3);
    final List<AccessListEntry> accessList = new ArrayList<>();
    return switch (txType) {
      case 0 -> ToyTransaction.builder()
          .sender(senderAccount)
          .keyPair(keyPair)
          .transactionType(TransactionType.FRONTIER)
          .gasLimit(10_000_000L)
          .value(Wei.of(BigInteger.valueOf(2_500)))
          .payload(Bytes.EMPTY)
          .build();
      case 1 -> ToyTransaction.builder()
          .sender(senderAccount)
          .keyPair(keyPair)
          .transactionType(TransactionType.ACCESS_LIST)
          .gasLimit(10_000_000L)
          .value(Wei.of(BigInteger.valueOf(2_500)))
          .payload(Bytes.EMPTY)
          .accessList(accessList)
          .build();

      case 2 -> ToyTransaction.builder()
          .sender(senderAccount)
          .keyPair(keyPair)
          .transactionType(TransactionType.EIP1559)
          .gasLimit(10_000_000L)
          .value(Wei.of(BigInteger.valueOf(2_500)))
          .payload(Bytes.EMPTY)
          .build();
      default -> throw new IllegalStateException("Unexpected value: " + txType);
    };
  }
}
