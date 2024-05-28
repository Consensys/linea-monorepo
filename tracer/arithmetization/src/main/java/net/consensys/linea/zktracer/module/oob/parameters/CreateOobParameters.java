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

import java.math.BigInteger;

import net.consensys.linea.zktracer.module.oob.Trace;
import net.consensys.linea.zktracer.types.EWord;

public record CreateOobParameters(
    EWord val, BigInteger bal, BigInteger nonce, boolean hasCode, BigInteger csd)
    implements OobParameters {

  public BigInteger valHi() {
    return val.hiBigInt();
  }

  public BigInteger valLo() {
    return val.loBigInt();
  }

  @Override
  public Trace trace(Trace trace) {
    return trace
        .data1(bigIntegerToBytes(valHi()))
        .data2(bigIntegerToBytes(valLo()))
        .data3(bigIntegerToBytes(bal))
        .data4(bigIntegerToBytes(nonce))
        .data5((hasCode ? ONE : ZERO))
        .data6(bigIntegerToBytes(csd))
        .data7(ZERO) // TODO: temporary value; to fill when oob update is complete
        .data8(ZERO); // TODO: temporary value; to fill when oob update is complete
  }
}
