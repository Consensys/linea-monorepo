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

import java.io.IOException;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.Optional;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.continoustracing.exception.InvalidBlockTraceException;
import net.consensys.linea.continoustracing.exception.TraceVerificationException;
import net.consensys.linea.corset.CorsetValidator;
import net.consensys.linea.zktracer.ZkTracer;
import org.apache.commons.io.FileUtils;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.plugin.data.BlockTraceResult;
import org.hyperledger.besu.plugin.data.TransactionTraceResult;
import org.hyperledger.besu.plugin.services.TraceService;

@Slf4j
public class ContinuousTracer {
  private static final Optional<Path> TRACES_PATH =
      Optional.ofNullable(System.getenv("TRACES_DIR")).map(Paths::get);
  private final TraceService traceService;
  private final CorsetValidator corsetValidator;

  public ContinuousTracer(final TraceService traceService, final CorsetValidator corsetValidator) {
    this.traceService = traceService;
    this.corsetValidator = corsetValidator;
  }

  public CorsetValidator.Result verifyTraceOfBlock(
      final Hash blockHash, final String zkEvmBin, final ZkTracer zkTracer)
      throws TraceVerificationException, InvalidBlockTraceException {
    zkTracer.traceStartConflation(1);

    final BlockTraceResult blockTraceResult;
    try {
      blockTraceResult = traceService.traceBlock(blockHash, zkTracer);
    } catch (final Exception e) {
      throw new TraceVerificationException(blockHash, e.getMessage());
    } finally {
      zkTracer.traceEndConflation();
    }

    for (final TransactionTraceResult transactionTraceResult :
        blockTraceResult.transactionTraceResults()) {
      if (transactionTraceResult.getStatus() != TransactionTraceResult.Status.SUCCESS) {
        throw new InvalidBlockTraceException(
            transactionTraceResult.getTxHash(),
            transactionTraceResult.errorMessage().orElse("Unknown error"));
      }
    }

    final CorsetValidator.Result result;
    try {
      result =
          corsetValidator.validate(
              TRACES_PATH.map(zkTracer::writeToTmpFile).orElseGet(zkTracer::writeToTmpFile),
              zkEvmBin);
      if (!result.isValid()) {
        log.error("Trace of block {} is not valid", blockHash.toHexString());
        return result;
      }
    } catch (RuntimeException e) {
      log.error(
          "Error while validating trace of block {}: {}", blockHash.toHexString(), e.getMessage());
      throw new TraceVerificationException(blockHash, e.getMessage());
    }

    try {
      FileUtils.delete(result.traceFile());
    } catch (IOException e) {
      log.warn("Error while deleting trace file {}: {}", result.traceFile(), e.getMessage());
    }

    log.info("Trace of block {} is valid", blockHash.toHexString());
    return result;
  }
}
