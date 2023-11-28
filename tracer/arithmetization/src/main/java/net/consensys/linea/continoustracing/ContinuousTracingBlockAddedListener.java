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

import java.util.concurrent.ArrayBlockingQueue;
import java.util.concurrent.ThreadPoolExecutor;
import java.util.concurrent.TimeUnit;

import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.continoustracing.exception.InvalidBlockTraceException;
import net.consensys.linea.continoustracing.exception.InvalidTraceHandlerException;
import net.consensys.linea.continoustracing.exception.TraceVerificationException;
import net.consensys.linea.corset.CorsetValidator;
import net.consensys.linea.zktracer.ZkTracer;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.plugin.data.AddedBlockContext;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.services.BesuEvents;

@Slf4j
public class ContinuousTracingBlockAddedListener implements BesuEvents.BlockAddedListener {
  final ContinuousTracer continuousTracer;
  final TraceFailureHandler traceFailureHandler;
  final String zkEvmBin;

  static final int BLOCK_PARALLELISM = 10; // Higher and IOs bite the dust
  final ThreadPoolExecutor pool =
      new ThreadPoolExecutor(
          BLOCK_PARALLELISM,
          BLOCK_PARALLELISM,
          0L,
          TimeUnit.SECONDS,
          new ArrayBlockingQueue<>(BLOCK_PARALLELISM),
          new ThreadPoolExecutor.CallerRunsPolicy());
  ;

  public ContinuousTracingBlockAddedListener(
      final ContinuousTracer continuousTracer,
      final TraceFailureHandler traceFailureHandler,
      final String zkEvmBin) {
    this.continuousTracer = continuousTracer;
    this.traceFailureHandler = traceFailureHandler;
    this.zkEvmBin = zkEvmBin;
  }

  @Override
  public void onBlockAdded(final AddedBlockContext addedBlockContext) {
    pool.submit(
        () -> {
          final BlockHeader blockHeader = addedBlockContext.getBlockHeader();
          final Hash blockHash = blockHeader.getBlockHash();
          log.info("Tracing block {} ({})", blockHeader.getNumber(), blockHash.toHexString());

          try {
            final CorsetValidator.Result traceResult =
                continuousTracer.verifyTraceOfBlock(blockHash, zkEvmBin, new ZkTracer());

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
          } finally {
            log.info("End of tracing block {}", blockHeader.getNumber());
          }
        });
  }
}
