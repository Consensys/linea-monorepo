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
package net.consensys.linea.sequencer.txselection;

import org.hyperledger.besu.plugin.data.TransactionSelectionResult;

public class LineaTransactionSelectionResult extends TransactionSelectionResult {
  private enum LineaStatus implements TransactionSelectionResult.Status {
    BLOCK_CALLDATA_OVERFLOW(false, false),
    BLOCK_MODULE_LINE_COUNT_FULL(true, false),
    TX_GAS_EXCEEDS_USER_MAX_BLOCK_GAS(false, true),
    TX_TOO_LARGE_FOR_REMAINING_USER_GAS(false, false),
    TX_MODULE_LINE_COUNT_OVERFLOW(false, true),
    TX_UNPROFITABLE(false, false);

    private final boolean stop;
    private final boolean discard;

    LineaStatus(boolean stop, boolean discard) {
      this.stop = stop;
      this.discard = discard;
    }

    @Override
    public boolean stop() {
      return stop;
    }

    @Override
    public boolean discard() {
      return discard;
    }
  }

  protected LineaTransactionSelectionResult(LineaStatus status) {
    super(status);
  }

  public static final TransactionSelectionResult BLOCK_CALLDATA_OVERFLOW =
      new LineaTransactionSelectionResult(LineaStatus.BLOCK_CALLDATA_OVERFLOW);
  public static final TransactionSelectionResult BLOCK_MODULE_LINE_COUNT_FULL =
      new LineaTransactionSelectionResult(LineaStatus.BLOCK_MODULE_LINE_COUNT_FULL);
  public static final TransactionSelectionResult TX_GAS_EXCEEDS_USER_MAX_BLOCK_GAS =
      new LineaTransactionSelectionResult(LineaStatus.TX_GAS_EXCEEDS_USER_MAX_BLOCK_GAS);
  public static final TransactionSelectionResult TX_TOO_LARGE_FOR_REMAINING_USER_GAS =
      new LineaTransactionSelectionResult(LineaStatus.TX_TOO_LARGE_FOR_REMAINING_USER_GAS);
  public static final TransactionSelectionResult TX_MODULE_LINE_COUNT_OVERFLOW =
      new LineaTransactionSelectionResult(LineaStatus.TX_MODULE_LINE_COUNT_OVERFLOW);

  public static final TransactionSelectionResult TX_UNPROFITABLE =
      new LineaTransactionSelectionResult(LineaStatus.TX_UNPROFITABLE);
}
