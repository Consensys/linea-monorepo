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

package net.consensys.linea.zktracer.module.rlp.txn;

import static net.consensys.linea.zktracer.module.Util.getTxTypeAsInt;
import static net.consensys.linea.zktracer.module.rlp.txn.RlpTxn.LLARGE;

import java.math.BigInteger;
import java.util.Optional;

import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.container.ModuleOperation;
import org.hyperledger.besu.datatypes.Transaction;

@Accessors(fluent = true)
@Getter
public final class RlpTxnChunk extends ModuleOperation {
  private final Transaction tx;
  private final boolean requireEvmExecution;
  private final Optional<Integer> id;

  public RlpTxnChunk(Transaction tx, boolean requireEvmExecution, Optional<Integer> id) {
    this.tx = tx;
    this.requireEvmExecution = requireEvmExecution;
    this.id = id;
  }

  public RlpTxnChunk(Transaction tx, boolean requireEvmExecution) {
    this(tx, requireEvmExecution, Optional.empty());
  }

  public RlpTxnChunk(Transaction tx, boolean requireEvmExecution, int codeIdentifierPreLexOrder) {
    this(tx, requireEvmExecution, Optional.of(codeIdentifierPreLexOrder));
  }

  @Override
  protected int computeLineCount() {
    final int txType = getTxTypeAsInt(this.tx.getType());
    // Phase 0 is always 17 rows long
    int rowSize = 17;

    // Phase 1: chainID
    if (txType == 1 || txType == 2) {
      if (this.tx.getChainId().orElseThrow().equals(BigInteger.ZERO)) {
        rowSize += 1;
      } else {
        rowSize += 8;
      }
    }

    // Phase 2: nonce
    if (this.tx.getNonce() == 0) {
      rowSize += 1;
    } else {
      rowSize += 8;
    }

    // Phase 3: gasPrice
    if (txType == 0 || txType == 1) {
      rowSize += 8;
    }

    // Phase 4: MaxPriorityFeeperGas
    if (txType == 2) {
      if (this.tx
          .getMaxPriorityFeePerGas()
          .orElseThrow()
          .getAsBigInteger()
          .equals(BigInteger.ZERO)) {
        rowSize += 1;
      } else {
        rowSize += 8;
      }
    }

    // Phase 5: MaxFeePerGas
    if (txType == 2) {
      if (this.tx.getMaxFeePerGas().orElseThrow().getAsBigInteger().equals(BigInteger.ZERO)) {
        rowSize += 1;
      } else {
        rowSize += 8;
      }
    }

    // Phase 6: GasLimit
    rowSize += 8;

    // Phase 7: To
    if (this.tx.getTo().isPresent()) {
      rowSize += 16;
    } else {
      rowSize += 1;
    }

    // Phase 8: Value
    if (this.tx.getValue().getAsBigInteger().equals(BigInteger.ZERO)) {
      rowSize += 1;
    } else {
      rowSize += 16;
    }

    // Phase 9: Data
    if (this.tx.getPayload().isEmpty()) {
      rowSize += 2; // 1 for prefix + 1 for padding
    } else {
      int dataSize = this.tx.getPayload().size();
      rowSize += 8 + LLARGE * ((dataSize - 1) / LLARGE + 1);
      rowSize += 2; // 2 lines of padding
    }

    // Phase 10: AccessList
    if (txType == 1 || txType == 2) {
      if (this.tx.getAccessList().orElseThrow().isEmpty()) {
        rowSize += 1;
      } else {
        // Rlp prefix of the AccessList list
        rowSize += 8;
        for (int i = 0; i < this.tx.getAccessList().orElseThrow().size(); i++) {
          rowSize += 8 + 16;
          if (this.tx.getAccessList().orElseThrow().get(i).storageKeys().isEmpty()) {
            rowSize += 1;
          } else {
            rowSize += 8 + 16 * this.tx.getAccessList().orElseThrow().get(i).storageKeys().size();
          }
        }
      }
    }

    // Phase 11: beta
    if (txType == 0) {
      rowSize += 8;
      if (this.tx.getV().compareTo(BigInteger.valueOf(28)) > 0) {
        rowSize += 9;
      }
    }

    // Phase 12: y
    if (txType == 1 || txType == 2) {
      rowSize += 1;
    }

    // Phase 13: r
    if (this.tx.getR().equals(BigInteger.ZERO)) {
      rowSize += 1;
    } else {
      rowSize += 16;
    }

    // Phase 14: s
    if (this.tx.getS().equals(BigInteger.ZERO)) {
      rowSize += 1;
    } else {
      rowSize += 16;
    }
    return rowSize;
  }
}
