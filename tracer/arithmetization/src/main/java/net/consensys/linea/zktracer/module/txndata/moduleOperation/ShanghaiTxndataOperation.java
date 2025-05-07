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

package net.consensys.linea.zktracer.module.txndata.moduleOperation;

import static net.consensys.linea.zktracer.Trace.*;

import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.txndata.TxnDataComparisonRecord;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.apache.tuweni.bytes.Bytes;

public class ShanghaiTxndataOperation extends LondonTxndataOperation {
  public static final Bytes MAX_INIT_CODE_SIZE_BYTES = Bytes.ofUnsignedInt(MAX_INIT_CODE_SIZE);
  private static final Bytes WORD_SIZE_BYTES = Bytes.ofUnsignedInt(WORD_SIZE);

  public ShanghaiTxndataOperation(Wcp wcp, Euc euc, TransactionProcessingMetadata tx) {
    super(wcp, euc, tx);
  }

  @Override
  void setShanghaiCallsToEucAndWcp() {
    if (tx.isDeployment()) {
      // row 2
      callsToEucAndWcp.add(
          TxnDataComparisonRecord.callToLeq(
              wcp,
              Bytes.of(tx.getBesuTransaction().getPayload().size()),
              MAX_INIT_CODE_SIZE_BYTES));
      // row 3
      final Bytes divisor =
          Bytes.minimalBytes(tx.getBesuTransaction().getPayload().size() + WORD_SIZE_MO);
      callsToEucAndWcp.add(TxnDataComparisonRecord.callToEuc(euc, divisor, WORD_SIZE_BYTES));
    } else {
      // rows 2 & 3
      callsToEucAndWcp.add(TxnDataComparisonRecord.empty());
      callsToEucAndWcp.add(TxnDataComparisonRecord.empty());
    }
  }
}
