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
package net.consensys.linea.zktracer.instructionprocessing.callTests.prc.modexp;

import static net.consensys.linea.zktracer.instructionprocessing.callTests.prc.CodeExecutionMethods.*;
import static net.consensys.linea.zktracer.instructionprocessing.callTests.prc.modexp.ByteSizeParameter.*;
import static net.consensys.linea.zktracer.opcode.OpCode.*;

import java.util.stream.Stream;

import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.zktracer.instructionprocessing.callTests.prc.*;
import net.consensys.linea.zktracer.instructionprocessing.callTests.prc.ecmul.ParameterGeneration;
import net.consensys.linea.zktracer.instructionprocessing.callTests.prc.framework.PrecompileCallTests;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.params.provider.Arguments;

@Disabled
@Tag("weekly")
public class Tests extends PrecompileCallTests<CallParameters> {

  public static Stream<Arguments> parameterGeneration() {
    return ParameterGeneration.parameterGeneration();
  }

  /** Non-parametric test to make sure things are working as expected. */
  @Test
  public void singleMessageCallTransactionTest() {
    MemoryContents memoryContents =
        new MemoryContents(
            MODERATE, // bbs
            MODERATE, // ebs
            MAX, // mbs
            CallDataSizeParameter.MODULUS_FULL // cds
            );
    CallParameters params =
        new CallParameters(
            CALL,
            GasParameter.COST_MO,
            memoryContents,
            ReturnAtParameter.FULL,
            RelativeRangePosition.OVERLAP,
            true);

    BytecodeCompiler rootCode = params.customPrecompileCallsSeparatedByReturnDataWipingOperation();
    runMessageCallTransactionWithProvidedCodeAsRootCode(rootCode);
  }

  /** Non-parametric test to make sure things are working as expected. */
  @Test
  public void singleMessageCallTransactionTest2() {
    MemoryContents memoryContents =
        new MemoryContents(
            MAX, // bbs
            SHORT, // ebs
            MODERATE, // mbs
            CallDataSizeParameter.MODULUS_FULL // cds
            );
    CallParameters params =
        new CallParameters(
            STATICCALL,
            GasParameter.COST,
            memoryContents,
            ReturnAtParameter.FULL,
            RelativeRangePosition.OVERLAP,
            true);

    BytecodeCompiler rootCode = params.customPrecompileCallsSeparatedByReturnDataWipingOperation();
    runMessageCallTransactionWithProvidedCodeAsRootCode(rootCode);
  }
}
