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

import java.util.ArrayList;
import java.util.List;

import lombok.Getter;
import net.consensys.linea.zktracer.module.hub.section.TraceSection;
import org.hyperledger.besu.evm.worldstate.WorldView;

/**
 * Stores all the trace sections associated to the same transaction, stored in chronological order
 * of creation.
 */
public class TxTrace {
  /** The {@link TraceSection} of which this transaction trace is made of */
  @Getter private final List<TraceSection> trace = new ArrayList<>();

  public int size() {
    return this.trace.size();
  }

  /**
   * Returns the latest trace section, i.e. the most recent.
   *
   * @return the most recent trace section
   */
  public TraceSection currentSection() {
    return this.trace.get(this.size() - 1);
  }

  public TraceSection getSection(int i) {
    return this.trace.get(i);
  }

  /**
   * @return whether this trace is empty
   */
  public boolean isEmpty() {
    return this.trace.isEmpty();
  }

  /**
   * Add a {@link TraceSection} to this transaction trace.
   *
   * @param section the section to append
   */
  public void add(TraceSection section) {
    this.trace.add(section);
  }

  public long refundedGas() {
    long refundedGas = 0;
    for (TraceSection section : this.trace) {
      if (!section.hasReverted()) {
        refundedGas += section.refundDelta();
      }
    }
    return refundedGas;
  }

  /**
   * Run the action required at the end of the transaction to finish this trace.
   *
   * @param hub the exection context
   */
  public void postTxRetcon(Hub hub) {
    long leftoverGas = hub.getRemainingGas();
    long refundedGas = this.refundedGas();
    for (TraceSection section : this.trace) {
      section.postTxRetcon(hub, leftoverGas, refundedGas);
      section.setFinalGasRefundCounter(refundedGas);
    }
  }

  /**
   * Run the action required at the end of the conflation to finish this trace.
   *
   * @param hub the exection context
   */
  public void postConflationRetcon(Hub hub, WorldView world) {
    for (TraceSection section : this.trace) {
      section.postConflationRetcon(hub, world);
    }
  }

  /**
   * Generate the final numeric trace from the accumulated information.
   *
   * @param hubTrace where to materialize the trace
   */
  public void commit(Trace hubTrace) {
    for (TraceSection opSection : this.trace) {
      for (TraceSection.TraceLine line : opSection.getLines()) {
        line.trace(hubTrace);
      }
    }
  }

  /**
   * @return the line number in this transaction trace
   */
  public int lineCount() {
    return this.trace.stream().mapToInt(section -> section.getLines().size()).sum();
  }
}
