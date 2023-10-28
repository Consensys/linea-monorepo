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

package net.consensys.linea.zktracer.module.hub.defer;

import java.util.ArrayList;
import java.util.List;

import net.consensys.linea.zktracer.module.hub.Hub;
import org.apache.commons.lang3.tuple.Pair;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.operation.Operation;
import org.hyperledger.besu.evm.worldstate.WorldView;

/**
 * Stores different categories of actions whose execution must be deferred later in the normal
 * transaction execution process.
 */
public class DeferRegistry {
  /** A list of actions deferred until the end of the current transaction */
  private final List<PostTransactionDefer> txDefers = new ArrayList<>();
  /** A list of actions deferred until the end of the current opcode execution */
  private final List<PostExecDefer> postExecDefers = new ArrayList<>();
  /** A list of actions deferred until the end of the current opcode execution */
  private final List<Pair<Integer, NextContextDefer>> contextReentry = new ArrayList<>();

  /** Schedule an action to be executed after the completion of the current opcode. */
  public void nextContext(NextContextDefer latch, int frameId) {
    this.contextReentry.add(Pair.of(frameId, latch));
  }

  /** Schedule an action to be executed after the completion of the current opcode. */
  public void postExec(PostExecDefer latch) {
    this.postExecDefers.add(latch);
  }

  /** Schedule an action to be executed at the end of the current transaction. */
  public void postTx(PostTransactionDefer defer) {
    this.txDefers.add(defer);
  }

  /**
   * Trigger the execution of the actions deferred to the next context.
   *
   * @param hub a {@link Hub} context
   * @param frame the new context {@link MessageFrame}
   */
  public void runNextContext(Hub hub, MessageFrame frame) {
    for (Pair<Integer, NextContextDefer> defer : this.contextReentry) {
      if (hub.currentFrame().parentFrame() == defer.getLeft()) {
        defer.getRight().runNextContext(hub, frame);
      }
    }
    this.contextReentry.clear();
  }

  /**
   * Trigger the execution of the actions deferred to the end of the transaction.
   *
   * @param hub the {@link Hub} context
   * @param world a {@link WorldView} on the state
   * @param tx the current {@link Transaction}
   */
  public void runPostTx(Hub hub, WorldView world, Transaction tx) {
    for (PostTransactionDefer defer : this.txDefers) {
      defer.runPostTx(hub, world, tx);
    }
    this.txDefers.clear();
  }

  /**
   * Trigger the execution of the actions deferred to the end of the current instruction execution.
   *
   * @param hub the {@link Hub} context
   * @param frame the {@link MessageFrame} of the transaction
   * @param result the {@link Operation.OperationResult} of the transaction
   */
  public void runPostExec(Hub hub, MessageFrame frame, Operation.OperationResult result) {
    for (PostExecDefer defer : this.postExecDefers) {
      defer.runPostExec(hub, frame, result);
    }
    this.postExecDefers.clear();
  }
}
