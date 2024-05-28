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
public class ModexpExtractParameters implements OobParameters {
  private final BigInteger cds;
  private final BigInteger bbs;
  private final BigInteger ebs;
  private final BigInteger mbs;

  @Setter private boolean extractBase;
  @Setter private boolean extractExponent;
  @Setter private boolean extractModulus;

  @Override
  public Trace trace(Trace trace) {
    return trace
        .data1(ZERO)
        .data2(bigIntegerToBytes(cds))
        .data3(bigIntegerToBytes(bbs))
        .data4(bigIntegerToBytes(ebs))
        .data5(bigIntegerToBytes(mbs))
        .data6(booleanToBytes(extractBase))
        .data7(booleanToBytes(extractExponent))
        .data8(booleanToBytes(extractModulus));
  }
}
