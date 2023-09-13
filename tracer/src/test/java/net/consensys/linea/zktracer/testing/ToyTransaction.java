/*
 * Copyright ConsenSys AG.
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

package net.consensys.linea.zktracer.testing;

import java.math.BigInteger;
import java.util.Optional;

import lombok.Builder;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.crypto.KeyPair;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.Transaction;

@Builder
public class ToyTransaction {
  private static final ToyAccount DEFAULT_SENDER =
      ToyAccount.builder()
          .nonce(1L)
          .address(Address.fromHexString("0xe8f1b89"))
          .balance(Wei.ONE)
          .build();

  private static final Wei DEFAULT_VALUE = Wei.ZERO;
  private static final Bytes DEFAULT_INPUT_DATA = Bytes.EMPTY;
  private static final long DEFAULT_GAS_LIMIT = 21_000L;
  private static final Wei DEFAULT_GAS_PRICE = Wei.of(10_000_000_000L);
  private static final BigInteger DEFAULT_CHAIN_ID = BigInteger.valueOf(23);
  private static final TransactionType DEFAULT_TX_TYPE = TransactionType.FRONTIER;

  private final ToyAccount to;
  private final ToyAccount sender;
  private final Wei gasPrice;
  private final Long gasLimit;
  private final Wei value;
  private final TransactionType transactionType;
  private final Bytes payload;
  private final BigInteger chainId;
  private final KeyPair keyPair;

  /** Customizations applied to the Lombok generated builder. */
  public static class ToyTransactionBuilder {

    /**
     * Builder method returning an instance of {@link Transaction}.
     *
     * @return an instance of {@link Transaction}
     */
    public Transaction build() {

      return Transaction.builder()
          .to(to.getAddress())
          //        .sender(sender != null ? sender.getAddress() : DEFAULT_SENDER.getAddress())
          //        .nonce(sender != null ? sender.getNonce() : DEFAULT_SENDER.getNonce())
          .nonce(sender.getNonce())
          .type(Optional.ofNullable(transactionType).orElse(DEFAULT_TX_TYPE))
          .gasPrice(Optional.ofNullable(gasPrice).orElse(DEFAULT_GAS_PRICE))
          .gasLimit(Optional.ofNullable(gasLimit).orElse(DEFAULT_GAS_LIMIT))
          .value(Optional.ofNullable(value).orElse(DEFAULT_VALUE))
          .payload(Optional.ofNullable(payload).orElse(DEFAULT_INPUT_DATA))
          .chainId(Optional.ofNullable(chainId).orElse(DEFAULT_CHAIN_ID))
          .signAndBuild(keyPair);
    }
  }
}
