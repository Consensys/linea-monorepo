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

package net.consensys.linea.blockcapture.snapshots;

import static org.hyperledger.besu.ethereum.core.Transaction.REPLAY_PROTECTED_V_BASE;
import static org.hyperledger.besu.ethereum.core.Transaction.REPLAY_PROTECTED_V_MIN;
import static org.hyperledger.besu.ethereum.core.Transaction.REPLAY_UNPROTECTED_V_BASE;

import java.math.BigInteger;
import java.util.List;
import java.util.Optional;

import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.DelegatingBytes;
import org.hyperledger.besu.crypto.SignatureAlgorithmFactory;
import org.hyperledger.besu.datatypes.AccessListEntry;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Quantity;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.evm.internal.Words;

public record TransactionSnapshot(
    String r,
    String s,
    String v,
    TransactionType type,
    String sender,
    Optional<String> to,
    long nonce,
    String value,
    String payload,
    Optional<String> gasPrice,
    Optional<String> maxPriorityFeePerGas,
    Optional<String> maxFeePerGas,
    Optional<String> maxFeePerBlobGas,
    long gasLimit,
    BigInteger chainId,
    Optional<List<AccessListEntrySnapshot>> accessList) {
  public static final BigInteger CHAIN_ID = BigInteger.valueOf(1337);

  public static TransactionSnapshot of(Transaction tx) {
    return new TransactionSnapshot(
        tx.getS().toString(16),
        tx.getR().toString(16),
        tx.getType() == TransactionType.FRONTIER
            ? tx.getV().toString(16)
            : tx.getYParity().toString(16),
        tx.getType(),
        tx.getSender().toHexString(),
        tx.getTo().map(DelegatingBytes::toHexString),
        tx.getNonce(),
        tx.getValue().toHexString(),
        tx.getPayload().toHexString(),
        tx.getGasPrice().map(Quantity::toHexString),
        tx.getMaxPriorityFeePerGas().map(Quantity::toHexString),
        tx.getMaxFeePerGas().map(Quantity::toHexString),
        tx.getMaxFeePerBlobGas().map(Quantity::toHexString),
        tx.getGasLimit(),
        tx.getChainId().orElse(CHAIN_ID),
        tx.getAccessList().map(l -> l.stream().map(AccessListEntrySnapshot::from).toList()));
  }

  public Transaction toTransaction() {
    BigInteger r = Bytes.fromHexStringLenient(this.r).toUnsignedBigInteger();
    BigInteger s = Bytes.fromHexStringLenient(this.s).toUnsignedBigInteger();
    BigInteger v = Bytes.fromHexStringLenient(this.v).toUnsignedBigInteger();

    if (this.type == TransactionType.FRONTIER) {
      if (v.compareTo(REPLAY_PROTECTED_V_MIN) > 0) {
        v = v.subtract(REPLAY_PROTECTED_V_BASE).subtract(chainId.multiply(BigInteger.TWO));
      } else {
        v = v.subtract(REPLAY_UNPROTECTED_V_BASE);
      }
    }

    final var tx =
        Transaction.builder()
            .type(this.type)
            .sender(Address.fromHexString(this.sender))
            .nonce(this.nonce)
            .value(Wei.fromHexString(this.value))
            .payload(Bytes.fromHexString(this.payload))
            .chainId(this.chainId)
            .gasLimit(this.gasLimit)
            .signature(
                SignatureAlgorithmFactory.getInstance().createSignature(r, s, v.byteValueExact()));

    this.to.ifPresent(to -> tx.to(Words.toAddress(Bytes.fromHexString(to))));
    this.gasPrice.ifPresent(gasPrice -> tx.gasPrice(Wei.fromHexString(gasPrice)));
    this.maxPriorityFeePerGas.ifPresent(
        maxPriorityFeePerGas -> tx.maxPriorityFeePerGas(Wei.fromHexString(maxPriorityFeePerGas)));
    this.maxFeePerGas.ifPresent(maxFeePerGas -> tx.maxFeePerGas(Wei.fromHexString(maxFeePerGas)));
    this.maxFeePerBlobGas.ifPresent(
        maxFeePerBlobGas -> tx.maxFeePerBlobGas(Wei.fromHexString(maxFeePerBlobGas)));
    this.accessList.ifPresent(
        l ->
            tx.accessList(
                l.stream()
                    .map(
                        e ->
                            AccessListEntry.createAccessListEntry(
                                Address.fromHexString(e.address()), e.storageKeys()))
                    .toList()));

    return tx.build();
  }
}
