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

package net.consensys.linea.zktracer.module.blockdata;

import static net.consensys.linea.zktracer.module.blockdata.Trace.GAS_LIMIT_MAXIMUM;
import static net.consensys.linea.zktracer.module.blockdata.Trace.nROWS_DEPTH;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LLARGE;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.*;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.module.Module;
import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.txndata.TxnData;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;

@RequiredArgsConstructor
public class Blockdata implements Module {
  private final Wcp wcp;
  private final Euc euc;
  private final TxnData txnData;

  private final Deque<BlockdataOperation> operations = new ArrayDeque<>();
  private long firstBlockNumber;
  private Bytes chainId;

  final OpCode[] opCodes = {
    OpCode.COINBASE,
    OpCode.TIMESTAMP,
    OpCode.NUMBER,
    OpCode.DIFFICULTY,
    OpCode.GASLIMIT,
    OpCode.CHAINID,
    OpCode.BASEFEE
  };

  public void setChainId(BigInteger chainId) {
    this.chainId = EWord.of(chainId).lo();
  }

  @Override
  public String moduleKey() {
    return "BLOCK_DATA";
  }

  @Override
  public void traceStartConflation(final long blockCount) {
    wcp.additionalRows.add(
        LLARGE // for COINBASE
            + 6
            + 6 // for TIMESTAMP
            + 1
            + 6 // for NUMBER
            + 1 // for DIFFICULTY
            + Bytes.minimalBytes(GAS_LIMIT_MAXIMUM).size() * 4 // for GASLIMIT
            + LLARGE // for CHAINID
            + LLARGE // for BASEFEE
        );

    euc.additionalRows.add(8);
  }

  @Override
  public void traceEndBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    final long blockNumber = blockHeader.getNumber();
    if (operations.isEmpty()) {
      firstBlockNumber = blockNumber;
    }
    final BlockHeader previousBlockHeader =
        operations.isEmpty() ? null : operations.getLast().blockHeader();
    for (OpCode opCode : opCodes) {
      final BlockdataOperation operation =
          new BlockdataOperation(
              txnData.hub(),
              blockHeader,
              previousBlockHeader,
              txnData.currentBlock().getNbOfTxsInBlock(),
              wcp,
              euc,
              chainId,
              opCode,
              firstBlockNumber);
      operations.addLast(operation);
    }
  }

  @Override
  public void enterTransaction() {}

  @Override
  public void popTransaction() {}

  @Override
  public int lineCount() {
    final int numberOfBlock = (operations.size() / opCodes.length);
    return numberOfBlock * nROWS_DEPTH;
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    for (BlockdataOperation blockData : operations) {
      blockData.trace(trace);
    }
  }
}
