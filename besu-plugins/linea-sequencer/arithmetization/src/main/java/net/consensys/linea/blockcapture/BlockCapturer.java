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

package net.consensys.linea.blockcapture;

import java.util.List;

import com.google.gson.Gson;
import net.consensys.linea.blockcapture.reapers.Reaper;
import net.consensys.linea.zktracer.ConflationAwareOperationTracer;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.worldstate.WorldUpdater;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;

public class BlockCapturer implements ConflationAwareOperationTracer {
  /**
   * The {@link Reaper} will collect all the data that will need to be mimicked to replay the block.
   */
  private final Reaper reaper = new Reaper();

  /**
   * This keeps a pointer to the initial state (i.e. ) to be used at the end of tracing to store the
   * minimal required information to replay the conflation.
   */
  private WorldUpdater worldUpdater;

  /**
   * Must be called **before** any tracing activity.
   *
   * @param worldUpdater the state of the world before the conflation is applied
   */
  public void setWorld(WorldUpdater worldUpdater) {
    this.worldUpdater = worldUpdater;
  }

  @Override
  public void traceStartConflation(long numBlocksInConflation) {}

  @Override
  public void traceEndConflation() {}

  @Override
  public void traceStartBlock(BlockHeader blockHeader, BlockBody blockBody) {
    this.reaper.enterBlock(blockHeader, blockBody);
  }

  @Override
  public void traceStartTransaction(WorldView worldView, Transaction transaction) {
    this.reaper.enterTransaction(transaction);
  }

  @Override
  public void traceEndTransaction(
      WorldView worldView,
      Transaction tx,
      boolean status,
      Bytes output,
      List<Log> logs,
      long gasUsed,
      long timeNs) {
    this.reaper.exitTransaction(status);
  }

  /**
   * This method only bothers with instruction putatively accessing the state as it was at the
   * beginning of the conflation.
   *
   * @param frame the frame
   */
  @Override
  public void tracePreExecution(MessageFrame frame) {
    final OpCode opCode = OpCode.of(frame.getCurrentOperation().getOpcode());

    switch (opCode) {
        // These access contracts potentially existing before the conflation played out.
      case EXTCODESIZE, EXTCODECOPY, EXTCODEHASH -> {
        if (frame.stackSize() > 0) {
          final Address target = Words.toAddress(frame.getStackItem(0));
          this.reaper.touchAddress(target);
        }
      }

        // SLOAD may access storage cells whose value was set before the conflation execution.
      case SLOAD -> {
        if (frame.stackSize() > 0) {
          final Account account = frame.getWorldUpdater().get(frame.getRecipientAddress());
          final Address address = account.getAddress();
          final UInt256 key = UInt256.fromBytes(frame.getStackItem(0));
          this.reaper.touchStorage(address, key);
        }
      }

        // SSTORE needs to know the previous storage value for correct gas computation
      case SSTORE -> {
        if (frame.stackSize() > 1) {
          final Account account = frame.getWorldUpdater().get(frame.getRecipientAddress());
          final Address address = account.getAddress();
          final UInt256 key = UInt256.fromBytes(frame.getStackItem(0));
          this.reaper.touchStorage(address, key);
        }
      }

        // These access contracts potentially existing before the conflation played out.
      case CALL, CALLCODE, DELEGATECALL, STATICCALL -> {
        if (frame.stackSize() > 1) {
          final Address target = Words.toAddress(frame.getStackItem(1));
          this.reaper.touchAddress(target);
        }
      }
    }
  }

  public String toJson() {
    Gson gson = new Gson();
    return gson.toJson(this.reaper.collapse(this.worldUpdater));
  }
}
