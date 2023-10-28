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

package net.consensys.linea.zktracer;

import java.util.List;

import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.services.tracer.BlockAwareOperationTracer;

/**
 * An extended operation tracer that can trace the start and end of a number of blocks in
 * conflation.
 */
public interface ZkBlockAwareOperationTracer extends BlockAwareOperationTracer {

  /**
   * Trace the start of conflation for a number of blocks.
   *
   * @param numBlocksInConflation blocks in conflation
   */
  void traceStartConflation(final long numBlocksInConflation);

  /** Trace the end of conflation for a number of blocks. */
  void traceEndConflation();

  /**
   * Get a JSON serialized version of the trace.
   *
   * @return a JSON string of the trace
   */
  String getJsonTrace();

  void traceStartTransaction(WorldView worldView, Transaction transaction);

  void traceEndTransaction(
      WorldView worldView,
      Transaction tx,
      boolean status,
      Bytes output,
      List<Log> logs,
      long gasUsed,
      long timeNs);
}
