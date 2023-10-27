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

package net.consensys.linea.zktracer.module.hub;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;

import lombok.Getter;
import lombok.experimental.Accessors;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.EWord;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.ModuleTrace;
import net.consensys.linea.zktracer.module.add.Add;
import net.consensys.linea.zktracer.module.ext.Ext;
import net.consensys.linea.zktracer.module.hub.defer.*;
import net.consensys.linea.zktracer.module.hub.fragment.*;
import net.consensys.linea.zktracer.module.hub.fragment.misc.MiscFragment;
import net.consensys.linea.zktracer.module.hub.section.*;
import net.consensys.linea.zktracer.module.hub.stack.ConflationInfo;
import net.consensys.linea.zktracer.module.hub.stack.StackContext;
import net.consensys.linea.zktracer.module.hub.stack.StackLine;
import net.consensys.linea.zktracer.module.mod.Mod;
import net.consensys.linea.zktracer.module.mul.Mul;
import net.consensys.linea.zktracer.module.mxp.Mxp;
import net.consensys.linea.zktracer.module.rlpAddr.RlpAddr;
import net.consensys.linea.zktracer.module.rlp_txn.RlpTxn;
import net.consensys.linea.zktracer.module.rlp_txrcpt.RlpTxrcpt;
import net.consensys.linea.zktracer.module.rom.Rom;
import net.consensys.linea.zktracer.module.romLex.RomLex;
import net.consensys.linea.zktracer.module.runtime.callstack.*;
import net.consensys.linea.zktracer.module.shf.Shf;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.*;
import net.consensys.linea.zktracer.opcode.gas.projector.GasProjector;
import org.apache.tuweni.bytes.Bytes;
import org.apache.tuweni.bytes.Bytes32;
import org.apache.tuweni.units.bigints.UInt256;
import org.hyperledger.besu.datatypes.AccessListEntry;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.account.Account;
import org.hyperledger.besu.evm.account.AccountState;
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

  // Revertible state of the hub
  private final State state = new State();

  // Long-lived states
  @Getter ConflationInfo conflation = new ConflationInfo();
  @Getter BlockInfo block = new BlockInfo();
  @Getter TxInfo tx = new TxInfo();
  @Getter CallStack callStack = new CallStack();
  private final DeferRegistry defers = new DeferRegistry();

  // These attributes are transient (opcode-specific) and do not need to be reversed.
  @Getter private Exceptions exceptions;
  @Getter private Aborts aborts;
  @Getter private Signals signals;

  private void resetSignals() {
    this.signals = new Signals(this);
  }

  public int stamp() {
    return this.state.stamps().hub();
  }

  public OpCodeData opCodeData() {
    return this.currentFrame().opCode().getData();
  }

  TraceSection currentTraceSection() {
    return this.state.currentTxTrace().currentSection();
  }

  public int lastPc() {
    if (this.state.currentTxTrace().isEmpty()) {
      return 0;
    } else {
      return this.state.currentTxTrace().currentSection().pc();
    }
  }

  public int lastContextNumber() {
    if (this.state.currentTxTrace().isEmpty()) {
      return 0;
    } else {
      return this.state.currentTxTrace().currentSection().contextNumber();
    }
  }

  public void addTraceSection(TraceSection section) {
    section.seal(this);
    this.state.currentTxTrace().add(section);
  }

  private final Module add = new Add();
  private final Module ext = new Ext();
  private final Module mod = new Mod();
  private final Module mul = new Mul();
  private final Module shf = new Shf();
  private final Module wcp = new Wcp();
  private final RlpTxn rlpTxn;
  private final Module mxp;
  private final RlpTxrcpt rlpTxrcpt = new RlpTxrcpt();
  private final RlpAddr rlpAddr = new RlpAddr();
  private final Rom rom;
  private final RomLex romLex;

  private final List<Module> modules;

  public Hub() {
    this.mxp = new Mxp(this);
    this.romLex = new RomLex(this);
    this.rom = new Rom(this.romLex);
    this.rlpTxn = new RlpTxn(this.romLex);

    this.modules =
        List.of(
            this.romLex, // romLex must be traced before modules requiring CodeFragmentIndex, like
            // RlpTxn, TxnData, etc
            this.add,
            this.ext,
            this.mod,
            this.mul,
            this.mxp,
            this.shf,
            this.wcp,
            this.rlpTxn,
            this.rlpTxrcpt,
            this.rlpAddr,
            this.rom);
  }

  /**
   * @return a list of all modules for which to generate traces
   */
  public List<Module> getModulesToTrace() {
    return List.of(
        this,
        this.romLex,
        this.add,
        this.ext,
        this.mod,
        this.mul,
        this.shf,
        this.wcp,
        this.mxp,
        this.rlpTxn,
        this.rlpTxrcpt,
        this.rlpAddr,
        this.rom);
  }

  @Override
  public String jsonKey() {
    return "hub_v2_off";
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
    this.state.stamps().stampHub();
    boolean isDeployment = this.tx.transaction().getTo().isEmpty();

    //
    // 3 sections -- account changes
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

    // Putatively updateCallerReturndata deployment number
    if (isDeployment) {
      this.conflation.deploymentInfo().deploy(toAddress);
    }

    this.defers.postTx(
        new SkippedPostTransactionDefer(
            oldFromAccount, oldToAccount, oldMinerAccount, this.tx.gasPrice(), this.block.baseFee));
  }

  /**
   * Traces the warm-up information of a transaction
   *
   * @param world a view onto the state
   */
  void processStateWarm(WorldView world) {
    this.state.stamps().stampHub();
    this.tx
        .transaction()
        .getAccessList()
        .ifPresent(
            preWarmed -> {
              if (!preWarmed.isEmpty()) {
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
                    EWord value =
                        Optional.ofNullable(world.get(address))
                            .map(account -> EWord.of(account.getStorageValue(key)))
                            .orElse(EWord.ZERO);
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
              }
            });
    this.tx.state(TxState.TX_INIT);
  }

  /**
   * Trace the preamble of a transaction
   *
   * @param world a view onto the state
   */
  void processStateInit(WorldView world) {
    this.state.stamps().stampHub();
    boolean isDeployment = this.tx.transaction().getTo().isEmpty();
    Address toAddress = effectiveToAddress(this.tx.transaction());
    this.callStack.newBedrock(
        this.state.stamps().hub(),
        this.tx.transaction().getSender(),
        toAddress,
        isDeployment ? CallFrameType.INIT_CODE : CallFrameType.STANDARD,
        new Bytecode(
            toAddress == null
                ? this.tx.transaction().getData().orElse(Bytes.EMPTY)
                : Optional.ofNullable(world.get(toAddress))
                    .map(AccountState::getCode)
                    .orElse(Bytes.EMPTY)), // TODO: see with Olivier
        Wei.of(this.tx.transaction().getValue().getAsBigInteger()),
        this.tx.transaction().getGasLimit(),
        this.tx.transaction().getData().orElse(Bytes.EMPTY),
        this.conflation.deploymentInfo().number(toAddress),
        toAddress.isEmpty() ? 0 : this.conflation.deploymentInfo().number(toAddress),
        this.conflation.deploymentInfo().isDeploying(toAddress));
    this.tx.state(TxState.TX_EXEC);
  }

  public CallFrame currentFrame() {
    // If the transaction is skipped, it has no relevant CF to trace.
    if (this.tx.state() == TxState.TX_SKIP) {
      return CallFrame.EMPTY;
    }
    return this.callStack.current();
  }

  public MessageFrame messageFrame() {
    return this.callStack.current().frame();
  }

  public long getRemainingGas() {
    return 0; // TODO:
  }

  private void handleStack(MessageFrame frame) {
    this.currentFrame()
        .stack()
        .processInstruction(frame, this.currentFrame(), TAU * this.state.stamps().hub());
  }

  void triggerModules(MessageFrame frame) {
    switch (this.opCodeData().instructionFamily()) {
      case ADD -> {
        if (this.exceptions.noStackException()) {
          this.add.tracePreOpcode(frame);
        }
      }
      case MOD -> {
        if (this.exceptions.noStackException()) {
          this.mod.tracePreOpcode(frame);
        }
      }
      case MUL -> {
        if (this.exceptions.noStackException()) {
          this.mul.tracePreOpcode(frame);
        }
      }
      case EXT -> {
        if (this.exceptions.noStackException()) {
          this.ext.tracePreOpcode(frame);
        }
      }
      case WCP -> {
        if (this.exceptions.noStackException()) {
          this.wcp.tracePreOpcode(frame);
        }
      }
      case BIN -> {}
      case SHF -> {
        if (this.exceptions.noStackException()) {
          this.shf.tracePreOpcode(frame);
        }
      }
      case KEC -> {
        if (this.exceptions.noStackException()) {
          this.mxp.tracePreOpcode(frame);
        }
      }
      case CONTEXT -> {}
      case ACCOUNT -> {}
      case COPY -> {
        if (this.exceptions.noStackException()) {
          if (this.currentFrame().opCode() == OpCode.RETURNDATACOPY) {
            if (!this.exceptions.returnDataCopyFault()) {
              this.mxp.tracePreOpcode(frame);
            }
          } else {
            this.mxp.tracePreOpcode(frame);
          }
        }
        if (!this.exceptions.any() && this.callStack().getDepth() < 1024) {
          this.romLex.tracePreOpcode(frame);
        }
      }
      case TRANSACTION -> {}
      case BATCH -> {}
      case STACK_RAM -> {
        if (this.exceptions.noStackException()
            && this.currentFrame().opCode() != OpCode.CALLDATALOAD) {
          this.mxp.tracePreOpcode(frame);
        }
      }
      case STORAGE -> {}
      case JUMP -> {}
      case MACHINE_STATE -> {
        if (this.exceptions.noStackException() && this.currentFrame().opCode() == OpCode.MSIZE) {
          this.mxp.tracePreOpcode(frame);
        }
      }
      case PUSH_POP -> {}
      case DUP -> {}
      case SWAP -> {}
      case LOG -> {
        if (this.exceptions.noStackException() && !this.exceptions.staticViolation()) {
          this.mxp.tracePreOpcode(frame);
        }
      }
      case CREATE -> {
        if (this.exceptions.noStackException() && !this.exceptions.staticViolation()) {
          this.mxp.tracePreOpcode(frame); // TODO: trigger in OoG
        }

        if (!this.exceptions.any() && this.callStack().getDepth() < 1024) {
          // TODO: check for failure: non empty byte code or non zero nonce (for the Deployed
          // Address)
          UInt256 value = UInt256.fromBytes(frame.getStackItem(0));
          if (frame
              .getWorldUpdater()
              .get(this.tx.transaction().getSender())
              .getBalance()
              .toUInt256()
              .greaterOrEqualThan(value)) {
            this.rlpAddr.tracePreOpcode(frame);
            this.romLex.tracePreOpcode(frame);
          }
        }
      }
      case CALL -> {
        if (!this.exceptions.any() && this.callStack().getDepth() < 1024) {
          this.romLex.tracePreOpcode(frame);
        }
      }
      case HALT -> {
        if (!this.exceptions.any() && this.callStack().getDepth() < 1024) {
          this.romLex.tracePreOpcode(frame);
        }
        if (this.exceptions.noStackException()
            && this.currentFrame().opCode() != OpCode.STOP
            && this.currentFrame().opCode() != OpCode.SELFDESTRUCT) {
          this.mxp.tracePreOpcode(frame);
        }
      }
      case INVALID -> {}
      default -> {}
    }
  }

  void processStateExec(MessageFrame frame) {
    this.currentFrame().frame(frame);
    this.state.stamps().stampHub();
    this.exceptions = Exceptions.forFrame(frame, Hub.gp);
    this.aborts = Aborts.forFrame(this);

    this.handleStack(frame);
    this.triggerModules(frame);
    if (this.exceptions.any() || this.currentFrame().opCode() == OpCode.REVERT) {
      this.callStack.revert(this.state.stamps().hub());
    }

    this.resetSignals();
    if (this.currentFrame().stack().isOk()) {
      this.traceOperation(frame);
    } else {
      this.addTraceSection(new StackOnlySection(this));
      // TODO: ‶return″ context line
    }
    this.state.stamps().stampSubmodules(this.signals());
  }

  void processStateFinal(WorldView worldView, Transaction tx, boolean isSuccess) {
    this.state.stamps().stampHub();

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
      // Trace the exceptions of a transaction that could not even start
      if (this.exceptions == null) {
        this.exceptions = Exceptions.fromOutOfGas();
      }

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
  public void enterTransaction() {
    this.state.enter();
    this.tx.enter();

    for (Module m : this.modules) {
      m.enterTransaction();
    }
  }

  @Override
  public void traceStartTx(final WorldView world, final Transaction tx) {
    this.enterTransaction();

    this.exceptions = Exceptions.empty();

    this.tx.update(tx);

    if (this.tx.shouldSkip(world)) {
      this.tx.state(TxState.TX_SKIP);
      this.processStateSkip(world);
      return;
    } else {
      this.tx.state(TxState.TX_WARM);
    }

    this.processStateWarm(world);
    this.processStateInit(world);

    for (Module m : this.modules) {
      m.traceStartTx(world, tx);
    }
  }

  @Override
  public void popTransaction() {
    this.tx.pop();
    this.state.pop();
    for (Module m : this.modules) {
      m.popTransaction();
    }
  }

  @Override
  public void traceEndTx(
      WorldView world, Transaction tx, boolean status, Bytes output, List<Log> logs, long gasUsed) {
    if (this.tx.state() != TxState.TX_SKIP) {
      this.tx.state(TxState.TX_FINAL);
    }
    this.tx.status(status);

    if (this.tx.state() != TxState.TX_SKIP) {
      this.processStateFinal(world, tx, status);
    }

    this.defers.runPostTx(this, world, tx);

    this.state.currentTxTrace().postTxRetcon(this);

    for (Module m : this.modules) {
      m.traceEndTx(world, tx, status, output, logs, gasUsed);
    }
  }

  private void unlatchStack(MessageFrame frame) {
    this.unlatchStack(frame, this.currentTraceSection());
  }

  public void unlatchStack(MessageFrame frame, TraceSection section) {
    if (this.currentFrame().pending() == null) {
      return;
    }

    StackContext pending = this.currentFrame().pending();
    for (int i = 0; i < pending.getLines().size(); i++) {
      StackLine line = pending.getLines().get(i);
      if (line.needsResult()) {
        EWord result = EWord.ZERO;
        // Only pop from the stack if no exceptions have been encountered
        if (!exceptions.any()) {
          result = EWord.of(frame.getStackItem(0));
        }

        // This works because we are certain that the stack chunks are the first.
        ((StackFragment) section.getLines().get(i).specific())
            .stackOps()
            .get(line.resultColumn() - 1)
            .setValue(result);
      }
    }

    if (this.exceptions.none()) {
      for (TraceSection.TraceLine line : section.getLines()) {
        if (line.specific() instanceof StackFragment stackFragment) {
          stackFragment.feedHashedValue(frame);
        }
      }
    }
  }

  @Override
  public void traceContextEnter(MessageFrame frame) {
    final boolean isDeployment = frame.getType() == MessageFrame.Type.CONTRACT_CREATION;
    final Address codeAddress = frame.getContractAddress();
    final CallFrameType frameType =
        frame.isStatic() ? CallFrameType.STATIC : CallFrameType.STANDARD;
    if (isDeployment) {
      this.conflation.deploymentInfo().markDeploying(codeAddress);
    }
    final int codeDeploymentNumber = this.conflation.deploymentInfo().number(codeAddress);
    this.callStack.enter(
        this.state.stamps().hub(),
        frame.getOriginatorAddress(), // TODO: check for all call types that it is correct
        frame.getContractAddress(),
        new Bytecode(frame.getCode().getBytes()),
        frameType,
        frame.getValue(),
        frame.getRemainingGas(),
        frame.getInputData(),
        this.conflation.deploymentInfo().number(codeAddress),
        codeDeploymentNumber,
        isDeployment);

    this.defers.runNextContext(this, frame);

    for (Module m : this.modules) {
      m.traceContextEnter(frame);
    }
  }

  public void traceContextReEnter(MessageFrame frame) {
    if (this.currentFrame().needsUnlatchingAtReEntry() != null) {
      this.unlatchStack(frame, this.currentFrame().needsUnlatchingAtReEntry());
      this.currentFrame().needsUnlatchingAtReEntry(null);
    }
  }

  @Override
  public void traceContextExit(MessageFrame frame) {
    conflation.deploymentInfo().unmarkDeploying(this.currentFrame().codeAddress());

    DeploymentExceptions contextExceptions =
        DeploymentExceptions.fromFrame(this.currentFrame(), frame);
    this.currentTraceSection().setContextExceptions(contextExceptions);
    if (contextExceptions.any()) {
      this.callStack.revert(this.state.stamps().hub());
    }

    this.callStack.exit(frame.getOutputData());

    for (Module m : this.modules) {
      m.traceContextExit(frame);
    }
  }

  @Override
  public void tracePreOpcode(final MessageFrame frame) {
    if (this.tx.state() == TxState.TX_SKIP) {
      return;
    }
    this.processStateExec(frame);
  }

  public void tracePostExecution(MessageFrame frame, Operation.OperationResult operationResult) {
    if (this.tx.state() == TxState.TX_SKIP) {
      return;
    }

    if (this.currentFrame().opCode().isCreate() && operationResult.getHaltReason() == null) {
      this.handleCreate(Words.toAddress(frame.getStackItem(0)));
    }

    this.defers.runPostExec(this, frame, operationResult);

    if (this.currentFrame().needsUnlatchingAtReEntry() == null) {
      this.unlatchStack(frame);
    }

    switch (this.opCodeData().instructionFamily()) {
      case ADD -> {
        if (this.exceptions.noStackException()) {
          this.add.tracePostOp(frame);
        }
      }
      case MOD -> {
        if (this.exceptions.noStackException()) {
          this.mod.tracePostOp(frame);
        }
      }
      case MUL -> {
        if (this.exceptions.noStackException()) {
          this.mul.tracePostOp(frame);
        }
      }
      case EXT -> {
        if (this.exceptions.noStackException()) {
          this.ext.tracePostOp(frame);
        }
      }
      case WCP -> {
        if (this.exceptions.noStackException()) {
          this.wcp.tracePostOp(frame);
        }
      }
      case BIN -> {}
      case SHF -> {
        if (this.exceptions.noStackException()) {
          this.shf.tracePostOp(frame);
        }
      }
      case KEC -> {}
      case CONTEXT -> {}
      case ACCOUNT -> {}
      case COPY -> {}
      case TRANSACTION -> {}
      case BATCH -> {}
      case STACK_RAM -> {
        if (this.exceptions.noStackException()) {
          this.mxp.tracePostOp(frame);
        }
      }
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

  private void handleCreate(Address target) {
    this.conflation.deploymentInfo().deploy(target);
  }

  @Override
  public void traceStartBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    this.block.update(blockHeader);
    for (Module m : this.modules) {
      m.traceStartBlock(blockHeader, blockBody);
    }
  }

  @Override
  public void traceStartConflation(long blockCount) {
    this.conflation.update();
    for (Module m : this.modules) {
      m.traceStartConflation(blockCount);
    }
  }

  @Override
  public void traceEndConflation() {
    this.state.postConflationRetcon(this);
    for (Module m : this.modules) {
      m.traceEndConflation();
    }
  }

  @Override
  public ModuleTrace commit() {
    final Trace.TraceBuilder trace = Trace.builder(this.lineCount());
    return new HubTrace(this.state.commit(trace).build());
  }

  public long refundedGas() {
    return this.state.currentTxTrace().refundedGas();
  }

  public long remainingGas() {
    return this.currentFrame().frame().getRemainingGas();
  }

  @Override
  public int lineCount() {
    return this.state.lineCount();
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
          BATCH,
          MACHINE_STATE,
          PUSH_POP,
          DUP,
          SWAP,
          HALT,
          INVALID -> this.addTraceSection(new StackOnlySection(this));
      case KEC -> {
        final boolean needsMmu = this.exceptions.none() && !frame.getStackItem(1).isZero();
        this.signals().wantMmu(needsMmu).wantMxp();

        this.addTraceSection(
            new KeccakSection(this, this.currentFrame(), new MiscFragment(this, frame)));
      }
      case CONTEXT, LOG -> this.addTraceSection(
          new ContextLogSection(
              this, new ContextFragment(this.callStack, this.currentFrame(), updateReturnData)));
      case ACCOUNT -> {
        TraceSection accountSection = new AccountSection(this);
        if (this.opCodeData().stackSettings().flag1()) {
          accountSection.addChunk(
              this,
              this.currentFrame(),
              new ContextFragment(this.callStack, this.currentFrame(), updateReturnData));
        }

        Address targetAddress =
            switch (this.currentFrame().opCode()) {
              case BALANCE, EXTCODESIZE, EXTCODEHASH -> Words.toAddress(frame.getStackItem(0));
              default -> Address.wrap(this.currentFrame().address());
            };
        Account targetAccount = frame.getWorldUpdater().getAccount(targetAddress);
        AccountSnapshot accountSnapshot =
            AccountSnapshot.fromAccount(
                targetAccount,
                frame.isAddressWarm(targetAddress),
                this.conflation.deploymentInfo().number(targetAddress),
                this.conflation.deploymentInfo().isDeploying(targetAddress));
        accountSection.addChunk(
            this,
            this.currentFrame(),
            new AccountFragment(accountSnapshot, accountSnapshot, false, 0, false));

        this.addTraceSection(accountSection);
      }
      case COPY -> { // TODO: call RomLex
        TraceSection copySection = new CopySection(this);
        if (this.opCodeData().stackSettings().flag1()) {
          Address targetAddress =
              switch (this.currentFrame().opCode()) {
                case CODECOPY -> this.currentFrame().codeAddress();
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
              this,
              this.currentFrame(),
              new AccountFragment(accountSnapshot, accountSnapshot, false, 0, false));
        } else {
          copySection.addChunk(
              this,
              this.currentFrame(),
              new ContextFragment(this.callStack, this.currentFrame(), updateReturnData));
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
        switch (this.currentFrame().opCode()) {
          case CALLDATALOAD -> {
            final long readOffset = Words.clampedToLong(frame.getStackItem(0));
            final boolean isOob = readOffset > this.currentFrame().callData().size();
            this.signals().wantMmu(!isOob && this.exceptions.none()).wantOob();

            final MiscFragment miscFragment = new MiscFragment(this, frame);
            this.defers.postExec(miscFragment);

            this.addTraceSection(
                new StackRam(
                    this,
                    miscFragment,
                    new ContextFragment(this.callStack(), this.currentFrame(), false)));
          }
          case MLOAD, MSTORE, MSTORE8 -> {
            this.signals().wantMmu(this.exceptions.none()).wantMxp();

            this.addTraceSection(new StackRam(this, new MiscFragment(this, frame)));
          }
          default -> throw new IllegalStateException("unexpected STACK_RAM opcode");
        }
      }
      case STORAGE -> {
        Address address = this.currentFrame().address();
        EWord key = EWord.of(frame.getStackItem(0));
        switch (this.currentFrame().opCode()) {
          case SSTORE -> {
            EWord valNext = EWord.of(frame.getStackItem(0));
            this.addTraceSection(
                new StorageSection(
                    this,
                    new ContextFragment(this.callStack, this.currentFrame(), updateReturnData),
                    new StorageFragment(
                        address,
                        this.currentFrame().accountDeploymentNumber(),
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
                        this.currentFrame().accountDeploymentNumber(),
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
        Address myAddress = this.currentFrame().address();
        Account myAccount = frame.getWorldUpdater().getAccount(myAddress);
        AccountSnapshot myAccountSnapshot =
            AccountSnapshot.fromAccount(
                myAccount,
                frame.isAddressWarm(myAddress),
                this.conflation.deploymentInfo().number(myAddress),
                this.conflation.deploymentInfo().isDeploying(myAddress));

        Address createdAddress = this.currentFrame().address();
        Account createdAccount = frame.getWorldUpdater().getAccount(createdAddress);
        AccountSnapshot createdAccountSnapshot =
            AccountSnapshot.fromAccount(
                createdAccount,
                frame.isAddressWarm(createdAddress),
                this.conflation.deploymentInfo().number(createdAddress),
                this.conflation.deploymentInfo().isDeploying(createdAddress));

        CreateSection createSection =
            new CreateSection(this, myAccountSnapshot, createdAccountSnapshot);
        // Will be traced in one (and only one!) of these depending on the success of the operation
        this.defers.postExec(createSection);
        this.defers.nextContext(createSection, currentFrame().id());
        this.addTraceSection(createSection);
        this.currentFrame().needsUnlatchingAtReEntry(createSection);
      }

      case CALL -> {
        final Address myAddress = this.currentFrame().address();
        final Account myAccount = frame.getWorldUpdater().getAccount(myAddress);
        final AccountSnapshot myAccountSnapshot =
            AccountSnapshot.fromAccount(
                myAccount,
                frame.isAddressWarm(myAddress),
                this.conflation.deploymentInfo().number(myAddress),
                this.conflation.deploymentInfo().isDeploying(myAddress));

        final Address calledAddress = Words.toAddress(frame.getStackItem(1));
        final Account calledAccount = frame.getWorldUpdater().getAccount(calledAddress);
        final boolean hasCode =
            Optional.ofNullable(calledAccount).map(AccountState::hasCode).orElse(false);

        final AccountSnapshot calledAccountSnapshot =
            AccountSnapshot.fromAccount(
                calledAccount,
                frame.isAddressWarm(myAddress),
                this.conflation.deploymentInfo().number(myAddress),
                this.conflation.deploymentInfo().isDeploying(myAddress));

        boolean targetIsPrecompile = isPrecompile(calledAddress);

        if (this.exceptions.any()) {
          if (this.exceptions.staticViolation()) {
            this.addTraceSection(
                new AbortedCallSection(
                    this,
                    this.currentFrame(),
                    new ContextFragment(this.callStack, this.currentFrame(), false),
                    new ContextFragment(this.callStack, this.callStack().parent(), true)));
          } else if (this.exceptions.outOfMemoryExpansion()) {
            this.signals().wantMxp();

            this.addTraceSection(
                new AbortedCallSection(
                    this,
                    this.currentFrame(),
                    new MiscFragment(this, frame),
                    new ContextFragment(this.callStack, this.callStack().parent(), true)));
          } else if (this.exceptions.outOfGas()) {
            this.signals().wantMxp().wantStipend();

            this.addTraceSection(
                new AbortedCallSection(
                    this,
                    this.currentFrame(),
                    new MiscFragment(this, frame),
                    new AccountFragment(calledAccountSnapshot, calledAccountSnapshot),
                    new ContextFragment(this.callStack, this.callStack().parent(), true)));
          }
        } else {
          Wei value = Wei.ZERO;
          if (this.currentFrame().opCode() == OpCode.CALL
              || this.currentFrame().opCode() == OpCode.CALLCODE) {
            value = Wei.wrap(frame.getStackItem(2));
          }
          final boolean abort =
              (myAccount.getBalance().lessThan(value)) || this.callStack().wouldOverflow();
          if (abort) {
            this.signals().wantMxp().wantOob().wantStipend();
            TraceSection abortedSection =
                new AbortedCallSection(
                    this,
                    this.currentFrame(),
                    new ScenarioFragment(
                        targetIsPrecompile,
                        hasCode,
                        true,
                        this.currentFrame().id(),
                        this.callStack().futureId()),
                    new ContextFragment(this.callStack, this.currentFrame(), false),
                    new MiscFragment(this, frame),
                    new AccountFragment(myAccountSnapshot, myAccountSnapshot),
                    new AccountFragment(calledAccountSnapshot, calledAccountSnapshot),
                    new ContextFragment(this.callStack, this.currentFrame(), true));
            this.addTraceSection(abortedSection);
          } else {
            this.signals().wantMxp().wantOob().wantStipend();
            final MiscFragment miscFragment = new MiscFragment(this, frame);

            if (hasCode) {
              final WithCodeCallSection section =
                  new WithCodeCallSection(
                      this, myAccountSnapshot, calledAccountSnapshot, miscFragment);
              this.defers.postExec(section);
              this.defers.nextContext(section, currentFrame().id());
              this.defers.postTx(section);
              this.addTraceSection(section);
              this.currentFrame().needsUnlatchingAtReEntry(section);
            } else {
              final NoCodeCallSection section =
                  new NoCodeCallSection(
                      this,
                      targetIsPrecompile,
                      myAccountSnapshot,
                      calledAccountSnapshot,
                      miscFragment);
              this.defers.postExec(section);
              this.defers.postTx(section);
              this.addTraceSection(section);
              this.currentFrame()
                  .needsUnlatchingAtReEntry(
                      section); // TODO: not sure there -- will we switch context?
            }
          }
        }
      }

      case JUMP -> {
        this.signals().wantOob();

        AccountSnapshot codeAccountSnapshot =
            AccountSnapshot.fromAccount(
                frame.getWorldUpdater().getAccount(this.currentFrame().codeAddress()),
                true,
                this.conflation.deploymentInfo().number(this.currentFrame().codeAddress()),
                this.currentFrame().underDeployment());

        JumpSection jumpSection =
            new JumpSection(
                this,
                new ContextFragment(this.callStack, this.currentFrame(), updateReturnData),
                new AccountFragment(codeAccountSnapshot, codeAccountSnapshot, false, 0, false),
                new MiscFragment(this, frame));

        this.addTraceSection(jumpSection);
      }
    }

    // In all cases, add a context fragment if an exception occurred
    if (this.exceptions.any()) {
      this.currentTraceSection()
          .addChunk(
              this,
              this.currentFrame(),
              new ContextFragment(this.callStack(), this.currentFrame(), true));
    }
  }

  public List<TraceFragment> makeStackChunks(CallFrame f) {
    List<TraceFragment> r = new ArrayList<>();
    if (f.pending().getLines().isEmpty()) {
      for (int i = 0; i < (this.opCodeData().stackSettings().twoLinesInstruction() ? 2 : 1); i++) {
        r.add(
            StackFragment.prepare(
                this.currentFrame().stack().snapshot(),
                new StackLine().asStackOperations(),
                this.exceptions.snapshot(),
                this.aborts.snapshot(),
                gp.of(f.frame(), f.opCode()),
                f.underDeployment()));
      }
    } else {
      for (StackLine line : f.pending().getLines()) {
        r.add(
            StackFragment.prepare(
                f.stack().snapshot(),
                line.asStackOperations(),
                this.exceptions.snapshot(),
                this.aborts.snapshot(),
                gp.of(f.frame(), f.opCode()),
                f.underDeployment()));
      }
    }
    return r;
  }
}
