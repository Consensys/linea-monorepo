/*
 * Copyright Consensys Software Inc.
 *
 * This file is dual-licensed under either the MIT license or Apache License 2.0.
 * See the LICENSE-MIT and LICENSE-APACHE files in the repository root for details.
 *
 * SPDX-License-Identifier: MIT OR Apache-2.0
 */
package net.consensys.linea.sequencer.txpoolvalidation.validators;

import static org.assertj.core.api.Assertions.assertThat;

import org.apache.tuweni.bytes.Bytes;
import io.grpc.ManagedChannel;
import io.grpc.inprocess.InProcessChannelBuilder;
import io.grpc.inprocess.InProcessServerBuilder;
import io.grpc.stub.StreamObserver;
import java.util.Optional;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import net.vac.prover.RlnProverGrpc;
import net.vac.prover.SendTransactionReply;
import net.vac.prover.SendTransactionRequest;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

class RlnProverForwarderValidatorEstimatedGasTest {

  private io.grpc.Server server;
  private ManagedChannel channel;

  private volatile SendTransactionRequest capturedRequest;

  @BeforeEach
  void setUp() throws Exception {
    final String serverName = InProcessServerBuilder.generateName();

    server =
        InProcessServerBuilder.forName(serverName)
            .directExecutor()
            .addService(
                new RlnProverGrpc.RlnProverImplBase() {
                  @Override
                  public void sendTransaction(
                      SendTransactionRequest request, StreamObserver<SendTransactionReply> resp) {
                    capturedRequest = request;
                    resp.onNext(SendTransactionReply.newBuilder().setResult(true).build());
                    resp.onCompleted();
                  }
                })
            .build()
            .start();

    channel = InProcessChannelBuilder.forName(serverName).directExecutor().build();
  }

  @AfterEach
  void tearDown() {
    if (channel != null) {
      channel.shutdownNow();
    }
    if (server != null) {
      server.shutdownNow();
    }
  }

  @Test
  void forwardsEstimatedGasUsed_21000_forSimpleEthTransfer() throws Exception {
    final var validator =
        new RlnProverForwarderValidator(
            /* rlnConfig */ null,
            /* enabled */ true,
            /* karmaServiceClient */ null,
            /* txSim */ null,
            /* blockchain */ null,
            /* tracerConfig */ null,
            /* l1l2 */ null,
            /* providedChannel */ channel);

    // Create a simple ETH transfer: to set, empty data, value > 0
    final org.hyperledger.besu.crypto.SECPSignature fakeSig =
        org.hyperledger.besu.crypto.SECPSignature.create(
            new java.math.BigInteger("1"),
            new java.math.BigInteger("2"),
            (byte) 0,
            new java.math.BigInteger("3"));

    final org.hyperledger.besu.ethereum.core.Transaction tx =
        org.hyperledger.besu.ethereum.core.Transaction.builder()
            .sender(Address.fromHexString("0x2222222222222222222222222222222222222222"))
            .to(Address.fromHexString("0x1111111111111111111111111111111111111111"))
            .gasLimit(21_000)
            .gasPrice(Wei.of(1))
            .payload(Bytes.EMPTY)
            .value(Wei.of(1))
            .signature(fakeSig)
            .build();

    final CountDownLatch latch = new CountDownLatch(1);
    // validateTransaction performs a blocking gRPC call; just invoke and then assert capture
    final var maybeError =
        validator.validateTransaction(
            (org.hyperledger.besu.datatypes.Transaction) tx, /* isLocal */ true, /* hasPriority */ false);
    assertThat(maybeError).isEmpty();
    latch.countDown();
    latch.await(100, TimeUnit.MILLISECONDS);

    assertThat(capturedRequest).isNotNull();
    assertThat(capturedRequest.getEstimatedGasUsed()).isEqualTo(21_000L);
  }
}


