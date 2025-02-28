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

package net.consensys.linea.zktracer.module.trm;

import static net.consensys.linea.zktracer.Trace.LLARGE;
import static net.consensys.linea.zktracer.module.trm.Trm.MAX_CT;
import static net.consensys.linea.zktracer.module.trm.Trm.PIVOT_BIT_FLIPS_TO_TRUE;
import static net.consensys.linea.zktracer.types.AddressUtils.isPrecompile;
import static net.consensys.linea.zktracer.types.Utils.bitDecomposition;
import static net.consensys.linea.zktracer.types.Utils.leftPadTo;

import java.util.List;

import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;

@Accessors(fluent = true)
@RequiredArgsConstructor
@EqualsAndHashCode(onlyExplicitlyIncluded = true, callSuper = false)
public class TrmOperation extends ModuleOperation {
  @EqualsAndHashCode.Include @Getter private final EWord rawAddress;

  void trace(Trace.Trm trace, final int stamp) {
    final Bytes trmHiBytes =
        leftPadTo(this.rawAddress.hi().slice(PIVOT_BIT_FLIPS_TO_TRUE, 4), LLARGE);
    final long trmHi = trmHiBytes.slice(PIVOT_BIT_FLIPS_TO_TRUE, 4).toLong();
    final Boolean isPrec = isPrecompile(Address.extract(this.rawAddress));
    final int accLastByte =
        isPrec ? 9 - (0xff & this.rawAddress.get(31)) : (0xff & this.rawAddress.get(31)) - 10;
    List<Boolean> ones = bitDecomposition(accLastByte, MAX_CT).bitDecList();

    for (int ct = 0; ct < MAX_CT; ct++) {
      trace
          .ct(ct)
          .stamp(stamp)
          .isPrecompile(isPrec)
          .plateauBit(ct >= PIVOT_BIT_FLIPS_TO_TRUE)
          .rawAddressHi(this.rawAddress.hi())
          .rawAddressLo(this.rawAddress.lo())
          .trmAddressHi(trmHi)
          .accHi(this.rawAddress.hi().slice(0, ct + 1))
          .accLo(this.rawAddress.lo().slice(0, ct + 1))
          .accT(trmHiBytes.slice(0, ct + 1).trimLeadingZeros().toLong())
          .byteHi(UnsignedByte.of(this.rawAddress.hi().get(ct)))
          .byteLo(UnsignedByte.of(this.rawAddress.lo().get(ct)))
          .one(ones.get(ct))
          .validateRow();
    }
  }

  @Override
  protected int computeLineCount() {
    return MAX_CT;
  }
}
