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

package net.consensys.linea.zktracer.module.hub.state;

import java.util.*;

import lombok.EqualsAndHashCode;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.stacked.CountOnlyOperation;
import net.consensys.linea.zktracer.container.stacked.StackedList;
import net.consensys.linea.zktracer.module.hub.HubProcessingPhase;
import net.consensys.linea.zktracer.module.hub.fragment.storage.StorageFragment;
import net.consensys.linea.zktracer.module.hub.state.State.HubTransactionState.Stamps;
import net.consensys.linea.zktracer.types.EWord;
import org.hyperledger.besu.datatypes.Address;

@RequiredArgsConstructor
public class State {
  private final StackedList<HubTransactionState> state = new StackedList<>();

  @Getter
  @Accessors(fluent = true)
  private final CountOnlyOperation lineCounter = new CountOnlyOperation();

  public HubTransactionState current() {
    return state.getLast();
  }

  public Stamps stamps() {
    return current().stamps;
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

  @Getter
  @Setter
  @Accessors(fluent = true)
  HubProcessingPhase processingPhase;

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
  public TraceSections currentTransactionHubSections() {
    return current().traceSections;
  }

  /**
   * Concretize the traces of all the accumulated transactions.
   *
   * @param hubTrace the trace builder to write to
   * @return the trace builder
   */
  public Trace.Hub commit(Trace.Hub hubTrace) {
    for (HubTransactionState state : state.getAll()) {
      state.traceSections().commit(hubTrace);
    }
    return hubTrace;
  }

  public int txCount() {
    return state.size();
  }

  public void enterTransaction() {
    if (state.isEmpty()) {
      state.add(new HubTransactionState());
    } else {
      state.add(current().spinOff());
    }
  }

  public void popTransactionBundle() {
    state.popTransactionBundle();
    lineCounter.popTransactionBundle();
  }

  public void commitTransactionBundle() {
    state.commitTransactionBundle();
    lineCounter.commitTransactionBundle();
  }

  /** Describes the Hub state during a given transaction. */
  @Accessors(fluent = true)
  @Getter
  public static class HubTransactionState {
    Stamps stamps;
    TraceSections traceSections;

    HubTransactionState() {
      stamps = new Stamps();
      traceSections = new TraceSections();
    }

    public HubTransactionState(Stamps stamps) {
      this.stamps = stamps;
      traceSections = new TraceSections();
    }

    HubTransactionState spinOff() {
      return new HubTransactionState(stamps.snapshot());
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
