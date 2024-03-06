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

package net.consensys.linea.sequencer.txvalidation.validators;

import static org.assertj.core.api.Assertions.assertThat;

import java.math.BigInteger;
import java.nio.file.Path;
import java.util.Optional;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.config.LineaProfitabilityCliOptions;
import org.apache.tuweni.bytes.Bytes;
import org.bouncycastle.asn1.sec.SECNamedCurves;
import org.bouncycastle.asn1.x9.X9ECParameters;
import org.bouncycastle.crypto.params.ECDomainParameters;
import org.hyperledger.besu.crypto.SECPSignature;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.plugin.data.BlockContext;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.services.BesuConfiguration;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.hyperledger.besu.plugin.services.storage.DataStorageFormat;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

@Slf4j
@RequiredArgsConstructor
public class ProfitabilityValidatorTest {
  public static final Address SENDER =
      Address.fromHexString("0x0000000000000000000000000000000000001000");
  public static final Address RECIPIENT =
      Address.fromHexString("0x0000000000000000000000000000000000001001");
  private static Wei PROFITABLE_GAS_PRICE = Wei.of(11000000);
  private static Wei UNPROFITABLE_GAS_PRICE = Wei.of(1000000);
  private static final SECPSignature FAKE_SIGNATURE;

  static {
    final X9ECParameters params = SECNamedCurves.getByName("secp256k1");
    final ECDomainParameters curve =
        new ECDomainParameters(params.getCurve(), params.getG(), params.getN(), params.getH());
    FAKE_SIGNATURE =
        SECPSignature.create(
            new BigInteger(
                "66397251408932042429874251838229702988618145381408295790259650671563847073199"),
            new BigInteger(
                "24729624138373455972486746091821238755870276413282629437244319694880507882088"),
            (byte) 0,
            curve.getN());
  }

  public static final double TX_POOL_MIN_MARGIN = 0.5;
  private ProfitabilityValidator profitabilityValidatorAlways;
  private ProfitabilityValidator profitabilityValidatorOnlyApi;
  private ProfitabilityValidator profitabilityValidatorOnlyP2p;
  private ProfitabilityValidator profitabilityValidatorNever;

  @BeforeEach
  public void initialize() {
    final var profitabilityConfBuilder =
        LineaProfitabilityCliOptions.create().toDomainObject().toBuilder()
            .txPoolMinMargin(TX_POOL_MIN_MARGIN);

    profitabilityValidatorAlways =
        new ProfitabilityValidator(
            new TestBesuConfiguration(),
            new TestBlockchainService(),
            profitabilityConfBuilder
                .txPoolCheckP2pEnabled(true)
                .txPoolCheckApiEnabled(true)
                .build());

    profitabilityValidatorNever =
        new ProfitabilityValidator(
            new TestBesuConfiguration(),
            new TestBlockchainService(),
            profitabilityConfBuilder
                .txPoolCheckP2pEnabled(false)
                .txPoolCheckApiEnabled(false)
                .build());

    profitabilityValidatorOnlyApi =
        new ProfitabilityValidator(
            new TestBesuConfiguration(),
            new TestBlockchainService(),
            profitabilityConfBuilder
                .txPoolCheckP2pEnabled(false)
                .txPoolCheckApiEnabled(true)
                .build());

    profitabilityValidatorOnlyP2p =
        new ProfitabilityValidator(
            new TestBesuConfiguration(),
            new TestBlockchainService(),
            profitabilityConfBuilder
                .txPoolCheckP2pEnabled(true)
                .txPoolCheckApiEnabled(false)
                .build());
  }

  @Test
  public void acceptPriorityRemoteWhenBelowMinProfitability() {
    final org.hyperledger.besu.ethereum.core.Transaction transaction =
        org.hyperledger.besu.ethereum.core.Transaction.builder()
            .sender(SENDER)
            .to(RECIPIENT)
            .gasLimit(21000)
            .gasPrice(PROFITABLE_GAS_PRICE)
            .payload(Bytes.EMPTY)
            .value(Wei.ONE)
            .signature(FAKE_SIGNATURE)
            .build();
    assertThat(profitabilityValidatorAlways.validateTransaction(transaction, false, true))
        .isEmpty();
  }

  @Test
  public void rejectRemoteWhenBelowMinProfitability() {
    final org.hyperledger.besu.ethereum.core.Transaction transaction =
        org.hyperledger.besu.ethereum.core.Transaction.builder()
            .sender(SENDER)
            .to(RECIPIENT)
            .gasLimit(21000)
            .gasPrice(UNPROFITABLE_GAS_PRICE)
            .payload(Bytes.EMPTY)
            .value(Wei.ONE)
            .signature(FAKE_SIGNATURE)
            .build();
    assertThat(profitabilityValidatorAlways.validateTransaction(transaction, false, false))
        .isPresent()
        .contains("Gas price too low");
  }

  @Test
  public void acceptRemoteWhenBelowMinProfitabilityIfCheckNeverEnabled() {
    final org.hyperledger.besu.ethereum.core.Transaction transaction =
        org.hyperledger.besu.ethereum.core.Transaction.builder()
            .sender(SENDER)
            .to(RECIPIENT)
            .gasLimit(21000)
            .gasPrice(UNPROFITABLE_GAS_PRICE)
            .payload(Bytes.EMPTY)
            .value(Wei.ONE)
            .signature(FAKE_SIGNATURE)
            .build();
    assertThat(profitabilityValidatorNever.validateTransaction(transaction, false, false))
        .isEmpty();
  }

  @Test
  public void acceptLocalWhenBelowMinProfitabilityIfCheckNeverEnabled() {
    final org.hyperledger.besu.ethereum.core.Transaction transaction =
        org.hyperledger.besu.ethereum.core.Transaction.builder()
            .sender(SENDER)
            .to(RECIPIENT)
            .gasLimit(21000)
            .gasPrice(UNPROFITABLE_GAS_PRICE)
            .payload(Bytes.EMPTY)
            .value(Wei.ONE)
            .signature(FAKE_SIGNATURE)
            .build();
    assertThat(profitabilityValidatorNever.validateTransaction(transaction, true, false)).isEmpty();
  }

  @Test
  public void acceptRemoteWhenBelowMinProfitabilityIfCheckDisabledForP2p() {
    final org.hyperledger.besu.ethereum.core.Transaction transaction =
        org.hyperledger.besu.ethereum.core.Transaction.builder()
            .sender(SENDER)
            .to(RECIPIENT)
            .gasLimit(21000)
            .gasPrice(UNPROFITABLE_GAS_PRICE)
            .payload(Bytes.EMPTY)
            .value(Wei.ONE)
            .signature(FAKE_SIGNATURE)
            .build();
    assertThat(profitabilityValidatorOnlyApi.validateTransaction(transaction, false, false))
        .isEmpty();
  }

  @Test
  public void rejectRemoteWhenBelowMinProfitabilityIfCheckEnableForP2p() {
    final org.hyperledger.besu.ethereum.core.Transaction transaction =
        org.hyperledger.besu.ethereum.core.Transaction.builder()
            .sender(SENDER)
            .to(RECIPIENT)
            .gasLimit(21000)
            .gasPrice(UNPROFITABLE_GAS_PRICE)
            .payload(Bytes.EMPTY)
            .value(Wei.ONE)
            .signature(FAKE_SIGNATURE)
            .build();
    assertThat(profitabilityValidatorOnlyP2p.validateTransaction(transaction, false, false))
        .isPresent()
        .contains("Gas price too low");
  }

  @Test
  public void acceptLocalWhenBelowMinProfitabilityIfCheckDisabledForApi() {
    final org.hyperledger.besu.ethereum.core.Transaction transaction =
        org.hyperledger.besu.ethereum.core.Transaction.builder()
            .sender(SENDER)
            .to(RECIPIENT)
            .gasLimit(21000)
            .gasPrice(UNPROFITABLE_GAS_PRICE)
            .payload(Bytes.EMPTY)
            .value(Wei.ONE)
            .signature(FAKE_SIGNATURE)
            .build();
    assertThat(profitabilityValidatorOnlyP2p.validateTransaction(transaction, true, false))
        .isEmpty();
  }

  @Test
  public void rejectLocalWhenBelowMinProfitabilityIfCheckEnableForApi() {
    final org.hyperledger.besu.ethereum.core.Transaction transaction =
        org.hyperledger.besu.ethereum.core.Transaction.builder()
            .sender(SENDER)
            .to(RECIPIENT)
            .gasLimit(21000)
            .gasPrice(UNPROFITABLE_GAS_PRICE)
            .payload(Bytes.EMPTY)
            .value(Wei.ONE)
            .signature(FAKE_SIGNATURE)
            .build();
    assertThat(profitabilityValidatorOnlyApi.validateTransaction(transaction, true, false))
        .isPresent()
        .contains("Gas price too low");
  }

  private static class TestBesuConfiguration implements BesuConfiguration {
    @Override
    public Path getStoragePath() {
      throw new UnsupportedOperationException();
    }

    @Override
    public Path getDataPath() {
      throw new UnsupportedOperationException();
    }

    @Override
    public DataStorageFormat getDatabaseFormat() {
      throw new UnsupportedOperationException();
    }

    @Override
    public Wei getMinGasPrice() {
      return Wei.of(100_000_000);
    }
  }

  private static class TestBlockchainService implements BlockchainService {

    @Override
    public Optional<BlockContext> getBlockByNumber(final long l) {
      throw new UnsupportedOperationException();
    }

    @Override
    public Hash getChainHeadHash() {
      throw new UnsupportedOperationException();
    }

    @Override
    public BlockHeader getChainHeadHeader() {
      throw new UnsupportedOperationException();
    }

    @Override
    public Optional<Wei> getNextBlockBaseFee() {
      return Optional.of(Wei.of(7));
    }
  }
}
