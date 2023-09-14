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

package net.consensys.linea.zktracer.module.hub.section;

import java.util.ArrayList;
import java.util.List;

import lombok.Getter;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.Trace;
import net.consensys.linea.zktracer.module.hub.chunks.CommonFragment;
import net.consensys.linea.zktracer.module.hub.chunks.StackFragment;
import net.consensys.linea.zktracer.module.hub.chunks.TraceFragment;

/** A TraceSection gather the trace lines linked to a single operation */
public abstract class TraceSection {
  /**
   * A TraceLine stores the information required to generate a trace line.
   *
   * @param common data required to trace shared columns
   * @param specific data required to trace perspective-specific columns
   */
  public record TraceLine(CommonFragment common, TraceFragment specific) {
    /**
     * Trace the line encoded within this chunk in the given trace builder.
     *
     * @param trace where to trace the line
     * @return the trace builder
     */
    public Trace.TraceBuilder trace(Trace.TraceBuilder trace) {
      assert common != null;
      assert specific != null;

      common.trace(trace);
      specific.trace(trace);

      return trace.fillAndValidateRow();
    }
  }

  /** Count the stack lines */
  private int stackRowsCounter;
  /** Count the non-stack lines */
  private int nonStackRowsCounter;

  /** A list of {@link TraceLine} representing the trace lines associated with this section. */
  @Getter List<TraceLine> lines = new ArrayList<>();

  /**
   * Fill the columns shared by all operations.
   *
   * @return a chunk representing the share columns
   */
  private CommonFragment traceCommon(Hub hub) {
    return new CommonFragment(
        hub.getTxNumber(),
        hub.getBatchNumber(),
        hub.getTxState(),
        hub.getStamp(),
        0, // retconned later on
        false, // retconned later on
        hub.opCodeData().instructionFamily(),
        hub.getExceptions().snapshot(),
        hub.currentFrame().getContextNumber(),
        hub.currentFrame().getContextNumber(),
        0, // TODO retconned
        false, // TODO
        false, // TODO
        hub.getPc(),
        hub.getPc(), // retconned later on
        hub.currentFrame().addressAsEWord(),
        hub.currentFrame().getCodeDeploymentNumber(),
        hub.currentFrame().isCodeDeploymentStatus(),
        hub.currentFrame().getAccountDeploymentNumber(),
        0,
        0,
        0,
        0,
        0, // TODO
        hub.opCodeData().stackSettings().twoLinesInstruction(),
        this.stackRowsCounter == 1,
        0, // retconned on sealing
        this.nonStackRowsCounter - 1);
  }

  /** Default creator for an empty section. */
  public TraceSection() {}

  /**
   * Add a fragment to the section while pairing it to its common piece.
   *
   * @param hub the execution context
   * @param fragment the fragment to insert
   */
  public final void addChunk(Hub hub, TraceFragment fragment) {
    assert !(fragment instanceof CommonFragment);
    if (fragment instanceof StackFragment) {
      this.stackRowsCounter++;
    } else {
      this.nonStackRowsCounter++;
    }

    this.lines.add(new TraceLine(traceCommon(hub), fragment));
  }

  /**
   * Create several {@link TraceLine} within this section for the specified fragments.
   *
   * @param hub the Hub linked to fragments execution
   * @param fragments the fragments to add to the section
   */
  public final void addChunks(Hub hub, TraceFragment... fragments) {
    for (TraceFragment chunk : fragments) {
      this.addChunk(hub, chunk);
    }
  }

  /**
   * Insert {@link TraceLine} related to the current state of the stack, then insert the provided
   * fragments in a single swoop.
   *
   * @param hub the execution context
   * @param fragments the fragments to insert
   */
  public final void addChunksAndStack(Hub hub, TraceFragment... fragments) {
    for (var stackChunk : hub.makeStackChunks()) {
      this.addChunk(hub, stackChunk);
    }
    this.addChunks(hub, fragments);
  }

  /**
   * Returns the context number associated with the operation encoded by this TraceChunk.
   *
   * @return the CN
   */
  public final int contextNumber() {
    return this.lines.get(0).common.contextNumber();
  }

  /**
   * Set the new context number associated with the operation encoded by this TraceChunk.
   *
   * @param contextNumber the new CN
   */
  public final void setContextNumber(int contextNumber) {
    for (TraceLine chunk : this.lines) {
      chunk.common.newContextNumber(contextNumber);
    }
  }

  /**
   * Returns the program counter associated with the operation encoded by this TraceSection.
   *
   * @return the PC
   */
  public final int pc() {
    return this.lines.get(0).common.pc();
  }

  /**
   * This method is called when the TraceSection is finished, to build required information
   * post-hoc.
   *
   * @param hub the linked {@link Hub} context
   */
  public void seal(Hub hub) {
    int nonStackLineNumbers =
        (int) this.lines.stream().filter(l -> !(l.specific instanceof StackFragment)).count();
    int nonStackLineCounter = 0;
    for (TraceLine chunk : this.lines) {
      if (!(chunk.specific instanceof StackFragment)) {
        nonStackLineCounter++;
        chunk.common.nonStackRowsCounter(nonStackLineCounter);
      }
      chunk.common.newPc(hub.lastPc());
      chunk.common.newContextNumber(hub.lastContextNumber());
      chunk.common.numberOfNonStackRows(nonStackLineNumbers);
    }
  }

  /**
   * This method is called when the transaction is finished to build required information post-hoc.
   *
   * @param hub the linked {@link Hub} context
   */
  public final void postTxRetcon(Hub hub) {
    for (TraceLine chunk : lines) {
      chunk.common().postTxRetcon(hub);
      chunk.specific().postTxRetcon(hub);
    }
  }

  /**
   * This method is called when the conflation is finished to build required information post-hoc.
   *
   * @param hub the linked {@link Hub} context
   */
  public final void postConflationRetcon(Hub hub) {
    for (TraceLine chunk : lines) {
      chunk.common().postConflationRetcon(hub);
      chunk.specific().postConflationRetcon(hub);
    }
  }
}
