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

package net.consensys.linea.zktracer.module.txn_data;

import static net.consensys.linea.zktracer.module.Util.getTxTypeAsInt;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.stream.Stream;

import lombok.RequiredArgsConstructor;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.hub.Hub;
import net.consensys.linea.zktracer.module.romLex.RomLex;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.gas.GasConstants;
import net.consensys.linea.zktracer.types.EWord;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.evm.account.AccountState;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;

@RequiredArgsConstructor
public class TxnData implements Module {
  private static final int LT = 16;
  static final int COMMON_RLP_TXN_PHASE_NUMBER_0 = 0;
  static final int COMMON_RLP_TXN_PHASE_NUMBER_1 = 7;
  static final int COMMON_RLP_TXN_PHASE_NUMBER_2 = 2;
  static final int COMMON_RLP_TXN_PHASE_NUMBER_3 = 8;
  static final int COMMON_RLP_TXN_PHASE_NUMBER_4 = 9;
  static final int COMMON_RLP_TXN_PHASE_NUMBER_5 = 6;
  static final int TYPE_0_RLP_TXN_PHASE_NUMBER_6 = 3;
  static final int TYPE_1_RLP_TXN_PHASE_NUMBER_6 = 3;
  static final int TYPE_1_RLP_TXN_PHASE_NUMBER_7 = 10;
  static final int TYPE_2_RLP_TXN_PHASE_NUMBER_6 = 5;
  static final int TYPE_2_RLP_TXN_PHASE_NUMBER_7 = 10;

  private final Hub hub;
  private final RomLex romLex;
  private final Wcp wcp;

  @Override
  public String jsonKey() {
    return "txnData";
  }

  private final List<BlockSnapshot> blocks = new ArrayList<>();

  @Override
  public void enterTransaction() {
    this.currentBlock().getTxs().enter();
  }

  @Override
  public void popTransaction() {
    this.currentBlock().getTxs().pop();
  }

  private BlockSnapshot currentBlock() {
    return this.blocks.get(this.blocks.size() - 1);
  }

  @Override
  public final void traceStartBlock(final ProcessableBlockHeader blockHeader) {
    this.blocks.add(new BlockSnapshot(this.blocks.size() + 1, blockHeader));
  }

  @Override
  public void traceStartTx(WorldView worldView, Transaction tx) {
    int codeIdBeforeLex = 0;
    if ((tx.getTo().isEmpty() && tx.getInit().isPresent() && !tx.getInit().orElseThrow().isEmpty()
        || tx.getTo().isPresent()
            && Optional.ofNullable(worldView.get(tx.getTo().orElseThrow()))
                .map(AccountState::hasCode)
                .orElse(false))) {
      codeIdBeforeLex = this.romLex.codeIdentifierBeforeLexOrder;
    }
    this.currentBlock().captureTx(codeIdBeforeLex, worldView, tx);
  }

  @Override
  public void traceEndTx(
      WorldView worldView,
      Transaction tx,
      boolean status,
      Bytes output,
      List<Log> logs,
      long cumulativeGasUsed) {
    final long refundCounter = hub.refundedGas();
    final long leftoverGas = hub.remainingGas();
    this.currentBlock().endTx(cumulativeGasUsed, leftoverGas, refundCounter, status);

    // Call the wcp module:
    if (!this.currentBlock().getTxs().isEmpty()) {
      this.callWcp(this.currentBlock(), this.currentBlock().currentTx());
    }
  }

  public void callWcp(BlockSnapshot block, TransactionSnapshot tx) {
    // ct = 0
    this.wcp.callLT(
        Bytes32.leftPad(bigIntegerToBytes(tx.initialSenderBalance())),
        Bytes32.leftPad(bigIntegerToBytes(tx.getMaximalUpfrontCost())));
    // ct = 1
    this.wcp.callLT(
        Bytes32.leftPad(Bytes.minimalBytes(tx.gasLimit())),
        Bytes32.leftPad(bigIntegerToBytes(tx.getMaximalUpfrontCost())));
    // ct = 2
    this.wcp.callLT(
        Bytes32.leftPad(bigIntegerToBytes(tx.getLimitMinusLeftoverGas())),
        Bytes32.leftPad(Bytes.minimalBytes(tx.refundCounter())));
    // ct = 3
    this.wcp.callLT(
        Bytes32.leftPad(bigIntegerToBytes(tx.getLimitMinusLeftoverGasDividedByTwo())),
        Bytes32.leftPad(bigIntegerToBytes(tx.getLimitMinusLeftoverGasDividedByTwo())));

    if (tx.type() == TransactionType.EIP1559) {
      // ct = 4
      this.wcp.callLT(
          Bytes32.leftPad(bigIntegerToBytes(tx.maxFeePerGas().orElseThrow().getAsBigInteger())),
          Bytes32.leftPad(bigIntegerToBytes(block.getBaseFee().orElseThrow().getAsBigInteger())));
      // ct = 5
      this.wcp.callLT(
          Bytes32.leftPad(bigIntegerToBytes(tx.maxFeePerGas().orElseThrow().getAsBigInteger())),
          Bytes32.leftPad(
              bigIntegerToBytes(tx.maxPriorityFeePerGas().orElseThrow().getAsBigInteger())));
      // ct = 6
      this.wcp.callLT(
          Bytes32.leftPad(bigIntegerToBytes(tx.maxFeePerGas().orElseThrow().getAsBigInteger())),
          Bytes32.leftPad(
              bigIntegerToBytes(
                  block
                      .getBaseFee()
                      .orElseThrow()
                      .getAsBigInteger()
                      .add(tx.maxPriorityFeePerGas().orElseThrow().getAsBigInteger()))));
    }
  }

  @Override
  public int lineCount() {
    int traceSize = 0;
    for (BlockSnapshot block : this.blocks) {
      for (TransactionSnapshot tx : block.getTxs()) {
        switch (tx.type()) {
          case FRONTIER -> traceSize += 7;
          case ACCESS_LIST, EIP1559 -> traceSize += 8;
          default -> throw new RuntimeException("Transaction type not supported:" + tx.type());
        }
      }
    }
    return traceSize;
  }

  private List<List<BigInteger>> setOutgoingHisAndLos(TransactionSnapshot tx) {
    final EWord toAddress = EWord.of(tx.to());
    boolean isDeployment = tx.isDeployment();

    List<BigInteger> commonHis =
        List.of(
            BigInteger.ZERO, // ct = 0
            isDeployment ? BigInteger.ZERO : toAddress.hiBigInt(), // ct = 1
            BigInteger.ZERO, // ct = 2
            isDeployment ? BigInteger.ONE : BigInteger.ZERO, // ct = 3
            BigInteger.valueOf(tx.dataCost()), // ct = 4
            BigInteger.ZERO // ct = 5
            );

    List<BigInteger> commonLos =
        List.of(
            BigInteger.valueOf(tx.typeAsInt()), // ct = 0
            isDeployment ? BigInteger.ZERO : toAddress.loBigInt(), // ct = 1
            BigInteger.valueOf(tx.nonce()), // ct = 2
            tx.value(), // ct = 3
            BigInteger.valueOf(tx.payload().size()), // ct = 4
            BigInteger.valueOf(tx.gasLimit()) // ct = 5
            );

    List<BigInteger> suffixHi;
    List<BigInteger> suffixLo;

    switch (tx.type()) {
      case FRONTIER -> {
        suffixHi =
            List.of(
                BigInteger.ZERO // ct = 6
                );
        suffixLo =
            List.of(
                tx.effectiveGasPrice() // ct = 6
                );
      }
      case ACCESS_LIST -> {
        suffixHi =
            List.of(
                BigInteger.ZERO, // ct = 6
                BigInteger.valueOf(tx.prewarmedStorageKeysCount()) // ct = 7
                );
        suffixLo =
            List.of(
                tx.effectiveGasPrice(), // ct = 6
                BigInteger.valueOf(tx.prewarmedAddressesCount()) // ct = 7
                );
      }
      case EIP1559 -> {
        suffixHi =
            List.of(
                tx.maxPriorityFeePerGas().orElseThrow().getAsBigInteger(), // ct = 6
                BigInteger.valueOf(tx.prewarmedStorageKeysCount()) // ct = 7
                );
        suffixLo =
            List.of(
                tx.maxFeePerGas().orElseThrow().getAsBigInteger(), // ct = 6
                BigInteger.valueOf(tx.prewarmedAddressesCount()) // ct = 7
                );
      }
      default -> throw new RuntimeException("transaction type not supported");
    }
    return List.of(
        Stream.concat(commonHis.stream(), suffixHi.stream()).toList(),
        Stream.concat(commonLos.stream(), suffixLo.stream()).toList());
  }

  private List<Integer> setWcpInst(TransactionSnapshot tx) {
    List<Integer> common =
        List.of(
            LT, // ct = 0
            LT, // ct = 1
            LT, // ct = 2
            LT // ct = 3
            );
    List<Integer> suffix =
        switch (tx.type()) {
          case FRONTIER -> List.of(
              0, // ct = 4
              0, // ct = 5
              0 // ct = 6
              );
          case ACCESS_LIST -> List.of(
              0, // ct = 4
              0, // ct = 5
              0, // ct = 6
              0 // ct = 7
              );
          case EIP1559 -> List.of(
              LT, // ct = 4
              LT, // ct = 5
              LT, // ct = 6
              0 // ct = 7
              );
          default -> throw new RuntimeException("transaction type not supported");
        };
    return Stream.concat(common.stream(), suffix.stream()).toList();
  }

  private List<List<BigInteger>> setWcpArguments(BlockSnapshot block, TransactionSnapshot tx) {

    List<BigInteger> commonOnes =
        List.of(
            tx.initialSenderBalance(), // ct = 0
            BigInteger.valueOf(tx.gasLimit()), // ct = 1
            tx.getLimitMinusLeftoverGas(), // ct = 2
            BigInteger.valueOf(tx.refundCounter()) // ct = 3
            );

    List<BigInteger> suffixOnes =
        switch (tx.type()) {
          case FRONTIER -> List.of(
              BigInteger.ZERO, // ct = 4
              BigInteger.ZERO, // ct = 5
              BigInteger.ZERO // ct = 6
              );
          case ACCESS_LIST -> List.of(
              BigInteger.ZERO, // ct = 4
              BigInteger.ZERO, // ct = 5
              BigInteger.ZERO, // ct = 6
              BigInteger.ZERO // ct = 7
              );
          case EIP1559 -> List.of(
              tx.maxFeePerGas().orElseThrow().getAsBigInteger(), // ct = 4
              tx.maxFeePerGas().orElseThrow().getAsBigInteger(), // ct = 5
              tx.maxFeePerGas().orElseThrow().getAsBigInteger(), // ct = 6
              BigInteger.ZERO // ct = 7
              );
          default -> throw new RuntimeException("transaction type not supported");
        };

    List<BigInteger> commonTwos =
        List.of(
            tx.getMaximalUpfrontCost(), // ct = 0
            BigInteger.valueOf(tx.getUpfrontGasCost()), // ct = 1
            tx.getLimitMinusLeftoverGasDividedByTwo(), // ct = 2
            tx.getLimitMinusLeftoverGasDividedByTwo() // ct = 3
            );

    List<BigInteger> suffixTwos =
        switch (tx.type()) {
          case FRONTIER -> List.of(
              BigInteger.ZERO, // ct = 4
              BigInteger.ZERO, // ct = 5
              BigInteger.ZERO // ct = 6
              );
          case ACCESS_LIST -> List.of(
              BigInteger.ZERO, // ct = 4
              BigInteger.ZERO, // ct = 5
              BigInteger.ZERO, // ct = 6
              BigInteger.ZERO // ct = 7
              );
          case EIP1559 -> List.of(
              block.getBaseFee().orElseThrow().getAsBigInteger(), // ct = 4
              tx.maxPriorityFeePerGas().orElseThrow().getAsBigInteger(), // ct = 5
              block
                  .getBaseFee()
                  .orElseThrow()
                  .getAsBigInteger()
                  .add(tx.maxPriorityFeePerGas().orElseThrow().getAsBigInteger()), // ct = 6
              BigInteger.ZERO // ct = 7
              );
          default -> throw new RuntimeException("transaction type not supported");
        };

    return List.of(
        Stream.concat(commonOnes.stream(), suffixOnes.stream()).toList(),
        Stream.concat(commonTwos.stream(), suffixTwos.stream()).toList());
  }

  private List<Boolean> setWcpRes(BlockSnapshot block, TransactionSnapshot tx) {
    return List.of(
        false, // ct = 0
        false, // ct = 1
        false, // ct = 2
        tx.getLimitMinusLeftoverGasDividedByTwo().compareTo(BigInteger.valueOf(tx.refundCounter()))
            >= 0, // ct = 3,
        false, // ct = 4
        false, // ct = 5
        tx.type() == TransactionType.EIP1559
            && tx.maxFeePerGas()
                    .orElseThrow()
                    .getAsBigInteger()
                    .compareTo(
                        block
                            .getBaseFee()
                            .orElseThrow()
                            .getAsBigInteger()
                            .add(tx.maxPriorityFeePerGas().orElseThrow().getAsBigInteger()))
                < 0, // ct = 6,
        false // ct = 7
        );
  }

  private List<Integer> setPhaseRlpTxnNumbers(TransactionSnapshot tx) {
    List<Integer> common =
        List.of(
            COMMON_RLP_TXN_PHASE_NUMBER_0, // ct = 0
            COMMON_RLP_TXN_PHASE_NUMBER_1, // ct = 1
            COMMON_RLP_TXN_PHASE_NUMBER_2, // ct = 2
            COMMON_RLP_TXN_PHASE_NUMBER_3, // ct = 3
            COMMON_RLP_TXN_PHASE_NUMBER_4, // ct = 4
            COMMON_RLP_TXN_PHASE_NUMBER_5 // ct = 5
            );

    List<Integer> phaseDependentSuffix;

    switch (tx.type()) {
      case FRONTIER -> phaseDependentSuffix =
          List.of(
              TYPE_0_RLP_TXN_PHASE_NUMBER_6 // ct = 6
              );
      case ACCESS_LIST -> phaseDependentSuffix =
          List.of(
              TYPE_1_RLP_TXN_PHASE_NUMBER_6, // ct = 6
              TYPE_1_RLP_TXN_PHASE_NUMBER_7 // ct = 7
              );
      case EIP1559 -> phaseDependentSuffix =
          List.of(
              TYPE_2_RLP_TXN_PHASE_NUMBER_6, // ct = 6
              TYPE_2_RLP_TXN_PHASE_NUMBER_7 // ct = 7
              );
      default -> throw new RuntimeException("transaction type not supported");
    }
    return Stream.concat(common.stream(), phaseDependentSuffix.stream()).toList();
  }

  private List<Integer> setPhaseRlpTxnRcpt(TransactionSnapshot tx) {
    return List.of(
        Trace.RLPRECEIPT_SUBPHASE_ID_TYPE, // ct =0
        Trace.RLPRECEIPT_SUBPHASE_ID_STATUS_CODE, // ct = 1
        Trace.RLPRECEIPT_SUBPHASE_ID_CUMUL_GAS, // ct = 2
        0, // ct = 3
        0, // ct = 4
        0, // ct = 5
        0, // ct = 6
        0 // ct = 7
        );
  }

  private List<Long> setOutgoingRlpTxnRcpt(TransactionSnapshot tx) {
    Long statusTx = 0L;
    if (tx.status()) {
      statusTx = 1L;
    }

    return List.of(
        (long) getTxTypeAsInt(tx.type()), // ct = 0
        statusTx, // ct = 1
        tx.cumulativeGasConsumption(), // ct = 2
        0L, // ct = 3
        0L, // ct = 4
        0L, // ct = 5
        0L, // ct = 6
        0L // ct = 7
        );
  }

  // getRefundCounter returns the sum of SSTORE related refunds
  // + the sum of SELFDESTRUCT related refunds.
  // Reference: [EYP] §6.2. Execution. Equation (71)
  long getRefundCounter(final MessageFrame frame) {
    long sstoreGasRefunds = frame.getGasRefund();
    long selfdestructGasRefunds =
        (long) frame.getSelfDestructs().size() * GasConstants.R_SELF_DESTRUCT.cost();
    return sstoreGasRefunds + selfdestructGasRefunds;
  }

  private void traceTx(
      Trace trace,
      BlockSnapshot block,
      TransactionSnapshot tx,
      int absTxNumMax,
      int absTxNum,
      int btcNumMax,
      int btcNum,
      int relTxNumMax,
      int relTxNum) {
    final EWord from = EWord.of(tx.from());
    final EWord to = EWord.of(tx.to());
    final EWord coinbase = EWord.of(block.getCoinbaseAddress());
    int codeFragmentIndex = 0;
    if (tx.codeIdBeforeLex() != 0) {
      codeFragmentIndex = this.romLex.getCFIById(tx.codeIdBeforeLex());
    }
    final List<BigInteger> outgoingHis = setOutgoingHisAndLos(tx).get(0);
    final List<BigInteger> outgoingLos = setOutgoingHisAndLos(tx).get(1);
    final List<Integer> wcpInsts = setWcpInst(tx);
    final List<BigInteger> wcpArgOnes = setWcpArguments(block, tx).get(0);
    final List<BigInteger> wcpArgTwos = setWcpArguments(block, tx).get(1);
    final List<Boolean> wcpRes = setWcpRes(block, tx);
    final List<Integer> phaseNumbers = setPhaseRlpTxnNumbers(tx);
    final List<Integer> phaseRlpTxnRcpt = setPhaseRlpTxnRcpt(tx);
    final List<Long> outgoingRlpTxnRcpt = setOutgoingRlpTxnRcpt(tx);
    for (int ct = 0; ct < tx.maxCounter(); ct++) {
      trace
          .absTxNumMax(Bytes.ofUnsignedInt(absTxNumMax))
          .absTxNum(Bytes.ofUnsignedInt(absTxNum))
          .btcNumMax(Bytes.ofUnsignedInt(btcNumMax))
          .btcNum(Bytes.ofUnsignedInt(btcNum))
          .relTxNumMax(Bytes.ofUnsignedInt(relTxNumMax))
          .relTxNum(Bytes.ofUnsignedInt(relTxNum))
          .ct(Bytes.ofUnsignedInt(ct))
          .fromHi(from.hi())
          .fromLo(from.lo())
          .nonce(Bytes.ofUnsignedLong(tx.nonce()))
          .initialBalance(bigIntegerToBytes(tx.initialSenderBalance()))
          .value(bigIntegerToBytes(tx.value()))
          .toHi(to.hi())
          .toLo(to.lo())
          .isDep(tx.isDeployment())
          .gasLimit(Bytes.ofUnsignedLong(tx.gasLimit()))
          .initialGas(Bytes.ofUnsignedLong(tx.gasLimit() - tx.getUpfrontGasCost()))
          .gasPrice(bigIntegerToBytes(tx.effectiveGasPrice()))
          .basefee(block.getBaseFee().orElseThrow())
          .coinbaseHi(coinbase.hi())
          .coinbaseLo(coinbase.lo())
          .callDataSize(tx.isDeployment() ? Bytes.EMPTY : Bytes.ofUnsignedInt(tx.payload().size()))
          .initCodeSize(tx.isDeployment() ? Bytes.ofUnsignedInt(tx.payload().size()) : Bytes.EMPTY)
          .type0(tx.type() == TransactionType.FRONTIER)
          .type1(tx.type() == TransactionType.ACCESS_LIST)
          .type2(tx.type() == TransactionType.EIP1559)
          .requiresEvmExecution(tx.requiresEvmExecution())
          .leftoverGas(Bytes.ofUnsignedLong(tx.leftoverGas()))
          .refundCounter(Bytes.ofUnsignedLong(tx.refundCounter()))
          .refundAmount(Bytes.ofUnsignedLong(tx.effectiveGasRefund()))
          .cumulativeConsumedGas(Bytes.ofUnsignedLong(tx.cumulativeGasConsumption()))
          .statusCode(tx.status())
          .codeFragmentIndex(Bytes.ofUnsignedInt(codeFragmentIndex))
          .phaseRlpTxn(Bytes.ofUnsignedInt(phaseNumbers.get(ct)))
          .outgoingHi(bigIntegerToBytes(outgoingHis.get(ct)))
          .outgoingLo(bigIntegerToBytes(outgoingLos.get(ct)))
          .wcpArgOneLo(bigIntegerToBytes(wcpArgOnes.get(ct)))
          .wcpArgTwoLo(bigIntegerToBytes(wcpArgTwos.get(ct)))
          .wcpResLo(wcpRes.get(ct))
          .wcpInst(Bytes.ofUnsignedInt(wcpInsts.get(ct)))
          .phaseRlpTxnrcpt(Bytes.ofUnsignedInt(phaseRlpTxnRcpt.get(ct)))
          .outgoingRlpTxnrcpt(Bytes.ofUnsignedLong(outgoingRlpTxnRcpt.get(ct)))
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
    int batchNumMax = 0;
    int btchNum = 0;
    for (BlockSnapshot block : this.blocks) {
      absTxNumMax += block.getTxs().size();
      if (!block.getTxs().isEmpty()) {
        batchNumMax += 1;
      }
    }
    for (BlockSnapshot block : this.blocks) {
      int relTxNumMax = block.getTxs().size();
      if (relTxNumMax != 0) {
        btchNum++;
        int relTxNum = 0;
        for (TransactionSnapshot tx : block.getTxs()) {
          absTxNum++;
          relTxNum++;
          this.traceTx(
              trace, block, tx, absTxNumMax, absTxNum, batchNumMax, btchNum, relTxNumMax, relTxNum);
        }
      }
    }
  }
}
