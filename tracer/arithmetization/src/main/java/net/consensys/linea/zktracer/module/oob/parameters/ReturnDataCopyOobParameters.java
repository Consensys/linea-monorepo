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
public class ReturnDataCopyOobParameters implements OobParameters {
  private final EWord offset;
  private final EWord size;
  private final BigInteger rds;
  @Setter private boolean rdcx;

  public BigInteger offsetHi() {
    return offset.hiBigInt();
  }

  public BigInteger offsetLo() {
    return offset.loBigInt();
  }

  public BigInteger sizeHi() {
    return size.hiBigInt();
  }

  public BigInteger sizeLo() {
    return size.loBigInt();
  }

  @Override
  public Trace trace(Trace trace) {
    return trace
        .data1(bigIntegerToBytes(offsetHi()))
        .data2(bigIntegerToBytes(offsetLo()))
        .data3(bigIntegerToBytes(sizeHi()))
        .data4(bigIntegerToBytes(sizeLo()))
        .data5(bigIntegerToBytes(rds))
        .data6(ZERO)
        .data7(booleanToBytes(rdcx))
        .data8(ZERO);
  }
}
