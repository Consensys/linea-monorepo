/*
 * Copyright ConsenSys Inc.
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

package net.consensys.linea.zktracer.module.mmu.instructions;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.EIP_3541_MARKER;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LLARGE;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LLARGEMO;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.MMIO_INST_RAM_TO_LIMB_ONE_SOURCE;

import java.util.ArrayList;
import java.util.List;

import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.euc.EucOperation;
import net.consensys.linea.zktracer.module.mmio.CallStackReader;
import net.consensys.linea.zktracer.module.mmu.MmuData;
import net.consensys.linea.zktracer.module.mmu.Trace;
import net.consensys.linea.zktracer.module.mmu.values.HubToMmuValues;
import net.consensys.linea.zktracer.module.mmu.values.MmuEucCallRecord;
import net.consensys.linea.zktracer.module.mmu.values.MmuOutAndBinValues;
import net.consensys.linea.zktracer.module.mmu.values.MmuToMmioConstantValues;
import net.consensys.linea.zktracer.module.mmu.values.MmuToMmioInstruction;
import net.consensys.linea.zktracer.module.mmu.values.MmuWcpCallRecord;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.runtime.callstack.CallStack;
import net.consensys.linea.zktracer.types.Bytes16;
import org.apache.tuweni.bytes.Bytes;

public class InvalidCodePrefix implements MmuInstruction {
  private final Euc euc;
  private final Wcp wcp;
  private List<MmuEucCallRecord> eucCallRecords;
  private List<MmuWcpCallRecord> wcpCallRecords;

  private long initialSourceLimbOffset;
  private short initialSourceByteOffset;
  private Bytes microLimb;

  public InvalidCodePrefix(Euc euc, Wcp wcp) {
    this.euc = euc;
    this.wcp = wcp;
    this.eucCallRecords = new ArrayList<>(Trace.NB_PP_ROWS_INVALID_CODE_PREFIX);
    this.wcpCallRecords = new ArrayList<>(Trace.NB_PP_ROWS_INVALID_CODE_PREFIX);
  }

  @Override
  public MmuData preProcess(MmuData mmuData, final CallStack callStack) {
    // Set mmuData.sourceRamBytes
    CallStackReader callStackReader = new CallStackReader(callStack);
    final Bytes sourceMemory =
        callStackReader.valueFromMemory(mmuData.hubToMmuValues().sourceId(), true);
    mmuData.sourceRamBytes(sourceMemory);

    // row n°1
    final long dividend1 = mmuData.hubToMmuValues().sourceOffsetLo().longValueExact();
    EucOperation eucOp = euc.callEUC(Bytes.ofUnsignedLong(dividend1), Bytes.of(16));
    final short rem = (short) eucOp.remainder().toInt();
    final long quot = eucOp.quotient().toLong();
    initialSourceLimbOffset = quot;
    initialSourceByteOffset = rem;

    eucCallRecords.add(
        MmuEucCallRecord.builder()
            .dividend(dividend1)
            .divisor((short) LLARGE)
            .quotient(quot)
            .remainder(rem)
            .build());

    microLimb =
        Bytes.of(
            mmuData
                .sourceRamBytes()
                .get((int) (LLARGE * initialSourceLimbOffset + initialSourceByteOffset)));
    Bytes arg1 = microLimb;
    Bytes arg2 = Bytes.of(EIP_3541_MARKER);
    boolean result = wcp.callEQ(arg1, arg2);

    wcpCallRecords.add(
        MmuWcpCallRecord.instEqBuilder().arg1Lo(arg1).arg2Lo(arg2).result(result).build());

    mmuData.eucCallRecords(eucCallRecords);
    mmuData.wcpCallRecords(wcpCallRecords);

    // setting Out and Bin values
    mmuData.outAndBinValues(MmuOutAndBinValues.builder().build()); // all 0

    mmuData.totalLeftZeroesInitials(0);
    mmuData.totalNonTrivialInitials(Trace.NB_MICRO_ROWS_TOT_INVALID_CODE_PREFIX);
    mmuData.totalRightZeroesInitials(0);

    return mmuData;
  }

  @Override
  public MmuData setMicroInstructions(MmuData mmuData) {
    HubToMmuValues hubToMmuValues = mmuData.hubToMmuValues();

    mmuData.mmuToMmioConstantValues(
        MmuToMmioConstantValues.builder().sourceContextNumber(hubToMmuValues.sourceId()).build());

    // First and only micro-instruction.
    mmuData.mmuToMmioInstruction(
        MmuToMmioInstruction.builder()
            .mmioInstruction(MMIO_INST_RAM_TO_LIMB_ONE_SOURCE)
            .size((short) 1)
            .sourceLimbOffset(initialSourceLimbOffset)
            .sourceByteOffset(initialSourceByteOffset)
            .targetByteOffset((short) LLARGEMO)
            .limb((Bytes16) microLimb)
            .build());

    return mmuData;
  }
}
