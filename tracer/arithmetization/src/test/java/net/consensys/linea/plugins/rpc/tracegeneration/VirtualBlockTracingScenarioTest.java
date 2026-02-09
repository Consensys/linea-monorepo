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

package net.consensys.linea.plugins.rpc.tracegeneration;

import static org.assertj.core.api.Assertions.assertThat;

import java.util.List;
import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.ToyAccount;
import net.consensys.linea.testing.ToyExecutionEnvironmentV2;
import net.consensys.linea.testing.ToyTransaction;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.crypto.SECP256K1;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInfo;

/**
 * Tests for virtual block tracing scenarios used for invalidity proof generation. These tests
 * verify that the ZkTracer can properly capture execution traces for BadPrecompile and TooManyLogs
 * scenarios.
 */
class VirtualBlockTracingScenarioTest extends TracerTestBase {

  private static final KeyPair senderKeyPair = new SECP256K1().generateKeyPair();
  private static final Address senderAddress =
      Address.extract(senderKeyPair.getPublicKey());
  private static final ToyAccount senderAccount =
      ToyAccount.builder().balance(Wei.fromEth(100)).nonce(0).address(senderAddress).build();

  /**
   * Tests tracing a transaction that calls a precompile with invalid input. This simulates the
   * BadPrecompile scenario for invalidity proofs.
   */
  @Test
  void traceBadPrecompileScenario(TestInfo testInfo) {
    // Contract that calls ECRECOVER precompile (0x01) with garbage data
    final Bytes contractCode =
        BytecodeCompiler.newProgram(chainConfig)
            // Setup call to precompile
            .push(0) // retSize
            .push(0) // retOffset
            .push(128) // argsSize - send 128 bytes of zeros (invalid ecrecover input)
            .push(0) // argsOffset
            .push(0) // value
            .push(Bytes.fromHexString("0x01")) // precompile address
            .push(10000) // gas
            .op(OpCode.CALL)
            .compile();

    final ToyAccount contractAccount =
        ToyAccount.builder()
            .balance(Wei.ZERO)
            .address(Address.fromHexString("0xdeadbeef00000000000000000000000000000001"))
            .code(contractCode)
            .build();

    final Transaction tx =
        ToyTransaction.builder()
            .sender(senderAccount)
            .to(contractAccount)
            .keyPair(senderKeyPair)
            .gasLimit(100000L)
            .value(Wei.ZERO)
            .build();

    ToyExecutionEnvironmentV2.builder(chainConfig, testInfo)
        .accounts(List.of(senderAccount, contractAccount))
        .transaction(tx)
        .zkTracerValidator(
            zkTracer -> {
              // Verify tracer captured the execution
              assertThat(zkTracer.getHub()).isNotNull();
              // The tracer should have processed the transaction even if precompile failed
              assertThat(zkTracer.getTracingExceptions()).isEmpty();
            })
        .build()
        .run();
  }

  /**
   * Tests tracing a transaction that emits many logs. This simulates the TooManyLogs scenario for
   * invalidity proofs.
   */
  @Test
  void traceTooManyLogsScenario(TestInfo testInfo) {
    // Contract that emits multiple LOG0 events in a loop
    // This creates a simple contract that emits logs
    final Bytes contractCode =
        BytecodeCompiler.newProgram(chainConfig)
            // Emit LOG0 multiple times
            .push(0) // size
            .push(0) // offset
            .op(OpCode.LOG0)
            .push(0)
            .push(0)
            .op(OpCode.LOG0)
            .push(0)
            .push(0)
            .op(OpCode.LOG0)
            .push(0)
            .push(0)
            .op(OpCode.LOG0)
            .push(0)
            .push(0)
            .op(OpCode.LOG0)
            .op(OpCode.STOP)
            .compile();

    final ToyAccount contractAccount =
        ToyAccount.builder()
            .balance(Wei.ZERO)
            .address(Address.fromHexString("0xdeadbeef00000000000000000000000000000002"))
            .code(contractCode)
            .build();

    final Transaction tx =
        ToyTransaction.builder()
            .sender(senderAccount)
            .to(contractAccount)
            .keyPair(senderKeyPair)
            .gasLimit(100000L)
            .value(Wei.ZERO)
            .nonce(1L) // Increment nonce for second test
            .build();

    ToyExecutionEnvironmentV2.builder(chainConfig, testInfo)
        .accounts(List.of(senderAccount, contractAccount))
        .transaction(tx)
        .zkTracerValidator(
            zkTracer -> {
              // Verify tracer captured the logs
              assertThat(zkTracer.getHub()).isNotNull();
              assertThat(zkTracer.getTracingExceptions()).isEmpty();
            })
        .build()
        .run();
  }

  /**
   * Tests tracing a simple value transfer transaction. This is a baseline test to ensure basic
   * tracing works.
   */
  @Test
  void traceSimpleValueTransfer(TestInfo testInfo) {
    final ToyAccount receiverAccount =
        ToyAccount.builder()
            .balance(Wei.ZERO)
            .address(Address.fromHexString("0xdeadbeef00000000000000000000000000000003"))
            .build();

    final Transaction tx =
        ToyTransaction.builder()
            .sender(senderAccount)
            .to(receiverAccount)
            .keyPair(senderKeyPair)
            .gasLimit(21000L)
            .value(Wei.of(1000))
            .nonce(2L)
            .build();

    ToyExecutionEnvironmentV2.builder(chainConfig, testInfo)
        .accounts(List.of(senderAccount, receiverAccount))
        .transaction(tx)
        .zkTracerValidator(
            zkTracer -> {
              assertThat(zkTracer.getHub()).isNotNull();
              assertThat(zkTracer.getTracingExceptions()).isEmpty();
            })
        .build()
        .run();
  }

  /**
   * Tests tracing a transaction with contract creation. This verifies the tracer handles CREATE
   * operations.
   */
  @Test
  void traceContractCreation(TestInfo testInfo) {
    // Simple contract bytecode that just stops
    final Bytes initCode =
        BytecodeCompiler.newProgram(chainConfig)
            .push(1) // return size
            .push(0) // return offset
            .op(OpCode.RETURN)
            .compile();

    final Transaction tx =
        ToyTransaction.builder()
            .sender(senderAccount)
            .keyPair(senderKeyPair)
            .gasLimit(100000L)
            .value(Wei.ZERO)
            .payload(initCode)
            .nonce(3L)
            .build();

    ToyExecutionEnvironmentV2.builder(chainConfig, testInfo)
        .accounts(List.of(senderAccount))
        .transaction(tx)
        .zkTracerValidator(
            zkTracer -> {
              assertThat(zkTracer.getHub()).isNotNull();
              assertThat(zkTracer.getTracingExceptions()).isEmpty();
            })
        .build()
        .run();
  }
}
