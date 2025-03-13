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
package net.consensys.linea.plugins.continoustracing;

import java.io.IOException;
import java.nio.file.Files;
import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.ThreadPoolExecutor;
import java.util.concurrent.TimeUnit;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.corset.CorsetValidator;
import net.consensys.linea.plugins.config.LineaL1L2BridgeSharedConfiguration;
import net.consensys.linea.plugins.exception.InvalidBlockTraceException;
import net.consensys.linea.plugins.exception.InvalidTraceHandlerException;
import net.consensys.linea.plugins.exception.TraceVerificationException;
import net.consensys.linea.zktracer.ZkTracer;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.plugin.data.AddedBlockContext;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.services.BesuEvents;

@Slf4j
@RequiredArgsConstructor
public class ContinuousTracingBlockAddedListener implements BesuEvents.BlockAddedListener {
  private final ContinuousTracer continuousTracer;
  private final TraceFailureHandler traceFailureHandler;

  static final int BLOCK_PARALLELISM = 5;
  final ThreadPoolExecutor pool =
      new ThreadPoolExecutor(
          BLOCK_PARALLELISM,
          BLOCK_PARALLELISM,
          0L,
          TimeUnit.SECONDS,
          new ArrayBlockingQueue<>(BLOCK_PARALLELISM),
          new ThreadPoolExecutor.CallerRunsPolicy());

  @Override
  public void onBlockAdded(final AddedBlockContext addedBlockContext) {
    pool.submit(
        () -> {
          final BlockHeader blockHeader = addedBlockContext.getBlockHeader();
          final Hash blockHash = blockHeader.getBlockHash();
          log.info("Tracing block {} ({})", blockHeader.getNumber(), blockHash.toHexString());

          try {
            final CorsetValidator.Result traceResult =
                continuousTracer.verifyTraceOfBlock(
                    blockHeader.getNumber(),
                    blockHash,
                    new ZkTracer(
                        LineaL1L2BridgeSharedConfiguration.TEST_DEFAULT, // FIXME: appropriate here?
                        Bytes.fromHexString("c0ffee").toUnsignedBigInteger()));
            Files.delete(traceResult.traceFile().toPath());

            if (!traceResult.isValid()) {
              log.error("Corset returned and error for block {}", blockHeader.getNumber());
              traceFailureHandler.handleCorsetFailure(blockHeader, traceResult);
              return;
            }
            log.info("Trace for block {} verified successfully", blockHeader.getNumber());
          } catch (InvalidBlockTraceException e) {
            log.error("Error while tracing block {}: {}", blockHeader.getNumber(), e.getMessage());
            traceFailureHandler.handleBlockTraceFailure(blockHeader.getNumber(), e.txHash(), e);
          } catch (TraceVerificationException e) {
            log.error(e.getMessage());
          } catch (InvalidTraceHandlerException e) {
            log.error("Error while handling invalid trace: {}", e.getMessage());
          } catch (IOException e) {
            log.error("IO error: {}", e.getMessage());
          } finally {
            log.info("End of tracing block {}", blockHeader.getNumber());
          }
        });
  }
}
