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

package net.consensys.linea.zktracer.testing;

import java.util.Optional;

import lombok.RequiredArgsConstructor;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.frame.BlockValues;

/** An implementation of {@link BlockValues} for testing purposes. */
@RequiredArgsConstructor
public class FakeBlockValues implements BlockValues {
  private static final Bytes DEFAULT_DIFFICULTY_BYTES = UInt256.ZERO;
  private static final long DEFAULT_NUMBER = 0L;
  private static final long DEFAULT_GAS_LIMIT = 0L;
  private static final long DEFAULT_TIMESTAMP = 0L;
  private static final Optional<Wei> DEFAULT_BASE_FEE = Optional.empty();

  private final long number;
  private final long gasLimit;
  private final long timestamp;
  private final Bytes difficultyBytes;
  private final Optional<Wei> baseFee;

  /** Constructor with sane defaults. */
  public FakeBlockValues() {
    this(
        DEFAULT_NUMBER,
        DEFAULT_GAS_LIMIT,
        DEFAULT_TIMESTAMP,
        DEFAULT_DIFFICULTY_BYTES,
        DEFAULT_BASE_FEE);
  }

  /**
   * Constructor allowing specification of block number argument.
   *
   * @param number block number
   */
  public FakeBlockValues(final long number) {
    this(number, DEFAULT_GAS_LIMIT, DEFAULT_TIMESTAMP, DEFAULT_DIFFICULTY_BYTES, DEFAULT_BASE_FEE);
  }

  /**
   * Constructor allowing specification of a base fee of the block.
   *
   * @param baseFee base fee of the block
   */
  public FakeBlockValues(final Optional<Wei> baseFee) {
    this(DEFAULT_NUMBER, DEFAULT_GAS_LIMIT, DEFAULT_TIMESTAMP, DEFAULT_DIFFICULTY_BYTES, baseFee);
  }

  @Override
  public long getNumber() {
    return number;
  }

  @Override
  public Optional<Wei> getBaseFee() {
    return baseFee;
  }

  @Override
  public Bytes getDifficultyBytes() {
    return difficultyBytes;
  }

  @Override
  public long getGasLimit() {
    return gasLimit;
  }

  @Override
  public long getTimestamp() {
    return timestamp;
  }
}
