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

package net.consensys.linea.zktracer.module.blockdata;

import static net.consensys.linea.zktracer.Trace.Blockdata.nROWS_DEPTH;
import static net.consensys.linea.zktracer.Trace.LLARGE;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import java.util.*;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ChainConfig;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.module.Module;
import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.txndata.TxnData;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;

@RequiredArgsConstructor
public class Blockdata implements Module {
  private final Wcp wcp;
  private final Euc euc;
  private final TxnData txnData;
  private final ChainConfig chain;

  private final List<BlockdataOperation> operations = new ArrayList<>();
  private long firstBlockNumber;

  private boolean conflationFinished = false;

  private static final OpCode[] opCodes = {
    OpCode.COINBASE,
    OpCode.TIMESTAMP,
    OpCode.NUMBER,
    OpCode.DIFFICULTY,
    OpCode.GASLIMIT,
    OpCode.CHAINID,
    OpCode.BASEFEE
  };

  @Override
  public String moduleKey() {
    return "BLOCK_DATA";
  }

  @Override
  public void traceStartConflation(final long blockCount) {
    wcp.additionalRows.add(
        LLARGE // for COINBASE
            + 6
            + 6 // for TIMESTAMP
            + 1
            + 6 // for NUMBER
            + 1 // for DIFFICULTY
            + (bigIntegerToBytes(chain.gasLimitMaximum).size() * 4) // for GASLIMIT
            + LLARGE // for CHAINID
            + LLARGE // for BASEFEE
        );

    euc.additionalRows.add(8);
  }

  @Override
  public void traceEndConflation(final WorldView state) {
    conflationFinished = true;
  }

  @Override
  public void traceEndBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    final long blockNumber = blockHeader.getNumber();
    if (operations.isEmpty()) {
      firstBlockNumber = blockNumber;
    }
    final BlockHeader previousBlockHeader =
        operations.isEmpty() ? null : operations.getLast().blockHeader();
    for (OpCode opCode : opCodes) {
      final BlockdataOperation operation =
          new BlockdataOperation(
              txnData.hub(),
              blockHeader,
              previousBlockHeader,
              txnData.currentBlock().getNbOfTxsInBlock(),
              wcp,
              euc,
              chain,
              opCode,
              firstBlockNumber);
      operations.addLast(operation);
    }
  }

  @Override
  public void commitTransactionBundle() {}

  @Override
  public void popTransactionBundle() {}

  @Override
  public int lineCount() {
    final int numberOfBlock = (operations.size() / opCodes.length) + (conflationFinished ? 0 : 1);
    return numberOfBlock * nROWS_DEPTH;
  }

  @Override
  public int spillage() {
    return Trace.Blockdata.SPILLAGE;
  }

  @Override
  public List<Trace.ColumnHeader> columnHeaders() {
    return Trace.Blockdata.headers(this.lineCount());
  }

  @Override
  public void commit(Trace trace) {
    for (BlockdataOperation blockData : operations) {
      blockData.trace(trace.blockdata);
    }
  }
}
