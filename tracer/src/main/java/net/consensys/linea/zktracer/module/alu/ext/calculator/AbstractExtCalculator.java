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

import net.consensys.linea.zktracer.OpCode;
import net.consensys.linea.zktracer.bytestheta.BaseTheta;
import net.consensys.linea.zktracer.bytestheta.BytesArray;
import net.consensys.linea.zktracer.module.alu.ext.calculator.addmod.AddModCalculator;
import net.consensys.linea.zktracer.module.alu.ext.calculator.mulmod.MulModCalculator;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;

/**
 * An abstract class representing a calculator for performing extended modular arithmetic
 * operations. It provides methods for computing the result of an extended modular arithmetic
 * operation, computing various intermediate variables such as comparison flags, deltas, h-values,
 * i-values, and j-values, as well as creating an instance of a calculator based on a given OpCode.
 */
public abstract class AbstractExtCalculator {

  /**
   * Computes the result of an extended modular arithmetic operation for the given arguments.
   *
   * @param arg1 the first argument.
   * @param arg2 the second argument.
   * @param arg3 the third argument.
   * @return the result of the extended modular arithmetic operation.
   */
  public abstract UInt256 computeResult(Bytes32 arg1, Bytes32 arg2, Bytes32 arg3);
  /**
   * Computes the Js array for the given arguments.
   *
   * @param arg1 the first argument.
   * @param arg2 the second argument.
   * @return the Js array.
   */
  public abstract BytesArray computeJs(Bytes32 arg1, Bytes32 arg2);
  /**
   * Computes the Qs array for the given arguments.
   *
   * @param arg1 the first argument.
   * @param arg2 the second argument.
   * @param arg3 the third argument.
   * @return the Qs array.
   */
  public abstract BytesArray computeQs(Bytes32 arg1, Bytes32 arg2, Bytes32 arg3);
  /**
   * Computes the overflow result for the given arguments.
   *
   * @param aBytes the aBytes value.
   * @param bBytes the bBytes value.
   * @param hBytes the hBytes value.
   * @param alpha the alpha value.
   * @param beta the beta value.
   * @return the overflow result.
   */
  public abstract boolean[] computeOverflowRes(
      BaseTheta aBytes, BaseTheta bBytes, BytesArray hBytes, UInt256 alpha, UInt256 beta);
  /**
   * Computes the comparison flags for the given arguments.
   *
   * @param cBytes the cBytes value.
   * @param rBytes the rBytes value.
   * @return the comparison flags.
   */
  public boolean[] computeComparisonFlags(BaseTheta cBytes, BaseTheta rBytes) {
    return CmpFlagsAndDeltaCalculator.computeComparisonFlags(cBytes, rBytes);
  }
  /**
   * Computes the delta values for the given arguments.
   *
   * @param cBytes the cBytes value.
   * @param rBytes the rBytes value.
   * @return the delta values.
   */
  public BaseTheta computeDeltas(BaseTheta cBytes, BaseTheta rBytes) {
    return CmpFlagsAndDeltaCalculator.computeDeltas(cBytes, rBytes);
  }
  /**
   * Computes the Hs array for the given arguments.
   *
   * @param aBytes the aBytes value.
   * @param bBytes the bBytes value.
   * @param hBytes the hBytes value.
   * @return the Hs array.
   */
  public boolean[] computeHs(BaseTheta aBytes, BaseTheta bBytes, BytesArray hBytes) {
    return HByteCalculator.computeHsAndOverflowH(aBytes, bBytes, hBytes);
  }
  /**
   * Computes the Is array for the given arguments.
   *
   * @param qBytes the qBytes value.
   * @param cBytes the cBytes value.
   * @param iBytes the iBytes value.
   * @return the Is array.
   */
  public boolean[] computeIs(BytesArray qBytes, BaseTheta cBytes, BytesArray iBytes) {
    return IByteCalculator.computeIsAndOverflowI(qBytes, cBytes, iBytes);
  }
  /**
   * Computes the overflow result for the given arguments.
   *
   * @param qBytes the qBytes value.
   * @param cBytes the cBytes value.
   * @param rBytes the rBytes value.
   * @param iBytes the iBytes value.
   * @param sigma the sigma value.
   * @param tau the tau value.
   * @return the overflow result.
   */
  public boolean[] computeOverflowJ(
      BytesArray qBytes,
      BaseTheta cBytes,
      BaseTheta rBytes,
      BytesArray iBytes,
      UInt256 sigma,
      UInt256 tau) {
    return JByteCalculator.computeJsAndOverflowJ(qBytes, cBytes, rBytes, iBytes, sigma, tau);
  }
  /**
   * Creates a new instance of a calculator based on the given OpCode.
   *
   * @param opCode the OpCode for which to create a calculator instance.
   * @return a new instance of a calculator.
   * @throws RuntimeException if the OpCode is not compatible with this calculator.
   */
  public static AbstractExtCalculator create(OpCode opCode) {
    return switch (opCode) {
      case MULMOD -> new MulModCalculator();
      case ADDMOD -> new AddModCalculator();
      default -> throw new RuntimeException(
          "Incompatible instruction for extended modular arithmetic module");
    };
  }
}
