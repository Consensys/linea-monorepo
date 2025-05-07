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
import net.consensys.linea.zktracer.module.hub.section.create.ShanghaiCreateSection;
import net.consensys.linea.zktracer.module.hub.section.txInitializationSection.ShanghaiInitializationSection;
import net.consensys.linea.zktracer.module.hub.state.ShanghaiTransactionStack;
import net.consensys.linea.zktracer.module.hub.state.TransactionStack;
import net.consensys.linea.zktracer.module.txndata.module.ShanghaiTxnData;
import net.consensys.linea.zktracer.module.txndata.module.TxnData;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.gascalculator.GasCalculator;
import org.hyperledger.besu.evm.gascalculator.ShanghaiGasCalculator;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class ShanghaiHub extends LondonHub {
  public ShanghaiHub(ChainConfig chain) {
    super(chain);
  }

  @Override
  protected GasCalculator setGasCalculator() {
    return new ShanghaiGasCalculator();
  }

  @Override
  protected TransactionStack setTransactionStack() {
    return new ShanghaiTransactionStack();
  }

  @Override
  protected TxnData setTxnData() {
    return new ShanghaiTxnData(this, wcp(), euc());
  }

  @Override
  protected void setInitializationSection(WorldView world) {
    new ShanghaiInitializationSection(this, world);
  }

  @Override
  protected boolean coinbaseWarmthAtTxEnd() {
    // since EIP-3651 (Shanghai), the coinbase address is warm at the beginning of the transaction,
    // so obviously at the end.
    return true;
  }

  @Override
  protected void setCreateSection(final Hub hub, final MessageFrame frame) {
    new ShanghaiCreateSection(hub, frame);
  }
}
