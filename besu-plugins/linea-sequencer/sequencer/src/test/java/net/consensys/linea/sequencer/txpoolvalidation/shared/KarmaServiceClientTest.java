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
package net.consensys.linea.sequencer.txpoolvalidation.shared;

import static org.assertj.core.api.Assertions.assertThat;

import com.google.protobuf.ByteString;
import java.io.IOException;
import io.grpc.ManagedChannel;
import io.grpc.Server;
import io.grpc.Status;
import io.grpc.StatusRuntimeException;
import io.grpc.inprocess.InProcessChannelBuilder;
import io.grpc.inprocess.InProcessServerBuilder;
import io.grpc.stub.StreamObserver;
import io.grpc.testing.GrpcCleanupRule;
import java.util.Optional;
import net.consensys.linea.sequencer.txpoolvalidation.shared.KarmaServiceClient.KarmaInfo;
import net.vac.prover.GetUserTierInfoReply;
import net.vac.prover.GetUserTierInfoRequest;
import net.vac.prover.RlnProverGrpc;
import net.vac.prover.Tier;
import net.vac.prover.UserTierInfoError;
import net.vac.prover.UserTierInfoResult;
import org.hyperledger.besu.datatypes.Address;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

/**
 * Comprehensive tests for KarmaServiceClient functionality.
 * 
 * Tests gRPC communication, error handling, timeouts, and karma info parsing.
 */
class KarmaServiceClientTest {

  @org.junit.Rule
  public final GrpcCleanupRule grpcCleanup = new GrpcCleanupRule();

  private static final Address TEST_USER = Address.fromHexString("0x1234567890123456789012345678901234567890");

  private KarmaServiceClient client;
  private Server mockServer;
  private ManagedChannel inProcessChannel;

  @BeforeEach
  void setUp() {
    // gRPC cleanup will handle channel cleanup
  }

  @AfterEach
  void tearDown() throws Exception {
    if (client != null) {
      client.close();
    }
    if (mockServer != null) {
      mockServer.shutdownNow();
    }
  }

  @Test
  void testSuccessfulKarmaInfoRetrieval() throws Exception {
    // Create mock server that returns valid karma info
    String serverName = InProcessServerBuilder.generateName();
    mockServer = InProcessServerBuilder
        .forName(serverName)
        .directExecutor()
        .addService(new MockKarmaService(MockResponseType.SUCCESS))
        .build()
        .start();

    inProcessChannel = InProcessChannelBuilder.forName(serverName).directExecutor().build();
    client = new KarmaServiceClient("Test", "localhost", 8545, false, 5000, inProcessChannel);

    Optional<KarmaInfo> result = client.fetchKarmaInfo(TEST_USER);

    assertThat(result).isPresent();
    KarmaInfo info = result.get();
    assertThat(info.tier()).isEqualTo("Regular");
    assertThat(info.epochTxCount()).isEqualTo(10);
    assertThat(info.dailyQuota()).isEqualTo(720);
    assertThat(info.epochId()).isEqualTo("12345");
    assertThat(info.karmaBalance()).isEqualTo(0L); // Always 0 in new schema
  }

  @Test
  void testUserNotFound() throws Exception {
    // Create mock server that returns NOT_FOUND error
    String serverName = InProcessServerBuilder.generateName();
    mockServer = InProcessServerBuilder
        .forName(serverName)
        .directExecutor()
        .addService(new MockKarmaService(MockResponseType.NOT_FOUND))
        .build()
        .start();

    inProcessChannel = InProcessChannelBuilder.forName(serverName).directExecutor().build();
    client = new KarmaServiceClient("Test", "localhost", 8545, false, 5000, inProcessChannel);

    Optional<KarmaInfo> result = client.fetchKarmaInfo(TEST_USER);

    assertThat(result).isEmpty();
  }

  @Test
  void testServiceError() throws Exception {
    // Create mock server that returns service error
    String serverName = InProcessServerBuilder.generateName();
    mockServer = InProcessServerBuilder
        .forName(serverName)
        .directExecutor()
        .addService(new MockKarmaService(MockResponseType.SERVICE_ERROR))
        .build()
        .start();

    inProcessChannel = InProcessChannelBuilder.forName(serverName).directExecutor().build();
    client = new KarmaServiceClient("Test", "localhost", 8545, false, 5000, inProcessChannel);

    Optional<KarmaInfo> result = client.fetchKarmaInfo(TEST_USER);

    assertThat(result).isEmpty();
  }

  @Test
  void testTimeout() throws Exception {
    // Create mock server that delays response
    String serverName = InProcessServerBuilder.generateName();
    mockServer = InProcessServerBuilder
        .forName(serverName)
        .directExecutor()
        .addService(new MockKarmaService(MockResponseType.TIMEOUT))
        .build()
        .start();

    inProcessChannel = InProcessChannelBuilder.forName(serverName).directExecutor().build();
    client = new KarmaServiceClient("Test", "localhost", 8545, false, 100, inProcessChannel); // 100ms timeout

    Optional<KarmaInfo> result = client.fetchKarmaInfo(TEST_USER);

    assertThat(result).isEmpty();
  }

  @Test
  void testServiceUnavailable() throws Exception {
    // Create mock server that throws UNAVAILABLE
    String serverName = InProcessServerBuilder.generateName();
    mockServer = InProcessServerBuilder
        .forName(serverName)
        .directExecutor()
        .addService(new MockKarmaService(MockResponseType.UNAVAILABLE))
        .build()
        .start();

    inProcessChannel = InProcessChannelBuilder.forName(serverName).directExecutor().build();
    client = new KarmaServiceClient("Test", "localhost", 8545, false, 5000, inProcessChannel);

    Optional<KarmaInfo> result = client.fetchKarmaInfo(TEST_USER);

    assertThat(result).isEmpty();
  }

  @Test
  void testEmptyResponse() throws Exception {
    // Create mock server that returns empty response
    String serverName = InProcessServerBuilder.generateName();
    mockServer = InProcessServerBuilder
        .forName(serverName)
        .directExecutor()
        .addService(new MockKarmaService(MockResponseType.EMPTY))
        .build()
        .start();

    inProcessChannel = InProcessChannelBuilder.forName(serverName).directExecutor().build();
    client = new KarmaServiceClient("Test", "localhost", 8545, false, 5000, inProcessChannel);

    Optional<KarmaInfo> result = client.fetchKarmaInfo(TEST_USER);

    assertThat(result).isEmpty();
  }

  @Test
  void testClientAvailability() throws IOException {
    // Test without channel
    client = new KarmaServiceClient("Test", "localhost", 8545, false, 5000);
    assertThat(client.isAvailable()).isTrue();

    // Close and test
    client.close();
    assertThat(client.isAvailable()).isFalse();
  }

  @Test
  void testNoKarmaTierInfo() throws Exception {
    // Create mock server that returns response without tier info
    String serverName = InProcessServerBuilder.generateName();
    mockServer = InProcessServerBuilder
        .forName(serverName)
        .directExecutor()
        .addService(new MockKarmaService(MockResponseType.NO_TIER))
        .build()
        .start();

    inProcessChannel = InProcessChannelBuilder.forName(serverName).directExecutor().build();
    client = new KarmaServiceClient("Test", "localhost", 8545, false, 5000, inProcessChannel);

    Optional<KarmaInfo> result = client.fetchKarmaInfo(TEST_USER);

    assertThat(result).isPresent();
    KarmaInfo info = result.get();
    assertThat(info.tier()).isEqualTo("Unknown");
    assertThat(info.dailyQuota()).isEqualTo(0);
  }

  private enum MockResponseType {
    SUCCESS,
    NOT_FOUND,
    SERVICE_ERROR,
    TIMEOUT,
    UNAVAILABLE,
    EMPTY,
    NO_TIER
  }

  /**
   * Mock gRPC service for testing different response scenarios
   */
  private static class MockKarmaService extends RlnProverGrpc.RlnProverImplBase {
    private final MockResponseType responseType;

    MockKarmaService(MockResponseType responseType) {
      this.responseType = responseType;
    }

    @Override
    public void getUserTierInfo(GetUserTierInfoRequest request, StreamObserver<GetUserTierInfoReply> responseObserver) {
      switch (responseType) {
        case SUCCESS:
          UserTierInfoResult result = UserTierInfoResult.newBuilder()
              .setTier(Tier.newBuilder().setName("Regular").setQuota(720).build())
              .setTxCount(10)
              .setCurrentEpoch(12345)
              .setCurrentEpochSlice(1)
              .build();
          
          GetUserTierInfoReply reply = GetUserTierInfoReply.newBuilder()
              .setRes(result)
              .build();
          
          responseObserver.onNext(reply);
          responseObserver.onCompleted();
          break;

        case NOT_FOUND:
          responseObserver.onError(new StatusRuntimeException(Status.NOT_FOUND));
          break;

        case SERVICE_ERROR:
          GetUserTierInfoReply errorReply = GetUserTierInfoReply.newBuilder()
              .setError(UserTierInfoError.newBuilder().setMessage("Service error").build())
              .build();
          
          responseObserver.onNext(errorReply);
          responseObserver.onCompleted();
          break;

        case TIMEOUT:
          // Don't respond - let it timeout
          break;

        case UNAVAILABLE:
          responseObserver.onError(new StatusRuntimeException(Status.UNAVAILABLE));
          break;

        case EMPTY:
          GetUserTierInfoReply emptyReply = GetUserTierInfoReply.newBuilder().build();
          responseObserver.onNext(emptyReply);
          responseObserver.onCompleted();
          break;

        case NO_TIER:
          UserTierInfoResult noTierResult = UserTierInfoResult.newBuilder()
              .setTxCount(5)
              .setCurrentEpoch(12345)
              .setCurrentEpochSlice(1)
              // No tier info
              .build();
          
          GetUserTierInfoReply noTierReply = GetUserTierInfoReply.newBuilder()
              .setRes(noTierResult)
              .build();
          
          responseObserver.onNext(noTierReply);
          responseObserver.onCompleted();
          break;
      }
    }
  }
}