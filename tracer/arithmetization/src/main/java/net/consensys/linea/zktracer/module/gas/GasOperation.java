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

package net.consensys.linea.zktracer.module.gas;

import static net.consensys.linea.zktracer.module.gas.Trace.CT_MAX;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;
import static net.consensys.linea.zktracer.types.Conversions.booleanToInt;

import java.math.BigInteger;

import lombok.EqualsAndHashCode;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;

@EqualsAndHashCode(onlyExplicitlyIncluded = true, callSuper = false)
public class GasOperation extends ModuleOperation {
  @EqualsAndHashCode.Include GasParameters gasParameters;
  Bytes acc1;
  Bytes acc2;

  public GasOperation(GasParameters gasParameters) {
    this.gasParameters = gasParameters;
    acc1 = bigIntegerToBytes(gasParameters.gasActl());
    acc2 =
        bigIntegerToBytes(
            (BigInteger.valueOf((2L * booleanToInt(gasParameters.oogx()) - 1))
                    .multiply(gasParameters.gasCost().subtract(gasParameters.gasActl())))
                .subtract(BigInteger.valueOf(booleanToInt(gasParameters.oogx()))));
  }

  @Override
  protected int computeLineCount() {
    return CT_MAX + 1;
  }

  public void trace(int stamp, Trace trace) {
    for (short i = 0; i < CT_MAX + 1; i++) {
      trace
          .stamp(stamp)
          .ct(i)
          .gasActl(gasParameters.gasActl().longValue())
          .gasCost(bigIntegerToBytes(gasParameters.gasCost()))
          .oogx(gasParameters.oogx())
          .byte1(UnsignedByte.of(acc1.get(i)))
          .byte2(UnsignedByte.of(acc2.get(i)))
          .acc1(acc1.slice(0, i + 1))
          .acc2(acc2.slice(0, i + 1))
          .validateRow();
    }
  }
}
