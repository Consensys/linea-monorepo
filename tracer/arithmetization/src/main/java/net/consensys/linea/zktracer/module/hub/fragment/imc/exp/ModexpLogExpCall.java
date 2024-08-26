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

package net.consensys.linea.zktracer.module.hub.fragment.imc.exp;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EXP_INST_MODEXPLOG;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import java.math.BigInteger;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.Setter;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.precompiles.ModexpMetadata;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;

@Setter
@Getter
@RequiredArgsConstructor
public class ModexpLogExpCall implements ExpCall {
  final ModexpMetadata modexpMetadata;
  EWord rawLeadingWord;
  int cdsCutoff;
  int ebsCutoff;
  BigInteger leadLog;

  @Override
  public int expInstruction() {
    return EXP_INST_MODEXPLOG;
  }

  @Override
  public Trace trace(Trace trace) {
    return trace
        .pMiscExpFlag(true)
        .pMiscExpData1(rawLeadingWord.hi())
        .pMiscExpData2(rawLeadingWord.lo())
        .pMiscExpData3(Bytes.ofUnsignedShort(cdsCutoff))
        .pMiscExpData4(Bytes.ofUnsignedShort(ebsCutoff))
        .pMiscExpData5(bigIntegerToBytes(this.leadLog));
  }
}
