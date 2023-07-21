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

import java.util.List;

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.module.ModuleTracer;
import org.hyperledger.besu.evm.frame.MessageFrame;

public class HubTracer implements ModuleTracer {
  @Override
  public String jsonKey() {
    return "hub";
  }

  @Override
  public List<OpCode> supportedOpCodes() {
    return List.of(OpCode.LT, OpCode.GT, OpCode.SLT, OpCode.SGT, OpCode.EQ, OpCode.ISZERO);
  }

  @Override
  public Object trace(final MessageFrame frame) {
    //      final OpCode opCode = OpCode.of(famelgetcurrentOperation().getOpcode());
    return null;
  }
}
