/*
 * Copyright ConsenSys Inc.
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

package net.consensys.linea.zktracer.types;

import java.io.IOException;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.Map;

import com.google.common.base.Preconditions;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.toml.Toml;
import org.apache.tuweni.toml.TomlTable;

public class Utils {

  /**
   * Add zeroes to the left of the {@link Bytes} to create {@link Bytes} of the given size. The
   * wantedSize must be at least the size of the Bytes.
   *
   * @param input
   * @param wantedSize
   * @return
   */
  public static Bytes leftPadTo(Bytes input, int wantedSize) {
    Preconditions.checkArgument(
        wantedSize >= input.size(), "wantedSize can't be shorter than the input size");
    return Bytes.concatenate(Bytes.repeat((byte) 0, wantedSize - input.size()), input);
  }

  public static Bytes rightPadTo(Bytes input, int wantedSize) {
    Preconditions.checkArgument(
        wantedSize >= input.size(), "wantedSize can't be shorter than the input size");
    return Bytes.concatenate(input, Bytes.repeat((byte) 0, wantedSize - input.size()));
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

  public static Map<String, Integer> computeSpillings() throws IOException {
    final Map<String, Integer> spillings = new HashMap<>();

    final TomlTable table =
        Toml.parse(Utils.class.getClassLoader().getResourceAsStream("spillings.toml"))
            .getTable("spillings");
    table.toMap().keySet().forEach(k -> spillings.put(k, Math.toIntExact(table.getLong(k))));

    return spillings;
  }
}
