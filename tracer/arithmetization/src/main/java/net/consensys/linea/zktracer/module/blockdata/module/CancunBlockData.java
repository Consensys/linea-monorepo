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

package net.consensys.linea.zktracer.module.blockdata.module;

import static net.consensys.linea.zktracer.TraceCancun.Blockdata.*;
import static net.consensys.linea.zktracer.opcode.OpCode.*;
import static net.consensys.linea.zktracer.opcode.OpCode.BASEFEE;
import static net.consensys.linea.zktracer.opcode.OpCode.CHAINID;
import static net.consensys.linea.zktracer.opcode.OpCode.GASLIMIT;
import static net.consensys.linea.zktracer.opcode.OpCode.PREVRANDAO;

import net.consensys.linea.zktracer.ChainConfig;
import net.consensys.linea.zktracer.module.blockdata.moduleOperation.BlockdataOperation;
import net.consensys.linea.zktracer.module.blockdata.moduleOperation.CancunBlockDataOperation;
import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.services.BlockchainService;

public class CancunBlockData extends ParisBlockData {
  private final BlockchainService blockchainService;

  public CancunBlockData(
      Hub hub, Wcp wcp, Euc euc, ChainConfig chain, BlockchainService blockchainService) {
    super(hub, wcp, euc, chain);
    this.blockchainService = blockchainService;
  }

  @Override
  protected OpCode[] setOpCodes() {
    return new OpCode[] {
      COINBASE, TIMESTAMP, NUMBER, PREVRANDAO, GASLIMIT, CHAINID, BASEFEE, BLOBBASEFEE
    };
  }

  @Override
  protected int numberOfLinesPerBlock() {
    return nROWS_DEPTH;
  }

  @Override
  protected BlockdataOperation setBlockDataOperation(
      Hub hub,
      BlockHeader blockHeader,
      BlockHeader previousBlockHeader,
      int nbOfTxsInBlock,
      Wcp wcp,
      Euc euc,
      ChainConfig chain,
      OpCode opCode,
      long firstBlockNumber) {
    return new CancunBlockDataOperation(
        blockchainService,
        hub,
        blockHeader,
        previousBlockHeader,
        txnData().currentBlock().getNbOfTxsInBlock(),
        wcp,
        euc,
        chain,
        opCode,
        firstBlockNumber);
  }
}
