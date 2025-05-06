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

package net.consensys.linea.zktracer.exceptions;

import static net.consensys.linea.testing.ToyExecutionEnvironmentV2.DEFAULT_BLOCK_NUMBER;
import static net.consensys.linea.zktracer.Trace.*;
import static net.consensys.linea.zktracer.module.hub.signals.TracedException.OUT_OF_GAS_EXCEPTION;
import static net.consensys.linea.zktracer.opcode.OpCodes.opCodeDataList;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotEquals;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.testing.ToyAccount;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.junit.jupiter.params.provider.ValueSource;

@Slf4j
@ExtendWith(UnitTestWatcher.class)
public class OutOfGasExceptionTest {

  @ParameterizedTest
  @MethodSource("outOfGasExceptionWithEmptyAccountsAndNoMemoryExpansionCostTestSource")
  void outOfGasExceptionWithEmptyAccountsAndNoMemoryExpansionCostTest(
      OpCode opCode, int nPushes, int cornerCase) {
    BytecodeCompiler program = BytecodeCompiler.newProgram();

    for (int i = 0; i < nPushes; i++) {
      // In order to disambiguate between empty stack items and writing a result of 0 on the stack
      // we push small integers to the stack which all produce non-zero results

      int pushedValue =
          switch (opCode) {
            case OpCode.BLOCKHASH -> Math.toIntExact(DEFAULT_BLOCK_NUMBER) - 1;
            case OpCode.EXP -> i == 0 ? 5 : 2; // EXP 2 5 (2 ** 5)
            default -> 7 * i + 11;
              // small integer but greater than 10, so as when it represents an address
              // it is not the one of a precompile contract
          };
      program.push(pushedValue);
    }
    program.op(opCode);
    Bytes pgCompile = program.compile();
    BytecodeRunner bytecodeRunner = BytecodeRunner.of(pgCompile);

    long gasCost = bytecodeRunner.runOnlyForGasCost();

    bytecodeRunner.run(gasCost + cornerCase);

    ExceptionUtils.assertEqualsOutOfGasIfCornerCaseMinusOneElseAssertNotEquals(
        cornerCase, bytecodeRunner);
  }

  static Stream<Arguments> outOfGasExceptionWithEmptyAccountsAndNoMemoryExpansionCostTestSource() {
    List<Arguments> arguments = new ArrayList<>();
    for (OpCodeData opCodeData : opCodeDataList) {
      if (opCodeData != null) {
        OpCode opCode = opCodeData.mnemonic();
        int nPushes = opCodeData.stackSettings().delta(); // number of items popped from the stack
        if (opCode != OpCode.CALLDATACOPY // CALLDATACOPY needs the memory expansion cost
            && opCode != OpCode.CODECOPY // CODECOPY needs the memory expansion cost
            && opCode != OpCode.EXTCODECOPY // EXTCODECOPY needs the memory expansion cost
            && opCode != OpCode.INVALID // INVALID consumes all gas left
            && opCode != OpCode.MLOAD // MLOAD needs the memory expansion cost
            && opCode != OpCode.MSTORE // MSTORE needs the memory expansion cost
            && opCode != OpCode.MSTORE8 // MSTORE8 needs the memory expansion cost
            && opCode != OpCode.RETURN // RETURN needs the memory expansion cost
            && opCode != OpCode.RETURNDATACOPY // RETURNDATACOPY needs the memory expansion cost
            && opCode != OpCode.REVERT // REVERT needs the memory expansion cost
            && opCode != OpCode.SHA3 // SHA3 needs the memory expansion cost ??
            && opCode != OpCode.STOP // STOP does not consume gas
            && opCode
                != OpCode
                    .JUMP // JUMP needs a valid bytecode to jump to, see outOfGasExceptionJump below
            && opCode
                != OpCode
                    .JUMPI // JUMPI needs a valid bytecode to jump to, see outOfGasExceptionJumpi
            // below
            && opCode != OpCode.SLOAD // SLOAD a non-zero value, see outOfGasExceptionSLoad below
            && !opCodeData.isCall() // CALL family is managed separately
            && !opCodeData.isCreate() // CREATE needs the memory expansion cost
            && !opCodeData.isLog() // LOG needs the memory expansion cost
        ) {
          arguments.add(Arguments.of(opCode, nPushes, -1));
          System.out.println(opCode);
          System.out.println(nPushes);
          arguments.add(Arguments.of(opCode, nPushes, 0));
          arguments.add(Arguments.of(opCode, nPushes, 1));
        }
      }
    }
    return arguments.stream();
  }

  @ParameterizedTest
  @MethodSource("outOfGasExceptionCallSource")
  /*
  When value is transferred
  -> Add additional call stipend (2300) to avoid OOGX in order to complete the call execution, even if no code is executed
   */
  void outOfGasExceptionCallTest(
      int value, boolean targetAddressExists, boolean isWarm, int cornerCase) {
    BytecodeCompiler program = BytecodeCompiler.newProgram();

    if (targetAddressExists && isWarm) {
      // Note: this is a possible way to warm the address
      program.push("ca11ee").op(OpCode.BALANCE);
    }

    program
        .push(0) // return at capacity
        .push(0) // return at offset
        .push(0) // call data size
        .push(0) // call data offset
        .push(value) // value
        .push("ca11ee") // address
        .push(0) // gas for subcontext (floored at 2300)
        .op(OpCode.CALL);

    Bytes pgCompile = program.compile();
    BytecodeRunner bytecodeRunner = BytecodeRunner.of(pgCompile);
    long gasCost;

    if (targetAddressExists) {
      final ToyAccount calleeAccount =
          ToyAccount.builder()
              .balance(Wei.fromEth(1))
              .nonce(10)
              .address(Address.fromHexString("ca11ee"))
              .build();
      gasCost = bytecodeRunner.runOnlyForGasCost(List.of(calleeAccount));
      bytecodeRunner.run(gasCost + cornerCase, List.of(calleeAccount));
    } else {
      gasCost = bytecodeRunner.runOnlyForGasCost();
      bytecodeRunner.run(gasCost + cornerCase);
    }

    if (value == 0) {
      ExceptionUtils.assertEqualsOutOfGasIfCornerCaseMinusOneElseAssertNotEquals(
          cornerCase, bytecodeRunner);
    } else {
      if (cornerCase == 2299) {
        assertEquals(
            OUT_OF_GAS_EXCEPTION,
            bytecodeRunner.getHub().previousTraceSection().commonValues.tracedException());
      } else {
        assertNotEquals(
            OUT_OF_GAS_EXCEPTION,
            bytecodeRunner.getHub().previousTraceSection().commonValues.tracedException());
      }
    }
  }

  static Stream<Arguments> outOfGasExceptionCallSource() {
    List<Arguments> arguments = new ArrayList<>();
    for (int value : new int[] {0, 1}) {
      int[] cornerCaseSet = value == 0 ? new int[] {-1, 0, 1} : new int[] {2299, 2300, 2301};
      for (int cornerCase : cornerCaseSet) {
        arguments.add(Arguments.of(value, true, true, cornerCase));
        arguments.add(Arguments.of(value, true, false, cornerCase));
        arguments.add(Arguments.of(value, false, false, cornerCase));
      }
    }
    return arguments.stream();
  }

  /**
   * We provide a non-zero value in storage so to disambiguate between writing the value in storage
   * to the stack and writing 0 to the stack.
   */
  @ParameterizedTest
  @ValueSource(ints = {-1, 0, 1})
  void outOfGasExceptionSLoad(int cornerCase) {
    BytecodeCompiler program = BytecodeCompiler.newProgram();

    program
        .push(2) // value
        .push(1) // key
        .op(OpCode.SSTORE);

    program
        .push(1)
        . // key
        op(OpCode.SLOAD);

    Bytes pgCompile = program.compile();
    BytecodeRunner bytecodeRunner = BytecodeRunner.of(pgCompile);

    long gasCost = bytecodeRunner.runOnlyForGasCost();

    bytecodeRunner.run(gasCost + cornerCase);

    ExceptionUtils.assertEqualsOutOfGasIfCornerCaseMinusOneElseAssertNotEquals(
        cornerCase, bytecodeRunner);
  }

  @ParameterizedTest
  @ValueSource(ints = {-1, 0, 1})
  void outOfGasExceptionJump(int cornerCase) {
    final Bytes bytecode =
        BytecodeCompiler.newProgram()
            .push(4)
            .op(OpCode.JUMP)
            .op(OpCode.INVALID)
            .op(OpCode.JUMPDEST)
            .push(OpCode.JUMPDEST.byteValue())
            .compile();

    BytecodeRunner bytecodeRunner = BytecodeRunner.of(bytecode);
    long gasCost;
    if (cornerCase == -1) {
      // JUMP needs JUMPDEST to jump to
      // Calculate the gas cost to trigger OOGX on JUMP and not on the last but one opcode
      // 21000L intrinsic gas cost + 3L PUSH + 8L JUMP, and we retrieve 1
      gasCost = GAS_CONST_G_TRANSACTION + GAS_CONST_G_VERY_LOW + GAS_CONST_G_MID - 1L;
    } else {
      gasCost = bytecodeRunner.runOnlyForGasCost();
    }
    bytecodeRunner.run(gasCost);

    ExceptionUtils.assertEqualsOutOfGasIfCornerCaseMinusOneElseAssertNotEquals(
        cornerCase, bytecodeRunner);
  }

  @ParameterizedTest
  @ValueSource(ints = {-1, 0, 1})
  void outOfGasExceptionJumpi(int cornerCase) {
    final Bytes bytecode =
        BytecodeCompiler.newProgram()
            .push(1) // pc = 0, 1
            .push(7) // pc = 2, 3
            .op(OpCode.JUMPI) // pc = 4
            .op(OpCode.JUMPDEST) // pc = 5
            .op(OpCode.INVALID) // pc = 6
            .op(OpCode.JUMPDEST) // pc = 7
            .push(1) // pc = 8
            .compile();

    BytecodeRunner bytecodeRunner = BytecodeRunner.of(bytecode);

    long gasCost;
    if (cornerCase == -1) {
      // JUMPI needs JUMPDEST to jump to
      // Calculate the gas cost to trigger OOGX on JUMPI and not on the last but one opcode
      // 21000L intrinsic gas cost + 3L PUSH + 10L JUMP, and we retrieve 1
      gasCost = GAS_CONST_G_TRANSACTION + 2 * GAS_CONST_G_VERY_LOW + GAS_CONST_G_HIGH - 1L;
    } else {
      gasCost = bytecodeRunner.runOnlyForGasCost();
    }

    bytecodeRunner.run(gasCost);

    ExceptionUtils.assertEqualsOutOfGasIfCornerCaseMinusOneElseAssertNotEquals(
        cornerCase, bytecodeRunner);
  }
}
