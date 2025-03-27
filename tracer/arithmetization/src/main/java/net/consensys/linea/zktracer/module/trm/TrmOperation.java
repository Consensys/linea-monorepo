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

import static net.consensys.linea.zktracer.Trace.*;
import static net.consensys.linea.zktracer.Trace.Trm.TRM_CT_MAX;
import static net.consensys.linea.zktracer.Trace.Trm.TRM_NB_ROWS;
import static net.consensys.linea.zktracer.module.wcp.WcpCall.*;
import static net.consensys.linea.zktracer.types.AddressUtils.isPrecompile;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import java.util.ArrayList;
import java.util.List;

import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.module.wcp.WcpCall;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;

@Accessors(fluent = true)
@EqualsAndHashCode(onlyExplicitlyIncluded = true, callSuper = false)
public class TrmOperation extends ModuleOperation {
  @EqualsAndHashCode.Include @Getter private final EWord rawAddress;
  private final List<WcpCall> wcpCalls = new ArrayList<>(TRM_NB_ROWS);
  private static final Bytes TWOFIFTYSIX_TO_THE_TWENTY_BYTES =
      bigIntegerToBytes(TWOFIFTYSIX_TO_THE_TWENTY);
  private static final Bytes TWOFIFTYSIX_TO_THE_TWELVE_MO_BYTES =
      bigIntegerToBytes(TWOFIFTYSIX_TO_THE_TWELVE_MO);

  public TrmOperation(EWord rawAddress, Wcp wcp) {
    this.rawAddress = rawAddress;
    final Bytes trmAddress = rawAddress.toAddress();

    wcpCalls.add(0, ltCall(wcp, trmAddress, TWOFIFTYSIX_TO_THE_TWENTY_BYTES));
    wcpCalls.add(1, leqCall(wcp, rawAddress.slice(0, 12), TWOFIFTYSIX_TO_THE_TWELVE_MO_BYTES));
    wcpCalls.add(2, isZeroCall(wcp, trmAddress));
    wcpCalls.add(3, leqCall(wcp, trmAddress, Bytes.ofUnsignedShort(MAX_PRC_ADDRESS)));
  }

  void trace(Trace.Trm trace) {
    final Address trmAddress = rawAddress.toAddress();
    final boolean isPrec = isPrecompile(trmAddress);
    final long trmAddrHi = trmAddress.slice(0, 4).toLong();

    for (int ct = 0; ct <= TRM_CT_MAX; ct++) {
      trace
          .iomf(true)
          .first(ct == 0)
          .ct(ct)
          .isPrecompile(isPrec)
          .rawAddressHi(rawAddress.hi())
          .rawAddressLo(rawAddress.lo())
          .trmAddressHi(trmAddrHi)
          .inst(wcpCalls.get(ct).instruction())
          .arg1Hi(wcpCalls.get(ct).arg1Hi())
          .arg1Lo(wcpCalls.get(ct).arg1Lo())
          .arg2Hi(wcpCalls.get(ct).arg2Hi())
          .arg2Lo(wcpCalls.get(ct).arg2Lo())
          .res(wcpCalls.get(ct).result())
          .validateRow();
    }
  }

  @Override
  protected int computeLineCount() {
    return TRM_NB_ROWS;
  }
}
