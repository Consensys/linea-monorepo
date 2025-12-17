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

package net.consensys.linea.zktracer.module.hub;

import net.consensys.linea.zktracer.ChainConfig;
import net.consensys.linea.zktracer.module.mxp.module.LondonMxp;
import net.consensys.linea.zktracer.module.mxp.module.Mxp;
import net.consensys.linea.zktracer.module.txndata.TxnData;
import net.consensys.linea.zktracer.module.txndata.london.LondonTxnData;
import net.consensys.linea.zktracer.types.PublicInputs;

public class LondonHub extends Hub {
  public LondonHub(ChainConfig chain, PublicInputs publicInputs) {
    super(chain, publicInputs);
  }

  @Override
  protected TxnData setTxnData() {
    return new LondonTxnData(this, wcp(), euc());
  }

  @Override
  protected Mxp setMxp() {
    return new LondonMxp();
  }
}
