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
import org.apache.tuweni.bytes.Bytes;

@Getter
@RequiredArgsConstructor
public class ModexpPricingParameters implements OobParameters {
  private final BigInteger callGas;
  private final BigInteger returnAtCapacity;
  @Setter private boolean success;
  private final BigInteger exponentLog;
  private final int maxMbsBbs;

  @Setter private BigInteger returnGas;
  @Setter private boolean returnAtCapacityNonZero;

  @Override
  public Trace trace(Trace trace) {
    return trace
        .data1(bigIntegerToBytes(callGas))
        .data2(ZERO)
        .data3(bigIntegerToBytes(returnAtCapacity))
        .data4(booleanToBytes(success))
        .data5(bigIntegerToBytes(returnGas))
        .data6(bigIntegerToBytes(exponentLog))
        .data7(Bytes.of(maxMbsBbs))
        .data8(booleanToBytes(returnAtCapacityNonZero));
  }
}
