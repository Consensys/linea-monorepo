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
import java.util.Set;

import com.google.gson.Gson;
import net.consensys.linea.blockcapture.reapers.Reaper;
import net.consensys.linea.zktracer.ConflationAwareOperationTracer;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.AddressUtils;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
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
  public static final int MAX_RELATIVE_BLOCK = 256;

  private static final int MAX_BLOCK_ARG_SIZE = 8;

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
  public void traceEndConflation(WorldView state) {}

  @Override
  public void traceStartBlock(
      BlockHeader blockHeader, BlockBody blockBody, Address miningBeneficiary) {
    this.reaper.enterBlock(blockHeader, blockBody, miningBeneficiary);
  }

  @Override
  public void tracePrepareTransaction(WorldView worldView, Transaction transaction) {
    this.reaper.prepareTransaction(transaction);
  }

  @Override
  public void traceEndTransaction(
      WorldView world,
      Transaction tx,
      boolean status,
      Bytes output,
      List<Log> logs,
      long gasUsed,
      Set<Address> selfDestructs,
      long timeNs) {
    this.reaper.exitTransaction(world, status, output, logs, gasUsed, selfDestructs);
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

      case BALANCE -> {
        if (frame.stackSize() > 0) {
          final Address target = Words.toAddress(frame.getStackItem(0));
          this.reaper.touchAddress(target);
        }
      }

        // Failure condition if created address already exists
      case CREATE, CREATE2 -> {
        if (frame.stackSize() > 0) {
          final Address target = AddressUtils.getDeploymentAddress(frame);
          this.reaper.touchAddress(target);
        }
      }

        // Funds of the selfdestruct account are sent to the target account
      case SELFDESTRUCT -> {
        if (frame.stackSize() > 0) {
          final Address target = Words.toAddress(frame.getStackItem(0));
          this.reaper.touchAddress(target);
        }
      }

      case BLOCKHASH -> {
        if (frame.stackSize() > 0) {
          // Determine current block number
          final long currentBlockNumber = frame.getBlockValues().getNumber();
          final Bytes arg = frame.getStackItem(0).trimLeadingZeros();
          // Check arguments fits within 8 bytes
          if (arg.size() <= 8) {
            // Determine block number requested
            final long blockNumber = arg.toLong();
            // Sanity check block within last 256 blocks.
            if (blockNumber < currentBlockNumber && (currentBlockNumber - blockNumber) <= 256) {
              // Use enclosing frame to determine hash
              Hash blockHash = frame.getBlockHashLookup().apply(frame, blockNumber);
              // Record it was seen
              this.reaper.touchBlockHash(blockNumber, blockHash);
            }
          }
        }
      }
    }
  }

  public String toJson() {
    Gson gson = new Gson();
    return gson.toJson(this.reaper.collapse(this.worldUpdater));
  }
}
