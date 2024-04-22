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

package net.consensys.linea.zktracer.module.mmu;

import java.nio.MappedByteBuffer;
import java.util.List;

import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.list.StackedList;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.blake2fmodexpdata.Blake2fModexpData;
import net.consensys.linea.zktracer.module.ec_data.EcData;
import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.mmu.MmuCall;
import net.consensys.linea.zktracer.module.mmu.values.HubToMmuValues;
import net.consensys.linea.zktracer.module.rlp.txn.RlpTxn;
import net.consensys.linea.zktracer.module.rlp.txrcpt.RlpTxrcpt;
import net.consensys.linea.zktracer.module.romLex.RomLex;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.runtime.callstack.CallStack;

@Accessors(fluent = true)
public class Mmu implements Module {
  @Getter private final StackedList<MmuOperation> mmuOperations = new StackedList<>();
  private final Euc euc;
  private final Wcp wcp;

  private final ExoSumDecoder exoSumDecoder;

  public Mmu(
      final Euc euc,
      final Wcp wcp,
      final RomLex romLex,
      final RlpTxn rlpTxn,
      final RlpTxrcpt rlpTxrcpt,
      final EcData ecData,
      final Blake2fModexpData blake2fModexpData,
      //  TODO: SHAKIRA module
      final CallStack callStack) {
    this.euc = euc;
    this.wcp = wcp;
    this.exoSumDecoder =
        new ExoSumDecoder(callStack, romLex, rlpTxn, rlpTxrcpt, ecData, blake2fModexpData);
  }

  @Override
  public String moduleKey() {
    return "MMU";
  }

  @Override
  public void enterTransaction() {
    this.mmuOperations.enter();
  }

  @Override
  public void popTransaction() {
    this.mmuOperations.pop();
  }

  @Override
  public int lineCount() {
    return this.mmuOperations.lineCount();
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    int mmuStamp = 0;
    int mmioStamp = 0;

    for (MmuOperation mmuOp : this.mmuOperations) {
      mmuOp.getCFI();
      mmuOp.setExoBytes(exoSumDecoder);
      mmuOp.fillLimb();

      mmuStamp += 1;
      mmuOp.trace(mmuStamp, mmioStamp, trace);
      mmioStamp += mmuOp.mmuData().numberMmioInstructions();
    }
  }

  public void call(final MmuCall mmuCall, final CallStack callStack) {
    MmuData mmuData = new MmuData(mmuCall, callStack);
    mmuData.hubToMmuValues(
        HubToMmuValues.fromMmuCall(mmuCall, mmuData.exoLimbIsSource(), mmuData.exoLimbIsTarget()));

    final MmuInstructions mmuInstructions = new MmuInstructions(euc, wcp);
    mmuData = mmuInstructions.compute(mmuData, callStack);

    this.mmuOperations.add(new MmuOperation(mmuData, callStack));
  }
}
