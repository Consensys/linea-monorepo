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
import net.consensys.linea.zktracer.module.blockdata.module.Blockdata;
import net.consensys.linea.zktracer.module.blockdata.module.ParisBlockData;
import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import org.hyperledger.besu.plugin.services.BlockchainService;

public class ParisHub extends LondonHub {
  public ParisHub(ChainConfig chain, BlockchainService blockchainService) {
    super(chain, blockchainService);
  }

  @Override
  protected Blockdata setBlockData(
      Hub hub, Wcp wcp, Euc euc, ChainConfig chain, BlockchainService blockchainService) {
    return new ParisBlockData(hub, wcp, euc, chain);
  }
}
