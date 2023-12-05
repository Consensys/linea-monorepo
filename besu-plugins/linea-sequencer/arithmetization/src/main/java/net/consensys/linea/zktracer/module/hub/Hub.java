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

import static net.consensys.linea.zktracer.types.AddressUtils.isPrecompile;

import java.nio.MappedByteBuffer;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.Set;
import java.util.stream.Stream;

import lombok.Getter;
import lombok.experimental.Accessors;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.add.Add;
import net.consensys.linea.zktracer.module.ext.Ext;
import net.consensys.linea.zktracer.module.hub.defer.*;
import net.consensys.linea.zktracer.module.hub.fragment.*;
import net.consensys.linea.zktracer.module.hub.fragment.misc.MiscFragment;
import net.consensys.linea.zktracer.module.hub.section.*;
import net.consensys.linea.zktracer.module.logData.LogData;
import net.consensys.linea.zktracer.module.logInfo.LogInfo;
import net.consensys.linea.zktracer.module.mmu.Mmu;
import net.consensys.linea.zktracer.module.mod.Mod;
import net.consensys.linea.zktracer.module.mul.Mul;
import net.consensys.linea.zktracer.module.mxp.Mxp;
import net.consensys.linea.zktracer.module.preclimits.Blake2f;
import net.consensys.linea.zktracer.module.preclimits.EcAdd;
import net.consensys.linea.zktracer.module.preclimits.EcMul;
import net.consensys.linea.zktracer.module.preclimits.EcPairingCall;
import net.consensys.linea.zktracer.module.preclimits.EcRec;
import net.consensys.linea.zktracer.module.preclimits.EcpairingWeightedCall;
import net.consensys.linea.zktracer.module.preclimits.Modexp;
import net.consensys.linea.zktracer.module.preclimits.Rip160;
import net.consensys.linea.zktracer.module.preclimits.Sha256;
import net.consensys.linea.zktracer.module.rlpAddr.RlpAddr;
import net.consensys.linea.zktracer.module.rlp_txn.RlpTxn;
import net.consensys.linea.zktracer.module.rlp_txrcpt.RlpTxrcpt;
import net.consensys.linea.zktracer.module.rom.Rom;
import net.consensys.linea.zktracer.module.romLex.RomLex;
import net.consensys.linea.zktracer.module.shf.Shf;
import net.consensys.linea.zktracer.module.stp.Stp;
import net.consensys.linea.zktracer.module.tables.instructionDecoder.InstructionDecoder;
import net.consensys.linea.zktracer.module.tables.shf.ShfRt;
import net.consensys.linea.zktracer.module.trm.Trm;
import net.consensys.linea.zktracer.module.txn_data.TxnData;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.*;
import net.consensys.linea.zktracer.opcode.gas.projector.GasProjector;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.runtime.callstack.CallFrameType;
import net.consensys.linea.zktracer.runtime.callstack.CallStack;
import net.consensys.linea.zktracer.runtime.stack.ConflationInfo;
import net.consensys.linea.zktracer.runtime.stack.StackContext;
import net.consensys.linea.zktracer.runtime.stack.StackLine;
import net.consensys.linea.zktracer.types.EWord;
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
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;

@Slf4j
@Accessors(fluent = true)
public class Hub implements Module {
  private static final int TAU = 8;

  public static Optional<Bytes> maybeStackItem(MessageFrame frame, int idx) {
    if (frame.stackSize() > idx) {
      return Optional.of(frame.getStackItem(idx));
    } else {
      return Optional.empty();
    }
  }

  public static final GasProjector gp = new GasProjector();

  // Revertible state of the hub
  @Getter private final State state = new State();

  // Long-lived states
  @Getter ConflationInfo conflation = new ConflationInfo();
  @Getter BlockInfo block = new BlockInfo();
  @Getter TxInfo tx = new TxInfo();
  @Getter CallStack callStack = new CallStack();
  private final DeferRegistry defers = new DeferRegistry();

  // These attributes are transient (opcode-specific) and do not need to be
  // reversed.
  @Getter private final PlatformController pch;

  public int stamp() {
    return this.state.stamps().hub();
  }

  public OpCodeData opCodeData() {
    return this.currentFrame().opCodeData();
  }

  public OpCode opCode() {
    return this.currentFrame().opCode();
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

  private final Module add = new Add(this);
  private final Module ext = new Ext();
  private final Mod mod = new Mod();
  private final Module mul = new Mul(this);
  private final Module shf = new Shf();
  private final Wcp wcp = new Wcp();
  private final RlpTxn rlpTxn;
  private final Module mxp;
  private final Mmu mmu;
  private final RlpTxrcpt rlpTxrcpt = new RlpTxrcpt();
  private final LogInfo logInfo = new LogInfo(rlpTxrcpt);
  private final LogData logData = new LogData(rlpTxrcpt);
  private final RlpAddr rlpAddr = new RlpAddr(this);
  private final Rom rom;
  private final RomLex romLex;
  private final TxnData txnData;
  private final Trm trm = new Trm();
  private final Modexp modexp;
  private final Stp stp = new Stp(this, wcp, mod);

  private final List<Module> modules;
  /* Those modules are not traced, we just compute the number of calls to those precompile to meet the prover limits */
  private final List<Module> precompileLimitModules;

  public Hub() {
    this.pch = new PlatformController(this);
    this.mmu = new Mmu(this.callStack);
    this.mxp = new Mxp(this);
    this.romLex = new RomLex(this);
    this.rom = new Rom(this.romLex);
    this.rlpTxn = new RlpTxn(this.romLex);
    this.txnData = new TxnData(this, this.romLex, this.wcp);

    this.modexp = new Modexp(this);
    final EcPairingCall ecpairingCall = new EcPairingCall(this);
    this.precompileLimitModules =
        List.of(
            new Sha256(this),
            new EcRec(this),
            new Rip160(this),
            this.modexp,
            new EcAdd(this),
            new EcMul(this),
            ecpairingCall,
            new EcpairingWeightedCall(ecpairingCall),
            new Blake2f(this));

    this.modules =
        Stream.concat(
                Stream.of(
                    this.romLex, // WARN: must be called first
                    this.add,
                    this.ext,
                    this.logData,
                    this.logInfo,
                    this.mod,
                    this.mul,
                    this.mxp,
                    this.rlpAddr,
                    this.rlpTxn,
                    this.rlpTxrcpt,
                    this.rom,
                    this.shf,
                    this.trm,
                    this.txnData,
                    this.stp,
                    this.wcp),
                this.precompileLimitModules.stream())
            .toList();
  }

  /**
   * @return a list of all modules for which to generate traces
   */
  public List<Module> getModulesToTrace() {
    return List.of(
        // Reference tables
        new InstructionDecoder(),
        new ShfRt(),
        // Modules
        this,
        this.add,
        this.ext,
        this.logData,
        this.logInfo,
        this.mod,
        this.mul,
        this.mxp,
        this.rlpAddr,
        this.rlpTxn,
        this.rlpTxrcpt,
        this.rom,
        this.romLex,
        this.shf,
        this.stp,
        this.txnData,
        this.wcp);
  }

  public List<Module> getModulesToCount() {
    return Stream.concat(
            Stream.of(
                this,
                this.romLex,
                this.add,
                this.ext,
                this.logData,
                this.logInfo,
                this.mod,
                this.mmu,
                this.mul,
                this.mxp,
                this.rlpAddr,
                this.rlpTxn,
                this.rlpTxrcpt,
                this.rom,
                this.shf,
                this.trm,
                this.txnData,
                this.wcp),
            this.precompileLimitModules.stream())
        .toList();
  }

  @Override
  public String moduleKey() {
    return "HUB";
  }

  public static boolean isValidPrecompileCall(MessageFrame frame, final OpCode opCode) {
    return switch (opCode) {
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
    if (isDeployment) {
      this.conflation().deploymentInfo().deploy(toAddress);
    }
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
    final boolean isDeployment = this.tx.transaction().getTo().isEmpty();
    final Address toAddress = effectiveToAddress(this.tx.transaction());
    if (isDeployment) {
      this.conflation().deploymentInfo().deploy(toAddress);
    }
    this.tx.state(TxState.TX_EXEC);
  }

  public CallFrame currentFrame() {
    if (this.callStack().isEmpty()) {
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
    for (Module precompileLimit : this.precompileLimitModules) {
      precompileLimit.tracePreOpcode(frame);
    }

    if (this.pch.signals().romLex()) {
      this.romLex.tracePreOpcode(frame);
    }
    if (this.pch.signals().add()) {
      this.add.tracePreOpcode(frame);
    }
    if (this.pch.signals().rlpAddr()) {
      this.rlpAddr.tracePreOpcode(frame);
    }
    if (this.pch.signals().mul()) {
      this.mul.tracePreOpcode(frame);
    }
    if (this.pch.signals().ext()) {
      this.ext.tracePreOpcode(frame);
    }
    if (this.pch.signals().mod()) {
      this.mod.tracePreOpcode(frame);
    }
    if (this.pch.signals().wcp()) {
      this.wcp.tracePreOpcode(frame);
    }
    if (this.pch.signals().shf()) {
      this.shf.tracePreOpcode(frame);
    }
    if (this.pch.signals().mmu()) {
      this.mmu.tracePreOpcode(frame);
    }

    if (this.pch.signals().mxp()) {
      this.mxp.tracePreOpcode(frame);
    }
    if (this.pch.signals().oob()) {
      // TODO: this.oob.tracePreOpcode(frame);
    }
    if (this.pch.signals().stp()) {
      this.stp.tracePreOpcode(frame);
    }
    if (this.pch.signals().exp()) {
      this.modexp.tracePreOpcode(frame);
    }
    if (this.pch.signals().trm()) {
      this.trm.tracePreOpcode(frame);
    }
    if (this.pch.signals().hashInfo()) {
      // TODO: this.hashInfo.tracePreOpcode(frame);
    }
  }

  void processStateExec(MessageFrame frame) {
    this.currentFrame().frame(frame);
    this.state.stamps().stampHub();
    this.pch.setup(frame);

    this.handleStack(frame);
    this.triggerModules(frame);
    if (this.pch().exceptions().any() || this.currentFrame().opCode() == OpCode.REVERT) {
      this.callStack.revert(this.state.stamps().hub());
    }

    if (this.currentFrame().stack().isOk()) {
      this.traceOperation(frame);
    } else {
      this.addTraceSection(new StackOnlySection(this));
      // TODO: ‶return″ context line
    }
    this.state.stamps().stampSubmodules(this.pch());
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
      // TODO: integrate with PCH
      // if (this.exceptions == null) {
      // this.exceptions = Exceptions.fromOutOfGas();
      // }

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
    this.pch.reset();
    this.tx.update(tx);

    this.enterTransaction();

    if (this.tx.shouldSkip(world)) {
      this.tx.state(TxState.TX_SKIP);
      this.processStateSkip(world);
    } else {
      this.tx.state(TxState.TX_WARM);
      this.processStateWarm(world);
      this.processStateInit(world);
    }

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
        Bytes result = Bytes.EMPTY;
        // Only pop from the stack if no exceptions have been encountered
        if (this.pch.exceptions().none()) {
          result = frame.getStackItem(0).copy();
        }

        // This works because we are certain that the stack chunks are the first.
        ((StackFragment) section.getLines().get(i).specific())
            .stackOps()
            .get(line.resultColumn() - 1)
            .setValue(result);
      }
    }

    if (this.pch.exceptions().none()) {
      for (TraceSection.TraceLine line : section.getLines()) {
        if (line.specific() instanceof StackFragment stackFragment) {
          stackFragment.feedHashedValue(frame);
        }
      }
    }
  }

  @Override
  public void traceContextEnter(MessageFrame frame) {
    this.pch.reset();

    if (frame.getDepth() == 0) {
      // Bedrock...
      final Address toAddress = effectiveToAddress(this.tx.transaction());
      final boolean isDeployment = this.tx.transaction().getTo().isEmpty();
      this.callStack.newBedrock(
          this.state.stamps().hub(),
          this.tx.transaction().getSender(),
          toAddress,
          isDeployment ? CallFrameType.INIT_CODE : CallFrameType.STANDARD,
          new Bytecode(
              toAddress == null
                  ? this.tx.transaction().getData().orElse(Bytes.EMPTY)
                  : Optional.ofNullable(frame.getWorldUpdater().get(toAddress))
                      .map(AccountState::getCode)
                      .orElse(Bytes.EMPTY)), // TODO: see with Olivier
          Wei.of(this.tx.transaction().getValue().getAsBigInteger()),
          this.tx.transaction().getGasLimit(),
          this.tx.transaction().getData().orElse(Bytes.EMPTY),
          this.conflation.deploymentInfo().number(toAddress),
          toAddress.isEmpty() ? 0 : this.conflation.deploymentInfo().number(toAddress),
          this.conflation.deploymentInfo().isDeploying(toAddress));
    } else {
      // ...or CALL
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
          frame.getRecipientAddress(), // TODO: check for all call types that it is correct
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
  }

  public void traceContextReEnter(MessageFrame frame) {
    if (this.currentFrame().needsUnlatchingAtReEntry() != null) {
      this.unlatchStack(frame, this.currentFrame().needsUnlatchingAtReEntry());
      this.currentFrame().needsUnlatchingAtReEntry(null);
    }
  }

  @Override
  public void traceContextExit(MessageFrame frame) {
    if (frame.getDepth() > 0) {
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
    this.romLex.tracePostExecution(frame, operationResult);

    if (this.currentFrame().needsUnlatchingAtReEntry() == null) {
      this.unlatchStack(frame);
    }

    switch (this.opCodeData().instructionFamily()) {
      case ADD -> {
        if (this.pch.exceptions().noStackException()) {
          this.add.tracePostOp(frame);
        }
      }
      case MOD -> {
        if (this.pch.exceptions().noStackException()) {
          this.mod.tracePostOp(frame);
        }
      }
      case MUL -> {
        if (this.pch.exceptions().noStackException()) {
          this.mul.tracePostOp(frame);
        }
      }
      case EXT -> {
        if (this.pch.exceptions().noStackException()) {
          this.ext.tracePostOp(frame);
        }
      }
      case WCP -> {
        if (this.pch.exceptions().noStackException()) {
          this.wcp.tracePostOp(frame);
        }
      }
      case BIN -> {}
      case SHF -> {
        if (this.pch.exceptions().noStackException()) {
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
        if (this.pch.exceptions().noStackException()) {
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
  public void traceStartBlock(final ProcessableBlockHeader processableBlockHeader) {
    this.block.update(processableBlockHeader);
    for (Module m : this.modules) {
      m.traceStartBlock(processableBlockHeader);
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
  public List<ColumnHeader> columnsHeaders() {
    return Trace.headers(this.lineCount());
  }

  @Override
  public void commit(List<MappedByteBuffer> buffers) {
    final Trace trace = new Trace(buffers);
    this.state.commit(trace);
  }

  public long refundedGas() {
    return this.state.currentTxTrace().refundedGas();
  }

  public long remainingGas() {
    return 0; // TODO:
  }

  @Override
  public int lineCount() {
    return this.state.lineCount();
  }

  void traceOperation(MessageFrame frame) {
    boolean updateReturnData =
        this.opCodeData().isHalt()
            || this.opCodeData().isInvalid()
            || this.pch().exceptions().any()
            || isValidPrecompileCall(frame, this.currentFrame().opCode());

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
        Account targetAccount = frame.getWorldUpdater().get(targetAddress);
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
          Account targetAccount = frame.getWorldUpdater().get(targetAddress);
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

            final MiscFragment miscFragment = new MiscFragment(this, frame);
            this.defers.postExec(miscFragment);

            this.addTraceSection(
                new StackRam(
                    this,
                    miscFragment,
                    new ContextFragment(this.callStack(), this.currentFrame(), false)));
          }
          case MLOAD, MSTORE, MSTORE8 -> {
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
        Account myAccount = frame.getWorldUpdater().get(myAddress);
        AccountSnapshot myAccountSnapshot =
            AccountSnapshot.fromAccount(
                myAccount,
                frame.isAddressWarm(myAddress),
                this.conflation.deploymentInfo().number(myAddress),
                this.conflation.deploymentInfo().isDeploying(myAddress));

        Address createdAddress = this.currentFrame().address();
        Account createdAccount = frame.getWorldUpdater().get(createdAddress);
        AccountSnapshot createdAccountSnapshot =
            AccountSnapshot.fromAccount(
                createdAccount,
                frame.isAddressWarm(createdAddress),
                this.conflation.deploymentInfo().number(createdAddress),
                this.conflation.deploymentInfo().isDeploying(createdAddress));

        CreateSection createSection =
            new CreateSection(this, myAccountSnapshot, createdAccountSnapshot);
        // Will be traced in one (and only one!) of these depending on the success of
        // the operation
        this.defers.postExec(createSection);
        this.defers.nextContext(createSection, currentFrame().id());
        this.addTraceSection(createSection);
        this.currentFrame().needsUnlatchingAtReEntry(createSection);
      }

      case CALL -> {
        final Address myAddress = this.currentFrame().address();
        final Account myAccount = frame.getWorldUpdater().get(myAddress);
        final AccountSnapshot myAccountSnapshot =
            AccountSnapshot.fromAccount(
                myAccount,
                frame.isAddressWarm(myAddress),
                this.conflation.deploymentInfo().number(myAddress),
                this.conflation.deploymentInfo().isDeploying(myAddress));

        final Address calledAddress = Words.toAddress(frame.getStackItem(1));
        final Account calledAccount = frame.getWorldUpdater().get(calledAddress);
        final boolean hasCode =
            Optional.ofNullable(calledAccount).map(AccountState::hasCode).orElse(false);

        final AccountSnapshot calledAccountSnapshot =
            AccountSnapshot.fromAccount(
                calledAccount,
                frame.isAddressWarm(myAddress),
                this.conflation.deploymentInfo().number(myAddress),
                this.conflation.deploymentInfo().isDeploying(myAddress));

        boolean targetIsPrecompile = isPrecompile(calledAddress);

        if (this.pch().exceptions().any()) {
          if (this.pch().exceptions().staticFault()) {
            this.addTraceSection(
                new AbortedCallSection(
                    this,
                    this.currentFrame(),
                    new ContextFragment(this.callStack, this.currentFrame(), false),
                    new ContextFragment(this.callStack, this.callStack().parent(), true)));
          } else if (this.pch().exceptions().outOfMemoryExpansion()) {
            this.pch().signals().wantMxp();
            this.addTraceSection(
                new AbortedCallSection(
                    this,
                    this.currentFrame(),
                    new MiscFragment(this, frame),
                    new ContextFragment(this.callStack, this.callStack().parent(), true)));
          } else if (this.pch().exceptions().outOfGas()) {
            this.pch().signals().wantMxp().wantStipend();
            this.addTraceSection(
                new AbortedCallSection(
                    this,
                    this.currentFrame(),
                    new MiscFragment(this, frame),
                    new AccountFragment(calledAccountSnapshot, calledAccountSnapshot),
                    new ContextFragment(this.callStack, this.callStack().parent(), true)));
          }
        } else {
          this.pch().signals().wantMxp().wantOob().wantStipend();
          if (this.pch.aborts().any()) {
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
        AccountSnapshot codeAccountSnapshot =
            AccountSnapshot.fromAccount(
                frame.getWorldUpdater().get(this.currentFrame().codeAddress()),
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
    if (this.pch().exceptions().any()) {
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
                this.pch.exceptions().snapshot(),
                this.pch.aborts().snapshot(),
                gp.of(f.frame(), f.opCode()),
                f.underDeployment()));
      }
    } else {
      for (StackLine line : f.pending().getLines()) {
        r.add(
            StackFragment.prepare(
                f.stack().snapshot(),
                line.asStackOperations(),
                this.pch.exceptions().snapshot(),
                this.pch.aborts().snapshot(),
                gp.of(f.frame(), f.opCode()),
                f.underDeployment()));
      }
    }
    return r;
  }
}
