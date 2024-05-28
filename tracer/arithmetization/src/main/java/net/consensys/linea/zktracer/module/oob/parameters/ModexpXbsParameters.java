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

package net.consensys.linea.zktracer.module.oob.parameters;

import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;
import static net.consensys.linea.zktracer.types.Conversions.booleanToBytes;

import java.math.BigInteger;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.Setter;
import net.consensys.linea.zktracer.module.oob.Trace;

@Getter
@RequiredArgsConstructor
public class ModexpXbsParameters implements OobParameters {
  private final BigInteger xbsHi;
  private final BigInteger xbsLo;
  private final BigInteger ybsLo;
  private final boolean computeMax;

  @Setter private BigInteger maxXbsYbs;
  @Setter private boolean xbsNonZero;

  @Override
  public Trace trace(Trace trace) {
    return trace
        .data1(bigIntegerToBytes(xbsHi))
        .data2(bigIntegerToBytes(xbsLo))
        .data3(bigIntegerToBytes(ybsLo))
        .data4(booleanToBytes(computeMax))
        .data5(ZERO)
        .data6(ZERO)
        .data7(bigIntegerToBytes(maxXbsYbs))
        .data8(booleanToBytes(xbsNonZero));
  }
}
