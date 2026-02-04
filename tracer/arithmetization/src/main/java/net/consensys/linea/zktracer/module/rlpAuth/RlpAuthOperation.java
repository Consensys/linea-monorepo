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

import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import java.math.BigInteger;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.ModuleOperation;
import org.apache.tuweni.bytes.Bytes;

public class RlpAuthOperation extends ModuleOperation {
  // TODO: create constructor
  CodeDelegationTuple codeDelegationTuple;
  long blkNumber;
  long userTxnNumber;
  Bytes txnFromAddress;
  Bytes networkChainId;
  boolean authorityEcrecoverSuccess;
  Bytes authorityNonce;
  boolean authorityHasEmptyCodeOrIsDelegated;
  long tupleIndex;
  long hubStamp;

  protected void trace(Trace.Rlpauth trace) {
    trace
        .chainId(bigIntegerToBytes(codeDelegationTuple.chainId()))
        .nonce(bigIntegerToBytes(BigInteger.valueOf(codeDelegationTuple.nonce())))
        .delegationAddress(bigIntegerToBytes(codeDelegationTuple.address().toUnsignedBigInteger()))
        .yParity(codeDelegationTuple.yParity())
        .r(bigIntegerToBytes(codeDelegationTuple.r()))
        .s(bigIntegerToBytes(codeDelegationTuple.s()))
        .authorityAddress(
            bigIntegerToBytes(codeDelegationTuple.authorizer().toUnsignedBigInteger()))
        .macro(false) // TODO: probably useless
        .blkNumber(blkNumber)
        .userTxnNumber(userTxnNumber)
        .txnFromAddress(txnFromAddress)
        .authorityIsSenderTot(false)
        .xtern(false) // TODO: probably useless
        .networkChainId(networkChainId)
        .authorityEcrecoverSuccess(authorityEcrecoverSuccess)
        .authorityNonce(authorityNonce)
        .authorityHasEmptyCodeOrIsDelegated(authorityHasEmptyCodeOrIsDelegated)
        .tupleIndex(tupleIndex)
        .hubStamp(hubStamp);
  }

  @Override
  protected int computeLineCount() {
    return 0;
  }
}
