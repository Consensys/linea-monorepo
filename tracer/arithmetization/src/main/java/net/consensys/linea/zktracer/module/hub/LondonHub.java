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
import net.consensys.linea.zktracer.module.hub.state.LondonTransactionStack;
import net.consensys.linea.zktracer.module.hub.state.TransactionStack;
import net.consensys.linea.zktracer.module.txndata.module.LondonTxnData;
import net.consensys.linea.zktracer.module.txndata.module.TxnData;
import org.hyperledger.besu.evm.gascalculator.GasCalculator;
import org.hyperledger.besu.evm.gascalculator.LondonGasCalculator;

public class LondonHub extends Hub {
  public LondonHub(ChainConfig chain) {
    super(chain);
  }

  @Override
  protected GasCalculator setGasCalculator() {
    return new LondonGasCalculator();
  }

  @Override
  protected TransactionStack setTransactionStack() {
    return new LondonTransactionStack();
  }

  @Override
  protected TxnData setTxnData() {
    return new LondonTxnData(this, wcp(), euc());
  }
}
