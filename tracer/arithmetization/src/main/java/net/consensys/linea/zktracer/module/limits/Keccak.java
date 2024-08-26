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

import java.util.List;

import com.google.common.base.Preconditions;
import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.module.limits.precompiles.EcRecoverEffectiveCall;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;

@RequiredArgsConstructor
public class Keccak extends CountingOnlyModule {
  private static final int ADDRESS_BYTES = Address.SIZE;
  private static final int HASH_BYTES = Hash.SIZE;
  private static final int L1_MSG_INDICES_BYTES = 8;
  private static final int L1_TIMESTAMPS_BYTES = 8;
  private static final int PUBKEY_BYTES = 64;
  private static final int KECCAK_BIT_RATE = 1088;
  private static final int KECCAK_BYTE_RATE = KECCAK_BIT_RATE / 8; // TODO: find correct name

  private final EcRecoverEffectiveCall ecRecoverEffectiveCall;
  private final L2Block l2Block;

  @Override
  public String moduleKey() {
    return "BLOCK_KECCAK";
  }

  @Override
  public void addPrecompileLimit(final int dataByteLength) {
    final int blockCount = numberOfKeccakBloc(dataByteLength);
    this.counts.add(blockCount);
  }

  @Override
  public int lineCount() {
    final int l2L1LogsCount = this.l2Block.l2l1LogSizes().stream().mapToInt(List::size).sum();
    final int txCount = this.l2Block.sizesRlpEncodedTxs().size();
    final int ecRecoverCount = ecRecoverEffectiveCall.lineCount();

    // From tx RLPs, used both for both the signature verification and the
    // public input computation.
    return this.l2Block.sizesRlpEncodedTxs().stream().mapToInt(Keccak::numberOfKeccakBloc).sum()
        // From ecRecover precompiles,
        // This accounts for the keccak of the recovered public keys to derive the
        // addresses. This also accounts for the transactions signatures
        // verifications.
        + (txCount + ecRecoverCount) * numberOfKeccakBloc(PUBKEY_BYTES)

        // From deployed contracts, SHA3 opcode, and CREATE2
        + counts.lineCount()

        /**
         * TODO: previous implem was missing CREATE2 and was the following: // From deployed
         * contracts, // @alex, this formula suggests that the same data is hashed twice. Is this //
         * accurate? If this is actually the same data then we should not need to // prove it twice.
         * If the second time the data is hashed with a few extra // bytes this should be accounted
         * for : numKeccak(l) + numKeccak(l + extra) +
         * this.deployedCodeSizes.stream().flatMap(List::stream).mapToInt(Keccak::numKeccak).sum()
         * // From SHA3 opcode +
         * this.sha3Sizes.stream().flatMap(List::stream).mapToInt(Keccak::numKeccak).sum()
         */

        // From public input computation. This accounts for the hashing of:
        // - The block data hash:
        // 		- hashing the list of the transaction hashes
        //		- hashing the list of the L2L1 messages hashes
        // 		- hashing the list of the from addresses of the transactions
        // 		- hashing the list of the batch reception indices
        //		- hashing the above resulting hashes together to obtain the hash
        //			for the current block data
        + txCount
            * (numberOfKeccakBloc(HASH_BYTES)
                + numberOfKeccakBloc(ADDRESS_BYTES)
                + numberOfKeccakBloc(L1_MSG_INDICES_BYTES))
        + l2L1LogsCount * numberOfKeccakBloc(HASH_BYTES)
        + numberOfKeccakBloc(4 * HASH_BYTES) // 4 because there are 4 fields above

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
        + 2 * numberOfKeccakBloc(HASH_BYTES) // one for the parent, one for the current block
        + 2 * numberOfKeccakBloc(L1_TIMESTAMPS_BYTES)
        + numberOfKeccakBloc(HASH_BYTES) // for the block data hash
        + numberOfKeccakBloc(4 * HASH_BYTES);
  }

  private static int numberOfKeccakBloc(final long dataByteLength) {
    final long r = (dataByteLength + KECCAK_BYTE_RATE - 1) / KECCAK_BYTE_RATE;
    Preconditions.checkState(r < Integer.MAX_VALUE, "demented KECCAK");
    return (int) r;
  }
}
