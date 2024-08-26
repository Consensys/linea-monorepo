/*
 * Copyright Consensys Software Inc.
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

package net.consensys.linea.zktracer.module.txndata;

import java.nio.MappedByteBuffer;
import java.util.ArrayList;
import java.util.List;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.list.StackedList;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;

@RequiredArgsConstructor
public class TxnData implements Module {

  private final Wcp wcp;
  private final Euc euc;

  private final List<BlockSnapshot> blocks = new ArrayList<>();
  private final StackedList<TxndataOperation> txs = new StackedList<>();

  @Override
  public String moduleKey() {
    return "TXN_DATA";
  }

  @Override
  public void enterTransaction() {
    this.txs.enter();
  }

  @Override
  public void popTransaction() {
    this.txs.pop();
  }

  @Override
  public void traceStartConflation(final long blockCount) {
    this.wcp.additionalRows.push(
        this.wcp.additionalRows.pop() + 4); /* 4 = byte length of LINEA_BLOCK_GAS_LIMIT */
  }

  @Override
  public final void traceStartBlock(final ProcessableBlockHeader blockHeader) {
    this.blocks.add(new BlockSnapshot(blockHeader));
  }

  @Override
  public void traceEndTx(TransactionProcessingMetadata tx) {
    this.txs.add(new TxndataOperation(wcp, euc, tx));
  }

  @Override
  public void traceEndBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    this.currentBlock().setNbOfTxsInBlock(this.currentTx().tx.getRelativeTransactionNumber());
    this.currentTx().setCallWcpLastTxOfBlock(this.currentBlock().getBlockGasLimit());
  }

  @Override
  public int lineCount() {
    // The last tx of each block has one more rows
    return this.txs.lineCount() + this.blocks.size();
  }

  public BlockSnapshot currentBlock() {
    return this.blocks.get(this.blocks.size() - 1);
  }

  private TxndataOperation currentTx() {
    return this.txs.getLast();
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    final int absTxNumMax = this.txs.size();

    for (TxndataOperation tx : this.txs) {
      tx.traceTx(trace, this.blocks.get(tx.getTx().getRelativeBlockNumber() - 1), absTxNumMax);
    }
  }
}
