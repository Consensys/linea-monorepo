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

import static net.consensys.linea.zktracer.module.blockdata.Trace.MAX_CT;
import static net.consensys.linea.zktracer.types.TransactionUtils.getChainIdFromTransaction;

import java.nio.MappedByteBuffer;
import java.util.ArrayDeque;
import java.util.Deque;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.rlptxn.RlpTxn;
import net.consensys.linea.zktracer.module.txndata.TxnData;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;

public class Blockdata implements Module {
  private final Wcp wcp;
  private final TxnData txnData;
  private final RlpTxn rlpTxn;
  private final Deque<BlockdataOperation> operations = new ArrayDeque<>();
  private boolean batchUnderConstruction;
  private final int TIMESTAMP_BYTESIZE = 4;
  private int previousTimestamp = 0;

  public Blockdata(Wcp wcp, TxnData txnData, RlpTxn rlpTxn) {
    this.wcp = wcp;
    this.txnData = txnData;
    this.rlpTxn = rlpTxn;
    this.batchUnderConstruction = true;
  }

  @Override
  public String moduleKey() {
    return "BLOCKDATA";
  }

  @Override
  public void traceStartBlock(final ProcessableBlockHeader processableBlockHeader) {
    this.batchUnderConstruction = true;
    this.wcp.additionalRows.push(this.wcp.additionalRows.pop() + TIMESTAMP_BYTESIZE);
  }

  @Override
  public void traceEndBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    final int currentTimestamp = (int) blockHeader.getTimestamp();
    this.operations.addLast(
        new BlockdataOperation(
            blockHeader.getCoinbase(),
            currentTimestamp,
            blockHeader.getNumber(),
            blockHeader.getDifficulty().getAsBigInteger(),
            this.txnData.currentBlock().getTxs().size()));

    this.batchUnderConstruction = false;
    this.wcp.callGT(currentTimestamp, previousTimestamp);
    this.wcp.additionalRows.push(this.wcp.additionalRows.pop() - TIMESTAMP_BYTESIZE);
    this.previousTimestamp = currentTimestamp;
  }

  @Override
  public void traceStartConflation(final long blockCount) {
    this.batchUnderConstruction = false; // Should be useless, but just to be sure
  }

  @Override
  public void enterTransaction() {}

  @Override
  public void popTransaction() {}

  @Override
  public int lineCount() {
    final int numberOfBlock =
        this.batchUnderConstruction ? this.operations.size() + 1 : this.operations.size();
    return numberOfBlock * (MAX_CT + 1);
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final long firstBlockNumber = this.operations.getFirst().absoluteBlockNumber();
    final long chainId = getChainIdFromTransaction(this.rlpTxn.chunkList.get(0).tx());
    final Trace trace = new Trace(buffers);
    int relblock = 0;
    for (BlockdataOperation blockData : this.operations) {
      if (blockData.relTxMax() != 0) {
        relblock += 1;
        blockData.trace(trace, relblock, firstBlockNumber, chainId);
      }
    }
  }
}
