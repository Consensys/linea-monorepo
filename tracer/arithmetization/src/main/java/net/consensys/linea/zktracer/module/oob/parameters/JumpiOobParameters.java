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

package net.consensys.linea.zktracer.module.oob.parameters;

import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;
import static net.consensys.linea.zktracer.types.Conversions.booleanToBytes;

import java.math.BigInteger;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.Setter;
import net.consensys.linea.zktracer.module.oob.Trace;
import net.consensys.linea.zktracer.types.EWord;

@Getter
@RequiredArgsConstructor
public class JumpiOobParameters implements OobParameters {
  private final EWord pcNew;
  private final EWord jumpCondition;
  private final BigInteger codeSize;
  @Setter boolean jumpNotAttempted;
  @Setter boolean jumpGuanranteedException;
  @Setter boolean jumpMustBeAttempted;

  public BigInteger pcNewHi() {
    return pcNew.hiBigInt();
  }

  public BigInteger pcNewLo() {
    return pcNew.loBigInt();
  }

  public BigInteger jumpConditionHi() {
    return jumpCondition.hiBigInt();
  }

  public BigInteger jumpConditionLo() {
    return jumpCondition.loBigInt();
  }

  @Override
  public Trace trace(Trace trace) {
    return trace
        .data1(bigIntegerToBytes(pcNewHi()))
        .data2(bigIntegerToBytes(pcNewLo()))
        .data3(bigIntegerToBytes(jumpConditionHi()))
        .data4(bigIntegerToBytes(jumpConditionLo()))
        .data5(bigIntegerToBytes(codeSize))
        .data6(booleanToBytes(jumpNotAttempted))
        .data7(booleanToBytes(jumpGuanranteedException))
        .data8(booleanToBytes(jumpMustBeAttempted));
  }
}
