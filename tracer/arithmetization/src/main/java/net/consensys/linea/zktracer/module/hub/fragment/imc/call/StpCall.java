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

package net.consensys.linea.zktracer.module.hub.fragment.imc.call;

import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.fragment.TraceSubFragment;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;

public record StpCall(
    byte opCode,
    EWord gas,
    EWord value,
    boolean exists,
    boolean warm,
    boolean outOfGasException,
    long upfront,
    long outOfPocket,
    long stipend)
    implements TraceSubFragment {
  @Override
  public Trace trace(Trace trace) {
    return trace
        .pMiscellaneousStpFlag(true)
        .pMiscellaneousStpInst(Bytes.of(opCode))
        .pMiscellaneousStpGasHi(gas.hi())
        .pMiscellaneousStpGasLo(gas.lo())
        .pMiscellaneousStpValHi(value.hi())
        .pMiscellaneousStpValLo(value.lo())
        .pMiscellaneousStpExists(exists)
        .pMiscellaneousStpWarmth(warm)
        .pMiscellaneousStpOogx(outOfGasException)
        .pMiscellaneousStpGasUpfrontGasCost(Bytes.ofUnsignedLong(upfront))
        .pMiscellaneousStpGasPaidOutOfPocket(Bytes.ofUnsignedLong(outOfPocket))
        .pMiscellaneousStpGasStipend(Bytes.ofUnsignedLong(stipend));
  }
}
