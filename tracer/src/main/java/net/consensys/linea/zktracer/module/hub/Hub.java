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

import lombok.Getter;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.EWord;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.add.Add;
import net.consensys.linea.zktracer.module.ext.Ext;
import net.consensys.linea.zktracer.module.hub.callstack.CallFrame;
import net.consensys.linea.zktracer.module.hub.callstack.CallFrameType;
import net.consensys.linea.zktracer.module.hub.callstack.CallStack;
import net.consensys.linea.zktracer.module.hub.chunks.AccountFragment;
import net.consensys.linea.zktracer.module.hub.chunks.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.chunks.CommonFragment;
import net.consensys.linea.zktracer.module.hub.chunks.ContextFragment;
import net.consensys.linea.zktracer.module.hub.chunks.StackFragment;
import net.consensys.linea.zktracer.module.hub.chunks.StorageFragment;
import net.consensys.linea.zktracer.module.hub.chunks.TraceFragment;
import net.consensys.linea.zktracer.module.hub.chunks.TransactionFragment;
import net.consensys.linea.zktracer.module.hub.defer.CallDefer;
import net.consensys.linea.zktracer.module.hub.defer.CreateDefer;
import net.consensys.linea.zktracer.module.hub.defer.NextContextDefer;
import net.consensys.linea.zktracer.module.hub.defer.PostExecDefer;
import net.consensys.linea.zktracer.module.hub.defer.SkippedTransactionDefer;
import net.consensys.linea.zktracer.module.hub.defer.TransactionDefer;
import net.consensys.linea.zktracer.module.hub.section.AccountSection;
import net.consensys.linea.zktracer.module.hub.section.ContextLogSection;
import net.consensys.linea.zktracer.module.hub.section.CopySection;
import net.consensys.linea.zktracer.module.hub.section.EndTransaction;
import net.consensys.linea.zktracer.module.hub.section.JumpSection;
import net.consensys.linea.zktracer.module.hub.section.StackOnlySection;
import net.consensys.linea.zktracer.module.hub.section.StackRam;
import net.consensys.linea.zktracer.module.hub.section.StorageSection;
import net.consensys.linea.zktracer.module.hub.section.TraceSection;
import net.consensys.linea.zktracer.module.hub.section.TransactionSection;
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
import org.hyperledger.besu.datatypes.AccessListEntry;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Quantity;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.gascalculator.GasCalculator;
import org.hyperledger.besu.evm.gascalculator.LondonGasCalculator;
import org.hyperledger.besu.evm.internal.Words;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.operation.Operation;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;

@Slf4j
public class Hub implements Module {
  private static final int TAU = 8;
  private static final GasCalculator gc = new LondonGasCalculator();

  public final Trace.TraceBuilder trace = Trace.builder();

  private int pc;
  private OpCode opCode;
  private int maxContextNumber;
  private Address minerAddress;
  private Wei baseFee;

  private OpCodeData opCodeData() {
    return this.opCode.getData();
  }

  final Map<Address, Map<EWord, EWord>> valOrigs = new HashMap<>();

  private Wei gasPrice() {
    return Wei.of(
        this.currentTx.getGasPrice().map(Quantity::getAsBigInteger).orElse(BigInteger.ZERO));
  }

  private EWord getValOrigOrUpdate(Address address, EWord key, EWord value) {
    EWord r = this.valOrigs.getOrDefault(address, new HashMap<>()).putIfAbsent(key, value);
    if (r == null) {
      return value;
    }
    return r;
  }

  private EWord getValOrigOrUpdate(Address address, EWord key) {
    return this.getValOrigOrUpdate(address, key, EWord.ZERO);
  }

  public final Map<Address, Integer> deploymentNumber = new HashMap<>();

  public final int deploymentNumber(Address address) {
    return this.deploymentNumber.getOrDefault(address, 0);
  }

  private void increaseDeploymentNumber(Address address) {
    this.deploymentNumber.put(address, this.deploymentNumber(address) + 1);
  }

  final Map<Address, Boolean> isDeploying = new HashMap<>();

  public final boolean isDeploying(Address address) {
    return this.isDeploying.getOrDefault(address, false);
  }

  public final void markDeploying(Address address) {
    this.increaseDeploymentNumber(address);
    this.isDeploying.put(address, true);
  }

  public final void unmarkDeploying(Address address) {
    this.isDeploying.put(address, false);
  }
  // Tx -> Opcode -> TraceSection
  private final List<List<TraceSection>> traceSections = new ArrayList<>();

  private int txChunksCount() {
    return this.traceSections.size();
  }

  private int opcodeChunksCount() {
    return this.traceSections.get(this.txChunksCount() - 1).size();
  }

  private List<TraceSection> currentTxTrace() {
    return this.traceSections.get(this.txChunksCount() - 1);
  }

  TraceSection currentTraceSection() {
    return this.traceSections.get(this.txChunksCount() - 1).get(this.opcodeChunksCount() - 1);
  }

  public void addTraceSection(TraceSection section) {
    if (!this.currentTxTrace().isEmpty()) {
      section.setContextNumber(
          this.currentTxTrace().get(this.currentTxTrace().size() - 1).contextNumber());
      section.setNewPc(this.currentTxTrace().get(this.currentTxTrace().size() - 1).pc());
    }
    section.seal(this);

    this.currentTxTrace().add(section);
  }

  void chunkNewTransaction() {
    this.traceSections.add(new ArrayList<>());
  }

  private Exceptions exceptions;

  TxState txState;
  @Getter Transaction currentTx;
  @Getter CallStack callStack;
  int txNumber = 0;
  @Getter int batchNumber = 0;
  @Getter int blockNumber = 0;
  @Getter int stamp = 0;

  /** A list of latches deferred until the end of the current transaction */
  private final List<TransactionDefer> txDefers = new ArrayList<>();
  /** Defers a latch to be executed at the end of the current transaction. */
  private void deferPostTx(TransactionDefer latch) {
    this.txDefers.add(latch);
  }

  /** A list of latches deferred until the end of the current opcode execution */
  private final List<PostExecDefer> postExecDefers = new ArrayList<>();
  /** Defers a latch to be executed after the completion of the current opcode. */
  private void deferPostExec(PostExecDefer latch) {
    this.postExecDefers.add(latch);
  }

  /** A list of latches deferred until the end of the current opcode execution */
  private final List<NextContextDefer> nextContextDefers = new ArrayList<>();
  /** Defers a latch to be executed after the completion of the current opcode. */
  private void deferNextContext(NextContextDefer latch) {
    this.nextContextDefers.add(latch);
  }

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

  public List<Module> getModules() {
    return List.of(add, ext, mod, mul, shf, trm, wcp);
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

  public static boolean isValidPrecompileCall(MessageFrame frame) {
    return switch (OpCode.of(frame.getCurrentOperation().getOpcode())) {
      case CALL, CALLCODE, STATICCALL, DELEGATECALL -> {
        if (frame.stackSize() < 2) {
          yield false; // invalid stack for a *CALL
        }

        Address targetAddress = Words.toAddress(frame.getStackItem(1));
        yield isPrecompile(targetAddress);
      }
      default -> false;
    };
  }

  /**
   * Traces a skipped transaction, i.e. a “pure” transaction without EVM execution.
   *
   * @param world a view onto the state
   */
  void processStateSkip(WorldView world) {
    log.error("TX_SKIP");
    this.stamp++;
    boolean isDeployment = this.currentTx.getTo().isEmpty();

    //
    // 3 lines -- account changes
    //
    // From account information
    Address fromAddress = this.currentTx.getSender();
    AccountSnapshot oldFromAccount =
        AccountSnapshot.fromAccount(
            world.get(fromAddress), false, this.deploymentNumber(fromAddress), false);

    // To account information
    Address toAddress = effectiveToAddress(this.currentTx, fromAddress, 0);
    boolean toIsWarm =
        (fromAddress == toAddress)
            || isPrecompile(toAddress); // should never happen – no TX to PC allowed
    AccountSnapshot oldToAccount =
        AccountSnapshot.fromAccount(
            world.get(toAddress), toIsWarm, this.deploymentNumber(toAddress), false);

    // Miner account information
    boolean minerIsWarm =
        (this.minerAddress == fromAddress)
            || (this.minerAddress == toAddress)
            || isPrecompile(this.minerAddress);
    AccountSnapshot oldMinerAccount =
        AccountSnapshot.fromAccount(
            world.get(this.minerAddress),
            minerIsWarm,
            this.deploymentNumber(this.minerAddress),
            false);

    // Putatively update deployment number
    if (isDeployment) {
      this.increaseDeploymentNumber(toAddress);
    }

    this.deferPostTx(
        new SkippedTransactionDefer(
            oldFromAccount, oldToAccount, oldMinerAccount, this.gasPrice(), this.baseFee));
  }

  public static BigInteger computeInitGas(Transaction tx) {
    boolean isDeployment = tx.getTo().isEmpty();
    return BigInteger.valueOf(
        tx.getGasLimit()
            - gc.transactionIntrinsicGasCost(tx.getPayload(), isDeployment)
            - tx.getAccessList().map(gc::accessListGasCost).orElse(0L));
  }

  /**
   * Fill the columns shared by all operations.
   *
   * @return a chunk representing the share columns
   */
  public CommonFragment traceCommon() {
    return new CommonFragment(
        this.txNumber,
        this.batchNumber,
        this.txState,
        this.stamp,
        0,
        false, // TODO
        this.opCodeData().instructionFamily(),
        this.exceptions.snapshot(),
        false, // TODO
        false, // TODO
        this.currentFrame().getContextNumber(),
        this.currentFrame().getContextNumber(),
        0, // TODO
        false, // TODO
        false, // TODO
        false, // TODO
        this.pc,
        this.pc,
        this.currentFrame().addressAsEWord(),
        this.currentFrame().getCodeDeploymentNumber(),
        this.currentFrame().isCodeDeploymentStatus(),
        this.currentFrame().getAccountDeploymentNumber(),
        0,
        0,
        0,
        0,
        0, // TODO
        this.opCodeData().stackSettings().twoLinesInstruction(),
        false, // TODO -- retcon with stack
        0, // TODO -- do it now
        0 // TODO -- do it now
        );
  }

  void processStateWarm() {
    this.stamp++;
    this.txState = TxState.TX_WARM;

    // reproduction ordonnée des préchauffages de la Tx
    this.currentTx
        .getAccessList()
        .ifPresent(
            preWarmed -> {
              for (AccessListEntry entry : preWarmed) {
                // TODO
              }
            });
  }

  void processStateInit(WorldView world) {
    this.stamp++;
    this.txState = TxState.TX_INIT;

    var fromAddress = this.currentTx.getSender();
    boolean isDeployment = this.currentTx.getTo().isEmpty();
    Address toAddress =
        effectiveToAddress(this.currentTx, fromAddress, world.get(fromAddress).getNonce());
    this.callStack =
        new CallStack(
            toAddress,
            isDeployment ? CallFrameType.INIT_CODE : CallFrameType.STANDARD,
            new Bytecode(world.get(toAddress).getCode()),
            Wei.of(this.currentTx.getValue().getAsBigInteger()),
            this.currentTx.getGasLimit(),
            this.currentTx.getData().orElse(Bytes.EMPTY),
            this.maxContextNumber,
            this.deploymentNumber(toAddress),
            toAddress.isEmpty() ? 0 : this.deploymentNumber.getOrDefault(toAddress, 0),
            this.isDeploying(toAddress));
  }

  private int currentLine() {
    return this.trace.size();
  }

  public CallFrame currentFrame() {
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
      this.traceOperation(frame);
    } else {
      this.addTraceSection(new StackOnlySection(this));
    }
  }

  void processStateFinal(WorldView worldView, Transaction tx, boolean isSuccess) {
    log.error("TX_FINAL");
    this.stamp++;

    Address fromAddress = this.currentTx.getSender();
    Account fromAccount = worldView.get(fromAddress);
    AccountSnapshot fromSnapshot =
        AccountSnapshot.fromAccount(
            fromAccount, true, this.deploymentNumber(fromAddress), this.isDeploying(fromAddress));

    Account minerAccount = worldView.get(this.minerAddress);
    AccountSnapshot minerSnapshot =
        AccountSnapshot.fromAccount(
            minerAccount,
            true,
            this.deploymentNumber(this.minerAddress),
            this.isDeploying(this.minerAddress));

    if (isSuccess) {
      // if no revert: 2 account rows (sender, coinbase) + 1 tx row
      this.addTraceSection(
          new EndTransaction(
              this,
              new AccountFragment(fromSnapshot, fromSnapshot, false, 0, false),
              new AccountFragment(minerSnapshot, minerSnapshot, false, 0, false),
              new TransactionFragment(
                  this.batchNumber, minerAddress, tx, true, this.gasPrice(), this.baseFee)));
    } else {
      // otherwise 4 account rows (sender, coinbase, sender, recipient) + 1 tx row
      Address toAddress = this.currentTx.getSender();
      Account toAccount = worldView.get(toAddress);
      AccountSnapshot toSnapshot =
          AccountSnapshot.fromAccount(
              toAccount, true, this.deploymentNumber(toAddress), this.isDeploying(toAddress));
      this.addTraceSection(
          new EndTransaction(
              this,
              new AccountFragment(fromSnapshot, fromSnapshot, false, 0, false),
              new AccountFragment(minerSnapshot, minerSnapshot, false, 0, false),
              new AccountFragment(fromSnapshot, fromSnapshot, false, 0, false),
              new AccountFragment(toSnapshot, toSnapshot, false, 0, false)));
    }
  }

  /**
   * Compute the effective address of a transaction target, i.e. the specified target if explicitly
   * set, or the to-be-deployed address otherwise.
   *
   * @return the effective target address of tx
   */
  private static Address effectiveToAddress(Transaction tx, Address fromAddress, long fromNonce) {
    return tx.getTo().map(x -> (Address) x).orElse(Address.contractAddress(fromAddress, fromNonce));
  }

  @Override
  public void traceStartTx(final WorldView world, final Transaction tx) {
    this.chunkNewTransaction();
    if (tx.getTo().isPresent() && isPrecompile(tx.getTo().get())) {
      throw new RuntimeException("Call to precompile forbidden");
    } else {
      this.txNumber += 1;
    }
    this.currentTx = tx;

    if ((this.currentTx.getTo().isPresent()
            && world.get(this.currentTx.getTo().get()).getCode().isEmpty()) // pure transaction
        || (this.currentTx.getTo().isEmpty()
            && this.currentTx.getInit().isEmpty())) { // contract creation without init code
      this.txState = TxState.TX_SKIP;
      this.processStateSkip(world);
      return;
    }

    this.processStateWarm();
    this.processStateInit(world);
    this.txState = TxState.TX_EXEC;
  }

  @Override
  public void traceEndTx(
      WorldView worldView,
      Transaction tx,
      boolean status,
      Bytes output,
      List<Log> logs,
      long gasUsed) {
    this.txState = TxState.TX_FINAL;
    this.processStateFinal(worldView, tx, status);

    for (TransactionDefer defer : this.txDefers) {
      defer.run(this, null, this.currentTx); // TODO
    }
    this.txDefers.clear();
  }

  private void unlatchStack(MessageFrame frame) {
    if (this.currentFrame().getPending() == null) {
      return;
    }

    StackContext pending = this.currentFrame().getPending();
    for (StackLine line : pending.getLines()) {
      if (line.needsResult()) {
        EWord result = EWord.ZERO;
        if (!exceptions.any()) {
          result = EWord.of(frame.getStackItem(0));
        }

        // This works because we are certain that the stack chunks are the first.
        ((StackFragment) this.currentTraceSection().getLines().get(line.ct()).specific())
            .stackOps()
            .get(line.resultColumn() - 1)
            .setValue(result);
      }
    }
  }

  @Override
  public void traceContextEnter(MessageFrame frame) {
    this.maxContextNumber += 1;

    final boolean isDeployment = frame.getType() == MessageFrame.Type.CONTRACT_CREATION;
    final Address codeAddress = frame.getContractAddress();
    final CallFrameType frameType =
        frame.isStatic() ? CallFrameType.STATIC : CallFrameType.STANDARD;
    if (isDeployment) {
      this.markDeploying(codeAddress);
    }
    final int codeDeploymentNumber = this.deploymentNumber.getOrDefault(codeAddress, 0);
    this.callStack.enter(
        frame.getContractAddress(),
        new Bytecode(frame.getCode().getBytes()),
        frameType,
        frame.getValue(),
        frame.getRemainingGas(),
        this.trace.size(),
        frame.getInputData(),
        this.stamp + 1,
        this.deploymentNumber.getOrDefault(codeAddress, 0),
        codeDeploymentNumber,
        isDeployment);

    for (NextContextDefer defer : this.nextContextDefers) {
      defer.run(this, frame);
    }
    this.nextContextDefers.clear();
  }

  @Override
  public void traceContextExit(MessageFrame frame) {
    unmarkDeploying(this.currentFrame().getCodeAddress());
    this.callStack.exit(this.trace.size() - 1, frame.getReturnData());
  }

  @Override
  public void trace(final MessageFrame frame) {
    if (this.txState == TxState.TX_SKIP) {
      return;
    }

    this.processStateExec(frame);
  }

  public void tracePostExecution(MessageFrame frame, Operation.OperationResult operationResult) {
    if (txState == TxState.TX_SKIP) {
      return;
    }

    this.unlatchStack(frame);

    if (this.opCode.isCreate() && operationResult.getHaltReason() == null) {
      this.handleCreate(Address.wrap(frame.getStackItem(0)));
    }

    for (PostExecDefer defer : this.postExecDefers) {
      defer.run(this, frame, operationResult);
    }
    this.postExecDefers.clear();
  }

  private void handleCreate(Address target) {
    this.deploymentNumber.put(target, this.deploymentNumber.getOrDefault(target, 0) + 1);
  }

  @Override
  public void traceStartBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    this.blockNumber++;
    this.minerAddress = blockHeader.getCoinbase();
    this.baseFee = Wei.of(blockHeader.getBaseFee().get().getAsBigInteger());
  }

  @Override
  public void traceStartConflation(long blockCount) {
    this.batchNumber++;
  }

  @Override
  public void traceEndConflation() {
    for (List<TraceSection> txSections : this.traceSections) {
      for (TraceSection section : txSections) {
        section.retcon(this);
      }
    }
  }

  @Override
  public Object commit() {
    for (var txSection : this.traceSections) {
      for (TraceSection opSection : txSection) {
        for (TraceSection.TraceLine line : opSection.getLines()) {
          line.trace(this.trace);
        }
      }
    }
    return new HubTrace(trace.build());
  }

  public int lineCount() {
    int count = 0;
    for (var txSection : this.traceSections) {
      for (TraceSection opSection : txSection) {
        count += opSection.getLines().size();
      }
    }
    return count;
  }

  void traceOperation(MessageFrame frame) {
    boolean updateReturnData =
        this.opCodeData().isHalt()
            || this.opCodeData().isInvalid()
            || exceptions.any()
            || isValidPrecompileCall(frame);

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
          INVALID -> this.addTraceSection(new StackOnlySection(this));
      case CONTEXT, LOG -> this.addTraceSection(
          new ContextLogSection(
              this, new ContextFragment(this.callStack, this.currentFrame(), updateReturnData)));
      case ACCOUNT -> {
        TraceSection accountSection = new AccountSection(this);
        if (this.opCodeData().stackSettings().flag1()) {
          accountSection.addChunk(
              this, new ContextFragment(this.callStack, this.currentFrame(), updateReturnData));
        }

        Address targetAddress =
            switch (this.opCode) {
              case BALANCE, EXTCODESIZE, EXTCODEHASH -> Address.wrap(frame.getStackItem(0));
              default -> Address.wrap(this.currentFrame().getAddress());
            };
        Account targetAccount = frame.getWorldUpdater().getAccount(targetAddress);
        AccountSnapshot accountSnapshot =
            AccountSnapshot.fromAccount(
                targetAccount,
                frame.isAddressWarm(targetAddress),
                this.deploymentNumber(targetAddress),
                this.isDeploying(targetAddress));
        accountSection.addChunk(
            this, new AccountFragment(accountSnapshot, accountSnapshot, false, 0, false));

        this.addTraceSection(accountSection);
      }
      case COPY -> {
        TraceSection copySection = new CopySection(this);
        if (this.opCodeData().stackSettings().flag1()) {
          Address targetAddress =
              switch (this.opCode) {
                case CODECOPY -> this.currentFrame().getCodeAddress();
                case EXTCODECOPY -> Words.toAddress(frame.getStackItem(0));
                default -> throw new IllegalStateException("unexpected opcode");
              };
          Account targetAccount = frame.getWorldUpdater().getAccount(targetAddress);
          AccountSnapshot accountSnapshot =
              AccountSnapshot.fromAccount(
                  targetAccount,
                  frame.isAddressWarm(targetAddress),
                  this.deploymentNumber(targetAddress),
                  this.isDeploying(targetAddress));

          copySection.addChunk(
              this, new AccountFragment(accountSnapshot, accountSnapshot, false, 0, false));
        } else {
          copySection.addChunk(
              this, new ContextFragment(this.callStack, this.currentFrame(), updateReturnData));
        }
        this.addTraceSection(copySection);
      }
      case TRANSACTION -> this.addTraceSection(
          new TransactionSection(
              this,
              new TransactionFragment(
                  this.batchNumber,
                  frame.getMiningBeneficiary(),
                  this.currentTx,
                  true,
                  frame.getGasPrice(),
                  frame.getBlockValues().getBaseFee().orElse(Wei.ZERO))));

      case STACK_RAM -> {
        TraceSection stackRamSection = new StackRam(this);
        if (this.opCodeData().stackSettings().flag2()) {
          stackRamSection.addChunk(
              this, new ContextFragment(this.callStack, this.currentFrame(), updateReturnData));
        }
        this.addTraceSection(stackRamSection);
      }
      case STORAGE -> {
        Address address = this.currentFrame().getAddress();
        EWord key = EWord.of(frame.getStackItem(0));
        switch (this.opCode) {
          case SSTORE -> {
            EWord valNext = EWord.of(frame.getStackItem(0));
            this.addTraceSection(
                new StorageSection(
                    this,
                    new ContextFragment(this.callStack, this.currentFrame(), updateReturnData),
                    new StorageFragment(
                        address,
                        this.currentFrame().getAccountDeploymentNumber(),
                        key,
                        getValOrigOrUpdate(address, key, valNext),
                        EWord.of(frame.getTransientStorageValue(address, key)),
                        valNext,
                        frame.isStorageWarm(address, key),
                        true)));
          }
          case SLOAD -> {
            EWord valCurrent = EWord.of(frame.getTransientStorageValue(address, key));
            this.addTraceSection(
                new StorageSection(
                    this,
                    new ContextFragment(this.callStack, this.currentFrame(), updateReturnData),
                    new StorageFragment(
                        address,
                        this.currentFrame().getAccountDeploymentNumber(),
                        key,
                        getValOrigOrUpdate(address, key),
                        valCurrent,
                        valCurrent,
                        frame.isStorageWarm(address, key),
                        true)));
          }
          default -> throw new IllegalStateException("invalid operation in family STORAGE");
        }
      }
      case CREATE -> {
        Address myAddress = this.currentFrame().getAddress();
        Account myAccount = frame.getWorldUpdater().getAccount(myAddress);
        AccountSnapshot myAccountSnapshot =
            AccountSnapshot.fromAccount(
                myAccount,
                frame.isAddressWarm(myAddress),
                this.deploymentNumber(myAddress),
                this.isDeploying(myAddress));

        Address createdAddress = this.currentFrame().getAddress();
        Account createdAccount = frame.getWorldUpdater().getAccount(createdAddress);
        AccountSnapshot createdAccountSnapshot =
            AccountSnapshot.fromAccount(
                createdAccount,
                frame.isAddressWarm(myAddress),
                this.deploymentNumber(myAddress),
                this.isDeploying(myAddress));

        CreateDefer protoCreateSection =
            new CreateDefer(
                myAccountSnapshot,
                createdAccountSnapshot,
                new ContextFragment(this.callStack, this.currentFrame(), updateReturnData));
        // Will be traced in one (and only one!) of these depending on the success of the operation
        this.deferPostExec(protoCreateSection);
        this.deferNextContext(protoCreateSection);
      }
      case CALL -> {
        Address myAddress = this.currentFrame().getAddress();
        Account myAccount = frame.getWorldUpdater().getAccount(myAddress);
        AccountSnapshot myAccountSnapshot =
            AccountSnapshot.fromAccount(
                myAccount,
                frame.isAddressWarm(myAddress),
                this.deploymentNumber(myAddress),
                this.isDeploying(myAddress));

        Address calledAddress = Address.wrap(frame.getStackItem(1));
        Account calledAccount = frame.getWorldUpdater().getAccount(calledAddress);
        AccountSnapshot calledAccountSnapshot =
            AccountSnapshot.fromAccount(
                calledAccount,
                frame.isAddressWarm(myAddress),
                this.deploymentNumber(myAddress),
                this.isDeploying(myAddress));

        CallDefer protoCallSection =
            new CallDefer(
                myAccountSnapshot,
                calledAccountSnapshot,
                new ContextFragment(this.callStack, this.currentFrame(), updateReturnData));

        this.deferPostExec(protoCallSection);
        this.deferNextContext(protoCallSection);
      }
      case JUMP -> {
        AccountSnapshot codeAccountSnapshot =
            AccountSnapshot.fromAccount(
                frame.getWorldUpdater().getAccount(this.currentFrame().getCodeAddress()),
                true,
                this.deploymentNumber(this.currentFrame().getCodeAddress()),
                this.currentFrame().isCodeDeploymentStatus());

        this.addTraceSection(
            new JumpSection(
                this,
                new ContextFragment(this.callStack, this.currentFrame(), updateReturnData),
                new AccountFragment(codeAccountSnapshot, codeAccountSnapshot, false, 0, false)));
      }
    }
  }

  public List<TraceFragment> makeStackChunks() {
    List<TraceFragment> r = new ArrayList<>();
    if (this.currentFrame().getPending().getLines().isEmpty()) {
      for (int i = 0; i < (this.opCodeData().stackSettings().twoLinesInstruction() ? 2 : 1); i++) {
        r.add(
            new StackFragment(
                this.currentFrame().getStack().snapshot(), new StackLine(i).asStackOperations()));
      }
    } else {
      for (StackLine line : this.currentFrame().getPending().getLines()) {
        r.add(
            new StackFragment(this.currentFrame().getStack().snapshot(), line.asStackOperations()));
      }
    }
    return r;
  }
}
