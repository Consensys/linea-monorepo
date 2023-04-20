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
package net.consensys.linea.zktracer.module.alu.ext.calculator;

import static net.consensys.linea.zktracer.module.Util.getBit;
import static net.consensys.linea.zktracer.module.Util.getOverflow;

import net.consensys.linea.zktracer.bytestheta.BaseTheta;
import net.consensys.linea.zktracer.bytestheta.BytesArray;
import org.apache.tuweni.units.bigints.UInt256;

/**
 * A utility class for computing the Hs and overflow values for extended modular arithmetic
 * operations.
 */
public class HByteCalculator {

  /**
   * Computes the Hs array and overflow values for the given arguments.
   *
   * @param aBytes the aBytes value.
   * @param bBytes the bBytes value.
   * @param hBytes the hBytes value.
   * @return the overflow values.
   */
  public static boolean[] computeHsAndOverflowH(
      BaseTheta aBytes, BaseTheta bBytes, BytesArray hBytes) {
    boolean[] overflow = new boolean[8];

    // Calculate alpha
    long alpha = calculateAlpha(aBytes, bBytes, hBytes);

    // Calculate Beta
    long beta = calculateBeta(aBytes, bBytes, hBytes);

    // Calculate gamma
    long gamma = calculateGamma(aBytes, bBytes, hBytes);

    overflow[0] = getBit(alpha, 0);
    overflow[1] = getBit(beta, 0);
    overflow[2] = getBit(beta, 1);
    overflow[3] = getBit(gamma, 0);
    return overflow;
  }

  /**
   * Calculates the alpha value for the given aBytes, bBytes, and hBytes values.
   *
   * @param aBytes the aBytes value.
   * @param bBytes the bBytes value.
   * @param hBytes the hBytes value.
   * @return the alpha value.
   */
  private static long calculateAlpha(BaseTheta aBytes, BaseTheta bBytes, BytesArray hBytes) {
    UInt256 prodA0xB1 = UInt256.fromBytes(aBytes.get(0)).multiply(UInt256.fromBytes(bBytes.get(1)));
    UInt256 prodA1xB0 = UInt256.fromBytes(aBytes.get(1)).multiply(UInt256.fromBytes(bBytes.get(0)));
    UInt256 sum1 = prodA0xB1.add(prodA1xB0);
    var truc = BaseTheta.fromBytes32(sum1);
    hBytes.set(0, truc.get(0));
    hBytes.set(1, truc.get(1));
    return getOverflow(sum1, 1, "alpha OOB");
  }

  /**
   * Calculates the beta value for the given aBytes, bBytes, and hBytes values.
   *
   * @param aBytes the aBytes value.
   * @param bBytes the bBytes value.
   * @param hBytes the hBytes value.
   * @return the beta value.
   */
  private static long calculateBeta(BaseTheta aBytes, BaseTheta bBytes, BytesArray hBytes) {
    UInt256 prodA0xB3 = UInt256.fromBytes(aBytes.get(0)).multiply(UInt256.fromBytes(bBytes.get(3)));
    UInt256 prodA1xB2 = UInt256.fromBytes(aBytes.get(1)).multiply(UInt256.fromBytes(bBytes.get(2)));
    UInt256 prodA2xB1 = UInt256.fromBytes(aBytes.get(2)).multiply(UInt256.fromBytes(bBytes.get(1)));
    UInt256 prodA3xB0 = UInt256.fromBytes(aBytes.get(3)).multiply(UInt256.fromBytes(bBytes.get(0)));
    UInt256 sum2 = prodA0xB3.add(prodA1xB2).add(prodA2xB1).add(prodA3xB0);
    var truc2 = BaseTheta.fromBytes32(sum2);
    hBytes.set(2, truc2.get(0));
    hBytes.set(3, truc2.get(1));
    return getOverflow(sum2, 3, "beta OOB");
  }
  /**
   * Calculates the gamma value for the given aBytes, bBytes, and hBytes values.
   *
   * @param aBytes the aBytes value.
   * @param bBytes the bBytes value.
   * @param hBytes the hBytes value.
   * @return the gamma value.
   */
  private static long calculateGamma(BaseTheta aBytes, BaseTheta bBytes, BytesArray hBytes) {
    UInt256 prodA2xB3 = UInt256.fromBytes(aBytes.get(2)).multiply(UInt256.fromBytes(bBytes.get(3)));
    UInt256 prodA3xB2 = UInt256.fromBytes(aBytes.get(3)).multiply(UInt256.fromBytes(bBytes.get(2)));
    UInt256 sum3 = prodA2xB3.add(prodA3xB2);
    var truc3 = BaseTheta.fromBytes32(sum3);
    hBytes.set(4, truc3.get(0));
    hBytes.set(5, truc3.get(1));
    return getOverflow(sum3, 1, "gamma OOB");
  }
}
