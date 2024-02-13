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
import net.consensys.linea.zktracer.types.Bytes16;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.UnsignedByte;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.evm.account.AccountState;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;
import org.jetbrains.annotations.NotNull;

@RequiredArgsConstructor
public class TxnData implements Module {
  private static final int N_ROWS_FRONTIER_TX = 7;
  private static final int N_ROWS_ACCESS_LIST_TX = 8;
  private static final int N_ROWS_EIP_1559_TX = 8;
  private static final int N_ROWS_TX_MAX =
      Math.max(Math.max(N_ROWS_FRONTIER_TX, N_ROWS_ACCESS_LIST_TX), N_ROWS_EIP_1559_TX);
  private static final int LT = 16;
  private static final int ISZERO = 21;
  static final int COMMON_RLP_TXN_PHASE_NUMBER_0 =
      net.consensys.linea.zktracer.module.rlp_txn.Trace.PHASE_RLP_PREFIX_VALUE;
  static final int COMMON_RLP_TXN_PHASE_NUMBER_1 =
      net.consensys.linea.zktracer.module.rlp_txn.Trace.PHASE_TO_VALUE;
  static final int COMMON_RLP_TXN_PHASE_NUMBER_2 =
      net.consensys.linea.zktracer.module.rlp_txn.Trace.PHASE_NONCE_VALUE;
  static final int COMMON_RLP_TXN_PHASE_NUMBER_3 =
      net.consensys.linea.zktracer.module.rlp_txn.Trace.PHASE_VALUE_VALUE;
  static final int COMMON_RLP_TXN_PHASE_NUMBER_4 =
      net.consensys.linea.zktracer.module.rlp_txn.Trace.PHASE_DATA_VALUE;
  static final int COMMON_RLP_TXN_PHASE_NUMBER_5 =
      net.consensys.linea.zktracer.module.rlp_txn.Trace.PHASE_GAS_LIMIT_VALUE;
  static final int TYPE_0_RLP_TXN_PHASE_NUMBER_6 =
      net.consensys.linea.zktracer.module.rlp_txn.Trace.PHASE_GAS_PRICE_VALUE;
  static final int TYPE_1_RLP_TXN_PHASE_NUMBER_6 =
      net.consensys.linea.zktracer.module.rlp_txn.Trace.PHASE_GAS_PRICE_VALUE;
  static final int TYPE_1_RLP_TXN_PHASE_NUMBER_7 =
      net.consensys.linea.zktracer.module.rlp_txn.Trace.PHASE_ACCESS_LIST_VALUE;
  static final int TYPE_2_RLP_TXN_PHASE_NUMBER_6 =
      net.consensys.linea.zktracer.module.rlp_txn.Trace.PHASE_MAX_FEE_PER_GAS_VALUE;
  static final int TYPE_2_RLP_TXN_PHASE_NUMBER_7 =
      net.consensys.linea.zktracer.module.rlp_txn.Trace.PHASE_ACCESS_LIST_VALUE;

  private final Hub hub;
  private final RomLex romLex;
  private final Wcp wcp;

  @Override
  public String moduleKey() {
    return "TXN_DATA";
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
      boolean isSuccessful,
      Bytes output,
      List<Log> logs,
      long cumulativeGasUsed) {
    final long refundCounter = hub.refundedGas();
    final long leftoverGas = hub.remainingGas();
    this.currentBlock().endTx(cumulativeGasUsed, leftoverGas, refundCounter, isSuccessful);

    // Call the wcp module:
    if (!this.currentBlock().getTxs().isEmpty()) {
      this.callWcp();
    }
  }

  public void callWcp() {
    final List<Bytes16> wcpArgOneLo = setWcpArgumentOne(this.currentBlock().currentTx());
    final List<Bytes16> wcpArgTwoLo =
        setWcpArgumentTwo(this.currentBlock(), this.currentBlock().currentTx());
    // wcp call row from 0 to 3
    for (int ct = 0; ct < 4; ct++) {
      this.wcp.callLT(Bytes32.leftPad(wcpArgOneLo.get(ct)), Bytes32.leftPad(wcpArgTwoLo.get(ct)));
    }
    // wcp call row 4
    this.wcp.callISZERO(Bytes32.leftPad(wcpArgOneLo.get(4)));
    // wcp call row 5 to 7 for Type 2 transaction only
    if (this.currentBlock().currentTx().type() == TransactionType.EIP1559) {
      for (int ct = 5; ct < 8; ct++) {
        this.wcp.callLT(Bytes32.leftPad(wcpArgOneLo.get(ct)), Bytes32.leftPad(wcpArgTwoLo.get(ct)));
      }
    }
  }

  @Override
  public int lineCount() {
    int traceSize = 0;
    for (BlockSnapshot block : this.blocks) {
      for (TransactionSnapshot tx : block.getTxs()) {
        switch (tx.type()) {
          case FRONTIER -> traceSize += N_ROWS_FRONTIER_TX;
          case ACCESS_LIST -> traceSize += N_ROWS_ACCESS_LIST_TX;
          case EIP1559 -> traceSize += N_ROWS_EIP_1559_TX;
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
            LT, // ct = 3
            ISZERO // ct = 4
            );
    List<Integer> suffix =
        switch (tx.type()) {
          case FRONTIER, ACCESS_LIST -> List.of(
              0, // ct = 5
              0, // ct = 6
              0 // ct = 7
              );
          case EIP1559 -> List.of(
              LT, // ct = 5
              LT, // ct = 6
              LT // ct = 7
              );
          default -> throw new RuntimeException("transaction type not supported");
        };
    return Stream.concat(common.stream(), suffix.stream()).toList();
  }

  private List<Bytes16> setWcpArgumentOne(TransactionSnapshot tx) {
    List<Bytes16> output = new ArrayList<>(N_ROWS_TX_MAX);

    output.add(0, Bytes16.leftPad(bigIntegerToBytes(tx.initialSenderBalance()))); // ct = 0
    output.add(1, Bytes16.leftPad(Bytes.ofUnsignedLong(tx.gasLimit()))); // ct = 1
    output.add(2, Bytes16.leftPad(bigIntegerToBytes(tx.getLimitMinusLeftoverGas()))); // ct = 2
    output.add(3, Bytes16.leftPad(Bytes.ofUnsignedLong(tx.refundCounter()))); // ct = 3
    output.add(4, Bytes16.leftPad(Bytes.ofUnsignedLong(tx.callDataSize()))); // ct = 4

    switch (tx.type()) {
      case FRONTIER, ACCESS_LIST -> {
        output.add(5, Bytes16.ZERO); // ct = 5
        output.add(6, Bytes16.ZERO); // ct = 6
        output.add(7, Bytes16.ZERO); // ct = 7
      }

      case EIP1559 -> {
        final Bytes16 maxFeePerGas =
            Bytes16.leftPad(bigIntegerToBytes(tx.maxFeePerGas().orElseThrow().getAsBigInteger()));
        output.add(5, maxFeePerGas); // ct = 5
        output.add(6, maxFeePerGas); // ct = 6
        output.add(7, maxFeePerGas); // ct = 7
      }
      default -> throw new RuntimeException("transaction type not supported");
    }
    return output;
  }

  private List<Bytes16> setWcpArgumentTwo(BlockSnapshot block, TransactionSnapshot tx) {
    final Bytes16 limitMinusLeftOverGasDividedByTwo =
        Bytes16.leftPad(bigIntegerToBytes(tx.getLimitMinusLeftoverGasDividedByTwo()));

    List<Bytes16> commonTwos =
        List.of(
            Bytes16.leftPad(bigIntegerToBytes(tx.getMaximalUpfrontCost())), // ct = 0
            Bytes16.leftPad(Bytes.ofUnsignedLong(tx.getUpfrontGasCost())), // ct = 1
            limitMinusLeftOverGasDividedByTwo, // ct = 2
            limitMinusLeftOverGasDividedByTwo, // ct = 3
            Bytes16.ZERO // ct = 4
            );

    List<Bytes16> suffixTwos =
        switch (tx.type()) {
          case FRONTIER, ACCESS_LIST -> List.of(
              Bytes16.ZERO, // ct = 5
              Bytes16.ZERO, // ct = 6
              Bytes16.ZERO // ct =7
              );
          case EIP1559 -> List.of(
              Bytes16.leftPad(
                  bigIntegerToBytes(block.getBaseFee().orElseThrow().getAsBigInteger())), // ct = 5
              Bytes16.leftPad(
                  bigIntegerToBytes(
                      tx.maxPriorityFeePerGas().orElseThrow().getAsBigInteger())), // ct = 6
              Bytes16.leftPad(
                  bigIntegerToBytes(
                      block
                          .getBaseFee()
                          .orElseThrow()
                          .getAsBigInteger()
                          .add(
                              tx.maxPriorityFeePerGas().orElseThrow().getAsBigInteger()))) // ct = 7
              );
          default -> throw new IllegalStateException("transaction type not supported:" + tx.type());
        };

    return Stream.concat(commonTwos.stream(), suffixTwos.stream()).toList();
  }

  private List<Boolean> setWcpRes(BlockSnapshot block, TransactionSnapshot tx) {
    return List.of(
        false, // ct = 0
        false, // ct = 1
        false, // ct = 2
        tx.getLimitMinusLeftoverGasDividedByTwo().compareTo(BigInteger.valueOf(tx.refundCounter()))
            >= 0, // ct = 3,
        tx.callDataSize() == 0, // ct = 4
        false, // ct = 5
        false, // ct = 6
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
                < 0 // ct = 7,
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

    List<Integer> phaseDependentSuffix = computePhaseDependentSuffix(tx);
    return Stream.concat(common.stream(), phaseDependentSuffix.stream()).toList();
  }

  @NotNull
  private static List<Integer> computePhaseDependentSuffix(TransactionSnapshot tx) {
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
      default -> throw new IllegalStateException("transaction type not supported:" + tx.type());
    }
    return phaseDependentSuffix;
  }

  private List<Integer> setPhaseRlpTxnRcpt() {
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
    long statusTx = 0L;
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
    final int codeFragmentIndex =
        tx.codeIdBeforeLex() == 0 ? 0 : this.romLex.getCFIById(tx.codeIdBeforeLex());
    final List<BigInteger> outgoingHis = setOutgoingHisAndLos(tx).get(0);
    final List<BigInteger> outgoingLos = setOutgoingHisAndLos(tx).get(1);
    final List<Bytes16> wcpArgOneLo = setWcpArgumentOne(tx);
    final List<Bytes16> wcpArgTwoLo = setWcpArgumentTwo(block, tx);
    final List<Integer> wcpInsts = setWcpInst(tx);
    final List<Boolean> wcpRes = setWcpRes(block, tx);
    final List<Integer> phaseNumbers = setPhaseRlpTxnNumbers(tx);
    final List<Integer> phaseRlpTxnRcpt = setPhaseRlpTxnRcpt();
    final List<Long> outgoingRlpTxnRcpt = setOutgoingRlpTxnRcpt(tx);
    final boolean copyTxCd = tx.requiresEvmExecution() && tx.callDataSize() != 0;
    for (int ct = 0; ct < tx.maxCounter(); ct++) {
      trace
          .absTxNumMax(Bytes.ofUnsignedInt(absTxNumMax))
          .absTxNum(Bytes.ofUnsignedInt(absTxNum))
          .btcNumMax(Bytes.ofUnsignedInt(btcNumMax))
          .btcNum(Bytes.ofUnsignedInt(btcNum))
          .relTxNumMax(Bytes.ofUnsignedInt(relTxNumMax))
          .relTxNum(Bytes.ofUnsignedInt(relTxNum))
          .ct(UnsignedByte.of(ct))
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
          .callDataSize(Bytes.ofUnsignedShort(tx.callDataSize()))
          .initCodeSize(tx.isDeployment() ? Bytes.ofUnsignedInt(tx.payload().size()) : Bytes.EMPTY)
          .type0(tx.type() == TransactionType.FRONTIER)
          .type1(tx.type() == TransactionType.ACCESS_LIST)
          .type2(tx.type() == TransactionType.EIP1559)
          .requiresEvmExecution(tx.requiresEvmExecution())
          .copyTxcdAtInitialisation(copyTxCd)
          .leftoverGas(Bytes.ofUnsignedLong(tx.leftoverGas()))
          .refundCounter(Bytes.ofUnsignedLong(tx.refundCounter()))
          .refundAmount(Bytes.ofUnsignedLong(tx.effectiveGasRefund()))
          .cumulativeConsumedGas(Bytes.ofUnsignedLong(tx.cumulativeGasConsumption()))
          .statusCode(tx.status())
          .codeFragmentIndex(Bytes.ofUnsignedInt(codeFragmentIndex))
          .phaseRlpTxn(UnsignedByte.of(phaseNumbers.get(ct)))
          .outgoingHi(bigIntegerToBytes(outgoingHis.get(ct)))
          .outgoingLo(bigIntegerToBytes(outgoingLos.get(ct)))
          .wcpArgOneLo(wcpArgOneLo.get(ct))
          .wcpArgTwoLo(wcpArgTwoLo.get(ct))
          .wcpRes(wcpRes.get(ct))
          .wcpInst(UnsignedByte.of(wcpInsts.get(ct)))
          .phaseRlpTxnrcpt(UnsignedByte.of(phaseRlpTxnRcpt.get(ct)))
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
