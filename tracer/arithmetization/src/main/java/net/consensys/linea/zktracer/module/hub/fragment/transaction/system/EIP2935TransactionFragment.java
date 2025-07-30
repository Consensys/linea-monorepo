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

package net.consensys.linea.zktracer.module.hub.fragment.transaction.system;

import static net.consensys.linea.zktracer.Trace.LLARGE;
import static net.consensys.linea.zktracer.module.hub.fragment.transaction.system.SystemTransactionFragmentType.EIP2935_HISTORICAL_HASH;

import net.consensys.linea.zktracer.Trace;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;

public class EIP2935TransactionFragment extends SystemTransactionFragment {
  final short previousBlockNumberModulo;
  final Bytes32 blockhash;

  public EIP2935TransactionFragment(short previousBlockNumberModulo, Bytes32 blockhash) {
    super(EIP2935_HISTORICAL_HASH);
    this.previousBlockNumberModulo = previousBlockNumberModulo;
    this.blockhash = blockhash;
  }

  @Override
  public Trace.Hub trace(Trace.Hub trace) {
    super.trace(trace);
    return trace
        .pTransactionEip2935(true)
        .pTransactionSystTxnData1(Bytes.ofUnsignedInt(previousBlockNumberModulo))
        .pTransactionSystTxnData2(blockhash.slice(0, LLARGE))
        .pTransactionSystTxnData3(blockhash.slice(LLARGE, LLARGE));
  }
}
