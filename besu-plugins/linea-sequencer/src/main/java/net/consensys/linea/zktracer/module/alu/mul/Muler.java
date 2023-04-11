package net.consensys.linea.zktracer.module.alu.mul;
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
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;

public class Muler {

  public static UInt256 operate(final OpCode opCode, final Bytes32 arg1, final Bytes32 arg2) {
    return switch (opCode) {
      case MUL -> UInt256.fromBytes(arg1).multiply(UInt256.fromBytes(arg2));
      case EXP -> UInt256.fromBytes(arg1).pow(UInt256.fromBytes(arg2));
      default -> UInt256.ZERO;
    };
  }
}
