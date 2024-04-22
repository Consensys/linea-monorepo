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

package net.consensys.linea.zktracer.module.mmio.instructions;

import static net.consensys.linea.zktracer.types.Bytecodes.readLimb;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.module.mmio.MmioData;
import net.consensys.linea.zktracer.module.mmu.MmuData;
import net.consensys.linea.zktracer.module.mmu.values.MmuToMmioConstantValues;
import net.consensys.linea.zktracer.module.mmu.values.MmuToMmioInstruction;
import net.consensys.linea.zktracer.types.Bytes16;

@RequiredArgsConstructor
public class RamToLimbTransplant implements MmioInstruction {
  private final MmuData mmuData;

  private final int instructionNumber;

  @Override
  public MmioData execute() {
    final MmuToMmioConstantValues mmuToMmioConstantValues = mmuData.mmuToMmioConstantValues();
    final MmuToMmioInstruction mmuToMmioInstruction =
        mmuData.mmuToMmioInstructions().get(instructionNumber);

    MmioData mmioData =
        new MmioData(
            mmuData.hubToMmuValues(),
            mmuToMmioConstantValues,
            mmuToMmioInstruction,
            mmuData.exoSumDecoder());

    mmioData.cnA(mmioData.sourceContext());
    mmioData.cnB(0);
    mmioData.cnC(0);

    mmioData.indexA(mmioData.sourceLimbOffset());
    mmioData.indexB(0);
    mmioData.indexC(0);
    mmioData.indexX(mmioData.targetLimbOffset());

    mmioData.valA(readLimb(mmuData.sourceRamBytes(), mmioData.indexA()));
    mmioData.valB(Bytes16.ZERO);
    mmioData.valC(Bytes16.ZERO);
    mmioData.limb(mmioData.limb());

    mmioData.valANew(mmioData.valA());
    mmioData.valBNew(Bytes16.ZERO);
    mmioData.valCNew(Bytes16.ZERO);

    return mmioData;
  }
}
