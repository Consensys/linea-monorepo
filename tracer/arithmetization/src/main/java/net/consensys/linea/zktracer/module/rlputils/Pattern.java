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

package net.consensys.linea.zktracer.module.rlputils;

import java.math.BigInteger;
import java.util.ArrayList;

import com.google.common.base.Preconditions;
import org.apache.tuweni.bytes.Bytes;

public class Pattern {
  /**
   * Returns the size of RLP(something) where something is of size inputSize (!=1) (it can be ZERO
   * though).
   */
  public static int outerRlpSize(int inputSize) {
    int rlpSize = inputSize;
    if (inputSize == 1) {
      // TODO panic
    } else {
      rlpSize += 1;
      if (inputSize >= 56) {
        rlpSize += Bytes.minimalBytes(inputSize).size();
      }
    }
    return rlpSize;
  }

  /**
   * Create the Power and AccSize list of the ByteCountAndPower RLP pattern.
   *
   * @param inputByteLen represents the number of meaningful bytes of inputByte, i.e. without the
   *     zero left padding
   * @param nbStep
   * @return
   */
  public static ByteCountAndPowerOutput byteCounting(int inputByteLen, int nbStep) {

    ArrayList<BigInteger> powerInit = new ArrayList<>(nbStep);
    ArrayList<Integer> acc = new ArrayList<>(nbStep);
    ByteCountAndPowerOutput output = new ByteCountAndPowerOutput(powerInit, acc);

    BigInteger power;
    int accByteSize = 0;
    int offset = 16 - nbStep;

    if (inputByteLen == nbStep) {
      power = BigInteger.valueOf(256).pow(offset);
      accByteSize = 1;
    } else {
      offset += 1;
      power = BigInteger.valueOf(256).pow(offset);
    }

    output.powerList().add(0, power);
    output.accByteSizeList().add(0, accByteSize);

    for (int i = 1; i < nbStep; i++) {
      if (inputByteLen + i < nbStep) {
        power = power.multiply(BigInteger.valueOf(256));
      } else {
        accByteSize += 1;
      }
      output.powerList().add(i, power);
      output.accByteSizeList().add(i, accByteSize);
    }
    return output;
  }

  /**
   * Create the Bit and BitDec list of the RLP pattern of an int.
   *
   * @param input
   * @param nbStep
   * @return
   */
  public static BitDecOutput bitDecomposition(int input, int nbStep) {
    final int nbStepMin = 8;
    Preconditions.checkArgument(
        nbStep >= nbStepMin, "Number of steps must be at least " + nbStepMin);

    ArrayList<Boolean> bit = new ArrayList<>(nbStep);
    ArrayList<Integer> acc = new ArrayList<>(nbStep);
    for (int i = 0; i < nbStep - nbStepMin; i++) {
      bit.add(i, false);
      acc.add(i, 0);
    }
    BitDecOutput output = new BitDecOutput(bit, acc);

    int bitAcc = 0;
    boolean bitDec = false;
    double div = 0;

    for (int i = nbStepMin - 1; i >= 0; i--) {
      div = Math.pow(2, i);
      bitAcc *= 2;

      if (input >= div) {
        bitDec = true;
        bitAcc += 1;
        input -= (int) div;
      } else {
        bitDec = false;
      }

      output.bitDecList().add(nbStep - i - 1, bitDec);
      output.bitAccList().add(nbStep - i - 1, bitAcc);
    }
    return output;
  }
}
