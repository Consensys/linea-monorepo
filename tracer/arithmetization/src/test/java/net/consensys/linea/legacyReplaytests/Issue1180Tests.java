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
package net.consensys.linea.legacyReplaytests;

import static net.consensys.linea.ReplayTestTools.replay;
import static net.consensys.linea.zktracer.ChainConfig.OLD_MAINNET_TESTCONFIG;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.junit.jupiter.api.Disabled;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.TestInfo;
import org.junit.jupiter.api.extension.ExtendWith;

/** This range broke the MOD module's mod.set-absolute-values constraint. */
@Tag("replay")
@Tag("weekly")
@Disabled
@ExtendWith(UnitTestWatcher.class)
public class Issue1180Tests extends TracerTestBase {

  @Test
  void split_range_2321470_2321479(TestInfo testInfo) {
    replay(OLD_MAINNET_TESTCONFIG, "legacy/2321470-2321479.mainnet.json.gz", testInfo);
  }

  @Test
  void failingSmodInstructionTest(TestInfo testInfo) {
    BytecodeCompiler program = BytecodeCompiler.newProgram(chainConfig);
    program
        .push("ffffffffffffffffffffffffffffffffffffffffffffffffffdc633cace676d7")
        .push("0000000000000000000000000000000000000000000000000000000000000000")
        .op(OpCode.SDIV);
    BytecodeRunner.of(program.compile()).run(chainConfig, testInfo);
  }
}
