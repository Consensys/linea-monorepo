/*
 * Copyright ConsenSys AG.
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

import net.consensys.linea.zktracer.module.hub.section.TraceSection;

public class TxTrace {
  private final List<TraceSection> trace = new ArrayList<>();

  public int size() {
    return this.trace.size();
  }

  public TraceSection currentSection() {
    return this.trace.get(this.size() - 1);
  }

  public boolean isEmpty() {
    return this.trace.isEmpty();
  }

  public void add(TraceSection section) {
    this.trace.add(section);
  }

  public void postTxRetcon(Hub hub) {
    long refundedGas = 0;
    for (TraceSection section : this.trace) {
      if (!section.hasReverted()) {
        refundedGas += section.refundDelta();
      }
      section.postTxRetcon(hub, refundedGas);
    }

    for (TraceSection section : this.trace) {
      section.setFinalGasRefundCounter(refundedGas);
    }
  }

  public void postConflationRetcon(Hub hub) {
    for (TraceSection section : this.trace) {
      section.postConflationRetcon(hub);
    }
  }

  public void commit(Trace.TraceBuilder hubTrace) {
    for (TraceSection opSection : this.trace) {
      for (TraceSection.TraceLine line : opSection.getLines()) {
        line.trace(hubTrace);
      }
    }
  }

  public int lineCount() {
    int count = 0;
    for (TraceSection opSection : this.trace) {
      count += opSection.getLines().size();
    }
    return count;
  }
}
