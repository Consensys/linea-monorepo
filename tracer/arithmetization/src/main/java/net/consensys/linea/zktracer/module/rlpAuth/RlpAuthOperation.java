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
import static org.hyperledger.besu.crypto.Hash.keccak256;
import static org.hyperledger.besu.ethereum.core.CodeDelegation.MAGIC;

import java.math.BigInteger;
import lombok.RequiredArgsConstructor;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.container.ModuleOperation;
import net.consensys.linea.zktracer.module.hub.fragment.AuthorizationFragment;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.CodeDelegation;
import org.hyperledger.besu.ethereum.rlp.BytesValueRLPOutput;

@Accessors(fluent = true)
@RequiredArgsConstructor
public class RlpAuthOperation extends ModuleOperation {
  final AuthorizationFragment authorizationFragment;
  final CodeDelegation delegation;
  final TransactionProcessingMetadata txMetadata;

  protected void trace(Trace.Rlpauth trace) {
    trace
        .chainId(bigIntegerToBytes(delegation.chainId()))
        .nonce(bigIntegerToBytes(BigInteger.valueOf(delegation.nonce())))
        .delegationAddress(delegation.address())
        .yParity(delegation.v() - 27)
        .r(bigIntegerToBytes(delegation.r()))
        .s(bigIntegerToBytes(delegation.s()))
        .msg(
            getMsg(
                delegation.chainId(),
                delegation.address(),
                delegation.nonce())) // predicted output from keccak256
        .authorityAddress(
            delegation.authorizer().orElse(Address.ZERO)) // predicted output from ecRecover
        .blkNumber(txMetadata.getRelativeBlockNumber())
        .userTxnNumber(txMetadata.getUserTransactionNumber())
        .txnFromAddress(txMetadata.getSender())
        .networkChainId(bigIntegerToBytes(authorizationFragment.networkChainId()))
        .authorityEcrecoverSuccess(delegation.authorizer().isPresent())
        .authorityNonce(
            bigIntegerToBytes(BigInteger.valueOf(authorizationFragment.authorityNonce())))
        .authorityHasEmptyCodeOrIsDelegated(
            authorizationFragment.authorityHasEmptyCodeOrIsDelegated())
        .tupleIndex(authorizationFragment.tupleIndex())
        .hubStamp(authorizationFragment.hubStamp())
        .validateRow();
  }

  @Override
  protected int computeLineCount() {
    return 0;
  } // TODO

  Bytes getMsg(BigInteger chainId, Address address, long nonce) {
    final BytesValueRLPOutput listRlp = new BytesValueRLPOutput();
    listRlp.startList();
    listRlp.writeBigIntegerScalar(chainId);
    listRlp.writeBytes(address);
    listRlp.writeLongScalar(nonce);
    listRlp.endList();
    final Bytes rlpOfListWithChainIdAddressNonce = listRlp.encoded();
    final Bytes keccakInput = Bytes.concatenate(MAGIC, rlpOfListWithChainIdAddressNonce);
    return keccak256(keccakInput);
    // msg = keccak(MAGIC || rlp([chain_id, address, nonce]))
  }

  /* Alternative method
  Bytes getMsg() {
    BytesValueRLPOutput rlpOutput = new BytesValueRLPOutput();
    CodeDelegationTransactionEncoder.encodeSingleCodeDelegationWithoutSignature(codeDelegation, rlpOutput);
    return Hash.hash(Bytes.concatenate(new Bytes[] {MAGIC, rlpOutput.encoded()}));
    msg = keccak(MAGIC || rlp([chain_id, address, nonce]))
  }
   */
}
