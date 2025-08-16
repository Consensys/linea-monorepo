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
package net.consensys.linea.sequencer.txpoolvalidation.validators;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.when;

import java.math.BigInteger;
import java.util.Optional;
import net.consensys.linea.config.LineaRlnValidatorConfiguration;
import net.consensys.linea.config.LineaSharedGaslessConfiguration;
import net.consensys.linea.sequencer.txpoolvalidation.shared.DenyListManager;
import net.consensys.linea.sequencer.txpoolvalidation.shared.KarmaServiceClient;
import net.consensys.linea.sequencer.txpoolvalidation.shared.NullifierTracker;
import org.apache.tuweni.bytes.Bytes;
import org.bouncycastle.asn1.sec.SECNamedCurves;
import org.bouncycastle.asn1.x9.X9ECParameters;
import org.bouncycastle.crypto.params.ECDomainParameters;
import org.hyperledger.besu.crypto.SECPSignature;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.services.BlockchainService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.mockito.junit.jupiter.MockitoSettings;
import org.mockito.quality.Strictness;

@ExtendWith(MockitoExtension.class)
@MockitoSettings(strictness = Strictness.LENIENT)
class RlnValidatorBasicTest {

  private static final Address TEST_SENDER = Address.fromHexString("0x1234567890123456789012345678901234567890");
  private static final String TEST_NULLIFIER = "0xa1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456";
  private static final String TEST_EPOCH = "0x1c61ef0b2ebc0235d85fe8537b4455549356e3895005ba7a03fbd4efc9ba3692";
  
  private static final SECPSignature FAKE_SIGNATURE;
  
  static {
    final X9ECParameters params = SECNamedCurves.getByName("secp256k1");
    final ECDomainParameters curve =
        new ECDomainParameters(params.getCurve(), params.getG(), params.getN(), params.getH());
    FAKE_SIGNATURE =
        SECPSignature.create(
            new BigInteger("66397251408932042429874251838229702988618145381408295790259650671563847073199"),
            new BigInteger("24729624138373455972486746091821238755870276413282629437244319694880507882088"),
            (byte) 0,
            curve.getN());
  }

  @Mock private BlockchainService blockchainService;
  @Mock private BlockHeader blockHeader;
  @Mock private DenyListManager denyListManager;
  @Mock private KarmaServiceClient karmaServiceClient;
  @Mock private NullifierTracker nullifierTracker;

  private LineaRlnValidatorConfiguration rlnConfig;
  private org.hyperledger.besu.ethereum.core.Transaction testTransaction;

  @BeforeEach
  void setUp() {
    // Setup mock blockchain service
    when(blockchainService.getChainHeadHeader()).thenReturn(blockHeader);
    when(blockHeader.getNumber()).thenReturn(12345L);

    // Create test configuration using constructor
    LineaSharedGaslessConfiguration sharedConfig = new LineaSharedGaslessConfiguration(
        "/tmp/test_deny_list.txt",
        300L, // denyListRefreshSeconds
        1L, // premiumGasPriceThresholdGWei  
        10L // denyListEntryMaxAgeMinutes
    );
    
    rlnConfig = new LineaRlnValidatorConfiguration(
        true, // rlnValidationEnabled
        "/tmp/test_vk.json", // verifyingKeyPath
        "localhost", // rlnProofServiceHost
        8545, // rlnProofServicePort
        false, // rlnProofServiceUseTls
        1000L, // rlnProofCacheMaxSize
        300L, // rlnProofCacheExpirySeconds
        3, // rlnProofStreamRetries
        1000L, // rlnProofStreamRetryIntervalMs
        1000L, // rlnProofLocalWaitTimeoutMs
        sharedConfig, // sharedGaslessConfig
        "localhost", // karmaServiceHost
        8546, // karmaServicePort
        false, // karmaServiceUseTls
        5000L, // karmaServiceTimeoutMs
        true, // exponentialBackoffEnabled
        30000L, // maxBackoffDelayMs
        "TEST", // defaultEpochForQuota
        Optional.empty() // rlnJniLibPath
    );

    // Create test transaction
    testTransaction = org.hyperledger.besu.ethereum.core.Transaction.builder()
        .sender(TEST_SENDER)
        .to(Address.fromHexString("0x9876543210987654321098765432109876543210"))
        .gasLimit(21000)
        .gasPrice(Wei.of(20_000_000_000L))
        .payload(Bytes.EMPTY)
        .value(Wei.ONE)
        .signature(FAKE_SIGNATURE)
        .build();
  }

  @Test
  void testConfigurationCreation() {
    assertThat(rlnConfig).isNotNull();
    assertThat(rlnConfig.rlnValidationEnabled()).isTrue();
    assertThat(rlnConfig.rlnProofServiceHost()).isEqualTo("localhost");
    assertThat(rlnConfig.rlnProofServicePort()).isEqualTo(8545);
    assertThat(rlnConfig.premiumGasPriceThresholdWei()).isEqualTo(1000000000L);
  }

  @Test
  void testValidatorCreationWithDisabledConfig() {
    LineaSharedGaslessConfiguration disabledSharedConfig = new LineaSharedGaslessConfiguration(
        "/tmp/test_deny_list.txt",
        300L, 1L, 10L
    );
    
    LineaRlnValidatorConfiguration disabledConfig = new LineaRlnValidatorConfiguration(
        false, // disabled
        "/tmp/test_vk.json",
        "localhost", 8545, false, 1000L, 300L, 3, 1000L, 1000L,
        disabledSharedConfig,
        "localhost", 8546, false, 5000L, true, 30000L, "TEST", Optional.empty()
    );

    RlnVerifierValidator validator = new RlnVerifierValidator(
        disabledConfig,
        blockchainService,
        denyListManager,
        karmaServiceClient,
        nullifierTracker,
        null,
        null);

    Optional<String> result = validator.validateTransaction(testTransaction, true, false);
    assertThat(result).isEmpty();
    
    try {
      validator.close();
    } catch (Exception e) {
      // Expected if resources already closed
    }
  }

  @Test
  void testForwarderValidatorCreation() {
    RlnProverForwarderValidator forwarder = new RlnProverForwarderValidator(
        rlnConfig, 
        false, // disabled in sequencer mode
        karmaServiceClient);

    assertThat(forwarder.isEnabled()).isFalse();
    assertThat(forwarder.getValidationCallCount()).isEqualTo(0);

    // Test validation with disabled validator
    Optional<String> result = forwarder.validateTransaction(testTransaction, true, false);
    assertThat(result).isEmpty();
    assertThat(forwarder.getValidationCallCount()).isEqualTo(1);

    try {
      forwarder.close();
    } catch (Exception e) {
      // Expected if resources already closed
    }
  }

  @Test
  void testSharedServicesConfiguration() {
    // Test that shared services are properly configured
    assertThat(rlnConfig.denyListPath()).contains("deny_list.txt");
    assertThat(rlnConfig.denyListRefreshSeconds()).isEqualTo(300L);
    assertThat(rlnConfig.denyListEntryMaxAgeMinutes()).isEqualTo(10L);
    assertThat(rlnConfig.premiumGasPriceThresholdWei()).isEqualTo(1_000_000_000L); // 1 GWei in Wei
    
    // Test karma service configuration
    assertThat(rlnConfig.karmaServiceHost()).isEqualTo("localhost");
    assertThat(rlnConfig.karmaServicePort()).isEqualTo(8546);
    assertThat(rlnConfig.karmaServiceTimeoutMs()).isEqualTo(5000L);
  }
}