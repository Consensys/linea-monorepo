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

package net.consensys.linea.zktracer.module.alu.add;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytestheta.BaseBytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;

/** A module responsible for addition operations. */
@Slf4j
public class Adder {

  /**
   * Performs addition or subtraction based on the opcode.
   *
   * @param opCode accepts {@link OpCode#ADD} or {@link OpCode#SUB}.
   * @param arg1 left argument.
   * @param arg2 right argument.
   * @return {@link BaseBytes} - 256-bit (32-byte) blocks data.
   */
  public static BaseBytes addSub(final OpCode opCode, final Bytes32 arg1, final Bytes32 arg2) {
    log.info("adding " + arg1 + " " + opCode.name() + " " + arg2);

    final BaseBytes resBytes = performOperation(opCode, arg1, arg2);

    // ensure result is correct length
    return BaseBytes.fromBytes32(resBytes.getBytes32());
  }

  private static BaseBytes performOperation(
      final OpCode opCode, final Bytes32 arg1, final Bytes32 arg2) {
    return switch (opCode) {
      case ADD -> BaseBytes.fromBytes32(UInt256.fromBytes(arg1).add(UInt256.fromBytes(arg2)));
      case SUB -> BaseBytes.fromBytes32(UInt256.fromBytes(arg1).subtract(UInt256.fromBytes(arg2)));
      default -> throw new RuntimeException("Modular arithmetic was given wrong opcode");
    };
  }
}
