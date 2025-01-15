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

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.module.OperationListModule;
import net.consensys.linea.zktracer.container.stacked.ModuleOperationStackedList;
import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;

@RequiredArgsConstructor
@Accessors(fluent = true)
public class TxnData implements OperationListModule<TxndataOperation> {
  @Getter
  private final ModuleOperationStackedList<TxndataOperation> operations =
      new ModuleOperationStackedList<>();

  @Getter private final Hub hub;
  private final Wcp wcp;
  private final Euc euc;

  private final List<BlockSnapshot> blocks = new ArrayList<>();

  @Override
  public String moduleKey() {
    return "TXN_DATA";
  }

  @Override
  public void traceStartConflation(final long blockCount) {
    wcp.additionalRows.add(4); /* 4 = byte length of LINEA_BLOCK_GAS_LIMIT */
  }

  @Override
  public final void traceStartBlock(
      final ProcessableBlockHeader blockHeader, final Address miningBeneficiary) {
    blocks.add(new BlockSnapshot(blockHeader));
  }

  @Override
  public void traceEndTx(TransactionProcessingMetadata tx) {
    operations.add(new TxndataOperation(hub, wcp, euc, tx));
  }

  @Override
  public void traceEndBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    currentBlock().setNbOfTxsInBlock(currentTx().tx.getRelativeTransactionNumber());
    currentTx().setCallWcpLastTxOfBlock(currentBlock().getBlockGasLimit());
  }

  @Override
  public int lineCount() {
    // The last tx of each block has one more rows
    return operations.lineCount() + blocks.size();
  }

  public BlockSnapshot currentBlock() {
    return blocks.getLast();
  }

  private TxndataOperation currentTx() {
    return operations.getLast();
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    final int absTxNumMax = operations.size();

    for (TxndataOperation tx : operations.getAll()) {
      tx.traceTx(trace, blocks.get(tx.getTx().getRelativeBlockNumber() - 1), absTxNumMax);
    }
  }
}
