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

package net.consensys.linea.zktracer.module.trm;

import java.util.List;

import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.OpCodes;
import org.hyperledger.besu.evm.frame.MessageFrame;

public class Trm implements Module {
  @Override
  public String jsonKey() {
    return "trm";
  }

  @Override
  public final List<OpCodeData> supportedOpCodes() {
    return OpCodes.of(
        OpCode.BALANCE,
        OpCode.EXTCODESIZE,
        OpCode.EXTCODECOPY,
        OpCode.EXTCODEHASH,
        OpCode.CALL,
        OpCode.CALLCODE,
        OpCode.DELEGATECALL,
        OpCode.STATICCALL,
        OpCode.SELFDESTRUCT);
  }

  @Override
  public void trace(MessageFrame frame) {}

  @Override
  public Object commit() {
    return null;
  }
}
