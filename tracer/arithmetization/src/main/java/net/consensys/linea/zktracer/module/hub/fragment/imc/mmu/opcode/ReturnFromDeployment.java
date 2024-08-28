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

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.MMU_INST_RAM_TO_EXO_WITH_PADDING;

import java.util.Optional;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.imc.mmu.MmuCall;
import net.consensys.linea.zktracer.module.romlex.ContractMetadata;
import net.consensys.linea.zktracer.types.EWord;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.evm.internal.Words;

/**
 * A specialization of {@link MmuCall} that addresses the fact that the MMU requires access to the
 * sorted Code Fragment Index of the deployed bytecode, which is only available post-conflation.
 */
public class ReturnFromDeployment extends MmuCall {
  private final Hub hub;
  private final ContractMetadata contract;

  public ReturnFromDeployment(final Hub hub) {
    super(hub, MMU_INST_RAM_TO_EXO_WITH_PADDING);

    this.hub = hub;

    // TODO: get the metaDaa directly from the hub
    final Address contractAddress = hub.currentFrame().frame().getContractAddress();
    final int depNumber = hub.transients().conflation().deploymentInfo().number(contractAddress);
    this.contract = ContractMetadata.underDeployment(contractAddress, depNumber);

    this.sourceId(hub.currentFrame().contextNumber())
        .sourceRamBytes(
            Optional.of(
                hub.currentFrame()
                    .frame()
                    .shadowReadMemory(0, hub.currentFrame().frame().memoryByteSize())))
        .exoBytes(Optional.of(hub.romLex().getCodeByMetadata(contract)))
        .auxId(hub.state().stamps().hub())
        .sourceOffset(EWord.of(hub.messageFrame().getStackItem(0)))
        .size(Words.clampedToLong(hub.messageFrame().getStackItem(1)))
        .referenceSize(Words.clampedToLong(hub.messageFrame().getStackItem(1)))
        .setKec()
        .setRom();
  }

  @Override
  public int targetId() {
    return this.hub.romLex().getCodeFragmentIndexByMetadata(this.contract);
  }
}
