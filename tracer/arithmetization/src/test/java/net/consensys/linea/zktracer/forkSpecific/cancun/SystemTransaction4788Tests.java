/*
 * Copyright ConsenSys Inc.
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

package net.consensys.linea.zktracer.forkSpecific.cancun;

import static net.consensys.linea.testing.ToyExecutionEnvironmentV2.DEFAULT_TIME_STAMP;
import static net.consensys.linea.zktracer.module.hub.section.systemTransaction.EIP4788BeaconBlockRoot.BEACONROOT_ADDRESS;

import net.consensys.linea.reporting.TracerTestBase;
import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.testing.BytecodeRunner;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.junit.jupiter.api.Test;

public class SystemTransaction4788Tests extends TracerTestBase {

  // This test checks the consistency of system account by calling the system account
  @Test
  void systemTransactionConsistencyTest() {
    BytecodeRunner.of(
            BytecodeCompiler.newProgram(testInfo)
                // prepare memory with TIMESTAMP left padded
                .push(Bytes32.leftPad(Bytes.minimalBytes(DEFAULT_TIME_STAMP))) // value
                .push(0) // offset
                .op(OpCode.MSTORE)
                // call system contract
                .push(0) // retSize
                .push(0) // retOffset
                .push(32) // argSize
                .push(0) // argOffset
                .push(0) // value
                .push(BEACONROOT_ADDRESS) // address
                .push(757575) // gas
                .op(OpCode.CALL)
                // just a stupid memory storing to check memory consistency
                .push(1) // offset
                .op(OpCode.MSTORE)
                .compile())
        .run(testInfo);
  }
}
