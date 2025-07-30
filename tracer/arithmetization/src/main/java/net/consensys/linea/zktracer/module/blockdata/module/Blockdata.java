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

import static net.consensys.linea.zktracer.Trace.LLARGE;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import java.util.*;

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ChainConfig;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.module.Module;
import net.consensys.linea.zktracer.module.blockdata.moduleOperation.BlockdataOperation;
import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.txndata.module.TxnData;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;

@RequiredArgsConstructor
public abstract class Blockdata implements Module {
  private final Hub hub;
  private final Wcp wcp;
  private final Euc euc;
  private final ChainConfig chain;
  @Getter private final List<BlockdataOperation> operations = new ArrayList<>();
  @Getter private long firstBlockNumber;

  private boolean conflationFinished = false;

  @Getter private final OpCode[] opCodes = setOpCodes();

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
            + 1 // for DIFFICULTY or PREVRANDAO
            + (bigIntegerToBytes(chain.gasLimitMaximum).size() * 4) // for GASLIMIT
            + LLARGE // for CHAINID
            + LLARGE // for BASEFEE
        // TODO: we should add +1 here for BLOBBASEFEE (post Cancun), but we don't as limitless
        // prover will be deployed before Cancun (and all this line counting will die)
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
          setBlockDataOperation(
              hub,
              blockHeader,
              previousBlockHeader,
              txnData().numberOfUserTransactionsInCurrentBlock(),
              wcp,
              euc,
              chain,
              opCode,
              firstBlockNumber);
      operations.addLast(operation);
    }
  }

  protected abstract BlockdataOperation setBlockDataOperation(
      Hub hub,
      BlockHeader blockHeader,
      BlockHeader previousBlockHeader,
      int nbOfTxsInBlock,
      Wcp wcp,
      Euc euc,
      ChainConfig chain,
      OpCode opCode,
      long firstBlockNumber);

  protected abstract OpCode[] setOpCodes();

  @Override
  public void commitTransactionBundle() {}

  @Override
  public void popTransactionBundle() {}

  @Override
  public int lineCount() {
    final int numberOfBlock = (operations.size() / opCodes.length) + (conflationFinished ? 0 : 1);
    return numberOfBlock * numberOfLinesPerBlock();
  }

  protected abstract int numberOfLinesPerBlock();

  @Override
  public int spillage(Trace trace) {
    return trace.blockdata().spillage();
  }

  @Override
  public List<Trace.ColumnHeader> columnHeaders(Trace trace) {
    return trace.blockdata().headers(this.lineCount());
  }

  @Override
  public void commit(Trace trace) {
    for (BlockdataOperation blockData : operations) {
      blockData.trace(trace.blockdata());
    }
  }

  TxnData txnData() {
    return hub.txnData();
  }
}
