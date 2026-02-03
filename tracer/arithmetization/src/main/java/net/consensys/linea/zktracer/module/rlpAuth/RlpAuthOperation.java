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

package net.consensys.linea.zktracer.module.rlpAuth;

import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.ModuleOperation;
import org.apache.tuweni.bytes.Bytes;

public class RlpAuthOperation extends ModuleOperation {

  protected void trace(Trace.Rlpauth trace) {
    trace
        .chainId(Bytes.EMPTY)
        .nonce(Bytes.EMPTY)
        .delegationAddress(Bytes.EMPTY)
        .yParity(0)
        .r(Bytes.EMPTY)
        .s(Bytes.EMPTY)
        .authorityAddress(Bytes.EMPTY)
        .macro(false) // TODO: probably useless
        .blkNumber(0)
        .userTxnNumber(0)
        .txnFromAddress(Bytes.EMPTY)
        .authorityIsSenderTot(false)
        .xtern(false) // TODO: probably useless
        .networkChainId(Bytes.EMPTY)
        .authorityEcrecoverSuccess(false)
        .authorityNonce(Bytes.EMPTY)
        .authorityHasEmptyCodeOrIsDelegated(false)
        .tupleIndex(0)
        .hubStamp(0);
  }

  @Override
  protected int computeLineCount() {
    return 0;
  }
}
