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

package org.hyperledger.besu.tests.acceptance.dsl;

/** Helper functions for creating and manipulating blocks and block headers. */
public class BlockUtils {

  /**
   * Creates a new block header.
   *
   * @param block Ethereum block
   * @param blockHeaderFunctions functions for calculating parts of a {@link BlockHeader}
   * @return a block header of type {@link BlockHeader}
   */
  //  public static BlockHeader createBlockHeader(
  //      final Block block, final BlockHeaderFunctions blockHeaderFunctions) {
  //    final Hash mixHash =
  //        block.getMixHash() == null
  //            ? Hash.fromHexStringLenient("0x0")
  //            : fromHexString(block.getMixHash());
  //
  //    return new BlockHeader(
  //        fromHexString(block.getParentHash()),
  //        fromHexString(block.getSha3Uncles()),
  //        Address.fromHexString(block.getMiner()),
  //        fromHexString(block.getStateRoot()),
  //        fromHexString(block.getTransactionsRoot()),
  //        fromHexString(block.getReceiptsRoot()),
  //        LogsBloomFilter.fromHexString(block.getLogsBloom()),
  //        Difficulty.fromHexString(block.getDifficultyRaw()),
  //        block.getNumber().longValue(),
  //        block.getGasLimit().longValue(),
  //        block.getGasUsed().longValue(),
  //        block.getTimestamp().longValue(),
  //        Bytes.fromHexString(block.getExtraData()),
  //        null,
  //        mixHash,
  //        new BigInteger(block.getNonceRaw().substring(2), 16).longValue(),
  //        null,
  //        0L,
  //        null,
  //        null,
  //        null,
  //        blockHeaderFunctions);
  //  }
}
