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

import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.module.Module;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.log.LogTopic;

@Accessors(fluent = true)
@RequiredArgsConstructor
public class L2Block implements Module {
  private final Address l2l1Address;
  private final LogTopic l2l1Topic;

  private static final int L1_MSG_INDICES_BYTES = 8;
  private static final int L1_TIMESTAMPS_BYTES = 8;
  private static final int ABI_OFFSET_BYTES = 32;
  private static final int ABI_LEN_BYTES = 32;

  /** The byte size of the RLP-encoded transaction of the conflation */
  @Getter private final Deque<Integer> sizesRlpEncodedTxs = new ArrayDeque<>();

  /** The byte size of the L2->L1 logs messages of the conflation */
  @Getter private final Deque<List<Integer>> l2l1LogSizes = new ArrayDeque<>();

  @Override
  public String moduleKey() {
    return "BLOCK_L1_SIZE";
  }

  @Override
  public void commitTransactionBundle() {
    this.sizesRlpEncodedTxs.push(0);
    this.l2l1LogSizes.push(new ArrayList<>());
  }

  @Override
  public void popTransactionBundle() {
    this.sizesRlpEncodedTxs.pop();
    this.l2l1LogSizes.pop();
  }

  @Override
  public int lineCount() {
    final int txCount = this.sizesRlpEncodedTxs.size();
    final int l2L1LogsCount = this.l2l1LogSizes.stream().mapToInt(List::size).sum();

    // This calculates the data size related to the transaction field of the
    // data sent on L1. This field is a double array of byte. Each subarray
    // corresponds to an RLP encoded transaction. The abi encoding incurs an
    // overhead for each transaction (32 bytes for an offset, and 32 bytes for
    // to encode the length of each sub bytes array). This overhead is also
    // incurred by the top-level array, hence the +1.
    int totalTxsRlpSize = (txCount + 1) * (ABI_OFFSET_BYTES + ABI_LEN_BYTES);
    for (int txRlpSize : this.sizesRlpEncodedTxs) {
      totalTxsRlpSize += txRlpSize;
    }

    // Calculates the data size related to the abi encoding of the list of the
    // from addresses. The field is a simple array of bytes20. We need to take
    // into account the offset and the length in the ABI encoding.
    final int totalFromSize = txCount * Address.SIZE + ABI_OFFSET_BYTES + ABI_LEN_BYTES;

    // Accumulates the data occupied for the hashes of the L2 to L1 messages
    // hashes each of them occupies 32 bytes. Also accounts for the overheads
    // of L2 and L1 messages encoding.
    final int totalL2L1Logs = Hash.SIZE * l2L1LogsCount + ABI_OFFSET_BYTES + ABI_LEN_BYTES;

    int l1Size = totalTxsRlpSize + totalL2L1Logs + totalFromSize;

    // Account for the overheads of sending the resulting root hash, the
    // timestamp and the L1 msg reception. For a sequence of conflated L2 blocks
    // , we will need to also need to send the initial timestamp and the parent
    // state root hash. Since we cannot forsee, at this point, the number of
    // blocks that will be conflated together with this block we make the worst
    // assumption that the block will be conflated alone. This corresponds to
    // counting twice the root hash and the timestamps. For the L1 messages, we
    // unfortunately do not have the data in the tracer yet. For that reason,
    // we also make a worst-case assumption that that every transaction is a
    // batch reception on layer 2. Finally, since what is sent on L1 is an array
    // of L2BlockData, we also make a worst-case assumption that the block will
    // be alone in the structure and account for the ABI encoding.
    l1Size +=
        2 * L1_TIMESTAMPS_BYTES
            + // the timestamp
            2 * Hash.SIZE
            + // the root hash
            L1_MSG_INDICES_BYTES * txCount
            + ABI_LEN_BYTES
            + ABI_OFFSET_BYTES
            + // the L1 messages
            ABI_LEN_BYTES
            + ABI_OFFSET_BYTES; // abi overheads for the blockdata struct.

    return l1Size;
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    throw new IllegalStateException("non-tracing module");
  }

  @Override
  public void traceEndTx(TransactionProcessingMetadata tx) {
    for (Log log : tx.getLogs()) {
      if (log.getLogger().equals(l2l1Address) && log.getTopics().contains(l2l1Topic)) {
        this.l2l1LogSizes.peek().add(log.getData().size());
      }
    }

    this.sizesRlpEncodedTxs.push(
        this.sizesRlpEncodedTxs.pop() + tx.getBesuTransaction().encoded().size());
  }

  public int l2l1LogsCount() {
    return this.l2l1LogSizes.stream().mapToInt(List::size).sum();
  }
}
