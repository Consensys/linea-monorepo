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

import static net.consensys.linea.zktracer.TraceCancun.Mxp.CT_MAX_UPDT_W;

import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.mxp.MxpExoCall;
import org.apache.tuweni.bytes.Bytes;

public class CancunStateUpdateWordPricingMxpCall extends CancunStateUpdateMxpCall {

  public CancunStateUpdateWordPricingMxpCall(Hub hub) {
    super(hub);
    if (this.isStateUpdate) {
      // if state has changed, an extra gas cost is incurred
      computeExtraGasCost(hub.euc());
      setGasMpxFromExtraGasCost();
    }
  }

  @Override
  public boolean isStateUpdateWordPricingScenario() {
    return true;
  }

  private void computeExtraGasCost(Euc euc) {
    // Row i + 11
    exoCalls[10] = MxpExoCall.callToEUC(euc, this.size1.lo(), Bytes.of(32));
    Bytes numberOfWords = exoCalls[10].resultB(); // result of row i + 11
    this.extraGasCost =
        numberOfWords
            .toUnsignedBigInteger()
            .multiply(this.gWord.toUnsignedBigInteger())
            .longValue();
  }

  @Override
  public int ctMax() {
    return CT_MAX_UPDT_W;
  }
}
