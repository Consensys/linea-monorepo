package net.consensys.linea.zktracer.module.alu.add;
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

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytestheta.BaseBytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class Adder {
  private static final Logger LOG = LoggerFactory.getLogger(Adder.class);

  public static BaseBytes addSub(final OpCode opCode, final Bytes32 arg1, final Bytes32 arg2) {
    LOG.info("adding " + arg1 + " " + opCode.name() + " " + arg2);
    final BaseBytes resBytes = x(opCode, arg1, arg2);
    // ensure result is correct length
    return BaseBytes.fromBytes32(resBytes.getBytes32());
  }

  private static BaseBytes x(final OpCode opCode, final Bytes32 arg1, final Bytes32 arg2) {
    {
      return switch (opCode) {
        case ADD -> BaseBytes.fromBytes32(UInt256.fromBytes(arg1).add(UInt256.fromBytes(arg2)));
        case SUB -> BaseBytes.fromBytes32(
            UInt256.fromBytes(arg1).subtract(UInt256.fromBytes(arg2)));
        default -> throw new RuntimeException("Modular arithmetic was given wrong opcode");
      };
    }
  }
}
