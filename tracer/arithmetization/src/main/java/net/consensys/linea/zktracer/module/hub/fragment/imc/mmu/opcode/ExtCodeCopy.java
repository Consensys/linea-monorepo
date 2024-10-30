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

package net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.opcode;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.MMU_INST_ANY_TO_RAM_WITH_PADDING;

import java.util.Optional;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.defer.PostConflationDefer;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall;
import net.consensys.linea.zktracer.module.romlex.ContractMetadata;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.internal.Words;
import org.hyperledger.besu.evm.worldstate.WorldView;

/**
 * A specialization of {@link MmuCall} that addresses the fact that the MMU requires access to the
 * sorted Code Fragment Index of the copied bytecode, which is only available post-conflation.
 */
public class ExtCodeCopy extends MmuCall implements PostConflationDefer {
  private final Hub hub;
  private final ContractMetadata contract;

  public ExtCodeCopy(final Hub hub) {
    super(hub, MMU_INST_ANY_TO_RAM_WITH_PADDING);
    this.hub = hub;
    hub.defers().scheduleForPostConflation(this);

    final Address foreignCodeAddress = Words.toAddress(hub.messageFrame().getStackItem(0));
    this.contract = ContractMetadata.canonical(hub, foreignCodeAddress);

    this.targetId(hub.currentFrame().contextNumber())
        .targetRamBytes(
            Optional.of(
                hub.currentFrame()
                    .frame()
                    .shadowReadMemory(0, hub.currentFrame().frame().memoryByteSize())))
        .sourceOffset(EWord.of(hub.messageFrame().getStackItem(2)))
        .targetOffset(EWord.of(hub.messageFrame().getStackItem(1)))
        .size(Words.clampedToLong(hub.messageFrame().getStackItem(3)))
        .setRom();
  }

  @Override
  public Optional<Bytes> exoBytes() {
    try {
      return Optional.of(hub.romLex().getCodeByMetadata(contract));
    } catch (Exception ignored) {
      // Can be empty Bytes in case the ext account is empty. In this case, no associated CFI
      return Optional.of(Bytes.EMPTY);
    }
  }

  @Override
  public long referenceSize() {
    try {
      return (hub.romLex().getCodeByMetadata(contract).size());
    } catch (Exception ignored) {
      // Can be 0 in case the ext account is empty. In this case, no associated CFI
      return 0;
    }
  }

  @Override
  public void resolvePostConflation(Hub hub, WorldView world) {
    try {
      sourceId(hub.romLex().getCodeFragmentIndexByMetadata(contract));
    } catch (Exception ignored) {
      // Can be 0 in case the ext account is empty. In this case, no associated CFI
      sourceId(0);
    }
  }
}
