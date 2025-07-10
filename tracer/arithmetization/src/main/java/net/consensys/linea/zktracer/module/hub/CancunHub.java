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
import net.consensys.linea.zktracer.module.hub.section.McopySection;
import net.consensys.linea.zktracer.module.hub.section.transients.TLoadSection;
import net.consensys.linea.zktracer.module.hub.section.transients.TStoreSection;
import net.consensys.linea.zktracer.module.mxp.module.CancunMxp;
import net.consensys.linea.zktracer.module.mxp.module.Mxp;
import net.consensys.linea.zktracer.module.rlpUtils.RlpUtils;
import net.consensys.linea.zktracer.module.rlptxn.RlpTxn;
import net.consensys.linea.zktracer.module.rlptxn.cancun.CancunRlpTxn;
import net.consensys.linea.zktracer.module.tables.PowerRt;
import net.consensys.linea.zktracer.module.tables.instructionDecoder.CancunInstructionDecoder;
import net.consensys.linea.zktracer.module.tables.instructionDecoder.InstructionDecoder;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import org.hyperledger.besu.evm.gascalculator.CancunGasCalculator;
import org.hyperledger.besu.evm.gascalculator.GasCalculator;

public class CancunHub extends ShanghaiHub {
  public CancunHub(ChainConfig chain) {
    super(chain);
  }

  @Override
  protected GasCalculator setGasCalculator() {
    return new CancunGasCalculator();
  }

  @Override
  protected Mxp setMxp() {
    return new CancunMxp();
  }

  @Override
  protected Blockdata setBlockData(Hub hub, Wcp wcp, Euc euc, ChainConfig chain) {
    return new CancunBlockData(hub, wcp, euc, chain);
  }

  @Override
  protected RlpTxn setRlpTxn(Hub hub) {
    return new CancunRlpTxn(hub.rlpUtils());
  }

  @Override
  protected RlpUtils setRlpUtils(Wcp wcp) {
    return new RlpUtils(wcp);
  }

  @Override
  protected InstructionDecoder setInstructionDecoder() {
    return new CancunInstructionDecoder();
  }

  @Override
  protected PowerRt setPower() {
    return new PowerRt();
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

  @Override
  protected void setMcopySection(Hub hub) {
    new McopySection(hub);
  }
}
