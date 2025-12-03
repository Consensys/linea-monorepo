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

package net.consensys.linea.zktracer.module.blockdata.moduleOperation;

import static net.consensys.linea.zktracer.opcode.OpCode.PREVRANDAO;

import net.consensys.linea.zktracer.ChainConfig;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.hyperledger.besu.plugin.data.BlockHeader;

public class ShanghaiBlockDataOperation extends ParisBlockDataOperation {
  public ShanghaiBlockDataOperation(
      Hub hub,
      BlockHeader blockHeader,
      BlockHeader prevBlockHeader,
      int relTxMax,
      Wcp wcp,
      Euc euc,
      ChainConfig chain,
      OpCode opCode,
      long firstBlockNumber) {
    super(hub, blockHeader, prevBlockHeader, relTxMax, wcp, euc, chain, opCode, firstBlockNumber);
  }

  @Override
  protected void traceIsDifficulty(Trace.Blockdata trace, OpCode opCode) {
    // OpCode in London fork only, not in Paris and after.
  }

  @Override
  protected void traceIsPrevRandao(Trace.Blockdata trace, OpCode opCode) {
    trace.isPrevrandao(opCode == PREVRANDAO);
  }
}
