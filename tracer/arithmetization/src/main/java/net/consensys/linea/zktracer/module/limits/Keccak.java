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

import com.google.common.base.Preconditions;
import lombok.Getter;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.container.module.CountingOnlyModule;
import net.consensys.linea.zktracer.container.stacked.CountOnlyOperation;
import net.consensys.linea.zktracer.module.limits.precompiles.EcRecoverEffectiveCall;

@RequiredArgsConstructor
@Getter
@Accessors(fluent = true)
public class Keccak implements CountingOnlyModule {
  private final CountOnlyOperation counts = new CountOnlyOperation();
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
  public void updateTally(final int count) {
    final int blockCount = numberOfKeccakBloc(count);
    counts.add(blockCount);
  }

  @Override
  public int lineCount() {
    final int txCount = l2Block.numberOfTransactions().lineCount();
    final int ecRecoverCount = ecRecoverEffectiveCall.lineCount();

    return
    // From tx RLPs, used both for both the signature verification and the
    // public input computation. As l2Block.sizesRlpEncodedTxs().lineCount() gives the size of the
    // concatenation of all
    // the RLP-encoded transactions, we add txCount to not miss keccak blocks.
    (numberOfKeccakBloc(l2Block.sizesRlpEncodedTxs().lineCount()) + txCount)
        // From ecRecover precompiles,
        // This accounts for the keccak of the recovered public keys to derive the
        // addresses. This also accounts for the transactions signatures
        // verifications.
        + (txCount + ecRecoverCount) * numberOfKeccakBloc(PUBKEY_BYTES)

        // From deployed contracts, number of Keccak block for SHA3 and CREATE2, and from RLP_ADDR
        + counts.lineCount();
  }

  public static int numberOfKeccakBloc(final long dataByteLength) {
    final long r = (dataByteLength + KECCAK_BYTE_RATE - 1) / KECCAK_BYTE_RATE;
    Preconditions.checkState(r < Integer.MAX_VALUE, "demented KECCAK");
    return (int) r;
  }
}
