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

package net.consensys.linea.zktracer.module.rlp.txn;

import static net.consensys.linea.zktracer.module.Util.getTxTypeAsInt;
import static net.consensys.linea.zktracer.module.rlputils.Pattern.byteCounting;
import static net.consensys.linea.zktracer.module.rlputils.Pattern.innerRlpSize;
import static net.consensys.linea.zktracer.module.rlputils.Pattern.outerRlpSize;
import static net.consensys.linea.zktracer.types.Conversions.bigIntegerToBytes;
import static net.consensys.linea.zktracer.types.Conversions.longToUnsignedBigInteger;
import static net.consensys.linea.zktracer.types.Utils.bitDecomposition;
import static net.consensys.linea.zktracer.types.Utils.leftPadTo;
import static net.consensys.linea.zktracer.types.Utils.rightPadTo;
import static org.hyperledger.besu.ethereum.core.encoding.EncodingContext.BLOCK_BODY;
import static org.hyperledger.besu.ethereum.core.encoding.TransactionEncoder.encodeOpaqueBytes;

import java.math.BigInteger;
import java.nio.MappedByteBuffer;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import java.util.function.Function;

import com.google.common.base.Preconditions;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.stacked.list.StackedList;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.rlputils.ByteCountAndPowerOutput;
import net.consensys.linea.zktracer.module.romLex.RomLex;
import net.consensys.linea.zktracer.types.BitDecOutput;
import net.consensys.linea.zktracer.types.UnsignedByte;
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
import org.hyperledger.besu.evm.account.AccountState;
import org.hyperledger.besu.evm.worldstate.WorldView;

public class RlpTxn implements Module {
  private final RomLex romLex;

  public RlpTxn(RomLex _romLex) {
    this.romLex = _romLex;
  }

  @Override
  public String moduleKey() {
    return "TX_RLP";
  }

  public static final int LLARGE = TxnrlpTrace.LLARGE.intValue();
  public static final Bytes BYTES_PREFIX_SHORT_INT =
      bigIntegerToBytes(BigInteger.valueOf(TxnrlpTrace.int_short.intValue()));
  public static final int INT_PREFIX_SHORT_INT =
      BYTES_PREFIX_SHORT_INT.toUnsignedBigInteger().intValueExact();
  public static final Bytes BYTES_PREFIX_LONG_INT =
      bigIntegerToBytes(BigInteger.valueOf(TxnrlpTrace.int_long.intValue()));
  public static final int INT_PREFIX_LONG_INT =
      BYTES_PREFIX_LONG_INT.toUnsignedBigInteger().intValueExact();
  public static final Bytes BYTES_PREFIX_SHORT_LIST =
      bigIntegerToBytes(BigInteger.valueOf(TxnrlpTrace.list_short.intValue()));
  public static final int INT_PREFIX_SHORT_LIST =
      BYTES_PREFIX_SHORT_LIST.toUnsignedBigInteger().intValueExact();
  public static final Bytes BYTES_PREFIX_LONG_LIST =
      bigIntegerToBytes(BigInteger.valueOf(TxnrlpTrace.list_long.intValue()));

  public static final int INT_PREFIX_LONG_LIST =
      BYTES_PREFIX_LONG_LIST.toUnsignedBigInteger().intValueExact();

  private final StackedList<RlpTxnChunk> chunkList = new StackedList<>();

  // Used to check the reconstruction of RLPs
  Bytes reconstructedRlpLt;

  Bytes reconstructedRlpLx;

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
    // Contract Creation
    if (tx.getTo().isEmpty() && !tx.getInit().orElseThrow().isEmpty()) {
      this.chunkList.add(new RlpTxnChunk(tx, true, romLex.codeIdentifierBeforeLexOrder));
    }

    // Call to a non-empty smart contract
    else if (tx.getTo().isPresent()
        && Optional.ofNullable(worldView.get(tx.getTo().orElseThrow()))
            .map(AccountState::hasCode)
            .orElse(false)) {
      this.chunkList.add(new RlpTxnChunk(tx, true));
    } else {
      // Contract doesn't require EVM execution
      this.chunkList.add(new RlpTxnChunk(tx, false));
    }
  }

  public void traceChunk(RlpTxnChunk chunk, int absTxNum, int codeFragmentIndex, Trace trace) {

    // Create the local row storage and specify transaction constant columns
    RlpTxnColumnsValue traceValue = new RlpTxnColumnsValue();
    traceValue.resetDataHiLo();
    traceValue.addrHi = bigIntegerToBytes(BigInteger.ZERO);
    traceValue.addrLo = bigIntegerToBytes(BigInteger.ZERO);
    traceValue.absTxNum = absTxNum;
    traceValue.requiresEvmExecution = chunk.requireEvmExecution();
    traceValue.codeFragmentIndex = codeFragmentIndex;
    traceValue.txType = getTxTypeAsInt(chunk.tx().getType());

    // Initialise RLP_LT and RLP_LX byte size + verify that we construct the right RLP
    this.reconstructedRlpLt = Bytes.EMPTY;
    this.reconstructedRlpLx = Bytes.EMPTY;
    Bytes besuRlpLt =
        encodeOpaqueBytes((org.hyperledger.besu.ethereum.core.Transaction) chunk.tx(), BLOCK_BODY);
    // the encodeOpaqueBytes method already concatenate with the first byte "transaction  type"
    if (traceValue.txType == 0) {
      traceValue.rlpLtByteSize = innerRlpSize(besuRlpLt.size());
    } else {
      traceValue.rlpLtByteSize = innerRlpSize(besuRlpLt.size() - 1);
    }

    Bytes besuRlpLx;
    switch (traceValue.txType) {
      case 0 -> {
        besuRlpLx =
            frontierPreimage(
                chunk.tx().getNonce(),
                (Wei) chunk.tx().getGasPrice().orElseThrow(),
                chunk.tx().getGasLimit(),
                chunk.tx().getTo().map(x -> (Address) x),
                (Wei) chunk.tx().getValue(),
                chunk.tx().getPayload(),
                chunk.tx().getChainId());
        traceValue.rlpLxByteSize = innerRlpSize(besuRlpLx.size());
      }
      case 1 -> {
        List<AccessListEntry> accessList = null;
        if (chunk.tx().getAccessList().isPresent()) {
          accessList = chunk.tx().getAccessList().orElseThrow();
        }
        besuRlpLx =
            accessListPreimage(
                chunk.tx().getNonce(),
                (Wei) chunk.tx().getGasPrice().orElseThrow(),
                chunk.tx().getGasLimit(),
                chunk.tx().getTo().map(x -> (Address) x),
                (Wei) chunk.tx().getValue(),
                chunk.tx().getPayload(),
                accessList,
                chunk.tx().getChainId());
        // the innerRlp method already concatenate with the first byte "transaction  type"
        traceValue.rlpLxByteSize = innerRlpSize(besuRlpLx.size() - 1);
      }
      case 2 -> {
        besuRlpLx =
            eip1559Preimage(
                chunk.tx().getNonce(),
                (Wei) chunk.tx().getMaxPriorityFeePerGas().orElseThrow(),
                (Wei) chunk.tx().getMaxFeePerGas().orElseThrow(),
                chunk.tx().getGasLimit(),
                chunk.tx().getTo().map(x -> (Address) x),
                (Wei) chunk.tx().getValue(),
                chunk.tx().getPayload(),
                chunk.tx().getChainId(),
                chunk.tx().getAccessList());
        // the innerRlp method already concatenate with the first byte "transaction  type"
        traceValue.rlpLxByteSize = innerRlpSize(besuRlpLx.size() - 1);
      }
      default -> throw new IllegalStateException(
          "Transaction Type not supported: " + traceValue.txType);
    }

    // Phase 0 : Global RLP prefix
    traceValue.dataLo = BigInteger.valueOf(traceValue.txType);
    handlePhaseGlobalRlpPrefix(traceValue, trace);

    // Phase 1 : ChainId
    if (traceValue.txType == 1 || traceValue.txType == 2) {
      Preconditions.checkArgument(
          bigIntegerToBytes(chunk.tx().getChainId().orElseThrow()).size() <= 8,
          "ChainId is longer than 8 bytes");
      handlePhaseInteger(traceValue, 1, chunk.tx().getChainId().get(), 8, trace);
    }

    // Phase 2 : Nonce
    BigInteger nonce = longToUnsignedBigInteger(chunk.tx().getNonce());
    traceValue.dataLo = nonce;
    handlePhaseInteger(traceValue, 2, nonce, 8, trace);

    // Phase 3 : GasPrice
    if (traceValue.txType == 0 || traceValue.txType == 1) {
      BigInteger gasPrice = chunk.tx().getGasPrice().orElseThrow().getAsBigInteger();
      Preconditions.checkArgument(
          bigIntegerToBytes(gasPrice).size() <= 8, "GasPrice is longer than 8 bytes");
      traceValue.dataLo = gasPrice;
      handlePhaseInteger(traceValue, 3, gasPrice, 8, trace);
    }

    // Phase 4 : max priority fee per gas (GasTipCap)
    if (traceValue.txType == 2) {
      BigInteger maxPriorityFeePerGas =
          chunk.tx().getMaxPriorityFeePerGas().orElseThrow().getAsBigInteger();
      Preconditions.checkArgument(
          bigIntegerToBytes(maxPriorityFeePerGas).size() <= 8,
          "Max Priority Fee per Gas is longer than 8 bytes");
      handlePhaseInteger(traceValue, 4, maxPriorityFeePerGas, 8, trace);
    }

    // Phase 5 : max fee per gas (GasFeeCap)
    if (traceValue.txType == 2) {
      traceValue.dataHi = chunk.tx().getMaxPriorityFeePerGas().orElseThrow().getAsBigInteger();
      BigInteger maxFeePerGas = chunk.tx().getMaxFeePerGas().orElseThrow().getAsBigInteger();
      Preconditions.checkArgument(
          bigIntegerToBytes(maxFeePerGas).size() <= 8, "Max Fee per Gas is longer than 8 bytes");
      traceValue.dataLo = maxFeePerGas;
      handlePhaseInteger(traceValue, 5, maxFeePerGas, 8, trace);
    }

    // Phase 6 : GasLimit
    BigInteger gasLimit = BigInteger.valueOf(chunk.tx().getGasLimit());
    traceValue.dataLo = gasLimit;
    handlePhaseInteger(traceValue, 6, gasLimit, 8, trace);

    // Phase 7 : To
    if (chunk.tx().getTo().isPresent()) {
      traceValue.dataHi = chunk.tx().getTo().orElseThrow().slice(0, 4).toUnsignedBigInteger();
      traceValue.dataLo = chunk.tx().getTo().orElseThrow().slice(4, 16).toUnsignedBigInteger();
    } else {
      traceValue.dataHi = BigInteger.ZERO;
      traceValue.dataLo = BigInteger.ZERO;
    }
    handlePhaseTo(traceValue, chunk.tx(), trace);

    // Phase 8 : Value
    BigInteger value = chunk.tx().getValue().getAsBigInteger();
    traceValue.dataLo = value;
    if (chunk.tx().getTo().isEmpty()) {
      traceValue.dataHi = BigInteger.ONE;
    } else {
      traceValue.dataHi = BigInteger.ZERO;
    }
    handlePhaseInteger(traceValue, 8, value, LLARGE, trace);

    // Phase 9 : Data
    handlePhaseData(traceValue, chunk.tx(), trace);

    // Phase 10 : AccessList
    if (traceValue.txType == 1 || traceValue.txType == 2) {
      handlePhaseAccessList(traceValue, chunk.tx(), trace);
    }

    // Phase 11 : Beta / w
    if (traceValue.txType == 0) {
      handlePhaseBeta(traceValue, chunk.tx(), trace);
    }

    // Phase 12 : y
    if (traceValue.txType == 1 || traceValue.txType == 2) {
      handlePhaseY(traceValue, chunk.tx(), trace);
    }

    // Phase 13 : r
    handle32BytesInteger(traceValue, 13, chunk.tx().getR(), trace);

    // Phase 14 : s
    handle32BytesInteger(traceValue, 14, chunk.tx().getS(), trace);

    Preconditions.checkArgument(
        this.reconstructedRlpLt.equals(besuRlpLt), "Reconstructed RLP LT and Besu RLP LT differ");
    Preconditions.checkArgument(
        this.reconstructedRlpLx.equals(besuRlpLx), "Reconstructed RLP LX and Besu RLP LX differ");
  }

  // Define each phase's constraints
  private void handlePhaseGlobalRlpPrefix(RlpTxnColumnsValue traceValue, Trace trace) {
    int phase = 0;
    // First, trace the Type prefix of the transaction
    traceValue.partialReset(phase, 1, true, true);
    if (traceValue.txType != 0) {
      traceValue.limbConstructed = true;
      traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(traceValue.txType));
      traceValue.nBytes = 1;
      traceRow(traceValue, trace);
    } else {
      traceValue.lcCorrection = true;
      traceRow(traceValue, trace);
    }

    // RLP prefix of RLP(LT)
    rlpByteString(
        0,
        traceValue.rlpLtByteSize,
        true,
        true,
        false,
        false,
        false,
        false,
        false,
        traceValue,
        trace);

    // RLP prefix of RLP(LT)
    rlpByteString(
        0,
        traceValue.rlpLxByteSize,
        true,
        false,
        true,
        false,
        false,
        false,
        true,
        traceValue,
        trace);
  }

  private void handlePhaseInteger(
      RlpTxnColumnsValue traceValue, int phase, BigInteger input, int nbstep, Trace trace) {
    if (input.equals(BigInteger.ZERO)) {
      traceZeroInt(traceValue, phase, true, true, false, true, trace);
    } else {
      rlpInt(phase, input, nbstep, true, true, false, true, false, traceValue, trace);
    }
  }

  private void handlePhaseTo(RlpTxnColumnsValue traceValue, Transaction tx, Trace trace) {
    int phase = 7;
    boolean lt = true;
    boolean lx = true;

    if (tx.getTo().isEmpty()) {
      traceZeroInt(traceValue, phase, lt, lx, false, true, trace);
    } else {
      handleAddress(traceValue, phase, tx.getTo().get(), trace);
    }
  }

  private void handlePhaseData(RlpTxnColumnsValue traceValue, Transaction tx, Trace trace) {
    int phase = 9;
    boolean lt = true;
    boolean lx = true;

    if (tx.getPayload().isEmpty()) {
      // Trivial case
      traceZeroInt(traceValue, phase, lt, lx, true, false, trace);

      // One row of padding
      traceValue.partialReset(phase, 1, lt, lx);
      traceValue.lcCorrection = true;
      traceValue.phaseEnd = true;
      traceRow(traceValue, trace);
    } else {
      // General case

      // Initialise DataSize and DataGasCost
      Bytes data = tx.getPayload();
      traceValue.phaseByteSize = data.size();
      for (int i = 0; i < traceValue.phaseByteSize; i++) {
        if (data.get(i) == 0) {
          traceValue.dataGasCost += TxnrlpTrace.G_txdatazero.intValue();
        } else {
          traceValue.dataGasCost += TxnrlpTrace.G_txdatanonzero.intValue();
        }
      }
      traceValue.dataHi = BigInteger.valueOf(traceValue.dataGasCost);
      traceValue.dataLo = BigInteger.valueOf(traceValue.phaseByteSize);

      // Trace
      // RLP prefix
      if (traceValue.phaseByteSize == 1) {
        rlpInt(
            phase,
            tx.getPayload().toUnsignedBigInteger(),
            8,
            lt,
            lx,
            true,
            false,
            true,
            traceValue,
            trace);
      } else {
        // General case
        rlpByteString(
            phase,
            traceValue.phaseByteSize,
            false,
            lt,
            lx,
            true,
            false,
            false,
            false,
            traceValue,
            trace);
      }

      // Tracing the Data: several 16-rows ct-loop
      int nbstep = 16;
      int nbloop = (traceValue.phaseByteSize - 1) / nbstep + 1;
      data = rightPadTo(data, nbstep * nbloop);
      for (int i = 0; i < nbloop; i++) {
        traceValue.partialReset(phase, nbstep, lt, lx);
        traceValue.input1 = data.slice(LLARGE * i, LLARGE);
        int accByteSize = 0;
        for (int ct = 0; ct < LLARGE; ct++) {
          traceValue.counter = ct;
          if (traceValue.phaseByteSize != 0) {
            accByteSize += 1;
          }
          traceValue.byte1 = traceValue.input1.get(ct);
          traceValue.acc1 = traceValue.input1.slice(0, ct + 1);
          traceValue.accByteSize = accByteSize;
          if (ct == nbstep - 1) {
            traceValue.limbConstructed = true;
            traceValue.limb = traceValue.input1;
            traceValue.nBytes = accByteSize;
          }
          traceRow(traceValue, trace);
        }
      }
      // Two rows of padding
      traceValue.partialReset(phase, 2, lt, lx);
      traceValue.lcCorrection = true;
      traceRow(traceValue, trace);

      traceValue.counter = 1;
      traceValue.phaseEnd = true;
      traceRow(traceValue, trace);
    }

    // Put INDEX_DATA to 0 at the end of the phase
    traceValue.indexData = 0;
  }

  private void handlePhaseAccessList(RlpTxnColumnsValue traceValue, Transaction tx, Trace trace) {
    int phase = 10;
    boolean lt = true;
    boolean lx = true;

    // Trivial case
    if (tx.getAccessList().get().isEmpty()) {
      traceVoidList(traceValue, phase, lt, lx, true, false, false, true, trace);
    } else {
      // Initialise traceValue
      int nbAddr = 0;
      int nbSto = 0;
      List<Integer> nbStoPerAddrList = new ArrayList<>();
      List<Integer> accessTupleByteSizeList = new ArrayList<>();
      int phaseByteSize = 0;
      for (int i = 0; i < tx.getAccessList().orElseThrow().size(); i++) {
        nbAddr += 1;
        nbSto += tx.getAccessList().orElseThrow().get(i).storageKeys().size();
        nbStoPerAddrList.add(tx.getAccessList().orElseThrow().get(i).storageKeys().size());
        accessTupleByteSizeList.add(
            21 + outerRlpSize(33 * tx.getAccessList().orElseThrow().get(i).storageKeys().size()));
        phaseByteSize += outerRlpSize(accessTupleByteSizeList.get(i));
      }

      traceValue.partialReset(phase, 0, lt, lx);
      traceValue.nbAddr = nbAddr;
      traceValue.dataLo = BigInteger.valueOf(nbAddr);
      traceValue.nbSto = nbSto;
      traceValue.dataHi = BigInteger.valueOf(nbSto);
      traceValue.phaseByteSize = phaseByteSize;

      // Trace RLP(Phase Byte Size)
      rlpByteString(
          phase,
          traceValue.phaseByteSize,
          true,
          lt,
          lx,
          true,
          false,
          false,
          false,
          traceValue,
          trace);

      // Loop Over AccessTuple
      for (int i = 0; i < nbAddr; i++) {

        // Update columns at the beginning of an AccessTuple entry
        traceValue.nbAddr -= 1;
        traceValue.nbStoPerAddr = nbStoPerAddrList.get(i);
        traceValue.addrHi = tx.getAccessList().orElseThrow().get(i).address().slice(0, 4);
        traceValue.addrLo = tx.getAccessList().orElseThrow().get(i).address().slice(4, LLARGE);
        traceValue.accessTupleByteSize = accessTupleByteSizeList.get(i);

        // Rlp(AccessTupleByteSize)
        rlpByteString(
            phase,
            traceValue.accessTupleByteSize,
            true,
            lt,
            lx,
            true,
            true,
            false,
            false,
            traceValue,
            trace);

        // RLP (address)
        handleAddress(traceValue, phase, tx.getAccessList().get().get(i).address(), trace);

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
              ((traceValue.nbSto == 0) && (traceValue.nbAddr == 0)),
              trace);
        } else {
          rlpByteString(
              phase,
              33L * traceValue.nbStoPerAddr,
              true,
              lt,
              lx,
              true,
              true,
              true,
              false,
              traceValue,
              trace);

          // Loop over StorageKey
          for (int j = 0; j < nbStoPerAddrList.get(i); j++) {
            traceValue.nbSto -= 1;
            traceValue.nbStoPerAddr -= 1;
            handleStorageKey(
                traceValue,
                ((traceValue.nbSto == 0) && (traceValue.nbAddr == 0)),
                tx.getAccessList().get().get(i).storageKeys().get(j),
                trace);
          }
        }
        traceValue.addrHi = bigIntegerToBytes(BigInteger.ZERO);
        traceValue.addrLo = bigIntegerToBytes(BigInteger.ZERO);
      }
    }
  }

  private void handlePhaseBeta(RlpTxnColumnsValue traceValue, Transaction tx, Trace trace) {
    final int phase = 11;
    final BigInteger V = tx.getV();
    Preconditions.checkArgument(bigIntegerToBytes(V).size() <= 8, "V is longer than 8 bytes");
    final boolean betaIsZero =
        V.equals(BigInteger.valueOf(27))
            || V.equals(BigInteger.valueOf(28)); // beta = ChainId = 0 iff (V == 27 or V == 28)

    // Rlp(w)
    rlpInt(
        phase,
        V,
        8,
        true,
        false,
        false,
        betaIsZero,
        false,
        traceValue,
        trace); // end of the phase if beta == 0

    // if beta != 0, then RLP(beta) and then one row with RLP().RLP ()
    if (!betaIsZero) {
      final BigInteger beta =
          BigInteger.valueOf(
              (V.longValueExact() - 35) / 2); // when b != 0, V = 2 beta + 35 or V = 2 beta + 36;

      rlpInt(phase, beta, 8, false, true, true, false, false, traceValue, trace);

      traceValue.partialReset(phase, 1, false, true);
      traceValue.limbConstructed = true;
      traceValue.limb = Bytes.concatenate(BYTES_PREFIX_SHORT_INT, BYTES_PREFIX_SHORT_INT);
      traceValue.nBytes = 2;
      traceValue.phaseEnd = true;
      traceRow(traceValue, trace);
    }
  }

  private void handlePhaseY(RlpTxnColumnsValue traceValue, Transaction tx, Trace trace) {
    traceValue.partialReset(12, 1, true, false);
    traceValue.input1 = bigIntegerToBytes(tx.getYParity());
    traceValue.limbConstructed = true;
    if (tx.getYParity().equals(BigInteger.ZERO)) {
      traceValue.limb = BYTES_PREFIX_SHORT_INT;
    } else {
      traceValue.limb = bigIntegerToBytes(BigInteger.ONE);
    }
    traceValue.nBytes = 1;
    traceValue.phaseEnd = true;
    traceRow(traceValue, trace);
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
      RlpTxnColumnsValue traceValue,
      Trace trace) {
    int lengthSize = bigIntegerToBytes(BigInteger.valueOf(length)).size();

    ByteCountAndPowerOutput byteCountingOutput = byteCounting(lengthSize, 8);

    traceValue.partialReset(phase, 8, lt, lx);
    traceValue.input1 = bigIntegerToBytes(BigInteger.valueOf(length));
    traceValue.isPrefix = isPrefix;
    traceValue.depth1 = depth1;
    traceValue.depth2 = depth2;

    Bytes input1RightShift = leftPadTo(traceValue.input1, 8);

    long acc2LastRow;
    if (length >= 56) {
      acc2LastRow = length - 56;
    } else {
      acc2LastRow = 55 - length;
    }
    Bytes acc2LastRowShift = leftPadTo(bigIntegerToBytes(BigInteger.valueOf(acc2LastRow)), 8);

    for (int ct = 0; ct < 8; ct++) {
      traceValue.counter = ct;
      traceValue.accByteSize = byteCountingOutput.accByteSizeList().get(ct);
      traceValue.power = byteCountingOutput.powerList().get(ct);
      traceValue.byte1 = input1RightShift.get(ct);
      traceValue.acc1 = input1RightShift.slice(0, ct + 1);
      traceValue.byte2 = acc2LastRowShift.get(ct);
      traceValue.acc2 = acc2LastRowShift.slice(0, ct + 1);

      if (length >= 56) {
        if (ct == 6) {
          traceValue.limbConstructed = true;
          traceValue.nBytes = 1;
          BigInteger tmp;
          if (isList) {
            tmp = BigInteger.valueOf(INT_PREFIX_LONG_LIST + lengthSize);
          } else {
            tmp = BigInteger.valueOf(INT_PREFIX_LONG_INT + lengthSize);
          }
          traceValue.limb = bigIntegerToBytes(tmp);
        }

        if (ct == 7) {
          traceValue.limb = traceValue.input1;
          traceValue.nBytes = lengthSize;
          traceValue.bit = true;
          traceValue.bitAcc = 1;
          traceValue.phaseEnd = endPhase;
        }
      } else {
        if (ct == 7) {
          traceValue.limbConstructed = true;
          Bytes tmp;
          if (isList) {
            tmp = bigIntegerToBytes(BigInteger.valueOf(INT_PREFIX_SHORT_LIST + length));
          } else {
            tmp = bigIntegerToBytes(BigInteger.valueOf(INT_PREFIX_SHORT_INT + length));
          }
          traceValue.limb = tmp;
          traceValue.nBytes = 1;
          traceValue.phaseEnd = endPhase;
        }
      }
      traceRow(traceValue, trace);
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
      RlpTxnColumnsValue traceValue,
      Trace trace) {

    traceValue.partialReset(phase, nStep, lt, lx);
    traceValue.isPrefix = isPrefix;

    Bytes inputByte = bigIntegerToBytes(input);
    int inputSize = inputByte.size();
    ByteCountAndPowerOutput byteCountingOutput = byteCounting(inputSize, nStep);

    Bytes inputBytePadded = leftPadTo(inputByte, nStep);
    BitDecOutput bitDecOutput =
        bitDecomposition(0xff & inputBytePadded.get(inputBytePadded.size() - 1), nStep);

    traceValue.input1 = inputByte;

    for (int ct = 0; ct < nStep; ct++) {
      traceValue.counter = ct;
      traceValue.byte1 = inputBytePadded.get(ct);
      traceValue.acc1 = inputBytePadded.slice(0, ct + 1);
      traceValue.power = byteCountingOutput.powerList().get(ct);
      traceValue.accByteSize = byteCountingOutput.accByteSizeList().get(ct);
      traceValue.bit = bitDecOutput.bitDecList().get(ct);
      traceValue.bitAcc = bitDecOutput.bitAccList().get(ct);

      if (input.compareTo(BigInteger.valueOf(128)) >= 0 && ct == nStep - 2) {
        traceValue.limbConstructed = true;
        traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(INT_PREFIX_SHORT_INT + inputSize));
        traceValue.nBytes = 1;
      }

      if (ct == nStep - 1) {
        if (onlyPrefix) {
          traceValue.lcCorrection = true;
          traceValue.limbConstructed = false;
          traceValue.limb = Bytes.ofUnsignedShort(0);
          traceValue.nBytes = 0;
        } else {
          traceValue.limbConstructed = true;
          traceValue.limb = inputByte;
          traceValue.nBytes = inputSize;
          traceValue.phaseEnd = endPhase;
        }
      }
      traceRow(traceValue, trace);
    }
  }

  private void handle32BytesInteger(
      RlpTxnColumnsValue traceValue, int phase, BigInteger input, Trace trace) {
    traceValue.partialReset(phase, LLARGE, true, false);
    if (input.equals(BigInteger.ZERO)) {
      // Trivial case
      traceZeroInt(traceValue, phase, true, false, false, true, trace);
    } else {
      // General case
      Bytes inputByte = bigIntegerToBytes(input);
      int inputLen = inputByte.size();
      Bytes inputByte32 = leftPadTo(inputByte, 32);
      traceValue.input1 = inputByte32.slice(0, LLARGE);
      traceValue.input2 = inputByte32.slice(LLARGE, LLARGE);

      ByteCountAndPowerOutput byteCounting;
      if (inputLen <= traceValue.nStep) {
        ByteCountAndPowerOutput byteCountingOutput = byteCounting(inputLen, traceValue.nStep);
        BitDecOutput bitDecOutput =
            bitDecomposition(inputByte.get(inputByte.size() - 1), traceValue.nStep);

        for (int ct = 0; ct < traceValue.nStep; ct++) {
          traceValue.counter = ct;
          traceValue.byte2 = traceValue.input2.get(ct);
          traceValue.acc2 = traceValue.input2.slice(0, ct + 1);
          traceValue.accByteSize = byteCountingOutput.accByteSizeList().get(ct);
          traceValue.power = byteCountingOutput.powerList().get(ct);
          traceValue.bit = bitDecOutput.bitDecList().get(ct);
          traceValue.bitAcc = bitDecOutput.bitAccList().get(ct);

          // if input >= 128, there is a RLP prefix, nothing if 0 < input < 128
          if (ct == traceValue.nStep - 2 && input.compareTo(BigInteger.valueOf(128)) >= 0) {
            traceValue.limbConstructed = true;
            traceValue.limb =
                bigIntegerToBytes(BigInteger.valueOf(INT_PREFIX_SHORT_INT + inputLen));
            traceValue.nBytes = 1;
          }
          if (ct == traceValue.nStep - 1) {
            traceValue.limbConstructed = true;
            traceValue.limb = traceValue.input2.slice(LLARGE - inputLen, inputLen);
            traceValue.nBytes = inputLen;
            traceValue.phaseEnd = true;
          }
          traceRow(traceValue, trace);
        }
      } else {
        inputLen -= traceValue.nStep;
        byteCounting = byteCounting(inputLen, traceValue.nStep);

        for (int ct = 0; ct < traceValue.nStep; ct++) {
          traceValue.counter = ct;
          traceValue.byte1 = traceValue.input1.get(ct);
          traceValue.acc1 = traceValue.input1.slice(0, ct + 1);
          traceValue.byte2 = traceValue.input2.get(ct);
          traceValue.acc2 = traceValue.input2.slice(0, ct + 1);
          traceValue.accByteSize = byteCounting.accByteSizeList().get(ct);
          traceValue.power = byteCounting.powerList().get(ct);

          if (ct == traceValue.nStep - 3) {
            traceValue.limbConstructed = true;
            traceValue.limb =
                bigIntegerToBytes(BigInteger.valueOf(INT_PREFIX_SHORT_INT + LLARGE + inputLen));
            traceValue.nBytes = 1;
          }
          if (ct == traceValue.nStep - 2) {
            traceValue.limb = traceValue.input1.slice(LLARGE - inputLen, inputLen);
            traceValue.nBytes = inputLen;
          }
          if (ct == traceValue.nStep - 1) {
            traceValue.limb = traceValue.input2;
            traceValue.nBytes = LLARGE;
            traceValue.phaseEnd = true;
          }
          traceRow(traceValue, trace);
        }
      }
    }
  }

  private void handleAddress(
      RlpTxnColumnsValue traceValue, int phase, Address address, Trace trace) {
    boolean lt = true;
    boolean lx = true;
    traceValue.partialReset(phase, LLARGE, lt, lx);
    traceValue.input1 = leftPadTo(address.slice(0, 4), LLARGE);
    traceValue.input2 = address.slice(4, LLARGE);

    if (phase == 10) {
      traceValue.depth1 = true;
    }

    for (int ct = 0; ct < traceValue.nStep; ct++) {
      traceValue.counter = ct;
      traceValue.byte1 = traceValue.input1.get(ct);
      traceValue.acc1 = traceValue.input1.slice(0, ct + 1);
      traceValue.byte2 = traceValue.input2.get(ct);
      traceValue.acc2 = traceValue.input2.slice(0, ct + 1);

      if (ct == traceValue.nStep - 3) {
        traceValue.limbConstructed = true;
        traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(INT_PREFIX_SHORT_INT + 20));
        traceValue.nBytes = 1;
      }

      if (ct == traceValue.nStep - 2) {
        traceValue.limb = address.slice(0, 4);
        traceValue.nBytes = 4;
      }
      if (ct == traceValue.nStep - 1) {
        traceValue.limb = traceValue.input2;
        traceValue.nBytes = LLARGE;

        if (phase == 7) {
          traceValue.phaseEnd = true;
        }
      }
      traceRow(traceValue, trace);
    }
  }

  private void handleStorageKey(
      RlpTxnColumnsValue traceValue, boolean end_phase, Bytes32 storage_key, Trace trace) {
    traceValue.partialReset(10, LLARGE, true, true);
    traceValue.depth1 = true;
    traceValue.depth2 = true;
    traceValue.input1 = storage_key.slice(0, LLARGE);
    traceValue.input2 = storage_key.slice(LLARGE, LLARGE);

    for (int ct = 0; ct < traceValue.nStep; ct++) {
      traceValue.counter = ct;
      traceValue.byte1 = traceValue.input1.get(ct);
      traceValue.acc1 = traceValue.input1.slice(0, ct + 1);
      traceValue.byte2 = traceValue.input2.get(ct);
      traceValue.acc2 = traceValue.input2.slice(0, ct + 1);

      if (ct == traceValue.nStep - 3) {
        traceValue.limbConstructed = true;
        traceValue.limb = bigIntegerToBytes(BigInteger.valueOf(INT_PREFIX_SHORT_INT + 32));
        traceValue.nBytes = 1;
      }

      if (ct == traceValue.nStep - 2) {
        traceValue.limb = traceValue.input1;
        traceValue.nBytes = LLARGE;
      }

      if (ct == traceValue.nStep - 1) {
        traceValue.limb = traceValue.input2;
        traceValue.phaseEnd = end_phase;
      }

      traceRow(traceValue, trace);
    }
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
            rlpOutput.writeBigIntegerScalar(chainId.orElseThrow());
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
      boolean phaseEnd,
      Trace trace) {
    traceValue.partialReset(phase, 1, lt, lx);
    traceValue.limbConstructed = true;
    traceValue.limb = BYTES_PREFIX_SHORT_INT;
    traceValue.nBytes = 1;
    traceValue.isPrefix = true;
    traceValue.phaseEnd = phaseEnd;
    traceRow(traceValue, trace);
  }

  private void traceVoidList(
      RlpTxnColumnsValue traceValue,
      int phase,
      boolean lt,
      boolean lx,
      boolean isPrefix,
      boolean depth1,
      boolean depth2,
      boolean phaseEnd,
      Trace trace) {
    traceValue.partialReset(phase, 1, lt, lx);
    traceValue.limbConstructed = true;
    traceValue.limb = BYTES_PREFIX_SHORT_LIST;
    traceValue.nBytes = 1;
    traceValue.isPrefix = isPrefix;
    traceValue.depth1 = depth1;
    traceValue.depth2 = depth2;
    traceValue.phaseEnd = phaseEnd;
    traceRow(traceValue, trace);
  }
  // Define the Tracer
  private void traceRow(RlpTxnColumnsValue traceValue, Trace builder) {
    // Decrements RLP_BYTESIZE
    if (traceValue.phase != 0) {
      if (traceValue.limbConstructed && traceValue.lt) {
        traceValue.rlpLtByteSize -= traceValue.nBytes;
      }
      if (traceValue.limbConstructed && traceValue.lx) {
        traceValue.rlpLxByteSize -= traceValue.nBytes;
      }
    }

    // Decrement phaseByteSize and accessTupleByteSize for Phase 10 (AccessList)
    if (traceValue.phase == 10) {
      // Decreases PhaseByteSize
      if (traceValue.depth1 && traceValue.limbConstructed) {
        traceValue.phaseByteSize -= traceValue.nBytes;
      }
      // Decreases AccessTupleSize
      if (traceValue.depth1
          && !(traceValue.isPrefix && !traceValue.depth2)
          && traceValue.limbConstructed) {
        traceValue.accessTupleByteSize -= traceValue.nBytes;
      }
    }

    builder
        .absTxNum(Bytes.ofUnsignedInt(traceValue.absTxNum))
        .absTxNumInfiny(Bytes.ofUnsignedInt(this.chunkList.size()))
        .acc1(traceValue.acc1)
        .acc2(traceValue.acc2)
        .accBytesize(Bytes.ofUnsignedInt(traceValue.accByteSize))
        .accessTupleBytesize(Bytes.ofUnsignedInt(traceValue.accessTupleByteSize))
        .addrHi(traceValue.addrHi)
        .addrLo(traceValue.addrLo)
        .bit(traceValue.bit)
        .bitAcc(Bytes.ofUnsignedInt(traceValue.bitAcc))
        .byte1(UnsignedByte.of(traceValue.byte1))
        .byte2(UnsignedByte.of(traceValue.byte2))
        .codeFragmentIndex(Bytes.ofUnsignedInt(traceValue.codeFragmentIndex))
        .counter(Bytes.ofUnsignedInt(traceValue.counter))
        .dataHi(bigIntegerToBytes(traceValue.dataHi))
        .dataLo(bigIntegerToBytes(traceValue.dataLo))
        .datagascost(Bytes.ofUnsignedInt(traceValue.dataGasCost))
        .depth1(traceValue.depth1)
        .depth2(traceValue.depth2);
    if (traceValue.counter == traceValue.nStep - 1) {
      builder.done(Boolean.TRUE);
    } else {
      builder.done(Boolean.FALSE);
    }
    builder
        .phaseEnd(traceValue.phaseEnd)
        .indexData(Bytes.ofUnsignedInt(traceValue.indexData))
        .indexLt(Bytes.ofUnsignedInt(traceValue.indexLt))
        .indexLx(Bytes.ofUnsignedInt(traceValue.indexLx))
        .input1(traceValue.input1)
        .input2(traceValue.input2)
        .lcCorrection(traceValue.lcCorrection)
        .isPrefix(traceValue.isPrefix)
        .limb(rightPadTo(traceValue.limb, LLARGE))
        .limbConstructed(traceValue.limbConstructed)
        .lt(traceValue.lt)
        .lx(traceValue.lx)
        .nBytes(Bytes.ofUnsignedInt(traceValue.nBytes))
        .nAddr(Bytes.ofUnsignedInt(traceValue.nbAddr))
        .nKeys(Bytes.ofUnsignedInt(traceValue.nbSto))
        .nKeysPerAddr(Bytes.ofUnsignedInt(traceValue.nbStoPerAddr))
        .nStep(Bytes.ofUnsignedInt(traceValue.nStep));
    List<Function<Boolean, Trace>> phaseColumns =
        List.of(
            builder::phase0,
            builder::phase1,
            builder::phase2,
            builder::phase3,
            builder::phase4,
            builder::phase5,
            builder::phase6,
            builder::phase7,
            builder::phase8,
            builder::phase9,
            builder::phase10,
            builder::phase11,
            builder::phase12,
            builder::phase13,
            builder::phase14);
    for (int i = 0; i < phaseColumns.size(); i++) {
      phaseColumns.get(i).apply(i == traceValue.phase);
    }
    builder
        .phaseSize(Bytes.ofUnsignedInt(traceValue.phaseByteSize))
        .power(bigIntegerToBytes(traceValue.power))
        .requiresEvmExecution(traceValue.requiresEvmExecution)
        .rlpLtBytesize(Bytes.ofUnsignedInt(traceValue.rlpLtByteSize))
        .rlpLxBytesize(Bytes.ofUnsignedInt(traceValue.rlpLxByteSize))
        .type(Bytes.ofUnsignedInt(traceValue.txType));

    // Increments Index
    if (traceValue.limbConstructed && traceValue.lt) {
      traceValue.indexLt += 1;
    }
    if (traceValue.limbConstructed && traceValue.lx) {
      traceValue.indexLx += 1;
    }

    // Increments IndexData (Phase 9)
    if (traceValue.phase == 9
        && !traceValue.isPrefix
        && (traceValue.limbConstructed || traceValue.lcCorrection)) {
      traceValue.indexData += 1;
    }

    // Decrements PhaseByteSize and DataGasCost in Data phase (phase 9)
    if (traceValue.phase == 9) {
      if (traceValue.phaseByteSize != 0 && !traceValue.isPrefix) {
        traceValue.phaseByteSize -= 1;
        if (traceValue.byte1 == 0) {
          traceValue.dataGasCost -= TxnrlpTrace.G_txdatazero.intValue();
        } else {
          traceValue.dataGasCost -= TxnrlpTrace.G_txdatanonzero.intValue();
        }
      }
    }
    if (traceValue.phaseEnd) {
      traceValue.resetDataHiLo();
    }
    builder.validateRow();

    // reconstruct RLPs
    if (traceValue.limbConstructed && traceValue.lt) {
      this.reconstructedRlpLt =
          Bytes.concatenate(this.reconstructedRlpLt, traceValue.limb.slice(0, traceValue.nBytes));
    }
    if (traceValue.limbConstructed && traceValue.lx) {
      this.reconstructedRlpLx =
          Bytes.concatenate(this.reconstructedRlpLx, traceValue.limb.slice(0, traceValue.nBytes));
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
      if (chunk.tx().getChainId().orElseThrow().equals(BigInteger.ZERO)) {
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
      if (chunk
          .tx()
          .getMaxPriorityFeePerGas()
          .orElseThrow()
          .getAsBigInteger()
          .equals(BigInteger.ZERO)) {
        rowSize += 1;
      } else {
        rowSize += 8;
      }
    }

    // Phase 5: MaxFeePerGas
    if (txType == 2) {
      if (chunk.tx().getMaxFeePerGas().orElseThrow().getAsBigInteger().equals(BigInteger.ZERO)) {
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
      rowSize += 8 + LLARGE * ((dataSize - 1) / LLARGE + 1);
      rowSize += 2; // 2 lines of padding
    }

    // Phase 10: AccessList
    if (txType == 1 || txType == 2) {
      if (chunk.tx().getAccessList().orElseThrow().isEmpty()) {
        rowSize += 1;
      } else {
        // Rlp prefix of the AccessList list
        rowSize += 8;
        for (int i = 0; i < chunk.tx().getAccessList().orElseThrow().size(); i++) {
          rowSize += 8 + 16;
          if (chunk.tx().getAccessList().orElseThrow().get(i).storageKeys().isEmpty()) {
            rowSize += 1;
          } else {
            rowSize +=
                8 + 16 * chunk.tx().getAccessList().orElseThrow().get(i).storageKeys().size();
          }
        }
      }
    }

    // Phase 11: beta
    if (txType == 0) {
      rowSize += 8;
      if (chunk.tx().getV().compareTo(BigInteger.valueOf(28)) > 0) {
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
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);
    int absTxNum = 0;
    for (RlpTxnChunk chunk : this.chunkList) {
      absTxNum += 1;
      final int codeFragmentIndex = chunk.id().map(romLex::getCFIById).orElse(0);
      traceChunk(chunk, absTxNum, codeFragmentIndex, trace);
    }
  }
}
