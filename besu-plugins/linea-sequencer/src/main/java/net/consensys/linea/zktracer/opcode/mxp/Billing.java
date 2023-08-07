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

package net.consensys.linea.zktracer.opcode.mxp;

import net.consensys.linea.zktracer.opcode.gas.BillingRate;
import net.consensys.linea.zktracer.opcode.gas.Gas;

/**
 * An ancillary class to compute gas billing of some instructions.
 *
 * @param perUnit gas cost of a unit
 * @param billingRate the unit used to bill gas
 */
public record Billing(Gas perUnit, BillingRate billingRate, MxpType type) {
  public Billing() {
    this(null, null, MxpType.None);
  }

  /**
   * Create a billing scheme only dependent on the Mxp.
   *
   * @param type the MXP type
   * @return the billing scheme
   */
  public static Billing byMxp(MxpType type) {
    return new Billing(null, null, type);
  }

  /**
   * Create a per-word billing scheme.
   *
   * @param type the MXP type
   * @param wordPrice gas cost of a word
   * @return the billing scheme
   */
  public static Billing byWord(MxpType type, Gas wordPrice) {
    return new Billing(wordPrice, BillingRate.ByWord, type);
  }

  /**
   * Create a per-byte billing scheme.
   *
   * @param type the MXP type
   * @param bytePrice gas cost of a byte
   * @return the billing scheme
   */
  public static Billing byByte(MxpType type, Gas bytePrice) {
    return new Billing(bytePrice, BillingRate.ByByte, type);
  }
}
