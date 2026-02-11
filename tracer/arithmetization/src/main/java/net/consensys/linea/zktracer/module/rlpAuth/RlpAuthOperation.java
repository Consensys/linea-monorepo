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
import net.consensys.linea.zktracer.module.ecdata.EcData;
import net.consensys.linea.zktracer.module.hub.fragment.AuthorizationFragment;
import net.consensys.linea.zktracer.module.hub.fragment.scenario.PrecompileScenarioFragment;
import net.consensys.linea.zktracer.module.shakiradata.ShakiraData;
import net.consensys.linea.zktracer.module.shakiradata.ShakiraDataOperation;
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
  final EcData ecData;
  final ShakiraData shakiraData;

  protected void trace(Trace.Rlpauth trace) {
    final Bytes magicConcatToRlpOfChainIdAddressNonceList =
        getMagicConcatToRlpOfChainIdAddressNonceList(
            delegation.chainId(), delegation.address(), delegation.nonce());
    final Bytes msg = getMsg(magicConcatToRlpOfChainIdAddressNonceList);
    final byte v = delegation.v();
    final Bytes r = bigIntegerToBytes(delegation.r());
    final Bytes s = bigIntegerToBytes(delegation.s());

    // Note:
    // msg = keccak(MAGIC || rlp([chain_id, address, nonce]))
    // authority = ecrecover(msg, y_parity, r, s)

    shakiraData.call(
        new ShakiraDataOperation(
            authorizationFragment.hubStamp(), magicConcatToRlpOfChainIdAddressNonceList));
    ecData.callEcData(
        authorizationFragment.hubStamp() + 1,
        PrecompileScenarioFragment.PrecompileFlag.PRC_ECRECOVER,
        Bytes.concatenate(msg, Bytes.of(v), r, s),
        delegation.authorizer().orElse(Address.ZERO));

    trace
        .chainId(bigIntegerToBytes(delegation.chainId()))
        .nonce(bigIntegerToBytes(BigInteger.valueOf(delegation.nonce())))
        .delegationAddress(delegation.address())
        .yParity(v - 27)
        .r(r)
        .s(s)
        .msg(msg) // predicted output from keccak256
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

  Bytes getMagicConcatToRlpOfChainIdAddressNonceList(
      BigInteger chainId, Address address, long nonce) {
    final BytesValueRLPOutput listRlp = new BytesValueRLPOutput();
    listRlp.startList();
    listRlp.writeBigIntegerScalar(chainId);
    listRlp.writeBytes(address);
    listRlp.writeLongScalar(nonce);
    listRlp.endList();
    final Bytes rlpOfListWithChainIdAddressNonce = listRlp.encoded();
    return Bytes.concatenate(MAGIC, rlpOfListWithChainIdAddressNonce);
  }

  Bytes getMsg(BigInteger chainId, Address address, long nonce) {
    return keccak256(getMagicConcatToRlpOfChainIdAddressNonceList(chainId, address, nonce));
  }

  Bytes getMsg(Bytes magicConcatToRlpOfChainIdAddressNonceList) {
    return keccak256(magicConcatToRlpOfChainIdAddressNonceList);
  }

  /* Alternative method
  Bytes getEcRecoverInput() {
    BytesValueRLPOutput rlpOutput = new BytesValueRLPOutput();
    CodeDelegationTransactionEncoder.encodeSingleCodeDelegationWithoutSignature(codeDelegation, rlpOutput);
    return Hash.hash(Bytes.concatenate(new Bytes[] {MAGIC, rlpOutput.encoded()}));
    msg = keccak(MAGIC || rlp([chain_id, address, nonce]))
  }
   */
}
