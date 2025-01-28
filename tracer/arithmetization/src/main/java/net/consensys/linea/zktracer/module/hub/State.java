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

package net.consensys.linea.zktracer.module.hub;

import java.util.*;

import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.container.StackedContainer;
import net.consensys.linea.zktracer.container.stacked.CountOnlyOperation;
import net.consensys.linea.zktracer.module.hub.State.TxState.Stamps;
import net.consensys.linea.zktracer.module.hub.fragment.storage.StorageFragment;
import net.consensys.linea.zktracer.types.EWord;
import org.hyperledger.besu.datatypes.Address;

public class State implements StackedContainer {
  private final Deque<TxState> state = new ArrayDeque<>();

  @Getter
  @Accessors(fluent = true)
  private final CountOnlyOperation lineCounter = new CountOnlyOperation();

  State() {}

  public TxState current() {
    return state.peek();
  }

  public Stamps stamps() {
    return this.current().stamps;
  }

  /** Increments at commit time */
  @Getter
  @Accessors(fluent = true)
  private int mxpStamp = 0;

  /** Increments at commit time */
  @Getter
  @Accessors(fluent = true)
  private int mmuStamp = 0;

  public void incrementMmuStamp() {
    mmuStamp++;
  }

  public void incrementMxpStamp() {
    mxpStamp++;
  }

  @Getter @Setter HubProcessingPhase processingPhase;

  @RequiredArgsConstructor
  @EqualsAndHashCode
  @Getter
  public static class StorageSlotIdentifier {
    final Address address;
    final int deploymentNumber;
    final EWord storageKey;
  }

  public void updateOrInsertStorageSlotOccurrence(
      StorageSlotIdentifier slotIdentifier, StorageFragment storageFragment) {
    final HashMap<StorageSlotIdentifier, StorageFragmentPair> current =
        firstAndLastStorageSlotOccurrences.getLast();
    if (current.containsKey(slotIdentifier)) {
      current.get(slotIdentifier).update(storageFragment);
    } else {
      current.put(slotIdentifier, new State.StorageFragmentPair(storageFragment));
    }
  }

  @Getter
  public static class StorageFragmentPair {
    final StorageFragment firstOccurrence;
    @Setter StorageFragment finalOccurrence;

    public StorageFragmentPair(StorageFragment firstOccurrence) {
      this.firstOccurrence = firstOccurrence;
      this.finalOccurrence = firstOccurrence;
    }

    public void update(StorageFragment current) {
      setFinalOccurrence(current);
    }
  }

  // initialized here
  public ArrayList<HashMap<StorageSlotIdentifier, StorageFragmentPair>>
      firstAndLastStorageSlotOccurrences = new ArrayList<>();

  /**
   * @return the current transaction trace elements
   */
  TxTrace currentTxTrace() {
    return current().txTrace;
  }

  /**
   * Concretize the traces of all the accumulated transactions.
   *
   * @param hubTrace the trace builder to write to
   * @return the trace builder
   */
  Trace commit(Trace hubTrace) {
    for (Iterator<TxState> it = state.descendingIterator(); it.hasNext(); ) {
      final TxState txState = it.next();
      txState.txTrace().commit(hubTrace);
    }
    return hubTrace;
  }

  int txCount() {
    return state.size();
  }

  @Override
  public void enter() {
    if (state.isEmpty()) {
      state.push(new TxState());
    } else {
      state.push(this.current().spinOff());
    }
    lineCounter.enter();
  }

  @Override
  public void pop() {
    state.pop();
    lineCounter.pop();
  }

  /** Describes the Hub state during a given transaction. */
  @Accessors(fluent = true)
  @Getter
  public static class TxState {
    Stamps stamps;
    TxTrace txTrace;

    TxState() {
      stamps = new Stamps();
      txTrace = new TxTrace();
    }

    public TxState(Stamps stamps) {
      this.stamps = stamps;
      txTrace = new TxTrace();
    }

    TxState spinOff() {
      return new TxState(stamps.snapshot());
    }

    /**
     * Stores the HUB and LOG stamps associated to the tracing of a transaction. As the MMU and MXP
     * stamps increment only at commit time and are not required during execution, they are not part
     * of this class.
     */
    @Accessors(fluent = true)
    @Getter
    public static class Stamps {
      private int hub = 0; // increments during execution
      private int log = 0; // increments at RunPostTx

      public Stamps() {}

      public Stamps(final int hubStamp, final int logStamp) {
        hub = hubStamp;
        log = logStamp;
      }

      public Stamps snapshot() {
        return new Stamps(hub, log);
      }

      public void incrementHubStamp() {
        hub++;
      }

      public int incrementLogStamp() {
        return ++log;
      }
    }
  }
}
