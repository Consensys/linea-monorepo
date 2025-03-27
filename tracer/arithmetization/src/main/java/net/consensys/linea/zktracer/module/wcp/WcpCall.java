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

package net.consensys.linea.zktracer.module.wcp;

import static net.consensys.linea.zktracer.Trace.*;

import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

@Getter()
@Accessors(fluent = true)
public class WcpCall {
  private final UnsignedByte instruction;
  private final Bytes arg1Hi;
  private final Bytes arg1Lo;
  private final Bytes arg2Hi;
  private final Bytes arg2Lo;
  private final boolean result;

  public WcpCall(Wcp wcp, byte instruction, Bytes arg1, Bytes arg2) {
    final Bytes32 arg1Bytes32 = Bytes32.leftPad(arg1);
    final Bytes32 arg2Bytes32 = Bytes32.leftPad(arg2);
    this.instruction = UnsignedByte.of(instruction);
    this.arg1Hi = arg1Bytes32.slice(0, LLARGE);
    this.arg1Lo = arg1Bytes32.slice(LLARGE, LLARGE);
    this.arg2Hi = arg2Bytes32.slice(0, LLARGE);
    this.arg2Lo = arg2Bytes32.slice(LLARGE, LLARGE);
    this.result =
        switch (instruction) {
          case EVM_INST_LT -> wcp.callLT(arg1Bytes32, arg2Bytes32);
          case EVM_INST_EQ -> wcp.callEQ(arg1Bytes32, arg2Bytes32);
          case EVM_INST_ISZERO -> wcp.callISZERO(arg1Bytes32);
          case EVM_INST_GT -> wcp.callGT(arg1Bytes32, arg2Bytes32);
          case WCP_INST_LEQ -> wcp.callLEQ(arg1Bytes32, arg2Bytes32);
          case WCP_INST_GEQ -> wcp.callGEQ(arg1Bytes32, arg2Bytes32);
          default -> throw new IllegalArgumentException(
              "Unexpected wcp instruction: " + instruction);
        };
  }

  public static WcpCall ltCall(Wcp wcp, Bytes arg1, Bytes arg2) {
    return new WcpCall(wcp, (byte) EVM_INST_LT, arg1, arg2);
  }

  public static WcpCall leqCall(Wcp wcp, Bytes arg1, Bytes arg2) {
    return new WcpCall(wcp, (byte) WCP_INST_LEQ, arg1, arg2);
  }

  public static WcpCall isZeroCall(Wcp wcp, Bytes arg1) {
    return new WcpCall(wcp, (byte) EVM_INST_ISZERO, arg1, Bytes.EMPTY);
  }
}
