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

import static net.consensys.linea.zktracer.module.ModuleName.*;

import java.util.Map;
import net.consensys.linea.zktracer.ChainConfig;
import net.consensys.linea.zktracer.container.module.CountingOnlyModule;
import net.consensys.linea.zktracer.container.module.Module;
import net.consensys.linea.zktracer.module.blockdata.module.BlockData;
import net.consensys.linea.zktracer.module.blockdata.module.LondonBlockData;
import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.hub.section.create.LondonCreateSection;
import net.consensys.linea.zktracer.module.hub.section.skip.LondonTxSkipSection;
import net.consensys.linea.zktracer.module.hub.transients.Transients;
import net.consensys.linea.zktracer.module.mxp.module.LondonMxp;
import net.consensys.linea.zktracer.module.mxp.module.Mxp;
import net.consensys.linea.zktracer.module.rlptxn.RlpTxn;
import net.consensys.linea.zktracer.module.rlptxn.london.LondonRlpTxn;
import net.consensys.linea.zktracer.module.tables.instructionDecoder.InstructionDecoder;
import net.consensys.linea.zktracer.module.tables.instructionDecoder.LondonInstructionDecoder;
import net.consensys.linea.zktracer.module.txndata.TxnData;
import net.consensys.linea.zktracer.module.txndata.london.LondonTxnData;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.types.PublicInputs;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class LondonHub extends Hub {
  public LondonHub(ChainConfig chain, PublicInputs publicInputs) {
    super(chain, publicInputs);
  }

  @Override
  protected TxnData setTxnData() {
    return new LondonTxnData(this, wcp(), euc());
  }

  @Override
  protected BlockData setBlockData(
      Hub hub, Wcp wcp, Euc euc, ChainConfig chain, Map<Long, Bytes> blobBaseFees) {
    return new LondonBlockData(hub, wcp, euc, chain, blobBaseFees);
  }

  @Override
  protected RlpTxn setRlpTxn(Hub hub) {
    return new LondonRlpTxn(hub.romLex());
  }

  @Override
  protected Module setRlpUtils() {
    // RlpUtils is not used in London, it is only used in Cancun
    return new CountingOnlyModule(RLP_UTILS);
  }

  @Override
  protected Mxp setMxp() {
    return new LondonMxp();
  }

  @Override
  protected InstructionDecoder setInstructionDecoder() {
    return new LondonInstructionDecoder(this.opCodes());
  }

  @Override
  protected void setSkipSection(
      Hub hub,
      WorldView world,
      TransactionProcessingMetadata transactionProcessingMetadata,
      Transients transients) {
    new LondonTxSkipSection(hub, world, transactionProcessingMetadata, transients);
  }

  @Override
  protected void setCreateSection(final Hub hub, final MessageFrame frame) {
    new LondonCreateSection(hub, frame);
  }
}
