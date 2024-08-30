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

package net.consensys.linea.zktracer.module.trm;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LLARGE;

import java.nio.MappedByteBuffer;
import java.util.List;

import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.set.StackedSet;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.AccessListEntry;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class Trm implements Module {
  static final int MAX_CT = LLARGE;
  static final int PIVOT_BIT_FLIPS_TO_TRUE = 12;

  private final StackedSet<TrmOperation> operations = new StackedSet<>();

  @Override
  public String moduleKey() {
    return "TRM";
  }

  @Override
  public void enterTransaction() {
    this.operations.enter();
  }

  @Override
  public void popTransaction() {
    this.operations.pop();
  }

  public Address callTrimming(Bytes32 rawHash) {
    operations.add(new TrmOperation(EWord.of(rawHash)));
    return Address.extract(rawHash);
  }

  public Address callTrimming(Bytes addressToTrim) {
    Bytes32 addressPadded = Bytes32.leftPad(addressToTrim);
    return callTrimming(addressPadded);
  }

  @Override
  public void traceStartTx(WorldView world, TransactionProcessingMetadata txMetaData) {
    // Add effective receiver Address
    operations.add(new TrmOperation(EWord.of(txMetaData.getEffectiveRecipient())));

    // Add Address in AccessList to warm
    final Transaction tx = txMetaData.getBesuTransaction();
    final TransactionType txType = tx.getType();

    switch (txType) {
      case ACCESS_LIST, EIP1559 -> {
        if (tx.getAccessList().isPresent()) {
          for (AccessListEntry entry : tx.getAccessList().get()) {
            operations.add(new TrmOperation(EWord.of(entry.address())));
          }
        }
      }
      case FRONTIER -> {
        return;
      }
      default -> {
        throw new IllegalStateException("TransactionType not supported: " + txType);
      }
    }
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    int stamp = 0;
    for (TrmOperation operation : this.operations) {
      stamp++;
      operation.trace(trace, stamp);
    }
  }

  @Override
  public int lineCount() {
    return this.operations.lineCount();
  }
}
