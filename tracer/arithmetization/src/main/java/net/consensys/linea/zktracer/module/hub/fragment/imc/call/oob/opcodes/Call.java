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

package net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.opcodes;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.OOB_INST_CALL;
import static net.consensys.linea.zktracer.types.Conversions.booleanToBytes;

import net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob.OobCall;
import net.consensys.linea.zktracer.module.oob.OobDataChannel;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;

public record Call(EWord value, EWord balance, int callStackDepth, boolean hasAbort)
    implements OobCall {
  @Override
  public Bytes data(OobDataChannel i) {
    return switch (i) {
      case DATA_1 -> value.hi();
      case DATA_2 -> value.lo();
      case DATA_3 -> balance.lo();
      case DATA_6 -> Bytes.ofUnsignedLong(callStackDepth);
      case DATA_7 -> booleanToBytes(!value.isZero());
      case DATA_8 -> booleanToBytes(hasAbort);
      default -> Bytes.EMPTY;
    };
  }

  @Override
  public int oobInstruction() {
    return OOB_INST_CALL;
  }
}
