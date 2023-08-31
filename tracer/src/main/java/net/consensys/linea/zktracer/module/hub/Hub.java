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

package net.consensys.linea.zktracer.module.hub;

import java.math.BigInteger;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

import lombok.Getter;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.EWord;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.add.Add;
import net.consensys.linea.zktracer.module.ext.Ext;
import net.consensys.linea.zktracer.module.hub.callstack.CallFrame;
import net.consensys.linea.zktracer.module.hub.callstack.CallFrameType;
import net.consensys.linea.zktracer.module.hub.callstack.CallStack;
import net.consensys.linea.zktracer.module.hub.chunks.StackChunk;
import net.consensys.linea.zktracer.module.hub.chunks.TraceChunk;
import net.consensys.linea.zktracer.module.hub.stack.StackContext;
import net.consensys.linea.zktracer.module.hub.stack.StackLine;
import net.consensys.linea.zktracer.module.mod.Mod;
import net.consensys.linea.zktracer.module.mul.Mul;
import net.consensys.linea.zktracer.module.shf.Shf;
import net.consensys.linea.zktracer.module.trm.Trm;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Hash;
import org.hyperledger.besu.datatypes.Quantity;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.Code;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.code.CodeV0;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.gascalculator.GasCalculator;
import org.hyperledger.besu.evm.gascalculator.LondonGasCalculator;
import org.hyperledger.besu.evm.operation.Operation;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;

@Slf4j
public class Hub implements Module {
  private static final Address ADDRESS_ZERO = Address.fromHexString("0x0");
  private static final int TAU = 8;

  public final Trace.TraceBuilder trace = Trace.builder();

  private int pc;
  private OpCode opCode;
  private int maxContextNumber;

  private OpCodeData opCodeData() {
    return this.opCode.getData();
  }

  private final Map<Address, Integer> deploymentNumber = new HashMap<>();
  private final Map<Address, Boolean> deploymentStatus = new HashMap<>();

  private final List<List<TraceChunk>> traceChunks = new ArrayList<>();
  private Exceptions exceptions;

  TxState txState;
  @Getter Transaction currentTx;
  CallStack callStack;
  int txNumber = 0;
  @Getter int batchNumber = 0;
  @Getter int blockNumber = 0;
  int stamp = 0;

  private final Module add;
  private final Module ext;
  private final Module mod;
  private final Module mul;
  private final Module shf;
  private final Module trm;
  private final Module wcp;

  public Hub(Add add, Ext ext, Mod mod, Mul mul, Shf shf, Trm trm, Wcp wcp) {
    this.add = add;
    this.ext = ext;
    this.mod = mod;
    this.mul = mul;
    this.shf = shf;
    this.trm = trm;
    this.wcp = wcp;
  }

  @Override
  public String jsonKey() {
    return "hub_v2_off";
  }

  @Override
  public final List<OpCode> supportedOpCodes() {
    // The Hub wants to catch all opcodes
    return List.of(OpCode.values());
  }

  public static boolean isPrecompile(Address to) {
    return List.of(
            Address.ECREC,
            Address.SHA256,
            Address.RIPEMD160,
            Address.ID,
            Address.MODEXP,
            Address.ALTBN128_ADD,
            Address.ALTBN128_MUL,
            Address.ALTBN128_PAIRING,
            Address.BLAKE2B_F_COMPRESSION)
        .contains(to);
  }

  /**
   * Traces a skipped transaction, i.e. a “pure” transaction without EVM execution.
   *
   * @param frame the frame of the transaction
   */
  void traceSkippedTx(MessageFrame frame) {
    this.stamp++;
    Quantity value = this.currentTx.getValue();
    boolean isDeployment = this.currentTx.getTo().isEmpty();

    // 3 lines -- account changes
    // From account information
    EWord from = EWord.of(this.currentTx.getSender());
    Address fromAddress = this.currentTx.getSender();
    Account fromAccount = frame.getWorldUpdater().get(fromAddress);
    long fromNonce = fromAccount.getNonce();
    Wei currentFromBalance = fromAccount.getBalance();
    Wei newFromBalance = currentFromBalance.subtract((Wei) value);
    Code fromCode = frame.getCode();
    EWord fromCodeHash = EWord.of(fromCode.getCodeHash());
    this.traceCommon()
        .peekAtAccount(true)
        .pAccountAddressHi(from.hiBigInt())
        .pAccountAddressLo(from.loBigInt())
        .pAccountBalance(currentFromBalance.toUnsignedBigInteger())
        .pAccountBalanceNew(newFromBalance.toUnsignedBigInteger())
        .pAccountCodeHashHi(fromCodeHash.hiBigInt())
        .pAccountCodeHashLo(fromCodeHash.loBigInt())
        .pAccountCodeHashHiNew(fromCodeHash.hiBigInt())
        .pAccountCodeHashLoNew(fromCodeHash.loBigInt())
        .pAccountCodeSize(BigInteger.valueOf(fromCode.getSize()))
        .pAccountCodeSizeNew(BigInteger.valueOf(fromCode.getSize()))
        .pAccountDeploymentNumber(BigInteger.ZERO)
        .pAccountDeploymentNumberInfty(BigInteger.ZERO)
        .pAccountDeploymentNumberNew(BigInteger.ZERO)
        .pAccountDeploymentStatus(BigInteger.ZERO)
        .pAccountDeploymentStatusInfty(false)
        .pAccountDeploymentStatusNew(BigInteger.ZERO)
        .pAccountExists(true)
        .pAccountExistsNew(true)
        .pAccountSufficientBalance(true)
        .pAccountHasCode(false)
        .pAccountHasCodeNew(false)
        .pAccountIsPrecompile(false)
        .pAccountNonce(BigInteger.valueOf(fromNonce))
        .pAccountNonceNew(BigInteger.valueOf(fromNonce + 1))
        .pAccountWarm(false)
        .pAccountWarmNew(true)
        .fillAndValidateRow();

    // To account information
    Address toAddress = effectiveToAddress(this.currentTx, fromAccount);
    EWord eToAddress = EWord.of(toAddress);
    Optional<Account> toAccount = Optional.of(frame.getWorldUpdater().get(toAddress));
    Optional<Long> toNonce = toAccount.map(Account::getNonce);
    Wei currentToBalance = frame.getWorldUpdater().getAccount(toAddress).getBalance();
    Wei newToBalance = currentToBalance.add((Wei) value);
    Bytes toCode = toAccount.map(Account::getCode).orElse(Bytes.EMPTY);
    EWord toCodeHash = EWord.of(toAccount.map(Account::getCodeHash).orElse(Hash.EMPTY));
    this.traceCommon()
        .peekAtAccount(true)
        .pAccountAddressHi(eToAddress.hiBigInt())
        .pAccountAddressLo(eToAddress.loBigInt())
        .pAccountBalance(currentToBalance.toUnsignedBigInteger())
        .pAccountBalanceNew(newToBalance.toUnsignedBigInteger())
        .pAccountCodeHashHi(toCodeHash.hiBigInt())
        .pAccountCodeHashLo(toCodeHash.loBigInt())
        .pAccountCodeHashHiNew(toCodeHash.hiBigInt())
        .pAccountCodeHashLoNew(toCodeHash.loBigInt())
        .pAccountCodeSize(BigInteger.valueOf(toCode.size()))
        .pAccountCodeSizeNew(BigInteger.valueOf(toCode.size()))
        .pAccountDeploymentNumber(BigInteger.ZERO)
        .pAccountDeploymentNumberInfty(BigInteger.ZERO)
        .pAccountDeploymentNumberNew(BigInteger.ZERO)
        .pAccountDeploymentStatus(BigInteger.ZERO)
        .pAccountDeploymentStatusInfty(false)
        .pAccountDeploymentStatusNew(BigInteger.ZERO)
        .pAccountExists(!isDeployment)
        .pAccountExistsNew(true)
        .pAccountSufficientBalance(true)
        .pAccountHasCode(!toCode.isEmpty())
        .pAccountHasCodeNew(!toCode.isEmpty())
        .pAccountIsPrecompile(false)
        .pAccountNonce(BigInteger.valueOf(isDeployment ? 0L : toNonce.orElse(0L)))
        .pAccountNonceNew(BigInteger.valueOf(isDeployment ? 1L : toNonce.orElse(0L)))
        .pAccountWarm(false)
        .pAccountWarmNew(true)
        .fillAndValidateRow();

    // Basecoin/miner information
    Address minerAddress = frame.getMiningBeneficiary();
    EWord eMinerAddress = EWord.of(minerAddress);
    Optional<Account> minerAccount = Optional.ofNullable(frame.getWorldUpdater().get(minerAddress));
    Optional<Long> minerNonce = minerAccount.map(Account::getNonce);
    Bytes minerCode = minerAccount.map(Account::getCode).orElse(Bytes.EMPTY);
    EWord minerCodeHash = EWord.of(minerAccount.map(Account::getCodeHash).orElse(Hash.EMPTY));

    Wei currentMinerBalance =
        frame.getWorldUpdater().get(frame.getMiningBeneficiary()).getBalance();
    Wei newMinerBalance = Wei.ZERO; // TODO: latch it
    this.traceCommon()
        .peekAtAccount(true)
        .pAccountAddressHi(eMinerAddress.hiBigInt())
        .pAccountAddressLo(eMinerAddress.loBigInt())
        .pAccountIsPrecompile(isPrecompile(minerAddress))
        .pAccountNonce(
            BigInteger.valueOf(
                (isDeployment && minerAddress == toAddress)
                    ? 1L
                    : minerAddress == fromAddress
                        ? fromNonce + 1
                        : minerNonce.orElse(0L))) // TODO: check == behaviour
        .pAccountNonceNew(
            BigInteger.valueOf(
                (isDeployment && minerAddress == toAddress)
                    ? 1L
                    : minerAddress == fromAddress ? fromNonce + 1 : minerNonce.orElse(0L)))
        .pAccountBalance(currentMinerBalance.toUnsignedBigInteger())
        .pAccountBalanceNew(newMinerBalance.toUnsignedBigInteger())
        .pAccountCodeSize(BigInteger.valueOf(minerCode.size()))
        .pAccountCodeSizeNew(BigInteger.valueOf(minerCode.size()))
        .pAccountCodeHashHi(minerCodeHash.hiBigInt())
        .pAccountCodeHashLo(minerCodeHash.loBigInt())
        .pAccountCodeHashHiNew(minerCodeHash.hiBigInt())
        .pAccountCodeHashLoNew(minerCodeHash.loBigInt())
        .pAccountHasCode(!minerCode.isEmpty())
        .pAccountHasCodeNew(!minerCode.isEmpty())
        .pAccountExists(minerAccount.isPresent() || (isDeployment && minerAddress == toAddress))
        .pAccountExistsNew(
            minerAccount.isPresent()
                || this.currentTx
                    .getGasPrice()
                    .map(g -> g.getAsBigInteger() != BigInteger.ZERO)
                    .orElse(false)) // TODO: latch it?
        .pAccountWarm(
            isPrecompile(minerAddress) || minerAddress == fromAddress || minerAddress == toAddress)
        .pAccountWarmNew(true)
        .pAccountDeploymentNumber(BigInteger.ZERO)
        .pAccountDeploymentNumberNew(BigInteger.ZERO)
        .pAccountDeploymentStatus(BigInteger.ZERO)
        .pAccountDeploymentStatusNew(BigInteger.ZERO)
        .pAccountSufficientBalance(true)
        .pAccountDeploymentStatusInfty(false)
        .fillAndValidateRow();

    // 1 line -- tx data
    this.traceCommon()
        .peekAtTransaction(true)
        .pTransactionBatchNumber(BigInteger.valueOf(this.batchNumber))
        .pTransactionNonce(BigInteger.valueOf(this.currentTx.getNonce()))
        .pTransactionIsDeployment(isDeployment)
        .pTransactionFromAddressHi(from.hiBigInt())
        .pTransactionFromAddressLo(from.loBigInt())
        .pTransactionToAddressHi(eToAddress.hiBigInt())
        .pTransactionToAddressLo(eToAddress.loBigInt())
        .pTransactionInitGas(computeInitGas(this.currentTx))
        .pTransactionValue(value.getAsBigInteger())
        .pTransactionGasFee(frame.getGasPrice().toUnsignedBigInteger())
        .fillAndValidateRow();
  }

  public static BigInteger computeInitGas(Transaction tx) {
    GasCalculator gc = new LondonGasCalculator();

    boolean isDeployment = tx.getTo().isEmpty();
    return BigInteger.valueOf(
        tx.getGasLimit()
            - gc.transactionIntrinsicGasCost(tx.getPayload(), isDeployment)
            - tx.getAccessList().map(gc::accessListGasCost).orElse(0L));
  }

  /**
   * Fill the columns shared by all operations.
   *
   * @return the partially filled trace row
   */
  public Trace.TraceBuilder traceCommon() {
    return this.trace
        .absoluteTransactionNumber(BigInteger.valueOf(this.txNumber))
        .batchNumber(BigInteger.valueOf(this.batchNumber))
        .txSkip(this.txState == TxState.TX_SKIP)
        .txWarm(this.txState == TxState.TX_WARM)
        .txInit(this.txState == TxState.TX_INIT)
        .txExec(this.txState == TxState.TX_STATE)
        .txFinl(this.txState == TxState.TX_FINAL)
        .hubStamp(BigInteger.valueOf(this.stamp))
        .transactionEndStamp(BigInteger.ZERO) // TODO
        .transactionReverts(BigInteger.ZERO) // TODO
        .contextMayChangeFlag(false) // TODO
        .exceptionAhoyFlag(false) // TODO
        .abortFlag(false) // TODO
        .failureConditionFlag(false) // TODO

        // Context data
        .contextNumber(BigInteger.valueOf(this.currentFrame().getContextNumber()))
        .contextNumberNew(BigInteger.ZERO) // TODO
        .contextRevertStamp(BigInteger.ZERO) // TODO
        .contextWillRevertFlag(false) // TODO
        .contextGetsRevrtdFlag(false) // TODO
        .contextSelfRevrtsFlag(false) // TODO
        .programCounter(BigInteger.valueOf(this.pc))
        .programCounterNew(BigInteger.ZERO) // TODO

        // Bytecode metadata
        .codeAddressHi(this.currentFrame().addressAsEWord().hiBigInt())
        .codeAddressLo(this.currentFrame().addressAsEWord().loBigInt())
        .codeDeploymentNumber(BigInteger.valueOf(this.currentFrame().getAccountDeploymentNumber()))
        .codeDeploymentStatus(false) // TODO
        .callerContextNumber(
            BigInteger.valueOf(this.callStack.caller().getAccountDeploymentNumber()))
        .gasExpected(BigInteger.ZERO)
        .gasActual(BigInteger.ZERO)
        .gasCost(BigInteger.ZERO)
        .gasNext(BigInteger.ZERO)
        .gasRefund(BigInteger.ZERO)
        .twoLineInstruction(this.opCodeData().stackSettings().twoLinesInstruction())
        .counterTli(false)
        .numberOfNonStackRows(BigInteger.ZERO)
        .counterNsr(BigInteger.ZERO);
  }

  void processStateWarm() {
    this.stamp++;
    // x lines - warm addresses
    // y lines - warm storage keys
  }

  void processStateInit() {
    this.stamp++;
    // 2 lines -- trace from & to accounts
    // 1 line  -- trace tx data
  }

  private int currentLine() {
    return this.trace.size();
  }

  private CallFrame currentFrame() {
    return this.callStack.top();
  }

  private void handleStack(MessageFrame frame) {
    this.currentFrame().getStack().processInstruction(frame, this.currentFrame(), TAU * this.stamp);
    this.currentFrame().getPending().setStartInTrace(this.currentLine());
  }

  void triggerModules(MessageFrame frame) {
    switch (this.opCodeData().instructionFamily()) {
      case ADD -> {
        if (this.exceptions.noStackException()) {
          this.add.trace(frame);
        }
      }
      case MOD -> {
        if (this.exceptions.noStackException()) {
          this.mod.trace(frame);
        }
      }
      case MUL -> {
        if (this.exceptions.noStackException()) {
          this.mul.trace(frame);
        }
      }
      case EXT -> {
        if (this.exceptions.noStackException()) {
          this.ext.trace(frame);
        }
      }
      case WCP -> {
        if (this.exceptions.noStackException()) {
          this.wcp.trace(frame);
        }
      }
      case BIN -> {}
      case SHF -> {
        if (this.exceptions.noStackException()) {
          this.shf.trace(frame);
        }
      }
      case KEC -> {}
      case CONTEXT -> {}
      case ACCOUNT -> {}
      case COPY -> {}
      case TRANSACTION -> {}
      case BATCH -> {}
      case STACK_RAM -> {}
      case STORAGE -> {}
      case JUMP -> {}
      case MACHINE_STATE -> {}
      case PUSH_POP -> {}
      case DUP -> {}
      case SWAP -> {}
      case LOG -> {}
      case CREATE -> {}
      case CALL -> {}
      case HALT -> {}
      case INVALID -> {}
    }
  }

  void processStateExec(MessageFrame frame) {
    this.opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    this.pc = frame.getPC();
    this.stamp++;
    this.exceptions = Exceptions.fromFrame(frame);
    this.handleStack(frame);
    this.triggerModules(frame);

    if (this.currentFrame().getStack().isOk()) {
      /*
      this.handleCallReturnData
      this.handleMemory
      this.handleRam
       */
    }

    this.updateTrace();
  }

  void processStateFinal() {
    this.stamp++;
    // if no revert: 2 account rows (sender, coinbase) + 1 tx row
    // otherwise 4 account rows (sender, coinbase, sender, recipient) + 1 tx row

    this.txState = TxState.TX_PRE_INIT;
  }

  /**
   * Compute the effective address of a transaction target, i.e. the specified target if explicitely
   * set, or the to-be-deployed address otherwise.
   *
   * @param tx the transaction to find the target for
   * @param senderAccount the transaction sender account
   * @return the effective target address of tx
   */
  private static Address effectiveToAddress(Transaction tx, Account senderAccount) {
    return tx.getTo()
        .map(x -> (Address) x)
        .orElse(Address.contractAddress(senderAccount.getAddress(), senderAccount.getNonce()));
  }

  @Override
  public void traceStartTx(Transaction tx) {
    if (tx.getTo().isPresent() && isPrecompile(tx.getTo().get())) {
      throw new RuntimeException("Call to precompile forbidden");
    } else {
      this.txNumber += 1;
    }
    this.currentTx = tx;

    // impossible to do the pre-init here -- missing information from the MessageFrame
    this.txState = TxState.TX_PRE_INIT;
  }

  @Override
  public void traceEndTx() {
    this.txState = TxState.TX_FINAL;
    this.processStateFinal();
  }

  private void txInit(final MessageFrame frame) {
    // TODO: check that this actually does what it should
    if ((!frame.getRecipientAddress().equals(ADDRESS_ZERO)
            && frame.getCode().getSize() == 0) // pure transaction
        || (frame.getRecipientAddress().equals(ADDRESS_ZERO)
            && frame.getInputData().isEmpty())) { // contract creation without initcode
      this.txState = TxState.TX_SKIP;
      this.traceSkippedTx(frame);
      return;
    }

    var to = this.currentTx.getTo();
    boolean isDeployment = to.isEmpty();
    Address toAddress =
        effectiveToAddress(this.currentTx, frame.getWorldUpdater().getSenderAccount(frame));
    this.callStack =
        new CallStack(
            toAddress,
            isDeployment ? CallFrameType.INIT_CODE : CallFrameType.STANDARD,
            CodeV0.EMPTY_CODE, // TODO
            Wei.of(this.currentTx.getValue().getAsBigInteger()),
            this.currentTx.getGasLimit(),
            this.currentTx.getData().orElse(Bytes.EMPTY),
            this.maxContextNumber,
            0, // TODO
            to.isEmpty() ? 0 : this.deploymentNumber.getOrDefault(to.get(), 0),
            false); // TODO

    if (false /* doWarm */) {
      this.processStateWarm();
    }
    this.processStateInit();
    this.txState = TxState.TX_STATE;
  }

  private void unlatchStack(MessageFrame frame, boolean failureState, boolean mxpx) {
    if (this.currentFrame().getPending() == null) {
      return;
    }

    StackContext pending = this.currentFrame().getPending();
    for (StackLine line : pending.getLines()) {
      if (line.needsResult()) {
        EWord result = EWord.ZERO;
        if (!failureState) {
          result = EWord.of(frame.getStackItem(0));
        }

        // This works because we are certain that the stack chunks are the first.
        ((StackChunk) traceChunks.get(traceChunks.size() - 1).get(line.ct()))
            .stackOps()
            .get(line.resultColumn() - 1)
            .setValue(result);
      }
    }
  }

  @Override
  public void traceContextEnter(MessageFrame frame) {
    this.maxContextNumber += 1;
    this.callStack.enter(
        frame.getContractAddress(),
        frame.getCode(),
        this.callStack.top().getType().ofOpCode(OpCode.of(frame.getCurrentOperation().getOpcode())),
        frame.getValue(),
        frame.getRemainingGas(),
        this.trace.size(),
        frame.getInputData(),
        this.maxContextNumber,
        0,
        this.deploymentNumber.getOrDefault(frame.getContractAddress(), 0),
        false);
  }

  @Override
  public void traceContextExit(MessageFrame frame) {
    this.callStack.exit(this.trace.size() - 1, frame.getReturnData());
  }

  @Override
  public void trace(final MessageFrame frame) {
    if (this.txState == TxState.TX_PRE_INIT) {
      this.txInit(frame);
    }

    if (this.txState == TxState.TX_SKIP) {
      return;
    }

    this.processStateExec(frame);
  }

  public void tracePostExecution(MessageFrame frame, Operation.OperationResult operationResult) {
    boolean mxpx = false;
    this.unlatchStack(frame, false, mxpx);

    if (this.opCode.isCreate() && operationResult.getHaltReason() == null) {
      this.handleCreate(Address.wrap(frame.getStackItem(0)));
    }
  }

  private void handleCreate(Address target) {
    this.deploymentNumber.put(target, this.deploymentNumber.getOrDefault(target, 0) + 1);
  }

  @Override
  public void traceStartBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    this.blockNumber++;
  }

  @Override
  public void traceStartConflation(long blockCount) {
    this.batchNumber++;
  }

  @Override
  public Object commit() {
    for (List<TraceChunk> opChunks : traceChunks) {
      for (TraceChunk chunk : opChunks) {
        this.traceCommon();
        chunk.trace(this.trace);
        this.trace.fillAndValidateRow();
      }
    }
    return new HubTrace(trace.build());
  }

  void updateTrace() {
    var opChunks = new ArrayList<TraceChunk>();
    this.makeStackChunks(opChunks);

    switch (this.opCodeData().instructionFamily()) {
      case ADD,
          MOD,
          MUL,
          SHF,
          BIN,
          WCP,
          EXT,
          KEC,
          BATCH,
          MACHINE_STATE,
          PUSH_POP,
          DUP,
          SWAP,
          HALT,
          INVALID -> {}
      case CONTEXT, LOG -> {
        //        opChunks.add(new ContextChunk());
      }
      case ACCOUNT -> {
        if (this.opCodeData().stackSettings().flag1()) {
          //          opChunks.add(new ContextChunk());
          //          opChunks.add(new AccountChunk());
        } else {
          //          opChunks.add(new AccountChunk());
        }
      }
      case COPY -> {
        if (this.opCodeData().stackSettings().flag1()) {
          //          opChunks.add(new AccountChunk());
        } else {
          //          opChunks.add(new ContextChunk());
        }
      }
      case TRANSACTION -> {
        //        opChunks.add(new TransactionChunk());
      }
      case STACK_RAM -> {
        if (this.opCodeData().stackSettings().flag2()) {
          //          opChunks.add(new ContextChunk());
        }
      }
      case STORAGE -> {
        //        opChunks.add(new ContextChunk());
        //        opChunks.add(new StorageChunk());
      }
      case CREATE, CALL -> {
        //        opChunks.add(new ContextChunk());
        //        opChunks.add(new AccountChunk());
        //        opChunks.add(new AccountChunk());
        //        opChunks.add(new AccountChunk());
      }
      case JUMP -> {
        //        opChunks.add(new ContextChunk());
        //        opChunks.add(new AccountChunk());
      }
    }

    this.traceChunks.add(opChunks);
  }

  void makeStackChunks(List<TraceChunk> currentChunks) {
    if (this.currentFrame().getPending().getLines().isEmpty()) {
      for (int i = 0; i < (this.opCodeData().stackSettings().twoLinesInstruction() ? 2 : 1); i++) {
        currentChunks.add(
            new StackChunk(
                this.currentFrame().getStack().snapshot(), new StackLine(i).asStackOperations()));
      }
    } else {
      for (StackLine line : this.currentFrame().getPending().getLines()) {
        currentChunks.add(
            new StackChunk(this.currentFrame().getStack().snapshot(), line.asStackOperations()));
      }
    }
  }
}
