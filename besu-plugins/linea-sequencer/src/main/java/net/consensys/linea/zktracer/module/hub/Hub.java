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

import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;

import lombok.Getter;
import lombok.experimental.Accessors;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.EWord;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.add.Add;
import net.consensys.linea.zktracer.module.ext.Ext;
import net.consensys.linea.zktracer.module.hub.defer.CallDefer;
import net.consensys.linea.zktracer.module.hub.defer.CreateDefer;
import net.consensys.linea.zktracer.module.hub.defer.DeferRegistry;
import net.consensys.linea.zktracer.module.hub.defer.SkippedTransactionDefer;
import net.consensys.linea.zktracer.module.hub.fragment.AccountFragment;
import net.consensys.linea.zktracer.module.hub.fragment.AccountSnapshot;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.fragment.StackFragment;
import net.consensys.linea.zktracer.module.hub.fragment.StorageFragment;
import net.consensys.linea.zktracer.module.hub.fragment.TraceFragment;
import net.consensys.linea.zktracer.module.hub.fragment.TransactionFragment;
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
import net.consensys.linea.zktracer.module.hub.section.WarmupSection;
import net.consensys.linea.zktracer.module.hub.stack.ConflationInfo;
import net.consensys.linea.zktracer.module.hub.stack.StackContext;
import net.consensys.linea.zktracer.module.hub.stack.StackLine;
import net.consensys.linea.zktracer.module.mod.Mod;
import net.consensys.linea.zktracer.module.mul.Mul;
import net.consensys.linea.zktracer.module.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.module.runtime.callstack.CallFrameType;
import net.consensys.linea.zktracer.module.runtime.callstack.CallStack;
import net.consensys.linea.zktracer.module.shf.Shf;
import net.consensys.linea.zktracer.module.trm.Trm;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.gas.projector.GasProjector;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.AccessListEntry;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.operation.Operation;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;

@Slf4j
@Accessors(fluent = true)
public class Hub implements Module {
  private static final int TAU = 8;
  public static final GasProjector gp = new GasProjector();

  public final Trace.TraceBuilder trace = Trace.builder();
  // Tx -> Opcode -> TraceSection
  private final List<TxTrace> traceSections = new ArrayList<>();

  @Getter ConflationInfo conflation = new ConflationInfo();
  @Getter BlockInfo block = new BlockInfo();
  @Getter TxInfo tx = new TxInfo();
  @Getter CallStack callStack = new CallStack();

  private final DeferRegistry defers = new DeferRegistry();
  @Getter private Exceptions exceptions;
  @Getter int stamp = 0;
  @Getter private int pc;
  @Getter private OpCode opCode;
  private int maxContextNumber;
  @Getter private MessageFrame frame;

  public OpCodeData opCodeData() {
    return this.opCode.getData();
  }

  private int txChunksCount() {
    return this.traceSections.size();
  }

  private TxTrace currentTxTrace() {
    return this.traceSections.get(this.txChunksCount() - 1);
  }

  TraceSection currentTraceSection() {
    return this.currentTxTrace().currentSection();
  }

  public int lastPc() {
    if (this.currentTxTrace().isEmpty()) {
      return 0;
    } else {
      return this.currentTxTrace().currentSection().pc();
    }
  }

  public int lastContextNumber() {
    if (this.currentTxTrace().isEmpty()) {
      return 0;
    } else {
      return this.currentTxTrace().currentSection().contextNumber();
    }
  }

  public void addTraceSection(TraceSection section) {
    section.seal(this);
    this.currentTxTrace().add(section);
  }

  void createNewTxTrace() {
    this.traceSections.add(new TxTrace());
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
    boolean isDeployment = this.tx.transaction().getTo().isEmpty();

    //
    // 3 lines -- account changes
    //
    // From account information
    Address fromAddress = this.tx.transaction().getSender();
    AccountSnapshot oldFromAccount =
        AccountSnapshot.fromAccount(
            world.get(fromAddress),
            false,
            this.conflation.deploymentInfo().number(fromAddress),
            false);

    // To account information
    Address toAddress = effectiveToAddress(this.tx.transaction());
    boolean toIsWarm =
        (fromAddress == toAddress)
            || isPrecompile(toAddress); // should never happen – no TX to PC allowed
    AccountSnapshot oldToAccount =
        AccountSnapshot.fromAccount(
            world.get(toAddress),
            toIsWarm,
            this.conflation.deploymentInfo().number(toAddress),
            false);

    // Miner account information
    boolean minerIsWarm =
        (this.block.minerAddress == fromAddress)
            || (this.block.minerAddress == toAddress)
            || isPrecompile(this.block.minerAddress);
    AccountSnapshot oldMinerAccount =
        AccountSnapshot.fromAccount(
            world.get(this.block.minerAddress),
            minerIsWarm,
            this.conflation.deploymentInfo().number(this.block.minerAddress),
            false);

    // Putatively update deployment number
    if (isDeployment) {
      this.conflation.deploymentInfo().deploy(toAddress);
    }

    this.defers.postTx(
        new SkippedTransactionDefer(
            oldFromAccount, oldToAccount, oldMinerAccount, this.tx.gasPrice(), this.block.baseFee));
  }

  void processStateWarm(WorldView world) {
    this.stamp++;
    this.tx
        .transaction()
        .getAccessList()
        .ifPresent(
            preWarmed -> {
              Set<Address> seenAddresses = new HashSet<>();
              Map<Address, Set<Bytes32>> seenKeys = new HashMap<>();
              List<TraceFragment> fragments = new ArrayList<>();

              for (AccessListEntry entry : preWarmed) {
                Address address = entry.address();
                AccountSnapshot snapshot =
                    AccountSnapshot.fromAccount(
                        world.get(address), seenAddresses.contains(address), 0, false);
                fragments.add(new AccountFragment(snapshot, snapshot, false, 0, false));
                seenAddresses.add(address);

                List<Bytes32> keys = entry.storageKeys();
                for (Bytes32 key_ : keys) {
                  UInt256 key = UInt256.fromBytes(key_);
                  EWord value = EWord.of(world.get(address).getStorageValue(key));
                  fragments.add(
                      new StorageFragment(
                          address,
                          this.conflation.deploymentInfo().number(address),
                          EWord.of(key),
                          value,
                          value,
                          value,
                          seenKeys.computeIfAbsent(address, x -> new HashSet<>()).contains(key),
                          true));
                  seenKeys.get(address).add(key);
                }
              }
              this.addTraceSection(new WarmupSection(this, fragments));
            });
    this.tx.state(TxState.TX_INIT);
  }

  void processStateInit(WorldView world) {
    this.stamp++;
    boolean isDeployment = this.tx.transaction().getTo().isEmpty();
    Address toAddress = effectiveToAddress(this.tx.transaction());
    this.callStack.newBedrock(
        toAddress,
        isDeployment ? CallFrameType.INIT_CODE : CallFrameType.STANDARD,
        new Bytecode(world.get(toAddress).getCode()),
        Wei.of(this.tx.transaction().getValue().getAsBigInteger()),
        this.tx.transaction().getGasLimit(),
        this.tx.transaction().getData().orElse(Bytes.EMPTY),
        this.maxContextNumber,
        this.conflation.deploymentInfo().number(toAddress),
        toAddress.isEmpty() ? 0 : this.conflation.deploymentInfo().number(toAddress),
        this.conflation.deploymentInfo().isDeploying(toAddress));
    this.tx.state(TxState.TX_EXEC);
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
      default -> {}
    }
  }

  void processStateExec(MessageFrame frame) {
    this.opCode = OpCode.of(frame.getCurrentOperation().getOpcode());
    this.pc = frame.getPC();
    this.stamp++;
    this.exceptions = Exceptions.fromFrame(frame, Hub.gp);
    this.frame = frame;

    this.handleStack(frame);
    this.triggerModules(frame);
    if (this.exceptions.any() || this.opCode == OpCode.REVERT) {
      this.callStack.revert(this.stamp);
    }

    if (this.currentFrame().getStack().isOk()) {
      this.traceOperation(frame);
    } else {
      this.addTraceSection(new StackOnlySection(this));
    }
  }

  void processStateFinal(WorldView worldView, Transaction tx, boolean isSuccess) {
    log.error("TX_FINAL");
    this.stamp++;

    Address fromAddress = this.tx.transaction().getSender();
    Account fromAccount = worldView.get(fromAddress);
    AccountSnapshot fromSnapshot =
        AccountSnapshot.fromAccount(
            fromAccount,
            true,
            this.conflation.deploymentInfo().number(fromAddress),
            this.conflation.deploymentInfo().isDeploying(fromAddress));

    Account minerAccount = worldView.get(this.block.minerAddress);
    AccountSnapshot minerSnapshot =
        AccountSnapshot.fromAccount(
            minerAccount,
            true,
            this.conflation.deploymentInfo().number(this.block.minerAddress),
            this.conflation.deploymentInfo().isDeploying(this.block.minerAddress));

    if (isSuccess) {
      // if no revert: 2 account rows (sender, coinbase) + 1 tx row
      this.addTraceSection(
          new EndTransaction(
              this,
              new AccountFragment(fromSnapshot, fromSnapshot, false, 0, false),
              new AccountFragment(minerSnapshot, minerSnapshot, false, 0, false),
              TransactionFragment.prepare(
                  this.conflation.number(),
                  this.block.minerAddress,
                  tx,
                  true,
                  this.tx.gasPrice(),
                  this.block.baseFee,
                  this.tx.initialGas())));
    } else {
      // otherwise 4 account rows (sender, coinbase, sender, recipient) + 1 tx row
      Address toAddress = this.tx.transaction().getSender();
      Account toAccount = worldView.get(toAddress);
      AccountSnapshot toSnapshot =
          AccountSnapshot.fromAccount(
              toAccount,
              true,
              this.conflation.deploymentInfo().number(toAddress),
              this.conflation.deploymentInfo().isDeploying(toAddress));
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
  public static Address effectiveToAddress(Transaction tx) {
    return tx.getTo()
        .map(x -> (Address) x)
        .orElse(Address.contractAddress(tx.getSender(), tx.getNonce()));
  }

  @Override
  public void traceStartTx(final WorldView world, final Transaction tx) {
    this.tx.update(tx);
    this.createNewTxTrace();

    if (this.tx.shouldSkip(world)) {
      this.tx.state(TxState.TX_SKIP);
      this.processStateSkip(world);
      return;
    } else {
      this.tx.state(TxState.TX_WARM);
    }

    this.processStateWarm(world);
    this.processStateInit(world);
  }

  @Override
  public void traceEndTx(
      WorldView world, Transaction tx, boolean status, Bytes output, List<Log> logs, long gasUsed) {
    this.tx.state(TxState.TX_FINAL);
    this.tx.status(status);
    this.processStateFinal(world, tx, status);

    this.defers.runPostTx(this, world, tx);

    this.currentTxTrace().postTxRetcon(this);
  }

  private void unlatchStack(MessageFrame frame) {
    if (this.currentFrame().getPending() == null) {
      return;
    }

    StackContext pending = this.currentFrame().getPending();
    for (int i = 0; i < pending.getLines().size(); i++) {
      StackLine line = pending.getLines().get(i);
      if (line.needsResult()) {
        EWord result = EWord.ZERO;
        // Only pop from the stack if no exceptions have been encountered
        if (!exceptions.any()) {
          result = EWord.of(frame.getStackItem(0));
        }

        // This works because we are certain that the stack chunks are the first.
        ((StackFragment) this.currentTraceSection().getLines().get(i).specific())
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
      this.conflation.deploymentInfo().markDeploying(codeAddress);
    }
    final int codeDeploymentNumber = this.conflation.deploymentInfo().number(codeAddress);
    this.callStack.enter(
        frame.getContractAddress(),
        new Bytecode(frame.getCode().getBytes()),
        frameType,
        frame.getValue(),
        frame.getRemainingGas(),
        frame.getInputData(),
        this.stamp + 1,
        this.conflation.deploymentInfo().number(codeAddress),
        codeDeploymentNumber,
        isDeployment);

    this.defers.runNextContext(this, frame);
  }

  @Override
  public void traceContextExit(MessageFrame frame) {
    conflation.deploymentInfo().unmarkDeploying(this.currentFrame().getCodeAddress());

    ContextExceptions contextExceptions = ContextExceptions.fromFrame(this.currentFrame(), frame);
    this.currentTraceSection().setContextExceptions(contextExceptions);
    if (contextExceptions.any()) {
      this.callStack.revert(this.stamp);
    }

    if (!frame.getReturnData().isEmpty() && !frame.getOutputData().isEmpty()) {
      throw new RuntimeException("both return data and output data set");
    }

    if (!frame.getOutputData().isEmpty()) {
      this.callStack.exit(this.trace.size() - 1, frame.getOutputData());
    } else {
      this.callStack.exit(this.trace.size() - 1, frame.getReturnData());
    }
  }

  @Override
  public void trace(final MessageFrame frame) {
    if (this.tx.state() == TxState.TX_SKIP) {
      return;
    }

    this.processStateExec(frame);
  }

  public void tracePostExecution(MessageFrame frame, Operation.OperationResult operationResult) {
    if (this.tx.state() == TxState.TX_SKIP) {
      return;
    }

    this.unlatchStack(frame);

    if (this.opCode.isCreate() && operationResult.getHaltReason() == null) {
      this.handleCreate(Words.toAddress(frame.getStackItem(0)));
    }

    this.defers.runPostExec(this, frame, operationResult);
  }

  private void handleCreate(Address target) {
    this.conflation.deploymentInfo().deploy(target);
  }

  @Override
  public void traceStartBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    this.block.update(blockHeader);
  }

  @Override
  public void traceStartConflation(long blockCount) {
    this.conflation.update();
  }

  @Override
  public void traceEndConflation() {
    for (TxTrace txTrace : this.traceSections) {
      txTrace.postConflationRetcon(this, null /* TODO WorldView */);
    }
  }

  @Override
  public Object commit() {
    for (TxTrace txTrace : this.traceSections) {
      txTrace.commit(this.trace);
    }
    return new HubTrace(trace.build());
  }

  public long refundedGas() {
    return this.currentTxTrace().refundedGas();
  }

  public long remainingGas() {
    return this.frame.getRemainingGas();
  }

  public int lineCount() {
    int count = 0;
    for (TxTrace txSection : this.traceSections) {
      count += txSection.lineCount();
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
              case BALANCE, EXTCODESIZE, EXTCODEHASH -> Words.toAddress(frame.getStackItem(0));
              default -> Address.wrap(this.currentFrame().getAddress());
            };
        Account targetAccount = frame.getWorldUpdater().getAccount(targetAddress);
        AccountSnapshot accountSnapshot =
            AccountSnapshot.fromAccount(
                targetAccount,
                frame.isAddressWarm(targetAddress),
                this.conflation.deploymentInfo().number(targetAddress),
                this.conflation.deploymentInfo().isDeploying(targetAddress));
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
                  this.conflation.deploymentInfo().number(targetAddress),
                  this.conflation.deploymentInfo().isDeploying(targetAddress));

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
              TransactionFragment.prepare(
                  this.conflation.number(),
                  frame.getMiningBeneficiary(),
                  this.tx.transaction(),
                  true,
                  frame.getGasPrice(),
                  frame.getBlockValues().getBaseFee().orElse(Wei.ZERO),
                  this.tx.initialGas())));

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
                        this.tx.storage().getOriginalValueOrUpdate(address, key, valNext),
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
                        this.tx.storage().getOriginalValueOrUpdate(address, key),
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
                this.conflation.deploymentInfo().number(myAddress),
                this.conflation.deploymentInfo().isDeploying(myAddress));

        Address createdAddress = this.currentFrame().getAddress();
        Account createdAccount = frame.getWorldUpdater().getAccount(createdAddress);
        AccountSnapshot createdAccountSnapshot =
            AccountSnapshot.fromAccount(
                createdAccount,
                frame.isAddressWarm(myAddress),
                this.conflation.deploymentInfo().number(myAddress),
                this.conflation.deploymentInfo().isDeploying(myAddress));

        CreateDefer protoCreateSection =
            new CreateDefer(
                myAccountSnapshot,
                createdAccountSnapshot,
                new ContextFragment(this.callStack, this.currentFrame(), updateReturnData));
        // Will be traced in one (and only one!) of these depending on the success of the operation
        this.defers.postExec(protoCreateSection);
        this.defers.nextContext(protoCreateSection);
      }
      case CALL -> {
        Address myAddress = this.currentFrame().getAddress();
        Account myAccount = frame.getWorldUpdater().getAccount(myAddress);
        AccountSnapshot myAccountSnapshot =
            AccountSnapshot.fromAccount(
                myAccount,
                frame.isAddressWarm(myAddress),
                this.conflation.deploymentInfo().number(myAddress),
                this.conflation.deploymentInfo().isDeploying(myAddress));

        Address calledAddress = Words.toAddress(frame.getStackItem(1));
        Account calledAccount = frame.getWorldUpdater().getAccount(calledAddress);
        AccountSnapshot calledAccountSnapshot =
            AccountSnapshot.fromAccount(
                calledAccount,
                frame.isAddressWarm(myAddress),
                this.conflation.deploymentInfo().number(myAddress),
                this.conflation.deploymentInfo().isDeploying(myAddress));

        CallDefer protoCallSection =
            new CallDefer(
                myAccountSnapshot,
                calledAccountSnapshot,
                new ContextFragment(this.callStack, this.currentFrame(), updateReturnData));

        this.defers.postExec(protoCallSection);
        this.defers.nextContext(protoCallSection);
      }
      case JUMP -> {
        AccountSnapshot codeAccountSnapshot =
            AccountSnapshot.fromAccount(
                frame.getWorldUpdater().getAccount(this.currentFrame().getCodeAddress()),
                true,
                this.conflation.deploymentInfo().number(this.currentFrame().getCodeAddress()),
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
            StackFragment.prepare(
                this.currentFrame().getStack().snapshot(),
                new StackLine().asStackOperations(),
                this.exceptions.snapshot(),
                gp.of(frame, this.opCode).staticGas()));
      }
    } else {
      for (StackLine line : this.currentFrame().getPending().getLines()) {
        r.add(
            StackFragment.prepare(
                this.currentFrame().getStack().snapshot(),
                line.asStackOperations(),
                this.exceptions.snapshot(),
                gp.of(frame, this.opCode).staticGas()));
      }
    }
    return r;
  }
}
