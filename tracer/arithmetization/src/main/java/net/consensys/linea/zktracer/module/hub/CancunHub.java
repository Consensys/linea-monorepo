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

package net.consensys.linea.zktracer.module.hub;

import net.consensys.linea.zktracer.ChainConfig;
import net.consensys.linea.zktracer.module.blockdata.module.Blockdata;
import net.consensys.linea.zktracer.module.blockdata.module.CancunBlockData;
import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.hub.section.transients.TLoadSection;
import net.consensys.linea.zktracer.module.hub.section.transients.TStoreSection;
import net.consensys.linea.zktracer.module.tables.instructionDecoder.CancunInstructionDecoder;
import net.consensys.linea.zktracer.module.tables.instructionDecoder.InstructionDecoder;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import org.hyperledger.besu.evm.gascalculator.CancunGasCalculator;
import org.hyperledger.besu.evm.gascalculator.GasCalculator;
import org.hyperledger.besu.plugin.services.BlockchainService;

public class CancunHub extends ShanghaiHub {
  public CancunHub(ChainConfig chain, BlockchainService blockchainService) {
    super(chain, blockchainService);
  }

  @Override
  protected GasCalculator setGasCalculator() {
    return new CancunGasCalculator();
  }

  @Override
  protected Blockdata setBlockData(
      Hub hub, Wcp wcp, Euc euc, ChainConfig chain, BlockchainService blockchainService) {
    return new CancunBlockData(hub, wcp, euc, chain, blockchainService);
  }

  @Override
  protected InstructionDecoder setInstructionDecoder() {
    return new CancunInstructionDecoder();
  }

  @Override
  protected void setTransientSection(final Hub hub) {
    switch (hub.opCode()) {
      case TLOAD -> new TLoadSection(hub);
      case TSTORE -> new TStoreSection(hub);
      default -> throw new IllegalStateException(
          "invalid operation in family TRANSIENT: " + hub.opCode());
    }
  }
}
