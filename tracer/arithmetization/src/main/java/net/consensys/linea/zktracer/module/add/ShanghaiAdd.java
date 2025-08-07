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

package net.consensys.linea.zktracer.module.add;

import static net.consensys.linea.zktracer.opcode.OpCode.ADD;

import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.module.Module;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;

/** Implementation of a {@link Module} for addition/subtraction. */
@RequiredArgsConstructor
@Getter
@Accessors(fluent = true)
public class ShanghaiAdd extends Add {
  @Override
  protected void addOperation(OpCode opcode, Bytes32 arg1, Bytes32 arg2) {
    operations.add(new Operation(opcode, arg1, arg2));
  }

  @Accessors(fluent = true)
  @EqualsAndHashCode(onlyExplicitlyIncluded = true, callSuper = true)
  private static final class Operation extends AddOperation {
    private Operation(OpCode opCode, Bytes32 arg1, Bytes32 arg2) {
      super(opCode, arg1, arg2);
    }

    public void trace(int stamp, Trace.Add trace) {
      UInt256 res;
      // Compute result of this operation.
      if (opCode == ADD) {
        res = UInt256.fromBytes(arg1).add(UInt256.fromBytes(arg2));
      } else {
        res = UInt256.fromBytes(arg1).subtract(UInt256.fromBytes(arg2));
      }
      // Trace it
      trace
          .arg1(arg1)
          .arg2(arg2)
          .inst(UnsignedByte.of(opCode.byteValue() & 0xff))
          .res(res)
          .validateRow();
    }

    @Override
    protected int computeLineCount() {
      return 1;
    }
  }
}
