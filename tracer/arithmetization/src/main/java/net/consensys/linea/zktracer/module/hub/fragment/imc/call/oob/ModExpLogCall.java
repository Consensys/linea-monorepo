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

package net.consensys.linea.zktracer.module.hub.fragment.imc.call.oob;

import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.TraceSubFragment;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;

public record ModExpLogCall(EWord rawLeadingWord, int cdsCutoff, int ebsCutoff)
    implements TraceSubFragment {
  public static int exponentLeadingWordLog(final Bytes exponent, int cdsCutoff, int ebsCutoff) {
    final int minCutoff = Math.min(cdsCutoff, ebsCutoff);
    final int leadBitLength =
        exponent.shiftRight(32 - minCutoff).shiftLeft(ebsCutoff - minCutoff).bitLength();

    return Math.max(leadBitLength - 1, 0);
  }

  @Override
  public Trace trace(Trace trace) {
    return trace
        .pMiscellaneousExpFlag(true)
        .pMiscellaneousExpData1(rawLeadingWord.hi())
        .pMiscellaneousExpData2(rawLeadingWord.lo())
        .pMiscellaneousExpData3(Bytes.ofUnsignedShort(cdsCutoff))
        .pMiscellaneousExpData4(Bytes.ofUnsignedShort(ebsCutoff))
        .pMiscellaneousExpData5(
            Bytes.ofUnsignedShort(exponentLeadingWordLog(rawLeadingWord, cdsCutoff, ebsCutoff)));
  }
}
