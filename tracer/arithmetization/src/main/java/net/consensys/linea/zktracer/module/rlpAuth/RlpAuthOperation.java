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
import java.util.List;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.ModuleOperation;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;

@Accessors(fluent = true)
@RequiredArgsConstructor
public class RlpAuthOperation extends ModuleOperation {
  final List<CodeDelegationTuple> codeDelegationTuple;
  final long blkNumber;
  final long userTxnNumber;
  final Bytes txnFromAddress;
  final Bytes networkChainId;
  final long hubStamp;

  Bytes authorityNonce;
  boolean authorityHasEmptyCodeOrIsDelegated;

  protected void trace(Trace.Rlpauth trace) {
    int tupleIndex = 0;
    for (CodeDelegationTuple codeDelegationTuple : codeDelegationTuple) {
      final Address authorityAddress = codeDelegationTuple.authorizer().orElseThrow();
      trace
          .chainId(bigIntegerToBytes(codeDelegationTuple.chainId()))
          .nonce(bigIntegerToBytes(BigInteger.valueOf(codeDelegationTuple.nonce())))
          .delegationAddress(codeDelegationTuple.address())
          .yParity(codeDelegationTuple.yParity())
          .r(bigIntegerToBytes(codeDelegationTuple.r()))
          .s(bigIntegerToBytes(codeDelegationTuple.s()))
          .msg(codeDelegationTuple.msg()) // predicted output from keccak256
          .authorityAddress(authorityAddress) // predicted output from ecRecover
          // .macro(false) // TODO: probably useless
          .blkNumber(blkNumber)
          .userTxnNumber(userTxnNumber)
          .txnFromAddress(txnFromAddress)
          // .authorityIsSenderTot(false)
          // .xtern(false) // TODO: probably useless
          .networkChainId(networkChainId)
          .authorityEcrecoverSuccess(codeDelegationTuple.authorityEcRecoverSuccess())
          // .authorityNonce(authorityNonce)
          // .authorityHasEmptyCodeOrIsDelegated(authorityHasEmptyCodeOrIsDelegated)
          .tupleIndex(tupleIndex++)
          .hubStamp(hubStamp)
          .validateRow();
    }
  }

  @Override
  protected int computeLineCount() {
    return 0;
  }
}
