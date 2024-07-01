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

package net.consensys.linea.zktracer.module.limits;

import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Deque;
import java.util.List;

import com.google.common.base.Preconditions;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;
import net.consensys.linea.zktracer.module.hub.signals.PlatformController;
import net.consensys.linea.zktracer.module.limits.precompiles.EcRecoverEffectiveCall;
import net.consensys.linea.zktracer.module.shakiradata.ShakiraData;
import net.consensys.linea.zktracer.module.shakiradata.ShakiraDataOperation;
import net.consensys.linea.zktracer.module.shakiradata.ShakiraPrecompileType;
import net.consensys.linea.zktracer.opcode.OpCode;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;

@RequiredArgsConstructor
public class Keccak implements Module {
  private static final int ADDRESS_BYTES = 20;
  private static final int HASH_BYTES = 32;
  private static final int L1_MSG_INDICES_BYTES = 8;
  private static final int L1_TIMESTAMPS_BYTES = 8;
  private static final int PUBKEY_BYTES = 64;
  private static final int KECCAK_BIT_RATE = 1088;
  private static final int KECCAK_BYTE_RATE = KECCAK_BIT_RATE / 8; // TODO: find correct name

  private final Hub hub;
  private final EcRecoverEffectiveCall ecRec;
  private final L2Block l2Block;

  @Getter private final ShakiraData shakiraData;

  private final Deque<List<Long>> deployedCodeSizes = new ArrayDeque<>();
  private final Deque<List<Long>> sha3Sizes = new ArrayDeque<>();
  private final Deque<List<Long>> create2Sizes = new ArrayDeque<>();

  @Override
  public String moduleKey() {
    return "BLOCK_KECCAK";
  }

  @Override
  public void enterTransaction() {
    this.deployedCodeSizes.push(new ArrayList<>());
    this.sha3Sizes.push(new ArrayList<>());
    this.create2Sizes.push(new ArrayList<>());
  }

  @Override
  public void popTransaction() {
    this.deployedCodeSizes.pop();
    this.sha3Sizes.pop();
    this.create2Sizes.pop();
  }

  private static int numKeccak(long x) {
    final long r = (x + KECCAK_BYTE_RATE - 1) / KECCAK_BYTE_RATE;
    Preconditions.checkState(r < Integer.MAX_VALUE, "demented KECCAK");
    return (int) r;
  }

  @Override
  public void tracePreOpcode(final MessageFrame frame) {
    final OpCode opCode = this.hub.opCode();

    final PlatformController pch = this.hub.pch();

    if (Exceptions.none(pch.exceptions())) {
      // Capture calls to SHA3.
      if (opCode == OpCode.SHA3) {
        callShakira(frame, 0, 1, this.sha3Sizes);
      }

      // Capture contract deployment
      // TODO: compute the gas cost if we are under deployment.
      if (opCode == OpCode.RETURN && hub.currentFrame().underDeployment()) {
        callShakira(frame, 0, 1, this.deployedCodeSizes);
      }

      if (opCode == OpCode.CREATE2 && pch.aborts().none()) {
        callShakira(frame, 1, 2, this.create2Sizes);
      }
    }
  }

  private void callShakira(
      final MessageFrame frame,
      final int codeOffsetStackItemOffset,
      final int codeSizeStackItemOffset,
      final Deque<List<Long>> codeSizes) {
    final long codeSize = Words.clampedToLong(frame.getStackItem(codeSizeStackItemOffset));
    codeSizes.peek().add(codeSize);

    if (codeSize != 0) {
      final long codeOffset = Words.clampedToLong(frame.getStackItem(codeOffsetStackItemOffset));
      final Bytes byteCode = frame.shadowReadMemory(codeOffset, codeSize);

      this.shakiraData.call(
          new ShakiraDataOperation(hub.stamp(), ShakiraPrecompileType.KECCAK, byteCode));
    }
  }

  @Override
  public int lineCount() {
    final int l2L1LogsCount = this.l2Block.l2l1LogSizes().stream().mapToInt(List::size).sum();
    final int txCount = this.l2Block.sizesRlpEncodedTxs().size();
    final int ecRecoverCount = ecRec.lineCount();

    // From tx RLPs, used both for both the signature verification and the
    // public input computation.
    return this.l2Block.sizesRlpEncodedTxs().stream().mapToInt(Keccak::numKeccak).sum()
        // From deployed contracts,
        // @alex, this formula suggests that the same data is hashed twice. Is this
        // accurate? If this is actually the same data then we should not need to
        // prove it twice. If the second time the data is hashed with a few extra
        // bytes this should be accounted for : numKeccak(l) + numKeccak(l + extra)
        + this.deployedCodeSizes.stream().flatMap(List::stream).mapToInt(Keccak::numKeccak).sum()
        // From ecRecover precompiles,
        // This accounts for the keccak of the recovered public keys to derive the
        // addresses. This also accounts for the transactions signatures
        // verifications.
        + (txCount + ecRecoverCount) * numKeccak(PUBKEY_BYTES)
        // From SHA3 opcode
        + this.sha3Sizes.stream().flatMap(List::stream).mapToInt(Keccak::numKeccak).sum()

        // From public input computation. This accounts for the hashing of:
        // - The block data hash:
        // 		- hashing the list of the transaction hashes
        //		- hashing the list of the L2L1 messages hashes
        // 		- hashing the list of the from addresses of the transactions
        // 		- hashing the list of the batch reception indices
        //		- hashing the above resulting hashes together to obtain the hash
        //			for the current block data
        + txCount
            * (numKeccak(HASH_BYTES) + numKeccak(ADDRESS_BYTES) + numKeccak(L1_MSG_INDICES_BYTES))
        + l2L1LogsCount * numKeccak(HASH_BYTES)
        + numKeccak(4 * HASH_BYTES) // 4 because there are 4 fields above

        // - The top-level structure (with the worst-case assumption, the
        // 	  current block is alone in the conflation). This includes:
        // 		- hashing concatenation of the state-root hashes for all blocks +1
        //			for the parent state-root hash.
        //		- hashing the list of the timestamps including the last finalized
        //			timestamp.
        //		- hashing the list of the block data hashes.
        //		- the first block number in the conflation list (formatted over 32
        // 			bytes). Note: it does not need to be hashed. It will just be
        //			included directly in the final hash.
        //		- the hash of the above fields, to obtain the final public input
        + 2 * numKeccak(HASH_BYTES) // one for the parent, one for the current block
        + 2 * numKeccak(L1_TIMESTAMPS_BYTES)
        + numKeccak(HASH_BYTES) // for the block data hash
        + numKeccak(4 * HASH_BYTES);
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    throw new IllegalStateException("non-tracing module");
  }
}
