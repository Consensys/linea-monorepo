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

package net.consensys.linea.zktracer;

import java.math.BigInteger;

import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.units.bigints.BaseUInt256Value;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.Quantity;

public final class EWord extends BaseUInt256Value<EWord> implements Quantity {
  /** The constant ZERO. */
  public static final EWord ZERO = of(0);

  /** The constant ONE. */
  public static final EWord ONE = of(1);

  /**
   * Instantiates a new EVM word.
   *
   * @param value the value
   */
  EWord(final UInt256 value) {
    super(value, EWord::new);
  }

  private EWord(final long v) {
    this(UInt256.valueOf(v));
  }

  private EWord(final BigInteger v) {
    this(UInt256.valueOf(v));
  }

  private EWord(final String hexString) {
    this(UInt256.fromHexString(hexString));
  }

  /**
   * EVM word of long.
   *
   * @param value the value
   * @return the EVM word
   */
  public static EWord of(final long value) {
    return new EWord(value);
  }

  /**
   * EVM word of {@link BigInteger}.
   *
   * @param value the value
   * @return the EVM word
   */
  public static EWord of(final BigInteger value) {
    return new EWord(value);
  }

  /**
   * EVM word of {@link UInt256}.
   *
   * @param value the value
   * @return the EVM word
   */
  public static EWord of(final UInt256 value) {
    return new EWord(value);
  }

  /**
   * EVM word of {@link Number}.
   *
   * @param value the value
   * @return the EVM word
   */
  public static EWord of(final Number value) {
    return new EWord((BigInteger) value);
  }

  /**
   * Wrap bytes into EVM word.
   *
   * @param value the value
   * @return the EVM word
   */
  public static EWord of(final Bytes value) {
    return new EWord(UInt256.fromBytes(value));
  }

  /**
   * From hex string to EVM word.
   *
   * @param str the str
   * @return the EVM word
   */
  public static EWord ofHexString(final String str) {
    return new EWord(str);
  }

  @Override
  public Number getValue() {
    return getAsBigInteger();
  }

  @Override
  public BigInteger getAsBigInteger() {
    return toBigInteger();
  }

  @Override
  public String toHexString() {
    return super.toHexString();
  }

  @Override
  public String toShortHexString() {
    return super.isZero() ? "0x0" : super.toShortHexString();
  }

  /**
   * Return the high half of the EWord
   *
   * @return the 16 high {@link Bytes}
   */
  public Bytes hi() {
    return this.toBytes().slice(0, 16);
  }

  /**
   * Return the low half of the EWord
   *
   * @return the 16 low {@link Bytes}
   */
  public Bytes lo() {
    return this.toBytes().slice(16);
  }

  /**
   * Returns a {@link BigInteger} containing the high half of the EWord
   *
   * @return the high bytes as a {@link BigInteger}
   */
  public BigInteger hiBigInt() {
    return this.hi().toUnsignedBigInteger();
  }

  /**
   * Returns a {@link BigInteger} containing the low half of the EWord
   *
   * @return the low bytes as a {@link BigInteger}
   */
  public BigInteger loBigInt() {
    return this.lo().toUnsignedBigInteger();
  }

  /**
   * From {@link Quantity} to EVM word.
   *
   * @param quantity the quantity
   * @return the EVM word
   */
  public static EWord ofQuantity(final Quantity quantity) {
    return EWord.of((Bytes) quantity);
  }
}
