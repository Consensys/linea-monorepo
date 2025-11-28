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
package net.consensys.linea.zktracer.instructionprocessing.callTests.prc.p256verify;

import static net.consensys.linea.zktracer.instructionprocessing.callTests.Utilities.simpleCall;
import static net.consensys.linea.zktracer.opcode.OpCode.*;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import org.hyperledger.besu.datatypes.Address;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.api.extension.ExtendWith;

@ExtendWith(UnitTestWatcher.class)
public class P256VerifyCallSuccessAndFailureCasesTest extends TracerTestBase {

  @Test
  void insufficientGasP256VerifyCall_ExpectedCallFailure(TestInfo testInfo) {

    BytecodeCompiler pg = BytecodeCompiler.newProgram(chainConfig);
    simpleCall(pg, STATICCALL, 6899, Address.P256_VERIFY, 0, 0, 0, 0, 32);

    BytecodeRunner.of(pg.compile()).run(chainConfig, testInfo);
  }

  @Test
  void sufficientGasWithStipendP256VerifyCall_ExpectedCallSuccess(TestInfo testInfo) {

    BytecodeCompiler pg = BytecodeCompiler.newProgram(chainConfig);
    simpleCall(pg, CALL, 6900 - 2300, Address.P256_VERIFY, 1, 0, 0, 0, 32);

    BytecodeRunner.of(pg.compile()).run(chainConfig, testInfo);
  }

  @Test
  void sufficientGasAndInvalidCDSP256VerifyCall_ExpectedCallSuccess(TestInfo testInfo) {

    BytecodeCompiler pg = BytecodeCompiler.newProgram(chainConfig);
    simpleCall(pg, STATICCALL, 6900, Address.P256_VERIFY, 0, 0, 0, 0, 32);

    BytecodeRunner.of(pg.compile()).run(chainConfig, testInfo);
  }

  @Test
  void sufficientGasAndValidCDSP256VerifyCall_ExpectedCallSuccess(TestInfo testInfo) {

    BytecodeCompiler pg = BytecodeCompiler.newProgram(chainConfig);
    simpleCall(pg, STATICCALL, 6900, Address.P256_VERIFY, 0, 0, 160, 0, 32);

    BytecodeRunner.of(pg.compile()).run(chainConfig, testInfo);
  }
}
