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
package net.consensys.linea.zktracer.instructionprocessing;

import net.consensys.linea.testing.BytecodeCompiler;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;

public class StpTests {

  private void forceWarmRecipient(
      BytecodeCompiler bytecode, Address recipient, boolean forceWarmth) {
    if (forceWarmth) {
      bytecode.push(recipient).op(OpCode.BALANCE).op(OpCode.POP);
    }
  }

  //  private void appendCall(BytecodeCompiler bytecode, Address recipient, Wei value) {
  //
  //      bytecode
  //              .push() // return at capacity
  //              .push() // return at offset
  //              .push() // call data size
  //              .push() // call data offset
  //              .push() // value
  //              .push() // address
  //              .push() // gas
  //  }
  private void appendCallcode(BytecodeCompiler bytecode, Address recipient, Wei value) {}

  private void appendDelegatecall(BytecodeCompiler bytecode, Address recipient) {}

  private void appendStaticcall(BytecodeCompiler bytecode, Address recipient, Wei value) {}
}
