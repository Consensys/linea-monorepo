/*
 * Copyright ConsenSys AG.
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

package net.consensys.linea.zktracer.module.hub;

import static org.assertj.core.api.AssertionsForClassTypes.assertThat;

import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.testutils.BytecodeCompiler;
import net.consensys.linea.zktracer.testutils.EvmExtension;
import net.consensys.linea.zktracer.testutils.TestCodeExecutor;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;

@ExtendWith(EvmExtension.class)
public class TwoPlusTwo {
  @Test
  void testAdd() {
    TestCodeExecutor.builder()
        .byteCode(BytecodeCompiler.newProgram().push(32).push(27).op(OpCode.ADD).compile())
        .frameAssertions(
            (frame) -> assertThat(frame.getState()).isEqualTo(MessageFrame.State.COMPLETED_SUCCESS))
        .build()
        .run();
  }
}
