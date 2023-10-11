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

package net.consensys.linea.zktracer.module.rlp_txn;

import static net.consensys.linea.zktracer.bytes.conversions.bigIntegerToBytes;
import static net.consensys.linea.zktracer.module.rlputils.Pattern.bitDecomposition;
import static net.consensys.linea.zktracer.module.rlputils.Pattern.byteCounting;
import static net.consensys.linea.zktracer.module.rlputils.Pattern.outerRlpSize;
import static net.consensys.linea.zktracer.module.rlputils.Pattern.padToGivenSizeWithLeftZero;
import static net.consensys.linea.zktracer.module.rlputils.Pattern.padToGivenSizeWithRightZero;
import static org.hyperledger.besu.ethereum.core.encoding.EncodingContext.BLOCK_BODY;
import static org.hyperledger.besu.ethereum.core.encoding.TransactionEncoder.encodeOpaqueBytes;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.function.Function;

import com.google.common.base.Preconditions;
import net.consensys.linea.zktracer.bytes.UnsignedByte;
import net.consensys.linea.zktracer.container.stacked.list.StackedList;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.rlputils.BitDecOutput;
import net.consensys.linea.zktracer.module.rlputils.ByteCountAndPowerOutput;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.AccessListEntry;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.TransactionType;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.ethereum.core.encoding.AccessListTransactionEncoder;
import org.hyperledger.besu.ethereum.rlp.RLP;
import org.hyperledger.besu.ethereum.rlp.RLPOutput;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class RlpTxn implements Module {
  final Trace.TraceBuilder builder = Trace.builder();
  public static final int llarge = TxnrlpTrace.LLARGE.intValue();
  public static final Bytes bytesPrefixShortInt =
      bigIntegerToBytes(BigInteger.valueOf(TxnrlpTrace.int_short.intValue()));
  public static final int intPrefixShortInt =
      bytesPrefixShortInt.toUnsignedBigInteger().intValueExact();
  public static final Bytes bytesPrefixLongInt =
      bigIntegerToBytes(BigInteger.valueOf(TxnrlpTrace.int_long.intValue()));
  public static final int intPrefixLongInt =
      bytesPrefixLongInt.toUnsignedBigInteger().intValueExact();
  public static final Bytes bytesPrefixShortList =
      bigIntegerToBytes(BigInteger.valueOf(TxnrlpTrace.list_short.intValue()));
  public static final int intPrefixShortList =
      bytesPrefixShortList.toUnsignedBigInteger().intValueExact();
  public static final Bytes bytesPrefixLongList =
      bigIntegerToBytes(BigInteger.valueOf(TxnrlpTrace.list_long.intValue()));
  public static final int intPrefixLongList =
      bytesPrefixLongList.toUnsignedBigInteger().intValueExact();

  private final StackedList<RlpTxnChunk> chunkList = new StackedList<>();

  // Used to check the reconstruction of RLPs
  Bytes reconstructedRlpLt;
  Bytes reconstructedRlpLx;

  @Override
  public String jsonKey() {
    return "rlpTxn";
  }

  @Override
  public void enterTransaction() {
    this.chunkList.enter();
  }

  @Override
  public void popTransaction() {
    this.chunkList.pop();
  }

  @Override
  public void traceStartTx(WorldView worldView, Transaction tx) {
    boolean requiresEvmExecution;
    if (tx.getTo().isEmpty()) {
      requiresEvmExecution = tx.getInit().isPresent();
    } else {
      requiresEvmExecution = worldView.get(tx.getTo().get()).hasCode();
    }
    this.chunkList.add(new RlpTxnChunk(tx, requiresEvmExecution));
  }

  public void traceChunk(RlpTxnChunk chunk, int absTxNum, int codeFragmentIndex) {

    // Create the local row storage and specify transaction constant columns
    RlpTxnColumnsValue traceValue = new RlpTxnColumnsValue();
    traceValue.DataHiLoReset();
    traceValue.ADDR_HI = bigIntegerToBytes(BigInteger.ZERO);
    traceValue.ADDR_LO = bigIntegerToBytes(BigInteger.ZERO);
    traceValue.absTxNum = absTxNum;
    traceValue.requiresEvmExecution = chunk.requireEvmExecution();
    traceValue.codeFragmentIndex = codeFragmentIndex;
    if (chunk.tx().getType() == TransactionType.FRONTIER) {
      traceValue.txType = 0;
    } else {
      traceValue.txType = chunk.tx().getType().getSerializedType();
    }

    // Initialise RLP_LT and RLP_LX byte size + verify that we construct the right RLP
    this.reconstructedRlpLt = Bytes.EMPTY;
    this.reconstructedRlpLx = Bytes.EMPTY;
    Bytes besuRlpLt =
        encodeOpaqueBytes((org.hyperledger.besu.ethereum.core.Transaction) chunk.tx(), BLOCK_BODY);
    // the encodeOpaqueBytes method already concatenate with the first byte "transaction  type"
    if (traceValue.txType == 0) {
      traceValue.RLP_LT_BYTESIZE = innerRlpSize(besuRlpLt.size());
    } else {
      traceValue.RLP_LT_BYTESIZE = innerRlpSize(besuRlpLt.size() - 1);
    }

    Bytes besuRlpLx = Bytes.EMPTY;
    switch (traceValue.txType) {
      case 0 -> {
        besuRlpLx =
            frontierPreimage(
                chunk.tx().getNonce(),
                (Wei) chunk.tx().getGasPrice().get(),
                chunk.tx().getGasLimit(),
                chunk.tx().getTo().map(x -> (Address) x),
                (Wei) chunk.tx().getValue(),
                chunk.tx().getPayload(),
                chunk.tx().getChainId());
        traceValue.RLP_LX_BYTESIZE = innerRlpSize(besuRlpLx.size());
      }
      case 1 -> {
        List<AccessListEntry> accessList = null;
        if (chunk.tx().getAccessList().isPresent()) {
          accessList = chunk.tx().getAccessList().get();
        }
        besuRlpLx =
            accessListPreimage(
                chunk.tx().getNonce(),
                (Wei) chunk.tx().getGasPrice().get(),
                chunk.tx().getGasLimit(),
                chunk.tx().getTo().map(x -> (Address) x),
                (Wei) chunk.tx().getValue(),
                chunk.tx().getPayload(),
                accessList,
                chunk.tx().getChainId());
        // the innerRlp method already concatenate with the first byte "transaction  type"
        traceValue.RLP_LX_BYTESIZE = innerRlpSize(besuRlpLx.size() - 1);
      }
      case 2 -> {
        besuRlpLx =
            eip1559Preimage(
                chunk.tx().getNonce(),
                (Wei) chunk.tx().getMaxPriorityFeePerGas().get(),
                (Wei) chunk.tx().getMaxFeePerGas().get(),
                chunk.tx().getGasLimit(),
                chunk.tx().getTo().map(x -> (Address) x),
                (Wei) chunk.tx().getValue(),
                chunk.tx().getPayload(),
                chunk.tx().getChainId(),
                chunk.tx().getAccessList());
        // the innerRlp method already concatenate with the first byte "transaction  type"
        traceValue.RLP_LX_BYTESIZE = innerRlpSize(besuRlpLx.size() - 1);
      }
      default -> throw new IllegalStateException(
          "Transaction Type not supported: " + traceValue.txType);
    }

    // Phase 0 : Global RLP prefix
    traceValue.DATA_LO = BigInteger.valueOf(traceValue.txType);
    handlePhaseGlobalRlpPrefix(traceValue);

    // Phase 1 : ChainId
    if (traceValue.txType == 1 || traceValue.txType == 2) {
      Preconditions.checkArgument(
          bigIntegerToBytes(chunk.tx().getChainId().get()).size() <= 8,
          "ChainId is longer than 8 bytes");
      handlePhaseInteger(traceValue, 1, chunk.tx().getChainId().get(), 8);
    }

    // Phase 2 : Nonce
    BigInteger nonce = BigInteger.valueOf(chunk.tx().getNonce());
    traceValue.DATA_LO = nonce;
    handlePhaseInteger(traceValue, 2, nonce, 8);

    // Phase 3 : GasPrice
    if (traceValue.txType == 0 || traceValue.txType == 1) {
      BigInteger gasPrice = chunk.tx().getGasPrice().get().getAsBigInteger();
      Preconditions.checkArgument(
          bigIntegerToBytes(gasPrice).size() <= 8, "GasPrice is longer than 8 bytes");
      traceValue.DATA_LO = gasPrice;
      handlePhaseInteger(traceValue, 3, gasPrice, 8);
    }

    // Phase 4 : max priority fee per gas (GasTipCap)
    if (traceValue.txType == 2) {
      BigInteger maxPriorityFeePerGas =
          chunk.tx().getMaxPriorityFeePerGas().get().getAsBigInteger();
      Preconditions.checkArgument(
          bigIntegerToBytes(maxPriorityFeePerGas).size() <= 8,
          "Max Priority Fee per Gas is longer than 8 bytes");
      handlePhaseInteger(traceValue, 4, maxPriorityFeePerGas, 8);
    }

    // Phase 5 : max fee per gas (GasFeeCap)
    if (traceValue.txType == 2) {
      traceValue.DATA_HI = chunk.tx().getMaxPriorityFeePerGas().get().getAsBigInteger();
      BigInteger maxFeePerGas = chunk.tx().getMaxFeePerGas().get().getAsBigInteger();
      Preconditions.checkArgument(
          bigIntegerToBytes(maxFeePerGas).size() <= 8, "Max Fee per Gas is longer than 8 bytes");
      traceValue.DATA_LO = maxFeePerGas;
      handlePhaseInteger(traceValue, 5, maxFeePerGas, 8);
    }

    // Phase 6 : GasLimit
    BigInteger gasLimit = BigInteger.valueOf(chunk.tx().getGasLimit());
    traceValue.DATA_LO = gasLimit;
    handlePhaseInteger(traceValue, 6, gasLimit, 8);

    // Phase 7 : To
    if (chunk.tx().getTo().isPresent()) {
      traceValue.DATA_HI = chunk.tx().getTo().get().slice(0, 4).toUnsignedBigInteger();
      traceValue.DATA_LO = chunk.tx().getTo().get().slice(4, 16).toUnsignedBigInteger();
    } else {
      traceValue.DATA_HI = BigInteger.ZERO;
      traceValue.DATA_LO = BigInteger.ZERO;
    }
    handlePhaseTo(traceValue, chunk.tx());

    // Phase 8 : Value
    BigInteger value = chunk.tx().getValue().getAsBigInteger();
    traceValue.DATA_LO = value;
    if (chunk.tx().getTo().isEmpty()) {
      traceValue.DATA_HI = BigInteger.ONE;
    } else {
      traceValue.DATA_HI = BigInteger.ZERO;
    }
    handlePhaseInteger(traceValue, 8, value, llarge);

    // Phase 9 : Data
    handlePhaseData(traceValue, chunk.tx());

    // Phase 10 : AccessList
    if (traceValue.txType == 1 || traceValue.txType == 2) {
      handlePhaseAccessList(traceValue, chunk.tx());
    }

    // Phase 11 : Beta / w
    if (traceValue.txType == 0) {
      handlePhaseBeta(traceValue, chunk.tx());
    }

    // Phase 12 : y
    if (traceValue.txType == 1 || traceValue.txType == 2) {
      handlePhaseY(traceValue, chunk.tx());
    }

    // Phase 13 : r
    handle32BytesInteger(traceValue, 13, chunk.tx().getR());

    // Phase 14 : s
    handle32BytesInteger(traceValue, 14, chunk.tx().getS());

    Preconditions.checkArgument(
        this.reconstructedRlpLt.equals(besuRlpLt), "Reconstructed RLP LT and Besu RLP LT differ");
    Preconditions.checkArgument(
        this.reconstructedRlpLx.equals(besuRlpLx), "Reconstructed RLP LX and Besu RLP LX differ");
  }

  // Define each phase's constraints
  private void handlePhaseGlobalRlpPrefix(RlpTxnColumnsValue traceValue) {
    int phase = 0;
    // First, trace the Type prefix of the transaction
    traceValue.partialReset(phase, 1, true, true);
    if (traceValue.txType != 0) {
      traceValue.LIMB_CONSTRUCTED = true;
      traceValue.LIMB = bigIntegerToBytes(BigInteger.valueOf(traceValue.txType));
      traceValue.nBYTES = 1;
      traceRow(traceValue);
    } else {
      traceValue.LC_CORRECTION = true;
      traceRow(traceValue);
    }

    // RLP prefix of RLP(LT)
    rlpByteString(
        0, traceValue.RLP_LT_BYTESIZE, true, true, false, false, false, false, false, traceValue);

    // RLP prefix of RLP(LT)
    rlpByteString(
        0, traceValue.RLP_LX_BYTESIZE, true, false, true, false, false, false, true, traceValue);
  }

  private void handlePhaseInteger(
      RlpTxnColumnsValue traceValue, int phase, BigInteger input, int nbstep) {
    if (input.equals(BigInteger.ZERO)) {
      traceZeroInt(traceValue, phase, true, true, false, true);
    } else {
      rlpInt(phase, input, nbstep, true, true, false, true, false, traceValue);
    }
  }

  private void handlePhaseTo(RlpTxnColumnsValue traceValue, Transaction tx) {
    int phase = 7;
    boolean lt = true;
    boolean lx = true;

    if (tx.getTo().isEmpty()) {
      traceZeroInt(traceValue, phase, lt, lx, false, true);
    } else {
      handleAddress(traceValue, phase, tx.getTo().get());
    }
  }

  private void handlePhaseData(RlpTxnColumnsValue traceValue, Transaction tx) {
    int phase = 9;
    boolean lt = true;
    boolean lx = true;

    if (tx.getPayload().isEmpty()) {
      // Trivial case
      traceZeroInt(traceValue, phase, lt, lx, true, false);

      // One row of padding
      traceValue.partialReset(phase, 1, lt, lx);
      traceValue.LC_CORRECTION = true;
      traceValue.PHASE_END = true;
      traceRow(traceValue);
    } else {
      // General case

      // Initialise DataSize and DataGasCost
      Bytes data = tx.getPayload();
      traceValue.PHASE_BYTESIZE = data.size();
      for (int i = 0; i < traceValue.PHASE_BYTESIZE; i++) {
        if (data.get(i) == 0) {
          traceValue.DATAGASCOST += TxnrlpTrace.G_txdatazero.intValue();
        } else {
          traceValue.DATAGASCOST += TxnrlpTrace.G_txdatanonzero.intValue();
        }
      }
      traceValue.DATA_HI = BigInteger.valueOf(traceValue.DATAGASCOST);
      traceValue.DATA_LO = BigInteger.valueOf(traceValue.PHASE_BYTESIZE);

      // Trace
      // RLP prefix
      if (traceValue.PHASE_BYTESIZE == 1) {
        rlpInt(
            phase,
            tx.getPayload().toUnsignedBigInteger(),
            8,
            lt,
            lx,
            true,
            false,
            true,
            traceValue);
      } else {
        // General case
        rlpByteString(
            phase, traceValue.PHASE_BYTESIZE, false, lt, lx, true, false, false, false, traceValue);
      }

      // Tracing the Data: several 16-rows ct-loop
      int nbstep = 16;
      int nbloop = (traceValue.PHASE_BYTESIZE - 1) / nbstep + 1;
      data = padToGivenSizeWithRightZero(data, nbstep * nbloop);
      for (int i = 0; i < nbloop; i++) {
        traceValue.partialReset(phase, nbstep, lt, lx);
        traceValue.INPUT_1 = data.slice(llarge * i, llarge);
        int accByteSize = 0;
        for (int ct = 0; ct < llarge; ct++) {
          traceValue.COUNTER = ct;
          if (traceValue.PHASE_BYTESIZE != 0) {
            accByteSize += 1;
          }
          traceValue.BYTE_1 = traceValue.INPUT_1.get(ct);
          traceValue.ACC_1 = traceValue.INPUT_1.slice(0, ct + 1);
          traceValue.ACC_BYTESIZE = accByteSize;
          if (ct == nbstep - 1) {
            traceValue.LIMB_CONSTRUCTED = true;
            traceValue.LIMB = traceValue.INPUT_1;
            traceValue.nBYTES = accByteSize;
          }
          traceRow(traceValue);
        }
      }
      // Two rows of padding
      traceValue.partialReset(phase, 2, lt, lx);
      traceValue.LC_CORRECTION = true;
      traceRow(traceValue);

      traceValue.COUNTER = 1;
      traceValue.PHASE_END = true;
      traceRow(traceValue);
    }

    // Put INDEX_DATA to 0 at the end of the phase
    traceValue.INDEX_DATA = 0;
  }

  private void handlePhaseAccessList(RlpTxnColumnsValue traceValue, Transaction tx) {
    int phase = 10;
    boolean lt = true;
    boolean lx = true;

    // Trivial case
    if (tx.getAccessList().get().isEmpty()) {
      traceVoidList(traceValue, phase, lt, lx, true, false, false, true);
    } else {
      // Initialise traceValue
      int nbAddr = 0;
      int nbSto = 0;
      List<Integer> nbStoPerAddrList = new ArrayList<>();
      List<Integer> accessTupleByteSizeList = new ArrayList<>();
      int phaseByteSize = 0;
      for (int i = 0; i < tx.getAccessList().get().size(); i++) {
        nbAddr += 1;
        nbSto += tx.getAccessList().get().get(i).storageKeys().size();
        nbStoPerAddrList.add(tx.getAccessList().get().get(i).storageKeys().size());
        accessTupleByteSizeList.add(
            21 + outerRlpSize(33 * tx.getAccessList().get().get(i).storageKeys().size()));
        phaseByteSize += outerRlpSize(accessTupleByteSizeList.get(i));
      }

      traceValue.partialReset(phase, 0, lt, lx);
      traceValue.nb_Addr = nbAddr;
      traceValue.DATA_LO = BigInteger.valueOf(nbAddr);
      traceValue.nb_Sto = nbSto;
      traceValue.DATA_HI = BigInteger.valueOf(nbSto);
      traceValue.PHASE_BYTESIZE = phaseByteSize;

      // Trace RLP(Phase Byte Size)
      rlpByteString(
          phase, traceValue.PHASE_BYTESIZE, true, lt, lx, true, false, false, false, traceValue);

      // Loop Over AccessTuple
      for (int i = 0; i < nbAddr; i++) {

        // Update columns at the beginning of an AccessTuple entry
        traceValue.nb_Addr -= 1;
        traceValue.nb_Sto_per_Addr = nbStoPerAddrList.get(i);
        traceValue.ADDR_HI = tx.getAccessList().get().get(i).address().slice(0, 4);
        traceValue.ADDR_LO = tx.getAccessList().get().get(i).address().slice(4, llarge);
        traceValue.ACCESS_TUPLE_BYTESIZE = accessTupleByteSizeList.get(i);

        // Rlp(AccessTupleByteSize)
        rlpByteString(
            phase,
            traceValue.ACCESS_TUPLE_BYTESIZE,
            true,
            lt,
            lx,
            true,
            true,
            false,
            false,
            traceValue);

        // RLP (address)
        handleAddress(traceValue, phase, tx.getAccessList().get().get(i).address());

        // Rlp prefix of the list of storage key
        if (nbStoPerAddrList.get(i) == 0) {
          traceVoidList(
              traceValue,
              phase,
              lt,
              lx,
              true,
              true,
              true,
              ((traceValue.nb_Sto == 0) && (traceValue.nb_Addr == 0)));
        } else {
          rlpByteString(
              phase,
              33L * traceValue.nb_Sto_per_Addr,
              true,
              lt,
              lx,
              true,
              true,
              true,
              false,
              traceValue);

          // Loop over StorageKey
          for (int j = 0; j < nbStoPerAddrList.get(i); j++) {
            traceValue.nb_Sto -= 1;
            traceValue.nb_Sto_per_Addr -= 1;
            handleStorageKey(
                traceValue,
                ((traceValue.nb_Sto == 0) && (traceValue.nb_Addr == 0)),
                tx.getAccessList().get().get(i).storageKeys().get(j));
          }
        }
        traceValue.ADDR_HI = bigIntegerToBytes(BigInteger.ZERO);
        traceValue.ADDR_LO = bigIntegerToBytes(BigInteger.ZERO);
      }
    }
  }

  private void handlePhaseBeta(RlpTxnColumnsValue traceValue, Transaction tx) {
    int phase = 11;
    BigInteger V = tx.getV();
    Preconditions.checkArgument(bigIntegerToBytes(V).size() <= 8, "V is longer than 8 bytes");

    // Rlp(w)
    boolean lt = true;
    boolean lx = false;
    rlpInt(phase, V, 8, lt, lx, false, false, false, traceValue);

    if (V.equals(BigInteger.valueOf(27)) || V.equals(BigInteger.valueOf(28))) {
      // One row of padding
      traceValue.partialReset(phase, 1, lt, lx);
      traceValue.IS_PREFIX = true;
      traceValue.LC_CORRECTION = true;
      traceValue.PHASE_END = true;
      traceRow(traceValue);
    } else {
      // RLP(ChainID) then one row with RLP().RLP()
      lt = false;
      lx = true;
      rlpInt(phase, tx.getChainId().get(), 8, lt, lx, true, false, false, traceValue);

      traceValue.partialReset(phase, 1, lt, lx);
      traceValue.LIMB_CONSTRUCTED = true;
      traceValue.LIMB = Bytes.concatenate(bytesPrefixShortInt, bytesPrefixShortInt);
      traceValue.nBYTES = 2;
      traceValue.PHASE_END = true;
      traceRow(traceValue);
    }
  }

  private void handlePhaseY(RlpTxnColumnsValue traceValue, Transaction tx) {
    traceValue.partialReset(12, 1, true, false);
    traceValue.INPUT_1 = bigIntegerToBytes(tx.getV());
    traceValue.LIMB_CONSTRUCTED = true;
    if (tx.getV().equals(BigInteger.ZERO)) {
      traceValue.LIMB = bytesPrefixShortInt;
    } else {
      traceValue.LIMB = bigIntegerToBytes(BigInteger.ONE);
    }
    traceValue.nBYTES = 1;
    traceValue.PHASE_END = true;
    traceRow(traceValue);
  }

  private void rlpByteString(
      int phase,
      long length,
      boolean isList,
      boolean lt,
      boolean lx,
      boolean isPrefix,
      boolean depth1,
      boolean depth2,
      boolean endPhase,
      RlpTxnColumnsValue traceValue) {
    int lengthSize = bigIntegerToBytes(BigInteger.valueOf(length)).size();

    ByteCountAndPowerOutput byteCountingOutput = byteCounting(lengthSize, 8);

    traceValue.partialReset(phase, 8, lt, lx);
    traceValue.INPUT_1 = bigIntegerToBytes(BigInteger.valueOf(length));
    traceValue.IS_PREFIX = isPrefix;
    traceValue.DEPTH_1 = depth1;
    traceValue.DEPTH_2 = depth2;

    Bytes input1RightShift = padToGivenSizeWithLeftZero(traceValue.INPUT_1, 8);

    long acc2LastRow;
    if (length >= 56) {
      acc2LastRow = length - 56;
    } else {
      acc2LastRow = 55 - length;
    }
    Bytes acc2LastRowShift =
        padToGivenSizeWithLeftZero(bigIntegerToBytes(BigInteger.valueOf(acc2LastRow)), 8);

    for (int ct = 0; ct < 8; ct++) {
      traceValue.COUNTER = ct;
      traceValue.ACC_BYTESIZE = byteCountingOutput.getAccByteSizeList().get(ct);
      traceValue.POWER = byteCountingOutput.getPowerList().get(ct);
      traceValue.BYTE_1 = input1RightShift.get(ct);
      traceValue.ACC_1 = input1RightShift.slice(0, ct + 1);
      traceValue.BYTE_2 = acc2LastRowShift.get(ct);
      traceValue.ACC_2 = acc2LastRowShift.slice(0, ct + 1);

      if (length >= 56) {
        if (ct == 6) {
          traceValue.LIMB_CONSTRUCTED = true;
          traceValue.nBYTES = 1;
          BigInteger tmp;
          if (isList) {
            tmp = BigInteger.valueOf(intPrefixLongList + lengthSize);
          } else {
            tmp = BigInteger.valueOf(intPrefixLongInt + lengthSize);
          }
          traceValue.LIMB = bigIntegerToBytes(tmp);
        }

        if (ct == 7) {
          traceValue.LIMB = traceValue.INPUT_1;
          traceValue.nBYTES = lengthSize;
          traceValue.BIT = true;
          traceValue.BIT_ACC = 1;
          traceValue.PHASE_END = endPhase;
        }
      } else {
        if (ct == 7) {
          traceValue.LIMB_CONSTRUCTED = true;
          Bytes tmp;
          if (isList) {
            tmp = bigIntegerToBytes(BigInteger.valueOf(intPrefixShortList + length));
          } else {
            tmp = bigIntegerToBytes(BigInteger.valueOf(intPrefixShortInt + length));
          }
          traceValue.LIMB = tmp;
          traceValue.nBYTES = 1;
          traceValue.PHASE_END = endPhase;
        }
      }
      traceRow(traceValue);
    }
  }

  private void rlpInt(
      int phase,
      BigInteger input,
      int nStep,
      boolean lt,
      boolean lx,
      boolean isPrefix,
      boolean endPhase,
      boolean onlyPrefix,
      RlpTxnColumnsValue traceValue) {

    traceValue.partialReset(phase, nStep, lt, lx);
    traceValue.IS_PREFIX = isPrefix;

    Bytes inputByte = bigIntegerToBytes(input);
    int inputSize = inputByte.size();
    ByteCountAndPowerOutput byteCountingOutput = byteCounting(inputSize, nStep);

    Bytes inputBytePadded = padToGivenSizeWithLeftZero(inputByte, nStep);
    BitDecOutput bitDecOutput =
        bitDecomposition(0xff & inputBytePadded.get(inputBytePadded.size() - 1), nStep);

    traceValue.INPUT_1 = inputByte;

    for (int ct = 0; ct < nStep; ct++) {
      traceValue.COUNTER = ct;
      traceValue.BYTE_1 = inputBytePadded.get(ct);
      traceValue.ACC_1 = inputBytePadded.slice(0, ct + 1);
      traceValue.POWER = byteCountingOutput.getPowerList().get(ct);
      traceValue.ACC_BYTESIZE = byteCountingOutput.getAccByteSizeList().get(ct);
      traceValue.BIT = bitDecOutput.getBitDecList().get(ct);
      traceValue.BIT_ACC = bitDecOutput.getBitAccList().get(ct);

      if (input.compareTo(BigInteger.valueOf(128)) >= 0 && ct == nStep - 2) {
        traceValue.LIMB_CONSTRUCTED = true;
        traceValue.LIMB = bigIntegerToBytes(BigInteger.valueOf(intPrefixShortInt + inputSize));
        traceValue.nBYTES = 1;
      }

      if (ct == nStep - 1) {
        if (onlyPrefix) {
          traceValue.LC_CORRECTION = true;
          traceValue.LIMB_CONSTRUCTED = false;
          traceValue.LIMB = Bytes.ofUnsignedShort(0);
          traceValue.nBYTES = 0;
        } else {
          traceValue.LIMB_CONSTRUCTED = true;
          traceValue.LIMB = inputByte;
          traceValue.nBYTES = inputSize;
          traceValue.PHASE_END = endPhase;
        }
      }
      traceRow(traceValue);
    }
  }

  private void handle32BytesInteger(RlpTxnColumnsValue traceValue, int phase, BigInteger input) {
    traceValue.partialReset(phase, llarge, true, false);
    if (input.equals(BigInteger.ZERO)) {
      // Trivial case
      traceZeroInt(traceValue, phase, true, false, false, true);
    } else {
      // General case
      Bytes inputByte = bigIntegerToBytes(input);
      int inputLen = inputByte.size();
      Bytes inputByte32 = padToGivenSizeWithLeftZero(inputByte, 32);
      traceValue.INPUT_1 = inputByte32.slice(0, llarge);
      traceValue.INPUT_2 = inputByte32.slice(llarge, llarge);

      ByteCountAndPowerOutput byteCounting;
      if (inputLen <= traceValue.nSTEP) {
        ByteCountAndPowerOutput byteCountingOutput = byteCounting(inputLen, traceValue.nSTEP);
        BitDecOutput bitDecOutput =
            bitDecomposition(inputByte.get(inputByte.size() - 1), traceValue.nSTEP);

        for (int ct = 0; ct < traceValue.nSTEP; ct++) {
          traceValue.COUNTER = ct;
          traceValue.BYTE_2 = traceValue.INPUT_2.get(ct);
          traceValue.ACC_2 = traceValue.INPUT_2.slice(0, ct + 1);
          traceValue.ACC_BYTESIZE = byteCountingOutput.getAccByteSizeList().get(ct);
          traceValue.POWER = byteCountingOutput.getPowerList().get(ct);
          traceValue.BIT = bitDecOutput.getBitDecList().get(ct);
          traceValue.BIT_ACC = bitDecOutput.getBitAccList().get(ct);

          // if input >= 128, there is a RLP prefix, nothing if 0 < input < 128
          if (ct == traceValue.nSTEP - 2 && input.compareTo(BigInteger.valueOf(128)) >= 0) {
            traceValue.LIMB_CONSTRUCTED = true;
            traceValue.LIMB = bigIntegerToBytes(BigInteger.valueOf(intPrefixShortInt + inputLen));
            traceValue.nBYTES = 1;
          }
          if (ct == traceValue.nSTEP - 1) {
            traceValue.LIMB_CONSTRUCTED = true;
            traceValue.LIMB = traceValue.INPUT_2.slice(llarge - inputLen, inputLen);
            traceValue.nBYTES = inputLen;
            traceValue.PHASE_END = true;
          }
          traceRow(traceValue);
        }
      } else {
        inputLen -= traceValue.nSTEP;
        byteCounting = byteCounting(inputLen, traceValue.nSTEP);

        for (int ct = 0; ct < traceValue.nSTEP; ct++) {
          traceValue.COUNTER = ct;
          traceValue.BYTE_1 = traceValue.INPUT_1.get(ct);
          traceValue.ACC_1 = traceValue.INPUT_1.slice(0, ct + 1);
          traceValue.BYTE_2 = traceValue.INPUT_2.get(ct);
          traceValue.ACC_2 = traceValue.INPUT_2.slice(0, ct + 1);
          traceValue.ACC_BYTESIZE = byteCounting.getAccByteSizeList().get(ct);
          traceValue.POWER = byteCounting.getPowerList().get(ct);

          if (ct == traceValue.nSTEP - 3) {
            traceValue.LIMB_CONSTRUCTED = true;
            traceValue.LIMB =
                bigIntegerToBytes(BigInteger.valueOf(intPrefixShortInt + llarge + inputLen));
            traceValue.nBYTES = 1;
          }
          if (ct == traceValue.nSTEP - 2) {
            traceValue.LIMB = traceValue.INPUT_1.slice(llarge - inputLen, inputLen);
            traceValue.nBYTES = inputLen;
          }
          if (ct == traceValue.nSTEP - 1) {
            traceValue.LIMB = traceValue.INPUT_2;
            traceValue.nBYTES = llarge;
            traceValue.PHASE_END = true;
          }
          traceRow(traceValue);
        }
      }
    }
  }

  private void handleAddress(RlpTxnColumnsValue traceValue, int phase, Address address) {
    boolean lt = true;
    boolean lx = true;
    traceValue.partialReset(phase, llarge, lt, lx);
    traceValue.INPUT_1 = padToGivenSizeWithLeftZero(address.slice(0, 4), llarge);
    traceValue.INPUT_2 = address.slice(4, llarge);

    if (phase == 10) {
      traceValue.DEPTH_1 = true;
    }

    for (int ct = 0; ct < traceValue.nSTEP; ct++) {
      traceValue.COUNTER = ct;
      traceValue.BYTE_1 = traceValue.INPUT_1.get(ct);
      traceValue.ACC_1 = traceValue.INPUT_1.slice(0, ct + 1);
      traceValue.BYTE_2 = traceValue.INPUT_2.get(ct);
      traceValue.ACC_2 = traceValue.INPUT_2.slice(0, ct + 1);

      if (ct == traceValue.nSTEP - 3) {
        traceValue.LIMB_CONSTRUCTED = true;
        traceValue.LIMB = bigIntegerToBytes(BigInteger.valueOf(intPrefixShortInt + 20));
        traceValue.nBYTES = 1;
      }

      if (ct == traceValue.nSTEP - 2) {
        traceValue.LIMB = address.slice(0, 4);
        traceValue.nBYTES = 4;
      }
      if (ct == traceValue.nSTEP - 1) {
        traceValue.LIMB = traceValue.INPUT_2;
        traceValue.nBYTES = llarge;

        if (phase == 7) {
          traceValue.PHASE_END = true;
        }
      }
      traceRow(traceValue);
    }
  }

  private void handleStorageKey(
      RlpTxnColumnsValue traceValue, boolean end_phase, Bytes32 storage_key) {
    traceValue.partialReset(10, llarge, true, true);
    traceValue.DEPTH_1 = true;
    traceValue.DEPTH_2 = true;
    traceValue.INPUT_1 = storage_key.slice(0, llarge);
    traceValue.INPUT_2 = storage_key.slice(llarge, llarge);

    for (int ct = 0; ct < traceValue.nSTEP; ct++) {
      traceValue.COUNTER = ct;
      traceValue.BYTE_1 = traceValue.INPUT_1.get(ct);
      traceValue.ACC_1 = traceValue.INPUT_1.slice(0, ct + 1);
      traceValue.BYTE_2 = traceValue.INPUT_2.get(ct);
      traceValue.ACC_2 = traceValue.INPUT_2.slice(0, ct + 1);

      if (ct == traceValue.nSTEP - 3) {
        traceValue.LIMB_CONSTRUCTED = true;
        traceValue.LIMB = bigIntegerToBytes(BigInteger.valueOf(intPrefixShortInt + 32));
        traceValue.nBYTES = 1;
      }

      if (ct == traceValue.nSTEP - 2) {
        traceValue.LIMB = traceValue.INPUT_1;
        traceValue.nBYTES = llarge;
      }

      if (ct == traceValue.nSTEP - 1) {
        traceValue.LIMB = traceValue.INPUT_2;
        traceValue.PHASE_END = end_phase;
      }

      traceRow(traceValue);
    }
  }

  public static int innerRlpSize(int rlpSize) {
    // If rlpSize >1, return size(something) where rlpSize = size(RLP(something)).
    Preconditions.checkArgument(rlpSize >= 2, "rlpSize should be at least 2 to get its inner size");
    int output = rlpSize;

    if (rlpSize < 57) {
      output -= 1;
    } else if (rlpSize == 57) {
      throw new RuntimeException("can't be of size 57");
    } else if (57 < rlpSize && rlpSize < 258) {
      output -= 2;
    } else if (rlpSize == 258) {
      throw new RuntimeException("can't be of size 258");
    } else {
      for (int i = 1; i < 8; i++) {
        if ((rlpSize - 2 - i >= Math.pow(2, 8 * i))
            && (rlpSize - i - 1 <= Math.pow(2, 8 * (i + 1)))) {
          output -= (2 + i);
        } else if (rlpSize == Math.pow(2, i) + 1 + i) {
          throw new RuntimeException("can't be this size");
        }
      }
    }
    return output;
  }

  private static Bytes frontierPreimage(
      final long nonce,
      final Wei gasPrice,
      final long gasLimit,
      final Optional<Address> to,
      final Wei value,
      final Bytes payload,
      final Optional<BigInteger> chainId) {
    return RLP.encode(
        rlpOutput -> {
          rlpOutput.startList();
          rlpOutput.writeLongScalar(nonce);
          rlpOutput.writeUInt256Scalar(gasPrice);
          rlpOutput.writeLongScalar(gasLimit);
          rlpOutput.writeBytes(to.map(Bytes::copy).orElse(Bytes.EMPTY));
          rlpOutput.writeUInt256Scalar(value);
          rlpOutput.writeBytes(payload);
          if (chainId.isPresent()) {
            rlpOutput.writeBigIntegerScalar(chainId.get());
            rlpOutput.writeUInt256Scalar(UInt256.ZERO);
            rlpOutput.writeUInt256Scalar(UInt256.ZERO);
          }
          rlpOutput.endList();
        });
  }

  private static Bytes accessListPreimage(
      final long nonce,
      final Wei gasPrice,
      final long gasLimit,
      final Optional<Address> to,
      final Wei value,
      final Bytes payload,
      final List<AccessListEntry> accessList,
      final Optional<BigInteger> chainId) {
    final Bytes encode =
        RLP.encode(
            rlpOutput -> {
              rlpOutput.startList();
              AccessListTransactionEncoder.encodeAccessListInner(
                  chainId, nonce, gasPrice, gasLimit, to, value, payload, accessList, rlpOutput);
              rlpOutput.endList();
            });
    return Bytes.concatenate(Bytes.of(TransactionType.ACCESS_LIST.getSerializedType()), encode);
  }

  private static Bytes eip1559Preimage(
      final long nonce,
      final Wei maxPriorityFeePerGas,
      final Wei maxFeePerGas,
      final long gasLimit,
      final Optional<Address> to,
      final Wei value,
      final Bytes payload,
      final Optional<BigInteger> chainId,
      final Optional<List<AccessListEntry>> accessList) {
    final Bytes encoded =
        RLP.encode(
            rlpOutput -> {
              rlpOutput.startList();
              eip1559PreimageFields(
                  nonce,
                  maxPriorityFeePerGas,
                  maxFeePerGas,
                  gasLimit,
                  to,
                  value,
                  payload,
                  chainId,
                  accessList,
                  rlpOutput);
              rlpOutput.endList();
            });
    return Bytes.concatenate(Bytes.of(TransactionType.EIP1559.getSerializedType()), encoded);
  }

  private static void eip1559PreimageFields(
      final long nonce,
      final Wei maxPriorityFeePerGas,
      final Wei maxFeePerGas,
      final long gasLimit,
      final Optional<Address> to,
      final Wei value,
      final Bytes payload,
      final Optional<BigInteger> chainId,
      final Optional<List<AccessListEntry>> accessList,
      final RLPOutput rlpOutput) {
    rlpOutput.writeBigIntegerScalar(chainId.orElseThrow());
    rlpOutput.writeLongScalar(nonce);
    rlpOutput.writeUInt256Scalar(maxPriorityFeePerGas);
    rlpOutput.writeUInt256Scalar(maxFeePerGas);
    rlpOutput.writeLongScalar(gasLimit);
    rlpOutput.writeBytes(to.map(Bytes::copy).orElse(Bytes.EMPTY));
    rlpOutput.writeUInt256Scalar(value);
    rlpOutput.writeBytes(payload);
    AccessListTransactionEncoder.writeAccessList(rlpOutput, accessList);
  }

  private void traceZeroInt(
      RlpTxnColumnsValue traceValue,
      int phase,
      boolean lt,
      boolean lx,
      boolean isPrefix,
      boolean phaseEnd) {
    traceValue.partialReset(phase, 1, lt, lx);
    traceValue.LIMB_CONSTRUCTED = true;
    traceValue.LIMB = bytesPrefixShortInt;
    traceValue.nBYTES = 1;
    traceValue.IS_PREFIX = true;
    traceValue.PHASE_END = phaseEnd;
    traceRow(traceValue);
  }

  private void traceVoidList(
      RlpTxnColumnsValue traceValue,
      int phase,
      boolean lt,
      boolean lx,
      boolean isPrefix,
      boolean depth1,
      boolean depth2,
      boolean phaseEnd) {
    traceValue.partialReset(phase, 1, lt, lx);
    traceValue.LIMB_CONSTRUCTED = true;
    traceValue.LIMB = bytesPrefixShortList;
    traceValue.nBYTES = 1;
    traceValue.IS_PREFIX = isPrefix;
    traceValue.DEPTH_1 = depth1;
    traceValue.DEPTH_2 = depth2;
    traceValue.PHASE_END = phaseEnd;
    traceRow(traceValue);
  }
  // Define the Tracer
  private void traceRow(RlpTxnColumnsValue traceValue) {
    // Decrements RLP_BYTESIZE
    if (traceValue.phase != 0) {
      if (traceValue.LIMB_CONSTRUCTED && traceValue.LT) {
        traceValue.RLP_LT_BYTESIZE -= traceValue.nBYTES;
      }
      if (traceValue.LIMB_CONSTRUCTED && traceValue.LX) {
        traceValue.RLP_LX_BYTESIZE -= traceValue.nBYTES;
      }
    }

    // Decrement phaseByteSize and accessTupleByteSize for Phase 10 (AccessList)
    if (traceValue.phase == 10) {
      // Decreases PhaseByteSize
      if (traceValue.DEPTH_1 && traceValue.LIMB_CONSTRUCTED) {
        traceValue.PHASE_BYTESIZE -= traceValue.nBYTES;
      }
      // Decreases AccessTupleSize
      if (traceValue.DEPTH_1
          && !(traceValue.IS_PREFIX && !traceValue.DEPTH_2)
          && traceValue.LIMB_CONSTRUCTED) {
        traceValue.ACCESS_TUPLE_BYTESIZE -= traceValue.nBYTES;
      }
    }

    this.builder
        .absTxNum(BigInteger.valueOf(traceValue.absTxNum))
        .absTxNumInfiny(BigInteger.valueOf(this.chunkList.size()))
        .acc1(traceValue.ACC_1.toUnsignedBigInteger())
        .acc2(traceValue.ACC_2.toUnsignedBigInteger())
        .accBytesize(BigInteger.valueOf(traceValue.ACC_BYTESIZE))
        .accessTupleBytesize(BigInteger.valueOf(traceValue.ACCESS_TUPLE_BYTESIZE))
        .addrHi(traceValue.ADDR_HI.toUnsignedBigInteger())
        .addrLo(traceValue.ADDR_LO.toUnsignedBigInteger())
        .bit(traceValue.BIT)
        .bitAcc(UnsignedByte.of(traceValue.BIT_ACC))
        .byte1(UnsignedByte.of(traceValue.BYTE_1))
        .byte2(UnsignedByte.of(traceValue.BYTE_2))
        .codeFragmentIndex(BigInteger.valueOf(traceValue.codeFragmentIndex))
        .counter(UnsignedByte.of(traceValue.COUNTER))
        .dataHi(traceValue.DATA_HI)
        .dataLo(traceValue.DATA_LO)
        .datagascost(BigInteger.valueOf(traceValue.DATAGASCOST))
        .depth1(traceValue.DEPTH_1)
        .depth2(traceValue.DEPTH_2);
    if (traceValue.COUNTER == traceValue.nSTEP - 1) {
      this.builder.done(Boolean.TRUE);
    } else {
      this.builder.done(Boolean.FALSE);
    }
    this.builder
        .phaseEnd(traceValue.PHASE_END)
        .indexData(BigInteger.valueOf(traceValue.INDEX_DATA))
        .indexLt(BigInteger.valueOf(traceValue.INDEX_LT))
        .indexLx(BigInteger.valueOf(traceValue.INDEX_LX))
        .input1(traceValue.INPUT_1.toUnsignedBigInteger())
        .input2(traceValue.INPUT_2.toUnsignedBigInteger())
        .lcCorrection(traceValue.LC_CORRECTION)
        .isPrefix(traceValue.IS_PREFIX)
        .limb(padToGivenSizeWithRightZero(traceValue.LIMB, llarge).toUnsignedBigInteger())
        .limbConstructed(traceValue.LIMB_CONSTRUCTED)
        .lt(traceValue.LT)
        .lx(traceValue.LX)
        .nBytes(UnsignedByte.of(traceValue.nBYTES))
        .nAddr(BigInteger.valueOf(traceValue.nb_Addr))
        .nKeys(BigInteger.valueOf(traceValue.nb_Sto))
        .nKeysPerAddr(BigInteger.valueOf(traceValue.nb_Sto_per_Addr))
        .nStep(UnsignedByte.of(traceValue.nSTEP));
    List<Function<Boolean, Trace.TraceBuilder>> phaseColumns =
        List.of(
            this.builder::phase0,
            this.builder::phase1,
            this.builder::phase2,
            this.builder::phase3,
            this.builder::phase4,
            this.builder::phase5,
            this.builder::phase6,
            this.builder::phase7,
            this.builder::phase8,
            this.builder::phase9,
            this.builder::phase10,
            this.builder::phase11,
            this.builder::phase12,
            this.builder::phase13,
            this.builder::phase14);
    for (int i = 0; i < phaseColumns.size(); i++) {
      phaseColumns.get(i).apply(i == traceValue.phase);
    }
    this.builder
        .phaseSize(BigInteger.valueOf(traceValue.PHASE_BYTESIZE))
        .power(traceValue.POWER)
        .requiresEvmExecution(traceValue.requiresEvmExecution)
        .rlpLtBytesize(BigInteger.valueOf(traceValue.RLP_LT_BYTESIZE))
        .rlpLxBytesize(BigInteger.valueOf(traceValue.RLP_LX_BYTESIZE))
        .type(UnsignedByte.of(traceValue.txType));

    // Increments Index
    if (traceValue.LIMB_CONSTRUCTED && traceValue.LT) {
      traceValue.INDEX_LT += 1;
    }
    if (traceValue.LIMB_CONSTRUCTED && traceValue.LX) {
      traceValue.INDEX_LX += 1;
    }

    // Increments IndexData (Phase 9)
    if (traceValue.phase == 9
        && !traceValue.IS_PREFIX
        && (traceValue.LIMB_CONSTRUCTED || traceValue.LC_CORRECTION)) {
      traceValue.INDEX_DATA += 1;
    }

    // Decrements PhaseByteSize and DataGasCost in Data phase (phase 9)
    if (traceValue.phase == 9) {
      if (traceValue.PHASE_BYTESIZE != 0 && !traceValue.IS_PREFIX) {
        traceValue.PHASE_BYTESIZE -= 1;
        if (traceValue.BYTE_1 == 0) {
          traceValue.DATAGASCOST -= TxnrlpTrace.G_txdatazero.intValue();
        } else {
          traceValue.DATAGASCOST -= TxnrlpTrace.G_txdatanonzero.intValue();
        }
      }
    }
    if (traceValue.PHASE_END) {
      traceValue.DataHiLoReset();
    }
    this.builder.validateRow();

    // reconstruct RLPs
    if (traceValue.LIMB_CONSTRUCTED && traceValue.LT) {
      this.reconstructedRlpLt =
          Bytes.concatenate(this.reconstructedRlpLt, traceValue.LIMB.slice(0, traceValue.nBYTES));
    }
    if (traceValue.LIMB_CONSTRUCTED && traceValue.LX) {
      this.reconstructedRlpLx =
          Bytes.concatenate(this.reconstructedRlpLx, traceValue.LIMB.slice(0, traceValue.nBYTES));
    }
  }

  private int ChunkRowSize(RlpTxnChunk chunk) {
    int txType;
    if (chunk.tx().getType() == TransactionType.FRONTIER) {
      txType = 0;
    } else {
      txType = chunk.tx().getType().getSerializedType();
    }
    // Phase 0 is always 17 rows long
    int rowSize = 17;

    // Phase 1: chainID
    if (txType == 1 || txType == 2) {
      if (chunk.tx().getChainId().get().equals(BigInteger.ZERO)) {
        rowSize += 1;
      } else {
        rowSize += 8;
      }
    }

    // Phase 2: nonce
    if (chunk.tx().getNonce() == 0) {
      rowSize += 1;
    } else {
      rowSize += 8;
    }

    // Phase 3: gasPrice
    if (txType == 0 || txType == 1) {
      rowSize += 8;
    }

    // Phase 4: MaxPriorityFeeperGas
    if (txType == 2) {
      if (chunk.tx().getMaxPriorityFeePerGas().get().getAsBigInteger().equals(BigInteger.ZERO)) {
        rowSize += 1;
      } else {
        rowSize += 8;
      }
    }

    // Phase 5: MaxFeePerGas
    if (txType == 2) {
      if (chunk.tx().getMaxFeePerGas().get().getAsBigInteger().equals(BigInteger.ZERO)) {
        rowSize += 1;
      } else {
        rowSize += 8;
      }
    }

    // Phase 6: GasLimit
    rowSize += 8;

    // Phase 7: To
    if (chunk.tx().getTo().isPresent()) {
      rowSize += 16;
    } else {
      rowSize += 1;
    }

    // Phase 8: Value
    if (chunk.tx().getValue().getAsBigInteger().equals(BigInteger.ZERO)) {
      rowSize += 1;
    } else {
      rowSize += 16;
    }

    // Phase 9: Data
    if (chunk.tx().getPayload().isEmpty()) {
      rowSize += 2; // 1 for prefix + 1 for padding
    } else {
      int dataSize = chunk.tx().getPayload().size();
      rowSize += 8 + llarge * ((dataSize - 1) / llarge + 1);
      rowSize += 2; // 2 lines of padding
    }

    // Phase 10: AccessList
    if (txType == 1 || txType == 2) {
      if (chunk.tx().getAccessList().get().isEmpty()) {
        rowSize += 1;
      } else {
        // Rlp prefix of the AccessList list
        rowSize += 8;
        for (int i = 0; i < chunk.tx().getAccessList().get().size(); i++) {
          rowSize += 8 + 16;
          if (chunk.tx().getAccessList().get().get(i).storageKeys().isEmpty()) {
            rowSize += 1;
          } else {
            rowSize += 8 + 16 * chunk.tx().getAccessList().get().get(i).storageKeys().size();
          }
        }
      }
    }

    // Phase 11: beta
    if (txType == 0) {
      rowSize += 8;
      if (chunk.tx().getChainId().get().equals(BigInteger.ZERO)) {
        rowSize += 1;
      } else {
        rowSize += 9;
      }
    }

    // Phase 12: y
    if (txType == 1 || txType == 2) {
      rowSize += 1;
    }

    // Phase 13: r
    if (chunk.tx().getR().equals(BigInteger.ZERO)) {
      rowSize += 1;
    } else {
      rowSize += 16;
    }

    // Phase 14: s
    if (chunk.tx().getS().equals(BigInteger.ZERO)) {
      rowSize += 1;
    } else {
      rowSize += 16;
    }
    return rowSize;
  }

  @Override
  public int lineCount() {
    int traceRowSize = 0;
    for (RlpTxnChunk chunk : this.chunkList) {
      traceRowSize += ChunkRowSize(chunk);
    }
    return traceRowSize;
  }

  @Override
  public Object commit() {
    int estTraceSize = 0;
    int absTxNum = 0;
    for (RlpTxnChunk chunk : this.chunkList) {
      absTxNum += 1;
      // TODO: recuperer les codeFragmentIndex ici
      int codeFragmentIndex = 0;
      traceChunk(chunk, absTxNum, codeFragmentIndex);

      estTraceSize += ChunkRowSize(chunk);
      if (this.builder.size() != estTraceSize) {
        throw new RuntimeException(
            "ChunkSize is not the right one, chunk n°: "
                + absTxNum
                + " estimated size ="
                + estTraceSize
                + " trace size ="
                + this.builder.size());
      }
    }
    return new RlpTxnTrace(builder.build());
  }
}
