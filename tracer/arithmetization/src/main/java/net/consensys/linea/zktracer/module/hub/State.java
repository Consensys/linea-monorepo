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

import java.util.ArrayDeque;
import java.util.Deque;
import java.util.Iterator;

import lombok.Getter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.container.StackedContainer;
import net.consensys.linea.zktracer.module.hub.State.TxState.Stamps;
import net.consensys.linea.zktracer.module.hub.signals.PlatformController;

public class State implements StackedContainer {
  private final Deque<TxState> state = new ArrayDeque<>(50);

  State() {}

  private TxState current() {
    return this.state.peek();
  }

  public Stamps stamps() {
    return this.current().stamps;
  }

  /**
   * @return the current transaction trace elements
   */
  TxTrace currentTxTrace() {
    return this.current().txTrace;
  }

  /**
   * Concretize the traces of all the accumulated transactions.
   *
   * @param hubTrace the trace builder to write to
   * @return the trace builder
   */
  Trace commit(Trace hubTrace) {
    for (Iterator<TxState> it = this.state.descendingIterator(); it.hasNext(); ) {
      final TxState txState = it.next();
      txState.txTrace().commit(hubTrace);
    }
    return hubTrace;
  }

  int txCount() {
    return this.state.size();
  }

  /**
   * @return the cumulated line numbers for all currently traced transactions
   */
  int lineCount() {
    int sum = 0;
    for (TxState s : this.state) {
      sum += s.txTrace.lineCount();
    }
    return sum;
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
    public static class Stamps {
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

      void stampSubmodules(final PlatformController platformController) {
        if (platformController.signals().mmu()) {
          this.mmu++;
        }
        if (platformController.signals().mxp()) {
          this.mxp++;
        }
      }
    }
  }
}
