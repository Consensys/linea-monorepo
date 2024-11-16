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
package net.consensys.linea.zktracer.instructionprocessing.callTests.abort;

import static net.consensys.linea.zktracer.instructionprocessing.callTests.Utilities.appendInsufficientBalanceCall;
import static net.consensys.linea.zktracer.opcode.OpCode.*;
import static net.consensys.linea.zktracer.opcode.OpCode.CALL;

import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.junit.jupiter.api.Test;

/**
 * The arithmetization has a two aborting scenarios for CALL's
 *
 * <p>- <b>scn/CALL_ABORT_WILL_REVERT</b>
 *
 * <p>- <b>scn/CALL_ABORT_WONT_REVERT</b> The main point being: (unexceptional) aborted CALL's warm
 * up the target account.
 */
public class BalanceAbortTests {

  final String eoaAddress = "abcdef0123456789";

  @Test
  void insufficientBalanceAbortWarmsUpTarget() {

    Bytes bytecode =
        BytecodeCompiler.newProgram()
            .push(1)
            .push(2)
            .push(3)
            .push(4)
            .op(SELFBALANCE)
            .push(1)
            .op(ADD) // our balance + 1
            .push(eoaAddress) // address
            .push(0) // gas
            .op(CALL) // CALL_ABORT_WONT_REVERT
            .op(POP)
            .push(eoaAddress) // address
            .op(EXTCODESIZE) // discounted pricing
            .compile();

    BytecodeRunner.of(bytecode).run();
  }

  /** scenario/CALL_ABORT_WILL_REVERT; reverts the warmth; */
  @Test
  void insufficientBalanceAbortWillRevert() {

    BytecodeCompiler program = BytecodeCompiler.newProgram();
    appendInsufficientBalanceCall(
        program, CALL, 1000, Address.fromHexString(eoaAddress), 0, 0, 0, 0);
    program.push(6).push(7).op(REVERT);
    Bytes bytecode = program.compile();
    BytecodeRunner.of(bytecode).run();
  }

  /** scenario/CALL_ABORT_WONT_REVERT */
  @Test
  void insufficientBalanceAbortWontRevert() {

    BytecodeCompiler program = BytecodeCompiler.newProgram();
    appendInsufficientBalanceCall(
        program, CALL, 1000, Address.fromHexString(eoaAddress), 0, 0, 0, 0);
    Bytes bytecode = program.compile();
    BytecodeRunner.of(bytecode).run();
  }
}
