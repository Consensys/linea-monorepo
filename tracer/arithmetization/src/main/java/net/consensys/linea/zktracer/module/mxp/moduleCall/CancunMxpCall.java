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

package net.consensys.linea.zktracer.module.mxp.moduleCall;

import static net.consensys.linea.zktracer.module.mxp.MxpUtils.memoryCost;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.imc.MxpCall;
import net.consensys.linea.zktracer.module.mxp.MxpExoCall;
import net.consensys.linea.zktracer.opcode.gas.BillingRate;
import org.apache.tuweni.bytes.Bytes;

/**
 *
 *
 * <h3>The parent class of this MXP Call is located in the Hub.</h3>
 *
 * The CancunMxpCall can follow 5 scenarii depending on the opcode, sizes and mxpx exception (see
 * diagram below taken from the specification). To implement this decision tree, each scenario (and
 * 3 intermediate states) is a class that executes computations and inherits from the previous
 * scenario as computations are cumulative. Intermediate states are parent to several scenarii to
 * factorize computations.
 *
 * <p>MSize scenario - no computation
 *
 * <p>Trivial scenario - computes size1IsZero and size2IsZero
 *
 * <p>Not MSize not Trivial (intermediate state) - computes mxpxExpression in addition to the above
 *
 * <p>Mxpx scenario - no additional computation
 *
 * <p>State update (intermediate state) - computes state update (wordsNew,cMemNew) in addition to
 * all the above
 *
 * <p>State update with word pricing scenario - computes extraGasCost for word pricing opcodes in
 * addition to State update computations
 *
 * <p>State update with byte pricing scenario - computes extraGasCost for byte pricing opcodes in
 * addition to State update computations
 *
 * <p><img src="./scenariiDiagram.png" />
 */
public class CancunMxpCall extends MxpCall {

  public final long words;
  public final long cMem;
  public final Bytes gWord;
  public final Bytes gByte;

  public CancunMxpCall(Hub hub) {
    super(hub);
    this.words = this.memorySizeInWords;
    this.cMem = memoryCost(this.memorySizeInWords);
    this.gWord = getCostBy(BillingRate.BY_WORD);
    this.gByte = getCostBy(BillingRate.BY_BYTE);
    // Initialization of the computed values of MxpCall
    this.gasMxp = 0L;
    setMxpxFromMxpxExpression();
  }

  /** Store all wcp and euc computations with params and results */
  public final MxpExoCall[] exoCalls = new MxpExoCall[ctMax()];

  /** Computed by CancunTrivialMxpCall */
  public boolean size1IsZero = false;

  public boolean size2IsZero = false;

  /** Computed by CancunNotMSizeNorTrivialMxpCall for CancunMxpxMxpCall */
  public int mxpxExpression = 0;

  /**
   * Computed in CancunStateUpdateMxpCall for State update scenarii CancunStateUpdtWPricingMxpCall
   * and CancunStateUpdtBPricingMxpCall
   */
  public boolean isStateUpdate = false;

  public long wordsNew = 0L;
  public long cMemNew = 0L;
  public long extraGasCost = 0L;

  public int ctMax() {
    return 0;
  }
  ;

  public boolean isMSizeScenario() {
    return false;
  }

  public boolean isTrivialScenario() {
    return false;
  }

  public boolean isMxpxScenario() {
    return false;
  }

  public boolean isStateUpdateWordPricingScenario() {
    return false;
  }

  public boolean isStateUpdateBytePricingScenario() {
    return false;
  }

  public void setMxpxFromMxpxExpression() {
    this.mxpx = this.mxpxExpression != 0;
  }

  public void setGasMpxFromExtraGasCost() {
    this.gasMxp = this.cMemNew - this.cMem + this.extraGasCost;
  }
}
