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
package net.consensys.linea.zktracer.module.alu.ext.calculator.addmod;

import static net.consensys.linea.zktracer.module.Util.getBit;
import static net.consensys.linea.zktracer.module.Util.getOverflow;

import net.consensys.linea.zktracer.bytestheta.BaseBytes;
import org.apache.tuweni.units.bigints.UInt256;

public class AddModOverflowResCalculator {
  static boolean[] calculateOverflow(BaseBytes arg1, BaseBytes arg2) {
    boolean[] overflowRes = new boolean[8];

    long lambda = calculateLambda(arg1, arg2);

    long mu = calculateMu(lambda, arg1, arg2);

    overflowRes[0] = getBit(lambda, 0);
    overflowRes[1] = getBit(mu, 0);
    return overflowRes;
  }

  private static long calculateLambda(BaseBytes arg1, BaseBytes arg2) {
    UInt256 sum = UInt256.fromBytes(arg1.getLow()).add(UInt256.fromBytes(arg2.getLow()));
    return getOverflow(sum, 1, "lambda out of range (ADDMOD)");
  }

  private static long calculateMu(long lambda, BaseBytes arg1, BaseBytes arg2) {
    UInt256 sum = UInt256.valueOf(lambda);
    sum = sum.add(UInt256.fromBytes(arg1.getHigh()));
    sum = sum.add(UInt256.fromBytes(arg2.getHigh()));
    return getOverflow(sum, 3, "mu out of range (ADDMOD)");
  }
}
