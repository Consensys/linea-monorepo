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
package net.consensys.linea.replaytests;

import static net.consensys.linea.replaytests.ReplayTestTools.replay;
import static net.consensys.linea.zktracer.ChainConfig.OLD_MAINNET_TESTCONFIG;

import net.consensys.linea.UnitTestWatcher;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

/** This range broke the MOD module's mod.set-absolute-values constraint. */
@Tag("nightly")
@Tag("replay")
@ExtendWith(UnitTestWatcher.class)
public class Issue1180Tests {

  @Test
  void split_range_2321470_2321479() {
    replay(OLD_MAINNET_TESTCONFIG, "2321470-2321479.mainnet.json.gz");
  }

  @Test
  void failingSmodInstructionTest() {
    BytecodeCompiler program = BytecodeCompiler.newProgram();
    program
        .push("ffffffffffffffffffffffffffffffffffffffffffffffffffdc633cace676d7")
        .push("0000000000000000000000000000000000000000000000000000000000000000")
        .op(OpCode.SDIV);
    BytecodeRunner.of(program.compile()).run();
  }
}
