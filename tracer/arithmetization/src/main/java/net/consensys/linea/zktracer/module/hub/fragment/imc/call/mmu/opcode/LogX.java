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

package net.consensys.linea.zktracer.module.hub.fragment.imc.call.mmu.opcode;

import static net.consensys.linea.zktracer.module.mmu.Trace.MMU_INST_RAM_TO_EXO_WITH_PADDING;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.imc.call.mmu.MmuCall;
import net.consensys.linea.zktracer.types.EWord;
import org.hyperledger.besu.evm.internal.Words;

/**
 * A specialization of the generic {@link MmuCall} that disables the MMU call if the LOG that
 * triggered eventually sees its context reverted.
 */
public class LogX extends MmuCall {
  private final int logId;

  public LogX(final Hub hub) {
    super(MMU_INST_RAM_TO_EXO_WITH_PADDING);
    this.logId = hub.transients().conflation().currentLogId();

    this.sourceId(hub.currentFrame().contextNumber())
        .sourceOffset(EWord.of(hub.messageFrame().getStackItem(0)))
        .size(Words.clampedToLong(hub.messageFrame().getStackItem(1)))
        .referenceSize(Words.clampedToLong(hub.messageFrame().getStackItem(1)))
        .setLog();
  }

  @Override
  public void postConflationRetcon(final Hub hub) {
    if (hub.transients().conflation().logs().get(this.logId).reverted()) this.enabled = false;
  }
}
