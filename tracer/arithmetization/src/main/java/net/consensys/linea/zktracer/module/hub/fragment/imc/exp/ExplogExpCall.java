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

package net.consensys.linea.zktracer.module.hub.fragment.imc.exp;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EXP_INST_EXPLOG;

import lombok.Getter;
import lombok.Setter;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;

@Getter
@Setter
public class ExplogExpCall implements ExpCall {
  EWord exponent;
  long dynCost;

  @Override
  public int expInstruction() {
    return EXP_INST_EXPLOG;
  }

  @Override
  public Trace trace(Trace trace) {
    return trace
        .pMiscExpFlag(true)
        .pMiscExpInst(EXP_INST_EXPLOG)
        .pMiscExpData1(exponent.hi())
        .pMiscExpData2(exponent.lo())
        .pMiscExpData5(Bytes.ofUnsignedLong(this.dynCost));
  }
}
