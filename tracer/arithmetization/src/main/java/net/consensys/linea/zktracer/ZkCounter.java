/*
 * Copyright ConsenSys Inc.
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

import static net.consensys.linea.zktracer.types.AddressUtils.isBlsPrecompile;
import static org.hyperledger.besu.datatypes.Address.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;

import net.consensys.linea.plugins.config.LineaL1L2BridgeSharedConfiguration;
import net.consensys.linea.zktracer.container.module.EventDetectorModule;
import net.consensys.linea.zktracer.container.module.Module;
import net.consensys.linea.zktracer.module.hub.precompiles.ModexpMetadata;
import net.consensys.linea.zktracer.module.limits.L1BlockSize;
import net.consensys.linea.zktracer.module.limits.L2L1Logs;
import net.consensys.linea.zktracer.types.MemoryRange;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;

public class ZkCounter implements LineCountingTracer {
  public static final String MODEXP = "MODEXP";
  public static final String RIP = "RIP";
  public static final String BLAKE = "BLAKE";

  final EventDetectorModule modexp = new EventDetectorModule(MODEXP) {};
  final EventDetectorModule rip = new EventDetectorModule(RIP) {};
  final EventDetectorModule blake = new EventDetectorModule(BLAKE) {};
  final EventDetectorModule pointEval = new EventDetectorModule("POINT_EVAL") {};
  final EventDetectorModule bls = new EventDetectorModule("BLS") {};
  final L1BlockSize l1BlockSize;
  final L2L1Logs l2l1Logs = new L2L1Logs();
  final List<Module> moduleToCount;

  public ZkCounter(LineaL1L2BridgeSharedConfiguration bridgeConfiguration) {
    l1BlockSize =
        new L1BlockSize(l2l1Logs, bridgeConfiguration.contract(), bridgeConfiguration.topic());
    moduleToCount = List.of(modexp, rip, blake, bls, pointEval, l1BlockSize, l2l1Logs);
  }

  @Override
  public void traceStartConflation(long numBlocksInConflation) {}

  @Override
  public void traceEndConflation(WorldView state) {}

  @Override
  public void traceStartBlock(
      final WorldView world,
      final BlockHeader blockHeader,
      final BlockBody blockBody,
      final Address miningBeneficiary) {
    l1BlockSize.traceStartBlock(world, blockHeader, miningBeneficiary);
  }

  @Override
  public void traceEndTransaction(
      WorldView worldView,
      Transaction tx,
      boolean status,
      Bytes output,
      List<Log> logs,
      long gasUsed,
      Set<Address> selfDestructs,
      long timeNs) {
    switch (tx.getType()) {
      case FRONTIER, ACCESS_LIST, EIP1559 -> l1BlockSize.traceEndTx(tx, logs);
      case BLOB, DELEGATE_CODE -> throw new IllegalStateException(
          "Unsupported tx type: " + tx.getType());
    }
  }

  @Override
  public void tracePrecompileCall(MessageFrame frame, long gasRequirement, Bytes output) {
    final Address precompileAddress = frame.getContractAddress();

    if (precompileAddress.equals(Address.MODEXP)) {
      final Bytes callData = frame.getInputData();
      final MemoryRange memoryRange = new MemoryRange(0, 0, callData.size(), callData);
      final ModexpMetadata modexpMetadata = new ModexpMetadata(memoryRange);
      if (modexpMetadata.unprovableModexp()) {
        modexp.detectEvent();
      }
      return;
    }

    if (precompileAddress.equals(KZG_POINT_EVAL)) {
      pointEval.detectEvent();
      return;
    }

    if (isBlsPrecompile(precompileAddress)) {
      bls.detectEvent();
      return;
    }

    if (precompileAddress.equals(RIPEMD160)) {
      // We COULD accept empty input data, as it implies no gnark circuit, so nothing to detect. We
      // don't do it for simplicity.
      // if (frame.getInputData().isEmpty()) {
      //   return;
      // }
      rip.detectEvent();
      return;
    }

    if (precompileAddress.equals(BLAKE2B_F_COMPRESSION)) {
      blake.detectEvent();
      return;
    }
    // No other precompiles are tracked
  }

  /** When called, erase all tracing related to the bundle of all transactions since the last. */
  @Override
  public void popTransactionBundle() {
    for (Module m : moduleToCount) {
      m.popTransactionBundle();
    }
  }

  @Override
  public void commitTransactionBundle() {
    for (Module m : moduleToCount) {
      m.commitTransactionBundle();
    }
  }

  @Override
  public Map<String, Integer> getModulesLineCount() {
    final HashMap<String, Integer> modulesLineCount = HashMap.newHashMap(moduleToCount.size());

    for (Module m : moduleToCount) {
      modulesLineCount.put(m.moduleKey(), m.lineCount());
    }
    return modulesLineCount;
  }

  @Override
  public List<Module> getModulesToCount() {
    return moduleToCount;
  }
}
