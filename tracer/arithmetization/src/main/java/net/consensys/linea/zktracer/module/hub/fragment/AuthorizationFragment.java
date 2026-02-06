/*
 * Copyright Consensys Software Inc.
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
package net.consensys.linea.zktracer.module.hub.fragment;

import static graphql.com.google.common.base.Preconditions.checkState;
import static net.consensys.linea.zktracer.Trace.*;

import lombok.Setter;
import lombok.experimental.Accessors;
import net.consensys.linea.zktracer.Trace;
import net.consensys.linea.zktracer.types.Bytecode;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.CodeDelegation;
import org.hyperledger.besu.datatypes.Hash;

@Accessors(fluent = true)
@Setter
public class AuthorizationFragment implements TraceFragment {

  final CodeDelegation delegation;
  final int tupleIndex;

  // fields below require successful authority recovery
  boolean isValidSenderIsAuthorityTuple;
  int validSenderIsAuthorityAccumulator;
  boolean authorityHasEmptyCodeOrIsDelegated;
  long authorityNonce;

  public AuthorizationFragment(
      CodeDelegation delegation,
      int tupleIndex,
      int validSenderIsAuthorityAccumulator,
      boolean isValidSenderIsAuthorityTuple,
      boolean authorityHasEmptyCodeOrIsDelegated,
      long authorityNonce) {
    this.delegation = delegation;
    this.tupleIndex = tupleIndex;
    this.isValidSenderIsAuthorityTuple = isValidSenderIsAuthorityTuple;
    this.validSenderIsAuthorityAccumulator = validSenderIsAuthorityAccumulator;
    this.authorityHasEmptyCodeOrIsDelegated = authorityHasEmptyCodeOrIsDelegated;
    this.authorityNonce = authorityNonce;
  }

  public Bytecode getBytecode() {
    if (delegation.address().equals(Address.ZERO)) {
      return Bytecode.EMPTY;
    }

    String bytecodeHexString =
        Bytes.ofUnsignedLong(EIP_7702_DELEGATION_INDICATOR).toHexString().substring(2)
            + delegation.address().toHexString().substring(2);

    checkState(
        bytecodeHexString.length() == 2 * EOA_DELEGATED_CODE_LENGTH,
        "Invalid delegated bytecode length");

    return new Bytecode(Bytes.fromHexString(bytecodeHexString));
  }

  public Hash getCodeHash() {
    return getBytecode().getCodeHash();
  }

  @Override
  public Trace.Hub trace(Trace.Hub trace) {

    final boolean ecrecoverSuccess = delegation.authorizer().isPresent();
    final Address authorityAddress = delegation.authorizer().orElse(Address.ZERO);

    return trace
        // .txAuth(true) // should be taken care of by the HUB phase setting
        .pAuthTupleIndex(tupleIndex)
        .pAuthDelegationAddressHi(delegation.address().slice(0, 4).toLong())
        .pAuthDelegationAddressLo(delegation.address().slice(4, LLARGE))
        .pAuthDelegationAddressIsZero(delegation.address().equals(Address.ZERO))
        .pAuthAuthorityNonce(Bytes.ofUnsignedLong(authorityNonce))
        .pAuthAuthorityEcrecoverSuccess(ecrecoverSuccess)
        .pAuthAuthorityAddressHi(authorityAddress.slice(0, 4).toLong())
        .pAuthAuthorityAddressLo(authorityAddress.slice(4, LLARGE))
        .pAuthAuthorityHasEmptyCodeOrIsDelegated(authorityHasEmptyCodeOrIsDelegated)
        .pAuthSenderIsAuthority(isValidSenderIsAuthorityTuple)
        .pAuthSenderIsAuthorityAcc(validSenderIsAuthorityAccumulator);
  }
}
