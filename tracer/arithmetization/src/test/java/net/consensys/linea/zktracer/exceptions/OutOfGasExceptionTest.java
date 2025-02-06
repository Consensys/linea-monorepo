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
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.GAS_CONST_G_CALL_VALUE;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.GAS_CONST_G_COLD_ACCOUNT_ACCESS;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.GAS_CONST_G_COLD_SLOAD;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.GAS_CONST_G_NEW_ACCOUNT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.GAS_CONST_G_SSET;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.GAS_CONST_G_TRANSACTION;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.GAS_CONST_G_VERY_LOW;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.GAS_CONST_G_WARM_ACCESS;
import static net.consensys.linea.zktracer.module.hub.signals.TracedException.OUT_OF_GAS_EXCEPTION;
import static net.consensys.linea.zktracer.opcode.OpCodes.opCodeToOpCodeDataMap;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotEquals;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Stream;

import com.google.common.base.Preconditions;
import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.testing.ToyAccount;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.junit.jupiter.params.provider.ValueSource;

@ExtendWith(UnitTestWatcher.class)
public class OutOfGasExceptionTest {

  @ParameterizedTest
  @MethodSource("outOfGasExceptionWithEmptyAccountsAndNoMemoryExpansionCostTestSource")
  void outOfGasExceptionWithEmptyAccountsAndNoMemoryExpansionCostTest(
      OpCode opCode, int opCodeStaticCost, int nPushes, int cornerCase) {
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
    BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());
    int opCodeDynamicCost =
        switch (opCode) {
          case OpCode.SELFDESTRUCT -> GAS_CONST_G_NEW_ACCOUNT
              + GAS_CONST_G_COLD_ACCOUNT_ACCESS; // since the account is empty
          case OpCode.EXP -> 50; // since the exponent requires 1 byte
          case OpCode.SLOAD -> GAS_CONST_G_COLD_SLOAD; // since the slot is cold
          case OpCode.BALANCE,
              OpCode.EXTCODEHASH,
              OpCode.EXTCODESIZE -> GAS_CONST_G_COLD_ACCOUNT_ACCESS; // since the account is cold
          case OpCode.SSTORE -> GAS_CONST_G_SSET
              + GAS_CONST_G_COLD_SLOAD; // value set from zero to non-zero and slot is cold
          default -> 0;
        };

    final long gasCost =
        GAS_CONST_G_TRANSACTION
            + (long) nPushes * GAS_CONST_G_VERY_LOW
            + opCodeStaticCost
            + opCodeDynamicCost;
    bytecodeRunner.run(gasCost + cornerCase);

    // TODO: this is to ensure that the gas cost is correctly calculated
    //  this may change when the gas accumulator will be associated to the frame
    assertEquals(gasCost, GAS_CONST_G_TRANSACTION + bytecodeRunner.getHub().gasCostAccumulator());

    if (cornerCase == -1) {
      assertEquals(
          OUT_OF_GAS_EXCEPTION,
          bytecodeRunner.getHub().previousTraceSection().commonValues.tracedException());
    } else {
      assertNotEquals(
          OUT_OF_GAS_EXCEPTION,
          bytecodeRunner.getHub().previousTraceSection().commonValues.tracedException());
    }
  }

  static Stream<Arguments> outOfGasExceptionWithEmptyAccountsAndNoMemoryExpansionCostTestSource() {
    List<Arguments> arguments = new ArrayList<>();
    for (OpCodeData opCodeData : opCodeToOpCodeDataMap.values()) {
      OpCode opCode = opCodeData.mnemonic();
      int opCodeStaticCost = opCodeData.stackSettings().staticGas().cost();
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
          && opCode != OpCode.SHA3 // SHA3 needs the memory expansion cost
          && opCode != OpCode.STOP // STOP does not consume gas
          && !opCodeData.isCall() // CALL family is managed separately
          && !opCodeData.isCreate() // CREATE needs the memory expansion cost
          && !opCodeData.isLog() // LOG needs the memory expansion cost
      ) {
        arguments.add(Arguments.of(opCode, opCodeStaticCost, nPushes, -1));
        arguments.add(Arguments.of(opCode, opCodeStaticCost, nPushes, 0));
        arguments.add(Arguments.of(opCode, opCodeStaticCost, nPushes, 1));
      }
    }
    return arguments.stream();
  }

  @ParameterizedTest
  @MethodSource("outOfGasExceptionCallSource")
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
        .push(1000) // gas
        .op(OpCode.CALL);

    BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());

    long gasCost =
        GAS_CONST_G_TRANSACTION
            + // base gas cost
            (isWarm ? GAS_CONST_G_VERY_LOW + GAS_CONST_G_COLD_ACCOUNT_ACCESS : 0) // PUSH + BALANCE
            + 7 * GAS_CONST_G_VERY_LOW // 7 PUSH
            + callGasCost(value != 0, targetAddressExists, isWarm); // CALL

    if (targetAddressExists) {
      final ToyAccount calleeAccount =
          ToyAccount.builder()
              .balance(Wei.fromEth(1))
              .nonce(10)
              .address(Address.fromHexString("ca11ee"))
              .build();
      bytecodeRunner.run(gasCost + cornerCase, List.of(calleeAccount));
    } else {
      bytecodeRunner.run(gasCost + cornerCase);
    }

    if (cornerCase == -1) {
      assertEquals(
          OUT_OF_GAS_EXCEPTION,
          bytecodeRunner.getHub().previousTraceSection().commonValues.tracedException());
    } else {
      assertNotEquals(
          OUT_OF_GAS_EXCEPTION,
          bytecodeRunner.getHub().previousTraceSection().commonValues.tracedException());
    }
  }

  static Stream<Arguments> outOfGasExceptionCallSource() {
    List<Arguments> arguments = new ArrayList<>();
    for (int value : new int[] {0, 1}) {
      for (int cornerCase : new int[] {-1, 0, 1}) {
        arguments.add(Arguments.of(value, true, true, cornerCase));
        arguments.add(Arguments.of(value, true, false, cornerCase));
        arguments.add(Arguments.of(value, false, false, cornerCase));
      }
    }
    return arguments.stream();
  }

  private long callGasCost(boolean transfersValue, boolean targetAddressExists, boolean isWarm) {
    Preconditions.checkArgument(
        !(isWarm && !targetAddressExists), "isWarm implies targetAddressExists");
    return (transfersValue ? GAS_CONST_G_CALL_VALUE : 0)
        + (targetAddressExists ? 0 : (transfersValue ? GAS_CONST_G_NEW_ACCOUNT : 0))
        + (isWarm ? GAS_CONST_G_WARM_ACCESS : GAS_CONST_G_COLD_ACCOUNT_ACCESS);
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

    BytecodeRunner bytecodeRunner = BytecodeRunner.of(program.compile());

    long gasCost =
        (long) GAS_CONST_G_TRANSACTION
            + (long) 2 * GAS_CONST_G_VERY_LOW // 2 PUSH
            + GAS_CONST_G_SSET
            // SSTORE cost since current_value == original_value
            // and original_value == 0 (20000)
            + GAS_CONST_G_COLD_SLOAD // SSTORE cost since slot is cold (2100)
            + (long) GAS_CONST_G_VERY_LOW // PUSH
            + GAS_CONST_G_WARM_ACCESS; // SLOAD (100)

    bytecodeRunner.run(gasCost + cornerCase);

    if (cornerCase == -1) {
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
