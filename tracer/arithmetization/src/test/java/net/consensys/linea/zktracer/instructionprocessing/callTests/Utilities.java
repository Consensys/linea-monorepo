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
package net.consensys.linea.zktracer.instructionprocessing.callTests;

import static com.google.common.base.Preconditions.checkArgument;
import static net.consensys.linea.zktracer.opcode.OpCode.*;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.hyperledger.besu.datatypes.Address;
import org.junit.jupiter.api.extension.ExtendWith;

@ExtendWith(UnitTestWatcher.class)
public class Utilities {

  public static final String fullEoaAddress = "000000000000000000000000abcdef0123456789";
  public static final String toTrim12 = "aaaaaaaaaaaaaaaaaaaaaaaa";
  public static final String untrimmedEoaAddress = toTrim12 + fullEoaAddress;
  public static final String eoaAddress = "c0ffeef00d";
  public static final String eoaAddress2 = "badbeef";

  public static void fullGasCall(
      BytecodeCompiler program,
      OpCode callOpcode,
      Address to,
      int value,
      int cdo,
      int cds,
      int rao,
      int rac) {
    program.push(rac).push(rao).push(cds).push(cdo);
    if (callOpcode.callHasValueArgument()) {
      program.push(value);
    }
    program.push(to).op(GAS).op(callOpcode).op(POP);
  }

  public static void simpleCall(
      BytecodeCompiler program,
      OpCode callOpcode,
      int gas,
      Address to,
      int value,
      int cdo,
      int cds,
      int rao,
      int rac) {
    program.push(rac).push(rao).push(cds).push(cdo);
    if (callOpcode.callHasValueArgument()) {
      program.push(value);
    }
    program.push(to).push(gas).op(callOpcode);
  }

  public static void callCaller(
      BytecodeCompiler program,
      OpCode callOpcode,
      int gas,
      int value,
      int cdo,
      int cds,
      int rao,
      int rac) {
    program.push(rac).push(rao).push(cds).push(cdo);
    if (callOpcode.callHasValueArgument()) {
      program.push(value);
    }
    program.op(CALLER).push(gas).op(callOpcode);
  }

  public static void appendInsufficientBalanceCall(
      BytecodeCompiler program,
      OpCode callOpcode,
      int gas,
      Address to,
      int cdo,
      int cds,
      int rao,
      int rac) {
    checkArgument(callOpcode.callHasValueArgument());
    program
        .push(rac)
        .push(rao)
        .push(cds)
        .push(cdo)
        .op(BALANCE)
        .push(1)
        .op(ADD) // puts balance + 1 on the stack
        .push(to)
        .push(gas)
        .op(callOpcode);
  }

  /**
   * Pushing, in order: h, v, r, s; values produce a valid signature.
   *
   * @param program
   */
  public static void validEcrecoverData(BytecodeCompiler program) {
    program
        .push("279d94621558f755796898fc4bd36b6d407cae77537865afe523b79c74cc680b")
        .push(0)
        .op(MSTORE)
        .push("1b")
        .push(32)
        .op(MSTORE)
        .push("c2ff96feed8749a5ad1c0714f950b5ac939d8acedbedcbc2949614ab8af06312")
        .push(64)
        .op(MSTORE)
        .push("1feecd50adc6273fdd5d11c6da18c8cfe14e2787f5a90af7c7c1328e7d0a2c42")
        .push(96)
        .op(MSTORE);
  }
}
