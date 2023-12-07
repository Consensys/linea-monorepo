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

package net.consensys.linea.zktracer.module.bin;

import com.google.common.base.Objects;
import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.bytestheta.BaseBytes;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.Bytes16;

@Getter
@Accessors(fluent = true)
public class BinOperation {
  private static final int LIMB_SIZE = 16;

  private final OpCode opCode;
  private final BaseBytes arg1;
  private final BaseBytes arg2;

  final Bytes16 arg1Hi;

  public BinOperation(OpCode opCode, BaseBytes arg1, BaseBytes arg2) {
    this.opCode = opCode;
    this.arg1 = arg1;
    this.arg2 = arg2;

    arg1Hi = arg1.getHigh();
  }

  @Override
  public int hashCode() {
    return Objects.hashCode(this.opCode, this.arg1, this.arg2);
  }

  @Override
  public boolean equals(Object o) {
    if (this == o) return true;
    if (o == null || getClass() != o.getClass()) return false;
    final BinOperation that = (BinOperation) o;
    return java.util.Objects.equals(opCode, that.opCode)
        && java.util.Objects.equals(arg1, that.arg1)
        && java.util.Objects.equals(arg2, that.arg2);
  }

  public boolean isOneLineInstruction() {
    return (opCode == OpCode.BYTE || opCode == OpCode.SIGNEXTEND) && !arg1Hi.isZero();
  }

  public int maxCt() {
    return isOneLineInstruction() ? 1 : LIMB_SIZE;
  }
}
