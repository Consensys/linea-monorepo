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

import java.util.Stack;

import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.container.StackedContainer;

public class State implements StackedContainer {
  private final Stack<TxState> state = new Stack<>();

  State() {}

  private TxState current() {
    return this.state.get(this.state.size() - 1);
  }

  TxState.Stamps stamps() {
    return this.current().stamps;
  }

  /**
   * @return the current transaction trace elements
   */
  TxTrace currentTxTrace() {
    return this.current().txTrace;
  }

  void postConflationRetcon(Hub hub) {
    for (TxState txState : this.state) {
      txState.txTrace.postConflationRetcon(hub, null /* TODO WorldView */);
    }
  }

  /**
   * Concretize the traces of all the accumulated transactions.
   *
   * @param hubTrace the trace builder to write to
   * @return the trace builder
   */
  Trace.TraceBuilder commit(Trace.TraceBuilder hubTrace) {
    for (TxState txState : this.state) {
      txState.txTrace().commit(hubTrace);
    }
    return hubTrace;
  }

  /**
   * @return the cumulated line numbers for all currently traced transactions
   */
  int lineCount() {
    return this.state.stream().mapToInt(s -> s.txTrace.lineCount()).sum();
  }

  @Override
  public void enter() {
    if (this.state.isEmpty()) {
      this.state.push(new TxState());
    } else {
      this.state.push(this.current().spinOff());
    }
  }

  @Override
  public void pop() {
    this.state.pop();
  }

  /** Describes the Hub state during a given transaction. */
  @Accessors(fluent = true)
  @Getter
  static class TxState {
    Stamps stamps;
    TxTrace txTrace;

    TxState() {
      this.stamps = new Stamps();
      this.txTrace = new TxTrace();
    }

    public TxState(Stamps stamps) {
      this.stamps = stamps;
      this.txTrace = new TxTrace();
    }

    TxState spinOff() {
      return new TxState(this.stamps.spinOff());
    }

    /** Stores all the stamps associated to the tracing of a transaction. */
    @Accessors(fluent = true)
    @Getter
    static class Stamps {
      private int hub = 0;
      private int mmu = 0;
      private int mxp = 0;
      private int hashInfo = 0;

      public Stamps() {}

      public Stamps(int hubStamp, int mmuStamp, int mxpStamp, int hashInfoStamp) {
        this.hub = hubStamp;
        this.mmu = mmuStamp;
        this.mxp = mxpStamp;
        this.hashInfo = hashInfoStamp;
      }

      Stamps spinOff() {
        return new Stamps(this.hub, this.mmu, this.mxp, this.hashInfo);
      }

      void stampHub() {
        this.hub++;
      }

      void stampSubmodules(final Signals signals) {
        if (signals.mmu()) {
          this.mmu++;
        }
        if (signals.mxp()) {
          this.mxp++;
        }
      }
    }
  }
}
