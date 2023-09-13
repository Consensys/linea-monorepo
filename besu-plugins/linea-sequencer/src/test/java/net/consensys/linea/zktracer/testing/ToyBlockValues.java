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

import lombok.Builder;
import lombok.RequiredArgsConstructor;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.frame.BlockValues;

/** An implementation of {@link BlockValues} for testing purposes. */
@Builder
@RequiredArgsConstructor
public class ToyBlockValues implements BlockValues {
  private static final Bytes DEFAULT_DIFFICULTY_BYTES = UInt256.ZERO;
  private static final long DEFAULT_NUMBER = 0L;
  private static final long DEFAULT_GAS_LIMIT = 0L;
  private static final long DEFAULT_TIMESTAMP = 0L;
  private static final Optional<Wei> DEFAULT_BASE_FEE = Optional.empty();

  private final Long number;
  private final Long gasLimit;
  private final Long timestamp;
  private final Bytes difficultyBytes;
  private final Optional<Wei> baseFee;

  public static ToyBlockValues defaultValues() {
    return builder().build();
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

  /** Customizations applied on the generated Lombok {@link Builder}. */
  public static class ToyBlockValuesBuilder {

    /**
     * Customizations applied on the generated Lombok {@link Builder}'s build method.
     *
     * @return an instance of {@link ToyBlockValues}
     */
    public ToyBlockValues build() {
      return new ToyBlockValues(
          Optional.ofNullable(number).orElse(DEFAULT_NUMBER),
          Optional.ofNullable(gasLimit).orElse(DEFAULT_GAS_LIMIT),
          Optional.ofNullable(timestamp).orElse(DEFAULT_TIMESTAMP),
          Optional.ofNullable(difficultyBytes).orElse(DEFAULT_DIFFICULTY_BYTES),
          Optional.ofNullable(baseFee).orElse(DEFAULT_BASE_FEE));
    }
  }
}
