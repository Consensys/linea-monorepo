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
package net.consensys.linea.zktracer.instructionprocessing.callTests.prc.framework;

import static net.consensys.linea.zktracer.Trace.WORD_SIZE;
import static net.consensys.linea.zktracer.instructionprocessing.callTests.Utilities.revertWith;
import static net.consensys.linea.zktracer.instructionprocessing.callTests.prc.CodeExecutionMethods.*;
import static net.consensys.linea.zktracer.instructionprocessing.callTests.prc.CodeExecutionMethods.runCreateDeployingForeignCodeAndCallIntoIt;

import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.zktracer.instructionprocessing.callTests.prc.CodeExecutionMethods;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.MethodSource;

public abstract class PrecompileCallTests<T extends PrecompileCallParameters> {

  /**
   * <b>MESSAGE_CALL_TRANSACTION</b> case.
   *
   * <p>See {@link CodeExecutionMethods} for documentation and context.
   */
  @ParameterizedTest
  @MethodSource("parameterGeneration")
  public void messageCallTransactionTest(T callParameter) {

    BytecodeCompiler rootCode =
        callParameter.customPrecompileCallsSeparatedByReturnDataWipingOperation();
    if (callParameter.willRevert()) revertWith(rootCode, 0, 5 * WORD_SIZE);

    runMessageCallTransactionWithProvidedCodeAsRootCode(rootCode);
  }

  /**
   * <b>CONTRACT_DEPLOYMENT_TRANSACTION</b> case.
   *
   * <p>See {@link CodeExecutionMethods} for documentation and context.
   */
  @ParameterizedTest
  @MethodSource("parameterGeneration")
  public void deploymentTransactionTest(T callParameter) {

    BytecodeCompiler txInitCode =
        callParameter.customPrecompileCallsSeparatedByReturnDataWipingOperation();
    if (callParameter.willRevert()) revertWith(txInitCode, 0, 0);

    runDeploymentTransactionWithProvidedCodeAsInitCode(txInitCode);
  }

  /**
   * <b>MESSAGE_CALL_FROM_ROOT</b> case.
   *
   * <p>See {@link CodeExecutionMethods} for documentation and context.
   */
  @ParameterizedTest
  @MethodSource("parameterGeneration")
  public void messageCallFromRootTest(T callParameter) {
    BytecodeCompiler chadPrcEnjoyerCode =
        callParameter.customPrecompileCallsSeparatedByReturnDataWipingOperation();
    runMessageCallToAccountEndowedWithProvidedCode(chadPrcEnjoyerCode, callParameter.willRevert());
  }

  /**
   * <b>DURING_DEPLOYMENT</b> case.
   *
   * <p>See {@link CodeExecutionMethods} for documentation and context.
   *
   * <p>The {@link CodeExecutionMethods#root} contract fully copies the code of the account whose
   * address is in the {@link CodeExecutionMethods#transaction} call data. This account is the
   * {@link CodeExecutionMethods#chadPrcEnjoyer}. That code is then used as the initialization code
   * of a <b>CREATE</b>. The whole operation optionally <b>REVERT</b>'s.
   */
  @ParameterizedTest
  @MethodSource("parameterGeneration")
  public void happyPathDuringCreate(T callParameter) {
    BytecodeCompiler foreignCode =
        callParameter.customPrecompileCallsSeparatedByReturnDataWipingOperation();
    runForeignByteCodeAsInitCode(foreignCode, callParameter.willRevert());
  }

  /**
   * <b>AFTER_DEPLOYMENT</b> case.
   *
   * <p>See {@link CodeExecutionMethods} for documentation and context.
   */
  @ParameterizedTest
  @MethodSource("parameterGeneration")
  public void happyPathAfterCreate(T callParameter) {
    BytecodeCompiler chadPrcEnjoyerCode =
        callParameter.customPrecompileCallsSeparatedByReturnDataWipingOperation();
    runCreateDeployingForeignCodeAndCallIntoIt(chadPrcEnjoyerCode, callParameter.willRevert());
  }
}
