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

package net.consensys.linea.zktracer.module.txndata;

import static net.consensys.linea.zktracer.module.txndata.Trace.NB_ROWS_TYPE_0;
import static net.consensys.linea.zktracer.module.txndata.Trace.NB_ROWS_TYPE_1;
import static net.consensys.linea.zktracer.module.txndata.Trace.NB_ROWS_TYPE_2;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import java.nio.MappedByteBuffer;
import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Deque;
import java.util.List;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.romlex.ContractMetadata;
import net.consensys.linea.zktracer.module.romlex.RomLex;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;

@RequiredArgsConstructor
public class TxnData implements Module {

  private static final int N_ROWS_TX_MAX =
      Math.max(Math.max(NB_ROWS_TYPE_0, NB_ROWS_TYPE_1), NB_ROWS_TYPE_2);

  private final Hub hub;
  private final RomLex romLex;
  private final Wcp wcp;
  private final Euc euc;

  @Override
  public String moduleKey() {
    return "TXN_DATA";
  }

  public final List<BlockSnapshot> blocks = new ArrayList<>();

  /** accumulate the gas used since the beginning of the current block */
  public final Deque<Integer> cumulatedGasUsed = new ArrayDeque<>();

  @Override
  public void enterTransaction() {
    this.currentBlock().getTxs().enter();
  }

  @Override
  public void popTransaction() {
    this.currentBlock().getTxs().pop();
    this.cumulatedGasUsed.pop();
  }

  @Override
  public void traceStartConflation(final long blockCount) {
    this.wcp.additionalRows.push(
        this.wcp.additionalRows.pop() + 4); /* 4 = byte length of LINEA_BLOCK_GAS_LIMIT */
  }

  public BlockSnapshot currentBlock() {
    return this.blocks.get(this.blocks.size() - 1);
  }

  @Override
  public final void traceStartBlock(final ProcessableBlockHeader blockHeader) {
    this.blocks.add(new BlockSnapshot(this.blocks.size() + 1, blockHeader));
    this.cumulatedGasUsed.push(0);
  }

  @Override
  public void traceStartTx(WorldView worldView, Transaction tx) {
    this.currentBlock().captureTx(wcp, euc, worldView, tx);
  }

  @Override
  public void traceEndTx(
      WorldView worldView,
      Transaction tx,
      boolean isSuccessful,
      Bytes output,
      List<Log> logs,
      long gasUsed) {
    final long leftoverGas = tx.getGasLimit() - gasUsed;
    final long refundCounter = hub.refundedGas();
    this.currentBlock().endTx(leftoverGas, refundCounter, isSuccessful);

    final TransactionSnapshot currentTx = this.currentBlock().currentTx();

    final int gasUsedMinusRefunded = (int) (currentTx.gasLimit() - currentTx.effectiveGasRefund());
    this.cumulatedGasUsed.push((this.cumulatedGasUsed.getFirst() + gasUsedMinusRefunded));
    this.currentBlock().currentTx().cumulativeGasConsumption(this.cumulatedGasUsed.getFirst());
  }

  @Override
  public void traceEndBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    this.currentBlock().currentTx().setCallWcpLastTxOfBlock(this.currentBlock().getBlockGasLimit());
  }

  @Override
  public int lineCount() {
    // The last tx of each block has one more rows
    int traceSize = this.blocks.size();

    // Count the number of rows of each tx, only depending on the type of the transaction
    for (BlockSnapshot block : this.blocks) {
      for (TransactionSnapshot tx : block.getTxs()) {
        switch (tx.type()) {
          case FRONTIER -> traceSize += NB_ROWS_TYPE_0;
          case ACCESS_LIST -> traceSize += NB_ROWS_TYPE_1;
          case EIP1559 -> traceSize += NB_ROWS_TYPE_2;
          default -> throw new RuntimeException("Transaction type not supported:" + tx.type());
        }
      }
    }
    return traceSize;
  }

  private void traceTx(
      Trace trace,
      BlockSnapshot block,
      TransactionSnapshot tx,
      int absTxNumMax,
      int absTxNum,
      int blockNum,
      int relTxNumMax,
      int relTxNum) {

    tx.setRlptxnValues();
    tx.setRlptxrcptValues();

    final EWord from = EWord.of(tx.from());
    final EWord to = EWord.of(tx.to());
    final long toHi = to.hi().slice(12, 4).toLong();
    final EWord coinbase = EWord.of(block.getCoinbaseAddress());
    final long coinbaseLo = coinbase.hi().trimLeadingZeros().toLong();
    final int codeFragmentIndex =
        tx.isDeployment() && tx.requiresEvmExecution()
            ? this.romLex.getCodeFragmentIndexByMetadata(
                ContractMetadata.underDeployment(tx.to(), 1))
            : 0;
    final boolean copyTxCd = tx.requiresEvmExecution() && tx.callDataSize() != 0;
    final long fromHi = from.hi().slice(12, 4).toLong();
    final Bytes gasPrice = tx.computeGasPriceColumn();
    final Bytes priorityFeePerGas =
        tx.type() == TransactionType.EIP1559
            ? bigIntegerToBytes(
                tx.maxFeePerGas()
                    .get()
                    .getAsBigInteger()
                    .subtract(tx.baseFee().get().getAsBigInteger()))
            : gasPrice;
    final Bytes nonce = Bytes.ofUnsignedLong(tx.nonce());
    final Bytes initialBalance = bigIntegerToBytes(tx.initialSenderBalance());
    final Bytes value = bigIntegerToBytes(tx.value());

    final boolean isLastTxOfTheBlock = relTxNum == relTxNumMax;
    final int ctMax = isLastTxOfTheBlock ? tx.maxCounter() + 1 : tx.maxCounter();

    for (int ct = 0; ct <= ctMax; ct++) {
      trace
          .absTxNumMax(absTxNumMax)
          .absTxNum(absTxNum)
          .relBlock(blockNum)
          .relTxNumMax(relTxNumMax)
          .relTxNum(relTxNum)
          .isLastTxOfBlock(isLastTxOfTheBlock)
          .ct(UnsignedByte.of(ct))
          .fromHi(fromHi)
          .fromLo(from.lo())
          .nonce(nonce)
          .initialBalance(initialBalance)
          .value(value)
          .toHi(toHi)
          .toLo(to.lo())
          .isDep(tx.isDeployment())
          .gasLimit(Bytes.ofUnsignedLong(tx.gasLimit()))
          .gasInitiallyAvailable(Bytes.ofUnsignedLong(tx.gasLimit() - tx.getUpfrontGasCost()))
          .gasPrice(gasPrice)
          .priorityFeePerGas(priorityFeePerGas)
          .basefee(block.getBaseFee().orElseThrow())
          .coinbaseHi(coinbaseLo)
          .coinbaseLo(coinbase.lo())
          .blockGasLimit(block.getBlockGasLimit())
          .callDataSize(tx.callDataSize())
          .initCodeSize(tx.isDeployment() ? tx.payload().size() : 0)
          .type0(tx.type() == TransactionType.FRONTIER)
          .type1(tx.type() == TransactionType.ACCESS_LIST)
          .type2(tx.type() == TransactionType.EIP1559)
          .requiresEvmExecution(tx.requiresEvmExecution())
          .copyTxcd(copyTxCd)
          .gasLeftover(Bytes.ofUnsignedLong(tx.leftoverGas()))
          .refundCounter(Bytes.ofUnsignedLong(tx.refundCounter()))
          .refundEffective(Bytes.ofUnsignedLong(tx.effectiveGasRefund()))
          .gasCumulative(Bytes.ofUnsignedLong(tx.cumulativeGasConsumption()))
          .statusCode(tx.status())
          .codeFragmentIndex(codeFragmentIndex)
          .phaseRlpTxn(UnsignedByte.of(tx.valuesToRlptxn().get(ct).phase()))
          .outgoingHi(tx.valuesToRlptxn().get(ct).outGoingHi())
          .outgoingLo(tx.valuesToRlptxn().get(ct).outGoingLo())
          .eucFlag(tx.callsToEucAndWcp().get(ct).eucFlag())
          .wcpFlag(tx.callsToEucAndWcp().get(ct).wcpFlag())
          .inst(UnsignedByte.of(tx.callsToEucAndWcp().get(ct).instruction()))
          .argOneLo(tx.callsToEucAndWcp().get(ct).arg1())
          .argTwoLo(tx.callsToEucAndWcp().get(ct).arg2())
          .res(tx.callsToEucAndWcp().get(ct).result())
          .phaseRlpTxnrcpt(UnsignedByte.of(tx.valuesToRlpTxrcpt().get(ct).phase()))
          .outgoingRlpTxnrcpt(Bytes.ofUnsignedLong(tx.valuesToRlpTxrcpt().get(ct).outgoing()))
          .validateRow();
    }
  }

  @Override
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);

    int absTxNumMax = 0;
    int absTxNum = 0;
    int blockNum = 0;
    for (BlockSnapshot block : this.blocks) {
      absTxNumMax += block.getTxs().size();
    }
    for (BlockSnapshot block : this.blocks) {
      final int relTxNumMax = block.getTxs().size();
      if (relTxNumMax != 0) {
        blockNum++;
        int relTxNum = 0;
        for (TransactionSnapshot tx : block.getTxs()) {
          absTxNum++;
          relTxNum++;
          this.traceTx(trace, block, tx, absTxNumMax, absTxNum, blockNum, relTxNumMax, relTxNum);
        }
      }
    }
  }
}
