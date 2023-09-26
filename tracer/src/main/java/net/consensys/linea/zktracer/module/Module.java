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

package net.consensys.linea.zktracer.module;

import java.util.List;

import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;

public interface Module {
  String jsonKey();

  default void traceStartConflation(final long blockCount) {}

  default void traceEndConflation() {}

  default void traceStartBlock(final BlockHeader blockHeader, final BlockBody blockBody) {}

  default void traceEndBlock(final BlockHeader blockHeader, final BlockBody blockBody) {}

  default void traceStartTx(WorldView worldView, Transaction tx) {}

  default void traceEndTx(
      WorldView worldView,
      Transaction tx,
      boolean status,
      Bytes output,
      List<Log> logs,
      long gasUsed) {}

  default void traceContextEnter(MessageFrame frame) {}

  default void traceContextExit(MessageFrame frame) {}

  default void trace(MessageFrame frame) {}

  Object commit();
}
