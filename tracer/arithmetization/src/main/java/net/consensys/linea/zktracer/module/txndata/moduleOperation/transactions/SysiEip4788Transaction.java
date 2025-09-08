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
import static net.consensys.linea.zktracer.Fork.isPostCancun;
import static net.consensys.linea.zktracer.Trace.HISTORY_BUFFER_LENGTH;
import static net.consensys.linea.zktracer.module.txndata.moduleOperation.TxnDataOperationPerspectivized.TransactionCategory.*;
import static net.consensys.linea.zktracer.module.txndata.rows.computationRows.EucRow.callToEuc;
import static net.consensys.linea.zktracer.module.txndata.rows.computationRows.WcpRow.smallCallToIszero;
import static net.consensys.linea.zktracer.module.txndata.rows.computationRows.WcpRow.smallCallToLeq;
import static net.consensys.linea.zktracer.module.txndata.rows.hubRows.Type.EIP4788;

import net.consensys.linea.zktracer.module.txndata.module.PerspectivizedTxnData;
import net.consensys.linea.zktracer.module.txndata.moduleOperation.TxnDataOperationPerspectivized;
import net.consensys.linea.zktracer.module.txndata.rows.computationRows.EucRow;
import net.consensys.linea.zktracer.module.txndata.rows.computationRows.WcpRow;
import net.consensys.linea.zktracer.module.txndata.rows.hubRows.HubRowForSystemTransactions;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes32;

public class SysiEip4788Transaction extends TxnDataOperationPerspectivized {

  private final long NONSENSE_CANCUN_HARDFORK_TIMESTAMP =
      0x1337L; // Placeholder for the actual Prague fork timestamp

  public SysiEip4788Transaction(final PerspectivizedTxnData txnData) {
    super(txnData, SYSI);
    checkState(isPostCancun(txnData.hub().fork));
    process();
  }

  private void process() {
    hubRow();
    computeTimestampModulo8191ComputationRow();
    detectTheGenesisBlockComputationRow();
    compareTimestampToLineaCancunForkTimestampComputationRow();
  }

  protected void hubRow() {
    HubRowForSystemTransactions hubRow = new HubRowForSystemTransactions(blockHeader, hub, EIP4788);

    long timestamp = blockHeader.getTimestamp();
    Bytes32 parentBeaconBlockRoot =
        blockHeader.getParentBeaconBlockRoot().isPresent()
            ? blockHeader.getParentBeaconBlockRoot().get()
            : Bytes32.ZERO;

    hubRow.systemTransactionData1 = EWord.of(timestamp);
    hubRow.systemTransactionData2 = EWord.of(timestamp % HISTORY_BUFFER_LENGTH);
    hubRow.systemTransactionData3 = EWord.of(EWord.of(parentBeaconBlockRoot).hi());
    hubRow.systemTransactionData4 = EWord.of(EWord.of(parentBeaconBlockRoot).lo());
    hubRow.systemTransactionData5 = EWord.of(blockHeader.getNumber() == 0 ? 1 : 0);

    rows.add(hubRow);
  }

  private void computeTimestampModulo8191ComputationRow() {
    // TODO: use the prime constant
    EucRow row = callToEuc(euc, blockHeader.getTimestamp(), HISTORY_BUFFER_LENGTH);
    rows.add(row);
  }

  private void detectTheGenesisBlockComputationRow() {
    WcpRow row = smallCallToIszero(wcp, blockHeader.getNumber());
    rows.add(row);
  }

  private void compareTimestampToLineaCancunForkTimestampComputationRow() {
    WcpRow row =
        smallCallToLeq(wcp, NONSENSE_CANCUN_HARDFORK_TIMESTAMP, blockHeader.getTimestamp());
    rows.add(row);
  }

  @Override
  protected int ctMax() {
    return 3;
  }
}
