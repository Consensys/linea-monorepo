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

import static net.consensys.linea.zktracer.TraceCancun.Mxp.CT_MAX_TRIV;
import static net.consensys.linea.zktracer.types.Conversions.bytesToBoolean;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.mxp.MxpExoCall;
import net.consensys.linea.zktracer.module.wcp.Wcp;

public class CancunTrivialMxpCall extends CancunMSizeMxpCall {

  public CancunTrivialMxpCall(Hub hub) {
    super(hub);
    computeSize1Size2IsZero(hub.wcp());
  }

  @Override
  public boolean isTrivialScenario() {
    return true;
  }

  private void computeSize1Size2IsZero(Wcp wcp) {
    // We compute and assign the computation's result for each row

    // Row i + 1
    // Compute size1IsZero
    exoCalls[0] = MxpExoCall.callToIsZero(wcp, this.size1);
    this.size1IsZero = bytesToBoolean(exoCalls[0].resultA());

    // Row i + 2
    // Compute size2IsZero
    exoCalls[1] = MxpExoCall.callToIsZero(wcp, this.size2);
    this.size2IsZero = bytesToBoolean(exoCalls[1].resultA());
  }

  @Override
  public int ctMax() {
    return CT_MAX_TRIV;
  }
}
