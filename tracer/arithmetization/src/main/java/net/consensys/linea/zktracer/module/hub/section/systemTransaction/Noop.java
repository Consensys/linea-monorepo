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

package net.consensys.linea.zktracer.module.hub.section.systemTransaction;

import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.fragment.transaction.system.NoopSystemTransactionFragment;
import net.consensys.linea.zktracer.module.hub.fragment.transaction.system.SystemTransactionFragment;
import net.consensys.linea.zktracer.module.hub.section.TraceSection;

public class Noop extends TraceSection {
  public Noop(Hub hub) {
    super(hub, (short) 2);

    final SystemTransactionFragment txFragment = new NoopSystemTransactionFragment();
    final ContextFragment contextFragment = ContextFragment.readZeroContextData(hub);

    addFragments(txFragment, contextFragment);

    hub.txnData().callTxnDataForSystemTransaction(txFragment);
  }
}
