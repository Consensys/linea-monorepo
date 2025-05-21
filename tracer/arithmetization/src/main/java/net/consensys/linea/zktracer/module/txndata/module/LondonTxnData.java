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

package net.consensys.linea.zktracer.module.txndata.module;

import static net.consensys.linea.zktracer.TraceLondon.Txndata.*;

import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.txndata.moduleOperation.LondonTxndataOperation;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;

public class LondonTxnData extends TxnData {

  private static final int N_ROWS_TX_MAX_LONDON =
      Math.max(Math.max(NB_ROWS_TYPE_0, NB_ROWS_TYPE_1), NB_ROWS_TYPE_2);

  public LondonTxnData(Hub hub, Wcp wcp, Euc euc) {
    super(hub, wcp, euc);
  }

  @Override
  public void traceEndTx(TransactionProcessingMetadata tx) {
    operations().add(new LondonTxndataOperation(wcp(), euc(), tx, N_ROWS_TX_MAX_LONDON));
  }
}
