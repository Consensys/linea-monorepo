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
package net.consensys.linea.continoustracing;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.matches;
import static org.mockito.Mockito.when;

import java.nio.file.Path;

import net.consensys.linea.continoustracing.exception.InvalidBlockTraceException;
import net.consensys.linea.continoustracing.exception.TraceVerificationException;
import net.consensys.linea.corset.CorsetValidator;
import net.consensys.linea.zktracer.ZkTracer;
import org.hyperledger.besu.plugin.services.TraceService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentMatchers;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

@ExtendWith(MockitoExtension.class)
public class ContinuousTracerTest {
  private ContinuousTracer continuousTracer;

  @Mock TraceService traceServiceMock;
  @Mock CorsetValidator corsetValidatorMock;
  @Mock ZkTracer zkTracerMock;

  @BeforeEach
  void setUp() {
    continuousTracer = new ContinuousTracer(traceServiceMock, corsetValidatorMock);
  }

  @Test
  void shouldReturnSuccessIfVerificationIsSuccessful()
      throws InvalidBlockTraceException, TraceVerificationException {
    when(corsetValidatorMock.validate(ArgumentMatchers.any(), matches("testZkEvmBin")))
        .thenReturn(
            new CorsetValidator.Result(
                true, Path.of("testTraceFile").toFile(), "testCorsetOutput"));

    when(zkTracerMock.writeToTmpFile()).thenReturn(Path.of(""));

    final CorsetValidator.Result validationResult =
        continuousTracer.verifyTraceOfBlock(1, "testZkEvmBin", zkTracerMock);
    assertThat(validationResult.isValid()).isTrue();
  }

  @Test
  void shouldReturnFailureIfVerificationIsNotSuccessful()
      throws InvalidBlockTraceException, TraceVerificationException {
    when(zkTracerMock.writeToTmpFile()).thenReturn(Path.of(""));

    when(corsetValidatorMock.validate(ArgumentMatchers.any(), matches("testZkEvmBin")))
        .thenReturn(
            new CorsetValidator.Result(
                false, Path.of("testTraceFile").toFile(), "testCorsetOutput"));

    final CorsetValidator.Result validationResult =
        continuousTracer.verifyTraceOfBlock(1, "testZkEvmBin", zkTracerMock);
    assertThat(validationResult.isValid()).isFalse();
  }
}
