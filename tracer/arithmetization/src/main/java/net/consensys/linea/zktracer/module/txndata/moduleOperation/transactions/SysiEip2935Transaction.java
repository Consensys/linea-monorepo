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
package net.consensys.linea.zktracer.module.txndata.moduleOperation.transactions;

import static com.google.common.base.Preconditions.checkState;
import static net.consensys.linea.zktracer.Fork.isPostPrague;
import static net.consensys.linea.zktracer.Trace.HISTORY_SERVE_WINDOW;
import static net.consensys.linea.zktracer.module.txndata.moduleOperation.TxnDataOperationPerspectivized.TransactionCategory.*;
import static net.consensys.linea.zktracer.module.txndata.rows.computationRows.WcpRow.smallCallToIszero;
import static net.consensys.linea.zktracer.module.txndata.rows.computationRows.WcpRow.smallCallToLeq;

import net.consensys.linea.zktracer.module.txndata.module.PerspectivizedTxnData;
import net.consensys.linea.zktracer.module.txndata.moduleOperation.TxnDataOperationPerspectivized;
import net.consensys.linea.zktracer.module.txndata.rows.computationRows.EucRow;
import net.consensys.linea.zktracer.module.txndata.rows.computationRows.WcpRow;
import net.consensys.linea.zktracer.module.txndata.rows.hubRows.HubRowForSystemTransactions;
import net.consensys.linea.zktracer.module.txndata.rows.hubRows.Type;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;

public class SysiEip2935Transaction extends TxnDataOperationPerspectivized {

  private final long nonsensePragueTimestamp =
      0x13370000L; // Placeholder for the actual Prague fork timestamp

  @Override
  protected int ctMax() {
    return 3;
  }

  public SysiEip2935Transaction(final PerspectivizedTxnData txnData) {
    super(txnData, SYSI);
    checkState(isPostPrague(txnData.hub().fork));
    process();
  }

  private void process() {
    hubRow();
    detectTheGenesisBlockComputationRow();
    computePreviousBlockNumberModulo8191ComputationRow();
    compareTimestampToLineaCancunForkTimestampComputationRow();
  }

  protected void hubRow() {

    HubRowForSystemTransactions hubRow =
        new HubRowForSystemTransactions(blockHeader, hub, Type.EIP2935);

    hubRow.systemTransactionData1 = EWord.of(previousBlockNumber());
    hubRow.systemTransactionData2 = EWord.of(previousBlockNumber() % HISTORY_SERVE_WINDOW);
    hubRow.systemTransactionData3 = EWord.of(EWord.of(previousBlockHash()).hi());
    hubRow.systemTransactionData4 = EWord.of(EWord.of(previousBlockHash()).lo());
    hubRow.systemTransactionData5 = EWord.of(currentBlockIsGenesisBlock() ? 1 : 0);

    rows.add(hubRow);
  }

  private void detectTheGenesisBlockComputationRow() {
    WcpRow row = smallCallToIszero(wcp, blockHeader.getNumber());
    rows.add(row);
  }

  private void computePreviousBlockNumberModulo8191ComputationRow() {
    EucRow row = EucRow.callToEuc(euc, previousBlockNumber(), HISTORY_SERVE_WINDOW);
    rows.add(row);
  }

  private void compareTimestampToLineaCancunForkTimestampComputationRow() {
    WcpRow row = smallCallToLeq(wcp, blockHeader.getTimestamp(), nonsensePragueTimestamp);
    rows.add(row);
  }

  private long previousBlockNumber() {
    return currentBlockIsGenesisBlock() ? 0 : blockHeader.getNumber() - 1;
  }

  private boolean currentBlockIsGenesisBlock() {
    return blockHeader.getNumber() == 0;
  }

  private Bytes previousBlockHash() {
    return currentBlockIsGenesisBlock() ? blockHeader.getParentHash() : Bytes.EMPTY;
  }
}
