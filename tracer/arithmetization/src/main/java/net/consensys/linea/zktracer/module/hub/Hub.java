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

import static net.consensys.linea.zktracer.module.hub.HubProcessingPhase.TX_EXEC;
import static net.consensys.linea.zktracer.module.hub.HubProcessingPhase.TX_FINL;
import static net.consensys.linea.zktracer.module.hub.HubProcessingPhase.TX_INIT;
import static net.consensys.linea.zktracer.module.hub.HubProcessingPhase.TX_SKIP;
import static net.consensys.linea.zktracer.module.hub.HubProcessingPhase.TX_WARM;
import static net.consensys.linea.zktracer.module.hub.Trace.MULTIPLIER___STACK_HEIGHT;
import static net.consensys.linea.zktracer.types.AddressUtils.effectiveToAddress;

import java.nio.MappedByteBuffer;
import java.util.HashMap;
import java.util.List;
import java.util.Optional;
import java.util.Set;
import java.util.stream.Stream;

import com.google.common.base.Preconditions;
import lombok.Getter;
import lombok.experimental.Accessors;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.ColumnHeader;
import net.consensys.linea.zktracer.container.module.Module;
import net.consensys.linea.zktracer.module.add.Add;
import net.consensys.linea.zktracer.module.bin.Bin;
import net.consensys.linea.zktracer.module.blake2fmodexpdata.BlakeModexpData;
import net.consensys.linea.zktracer.module.blockdata.Blockdata;
import net.consensys.linea.zktracer.module.blockhash.Blockhash;
import net.consensys.linea.zktracer.module.ecdata.EcData;
import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.exp.Exp;
import net.consensys.linea.zktracer.module.ext.Ext;
import net.consensys.linea.zktracer.module.gas.Gas;
import net.consensys.linea.zktracer.module.hub.defer.DeferRegistry;
import net.consensys.linea.zktracer.module.hub.fragment.ContextFragment;
import net.consensys.linea.zktracer.module.hub.fragment.StackFragment;
import net.consensys.linea.zktracer.module.hub.section.AccountSection;
import net.consensys.linea.zktracer.module.hub.section.CallDataLoadSection;
import net.consensys.linea.zktracer.module.hub.section.ContextSection;
import net.consensys.linea.zktracer.module.hub.section.CreateSection;
import net.consensys.linea.zktracer.module.hub.section.ExpSection;
import net.consensys.linea.zktracer.module.hub.section.JumpSection;
import net.consensys.linea.zktracer.module.hub.section.KeccakSection;
import net.consensys.linea.zktracer.module.hub.section.LogSection;
import net.consensys.linea.zktracer.module.hub.section.SloadSection;
import net.consensys.linea.zktracer.module.hub.section.SstoreSection;
import net.consensys.linea.zktracer.module.hub.section.StackOnlySection;
import net.consensys.linea.zktracer.module.hub.section.StackRamSection;
import net.consensys.linea.zktracer.module.hub.section.TraceSection;
import net.consensys.linea.zktracer.module.hub.section.TransactionSection;
import net.consensys.linea.zktracer.module.hub.section.TxFinalizationSection;
import net.consensys.linea.zktracer.module.hub.section.TxInitializationSection;
import net.consensys.linea.zktracer.module.hub.section.TxPreWarmingMacroSection;
import net.consensys.linea.zktracer.module.hub.section.TxSkippedSection;
import net.consensys.linea.zktracer.module.hub.section.call.CallSection;
import net.consensys.linea.zktracer.module.hub.section.copy.CallDataCopySection;
import net.consensys.linea.zktracer.module.hub.section.copy.CodeCopySection;
import net.consensys.linea.zktracer.module.hub.section.copy.ExtCodeCopySection;
import net.consensys.linea.zktracer.module.hub.section.copy.ReturnDataCopySection;
import net.consensys.linea.zktracer.module.hub.section.halt.ReturnSection;
import net.consensys.linea.zktracer.module.hub.section.halt.RevertSection;
import net.consensys.linea.zktracer.module.hub.section.halt.SelfdestructSection;
import net.consensys.linea.zktracer.module.hub.section.halt.StopSection;
import net.consensys.linea.zktracer.module.hub.signals.Exceptions;
import net.consensys.linea.zktracer.module.hub.signals.PlatformController;
import net.consensys.linea.zktracer.module.hub.transients.Transients;
import net.consensys.linea.zktracer.module.limits.Keccak;
import net.consensys.linea.zktracer.module.limits.L2Block;
import net.consensys.linea.zktracer.module.limits.L2L1Logs;
import net.consensys.linea.zktracer.module.limits.precompiles.BlakeEffectiveCall;
import net.consensys.linea.zktracer.module.limits.precompiles.BlakeRounds;
import net.consensys.linea.zktracer.module.limits.precompiles.EcAddEffectiveCall;
import net.consensys.linea.zktracer.module.limits.precompiles.EcMulEffectiveCall;
import net.consensys.linea.zktracer.module.limits.precompiles.EcPairingFinalExponentiations;
import net.consensys.linea.zktracer.module.limits.precompiles.EcPairingG2MembershipCalls;
import net.consensys.linea.zktracer.module.limits.precompiles.EcPairingMillerLoops;
import net.consensys.linea.zktracer.module.limits.precompiles.EcRecoverEffectiveCall;
import net.consensys.linea.zktracer.module.limits.precompiles.ModexpEffectiveCall;
import net.consensys.linea.zktracer.module.limits.precompiles.RipemdBlocks;
import net.consensys.linea.zktracer.module.limits.precompiles.Sha256Blocks;
import net.consensys.linea.zktracer.module.logdata.LogData;
import net.consensys.linea.zktracer.module.loginfo.LogInfo;
import net.consensys.linea.zktracer.module.mmio.Mmio;
import net.consensys.linea.zktracer.module.mmu.Mmu;
import net.consensys.linea.zktracer.module.mod.Mod;
import net.consensys.linea.zktracer.module.mul.Mul;
import net.consensys.linea.zktracer.module.mxp.Mxp;
import net.consensys.linea.zktracer.module.oob.Oob;
import net.consensys.linea.zktracer.module.rlpaddr.RlpAddr;
import net.consensys.linea.zktracer.module.rlptxn.RlpTxn;
import net.consensys.linea.zktracer.module.rlptxrcpt.RlpTxnRcpt;
import net.consensys.linea.zktracer.module.rom.Rom;
import net.consensys.linea.zktracer.module.romlex.ContractMetadata;
import net.consensys.linea.zktracer.module.romlex.RomLex;
import net.consensys.linea.zktracer.module.shakiradata.ShakiraData;
import net.consensys.linea.zktracer.module.shf.Shf;
import net.consensys.linea.zktracer.module.stp.Stp;
import net.consensys.linea.zktracer.module.tables.bin.BinRt;
import net.consensys.linea.zktracer.module.tables.instructionDecoder.InstructionDecoder;
import net.consensys.linea.zktracer.module.tables.shf.ShfRt;
import net.consensys.linea.zktracer.module.trm.Trm;
import net.consensys.linea.zktracer.module.txndata.TxnData;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.OpCode;
import net.consensys.linea.zktracer.opcode.OpCodeData;
import net.consensys.linea.zktracer.opcode.gas.projector.GasProjector;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.runtime.callstack.CallFrameType;
import net.consensys.linea.zktracer.runtime.callstack.CallStack;
import net.consensys.linea.zktracer.runtime.stack.StackContext;
import net.consensys.linea.zktracer.runtime.stack.StackLine;
import net.consensys.linea.zktracer.types.Bytecode;
import net.consensys.linea.zktracer.types.MemorySpan;
import net.consensys.linea.zktracer.types.TransactionProcessingMetadata;
import org.apache.tuweni.bytes.Bytes;
import org.hyperledger.besu.datatypes.Address;
import org.hyperledger.besu.datatypes.Transaction;
import org.hyperledger.besu.datatypes.Wei;
import org.hyperledger.besu.evm.account.AccountState;
import org.hyperledger.besu.evm.frame.MessageFrame;
import org.hyperledger.besu.evm.internal.Words;
import org.hyperledger.besu.evm.log.Log;
import org.hyperledger.besu.evm.log.LogTopic;
import org.hyperledger.besu.evm.operation.Operation;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.data.BlockBody;
import org.hyperledger.besu.plugin.data.BlockHeader;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;

@Slf4j
@Accessors(fluent = true)
public class Hub implements Module {

  public static final GasProjector GAS_PROJECTOR = new GasProjector();

  /** accumulate the trace information for the Hub */
  @Getter public final State state = new State();

  /** contain the factories for trace segments that need complex initialization */
  @Getter private final Factories factories = new Factories(this);

  /** provides phase-related volatile information */
  @Getter Transients transients = new Transients(this);

  /**
   * Long-lived states, not used in tracing per se but keeping track of data of the associated
   * lifetime
   */
  @Getter CallStack callStack = new CallStack();

  /** Stores the transaction Metadata of all the transaction of the conflated block */
  @Getter TransactionStack txStack = new TransactionStack();

  /** Stores all the actions that must be deferred to a later time */
  @Getter private final DeferRegistry defers = new DeferRegistry();

  /** stores all data related to failure states & module activation */
  @Getter private final PlatformController pch = new PlatformController(this);

  @Override
  public String moduleKey() {
    return "HUB";
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

  @Override
  public int lineCount() {
    return state.lineCounter().lineCount();
  }

  /** List of all modules of the ZK-evm */
  // stateless modules
  @Getter private final Wcp wcp = new Wcp();

  private final Add add = new Add();
  private final Bin bin = new Bin();
  private final Blockhash blockhash = new Blockhash(wcp);
  private final Euc euc = new Euc(wcp);
  @Getter private final Ext ext = new Ext(this);
  private final Gas gas = new Gas();
  private final Mul mul = new Mul(this);
  private final Mod mod = new Mod();
  private final Shf shf = new Shf();
  @Getter private final Trm trm = new Trm();

  // other
  private final Blockdata blockdata;
  @Getter private final RomLex romLex = new RomLex(this);
  private final Rom rom = new Rom(romLex);
  private final RlpTxn rlpTxn = new RlpTxn(romLex);
  private final Mmio mmio;

  private final TxnData txnData = new TxnData(wcp, euc);
  private final RlpTxnRcpt rlpTxnRcpt = new RlpTxnRcpt();
  private final LogInfo logInfo = new LogInfo(rlpTxnRcpt);
  private final LogData logData = new LogData(rlpTxnRcpt);
  private final RlpAddr rlpAddr = new RlpAddr(this, trm);

  // modules triggered by sub-fragments of the MISCELLANEOUS / IMC perspective
  @Getter private final Mxp mxp = new Mxp();
  @Getter private final Oob oob = new Oob(this, add, mod, wcp);
  @Getter private final Mmu mmu;
  @Getter private final Stp stp = new Stp(wcp, mod);
  @Getter private final Exp exp = new Exp(this, wcp);

  /*
   * Those modules are not traced, we just compute the number of calls to those
   * precompile to meet the prover limits
   */
  private final Keccak keccak;

  private final Sha256Blocks sha256Blocks = new Sha256Blocks();

  private final EcAddEffectiveCall ecAddEffectiveCall = new EcAddEffectiveCall();
  private final EcMulEffectiveCall ecMulEffectiveCall = new EcMulEffectiveCall();
  private final EcRecoverEffectiveCall ecRecoverEffectiveCall = new EcRecoverEffectiveCall();

  private final EcPairingG2MembershipCalls ecPairingG2MembershipCalls =
      new EcPairingG2MembershipCalls();
  private final EcPairingMillerLoops ecPairingMillerLoops = new EcPairingMillerLoops();
  private final EcPairingFinalExponentiations ecPairingFinalExponentiations =
      new EcPairingFinalExponentiations();

  @Getter private final ModexpEffectiveCall modexpEffectiveCall = new ModexpEffectiveCall();

  private final RipemdBlocks ripemdBlocks = new RipemdBlocks();

  private final BlakeEffectiveCall blakeEffectiveCall = new BlakeEffectiveCall();
  private final BlakeRounds blakeRounds = new BlakeRounds();

  private List<Module> precompileLimitModules() {

    return List.of(
        keccak,
        sha256Blocks,
        ecAddEffectiveCall,
        ecMulEffectiveCall,
        ecRecoverEffectiveCall,
        ecPairingG2MembershipCalls,
        ecPairingMillerLoops,
        ecPairingFinalExponentiations,
        modexpEffectiveCall,
        ripemdBlocks,
        blakeEffectiveCall,
        blakeRounds);
  }

  /*
   * precompile-data modules
   * those module are traced (and could be count)
   */
  @Getter private final ShakiraData shakiraData;

  @Getter
  private final BlakeModexpData blakeModexpData =
      new BlakeModexpData(this.wcp, modexpEffectiveCall, blakeEffectiveCall, blakeRounds);

  @Getter
  public final EcData ecData =
      new EcData(
          wcp,
          ext,
          ecAddEffectiveCall,
          ecMulEffectiveCall,
          ecRecoverEffectiveCall,
          ecPairingG2MembershipCalls,
          ecPairingMillerLoops,
          ecPairingFinalExponentiations);

  private final L2Block l2Block;
  private final L2L1Logs l2L1Logs;

  /** list of module than can be modified during execution */
  private final List<Module> modules;

  /** reference table modules */
  private final List<Module> refTableModules;

  /**
   * @return a list of all modules for which to generate traces
   */
  public List<Module> getModulesToTrace() {
    return Stream.concat(
            this.refTableModules.stream(),
            // Modules
            Stream.of(
                this,
                add,
                bin,
                blakeModexpData,
                ecData,
                blockdata,
                blockhash,
                ext,
                euc,
                exp,
                logData,
                logInfo,
                mmu, // WARN: must be traced before the MMIO
                mmio,
                mod,
                mul,
                mxp,
                oob,
                rlpAddr,
                rlpTxn,
                rlpTxnRcpt,
                rom,
                romLex,
                shakiraData,
                shf,
                stp,
                trm,
                txnData,
                wcp))
        .toList();
  }

  /**
   * List all the modules for which to generate counters. Intersects with, but is not equal to
   * {@code getModulesToTrace}.
   *
   * @return the modules to count
   */
  public List<Module> getModulesToCount() {
    return Stream.concat(
            Stream.of(
                this,
                romLex,
                add,
                bin,
                blockdata,
                blockhash,
                ext,
                ecData,
                euc,
                mmu,
                mmio,
                logData,
                logInfo,
                mod,
                mul,
                mxp,
                oob,
                exp,
                rlpAddr,
                rlpTxn,
                rlpTxnRcpt,
                rom,
                shf,
                trm,
                txnData,
                wcp,
                l2Block),
            precompileLimitModules().stream())
        .toList();
  }

  public Hub(final Address l2l1ContractAddress, final Bytes l2l1Topic) {
    this.l2Block = new L2Block(l2l1ContractAddress, LogTopic.of(l2l1Topic));
    this.l2L1Logs = new L2L1Logs(l2Block); // TODO: we never use it, to delete ?
    this.keccak = new Keccak(ecRecoverEffectiveCall, l2Block);
    this.shakiraData = new ShakiraData(wcp, sha256Blocks, keccak, ripemdBlocks);
    this.blockdata = new Blockdata(wcp, txnData, rlpTxn);
    this.mmu = new Mmu(euc, wcp);
    this.mmio = new Mmio(mmu);

    this.refTableModules = List.of(new BinRt(), new InstructionDecoder(), new ShfRt());

    this.modules =
        Stream.concat(
                Stream.of(
                    add,
                    bin,
                    blakeModexpData,
                    blockhash,
                    ecData,
                    euc,
                    ext,
                    gas,
                    mmio,
                    mmu,
                    mod,
                    mul,
                    mxp,
                    oob,
                    exp,
                    rlpAddr,
                    rlpTxn,
                    rlpTxnRcpt,
                    logData, /* WARN: must be called AFTER rlpTxnRcpt */
                    logInfo, /* WARN: must be called AFTER rlpTxnRcpt */
                    rom,
                    romLex,
                    shakiraData,
                    shf,
                    stp,
                    trm,
                    wcp, /* WARN: must be called BEFORE txnData */
                    txnData,
                    blockdata /* WARN: must be called AFTER txnData */),
                precompileLimitModules().stream())
            .toList();
  }

  @Override
  public void enterTransaction() {
    for (Module m : this.modules) {
      m.enterTransaction();
    }
  }

  @Override
  public void popTransaction() {
    this.txStack.pop();
    this.state.pop();
    for (Module m : this.modules) {
      m.popTransaction();
    }
  }

  /** Tracing Operation, triggered by Besu hook */
  @Override
  public void traceStartConflation(long blockCount) {
    for (Module m : this.modules) {
      m.traceStartConflation(blockCount);
    }
  }

  @Override
  public void traceEndConflation(final WorldView world) {
    this.romLex.determineCodeFragmentIndex();
    this.txStack.setCodeFragmentIndex(this);
    this.defers.resolvePostConflation(this, world);

    for (Module m : this.modules) {
      m.traceEndConflation(world);
    }
  }

  @Override
  public void traceStartBlock(final ProcessableBlockHeader processableBlockHeader) {
    this.state.firstAndLastStorageSlotOccurrences.add(new HashMap<>());
    this.transients().block().update(processableBlockHeader);
    this.txStack.resetBlock();
    for (Module m : this.modules) {
      m.traceStartBlock(processableBlockHeader);
    }
  }

  @Override
  public void traceEndBlock(final BlockHeader blockHeader, final BlockBody blockBody) {
    for (Module m : this.modules) {
      m.traceEndBlock(blockHeader, blockBody);
    }
  }

  public void traceStartTransaction(final WorldView world, final Transaction tx) {
    pch.reset();
    state.enter();
    txStack.enterTransaction(world, tx, transients.block());

    defers.scheduleForPostTransaction(txStack.current());

    this.enterTransaction();

    if (!txStack.current().requiresEvmExecution()) {
      state.setProcessingPhase(TX_SKIP);
      new TxSkippedSection(this, world, this.txStack.current(), this.transients);
    } else {
      if (txStack.current().requiresPrewarming()) {
        state.setProcessingPhase(TX_WARM);
        new TxPreWarmingMacroSection(world, this);
      }
      state.setProcessingPhase(TX_INIT);
      new TxInitializationSection(this, world);
    }

    /*
     * TODO: the ID = 0 (universal parent context) context should
     *  1. be shared by all transactions in a conflation (OK)
     *  2. should be the father of all root contexts
     *  3. should have the current root context as its lastCallee()
     */
    callStack.getById(0).universalParentReturnDataContextNumber(this.stamp() + 1);

    for (Module m : this.modules) {
      m.traceStartTx(world, this.txStack().current());
    }
  }

  public void traceEndTransaction(
      WorldView world,
      Transaction tx,
      boolean isSuccessful,
      List<Log> logs,
      Set<Address> selfDestructs) {
    // TODO: see issue #875. It is currently unclear which, if any,
    //  rollbacks already took place at traceEndTransaction.

    // TODO: add the following resolution this.defers.resolvePostRollback(this, ...

    txStack.current().completeLineaTransaction(this, isSuccessful, logs, selfDestructs);

    defers.resolvePostTransaction(this, world, tx, isSuccessful);

    // Warn: we need to call MMIO after resolving the defers
    for (Module m : modules) {
      m.traceEndTx(txStack.current());
    }

    // Compute the line counting of the HUB of the current transaction
    state.lineCounter().add(state.currentTxTrace().lineCount());
  }

  @Override
  public void traceContextEnter(MessageFrame frame) {
    this.pch.reset();

    if (frame.getDepth() == 0) {
      // Root context
      final TransactionProcessingMetadata currentTx = transients().tx();
      final Address toAddress = effectiveToAddress(currentTx.getBesuTransaction());
      final boolean isDeployment = this.transients.tx().getBesuTransaction().getTo().isEmpty();

      final boolean shouldCopyTxCallData = !isDeployment && currentTx.requiresEvmExecution();
      // TODO simplify this, the same bedRock context ( = root context ??) seems to be
      // generated in
      // both case
      if (shouldCopyTxCallData) {
        this.callStack.newMantleAndBedrock(
            this.state.stamps().hub(),
            this.transients.tx().getBesuTransaction().getSender(),
            toAddress,
            CallFrameType.TRANSACTION_CALL_DATA_HOLDER,
            new Bytecode(
                toAddress == null
                    ? this.transients.tx().getBesuTransaction().getData().orElse(Bytes.EMPTY)
                    : Optional.ofNullable(frame.getWorldUpdater().get(toAddress))
                        .map(AccountState::getCode)
                        .orElse(Bytes.EMPTY)),
            Wei.of(this.transients.tx().getBesuTransaction().getValue().getAsBigInteger()),
            this.transients.tx().getBesuTransaction().getGasLimit(),
            this.transients.tx().getBesuTransaction().getData().orElse(Bytes.EMPTY),
            this.transients.conflation().deploymentInfo().deploymentNumber(toAddress),
            toAddress.isEmpty()
                ? 0
                : this.transients.conflation().deploymentInfo().deploymentNumber(toAddress),
            this.transients.conflation().deploymentInfo().getDeploymentStatus(toAddress));
      } else {
        this.callStack.newBedrock(
            this.state.stamps().hub(),
            this.txStack.current().getBesuTransaction().getSender(),
            toAddress,
            CallFrameType.ROOT,
            new Bytecode(
                toAddress == null
                    ? this.transients.tx().getBesuTransaction().getData().orElse(Bytes.EMPTY)
                    : Optional.ofNullable(frame.getWorldUpdater().get(toAddress))
                        .map(AccountState::getCode)
                        .orElse(Bytes.EMPTY)),
            Wei.of(this.transients.tx().getBesuTransaction().getValue().getAsBigInteger()),
            this.transients.tx().getBesuTransaction().getGasLimit(),
            this.transients.tx().getBesuTransaction().getData().orElse(Bytes.EMPTY),
            this.transients.conflation().deploymentInfo().deploymentNumber(toAddress),
            toAddress.isEmpty()
                ? 0
                : this.transients.conflation().deploymentInfo().deploymentNumber(toAddress),
            this.transients.conflation().deploymentInfo().getDeploymentStatus(toAddress));
      }
    } else {
      // ...or CALL or CREATE
      final OpCode currentOpCode = callStack.current().opCode();
      final boolean isDeployment = frame.getType() == MessageFrame.Type.CONTRACT_CREATION;
      final Address codeAddress = frame.getContractAddress();
      final CallFrameType frameType =
          frame.isStatic() ? CallFrameType.STATIC : CallFrameType.STANDARD;
      if (isDeployment) {
        this.transients.conflation().deploymentInfo().newDeploymentAt(codeAddress);
      }
      final int codeDeploymentNumber =
          this.transients.conflation().deploymentInfo().deploymentNumber(codeAddress);

      final long callDataOffset =
          isDeployment
              ? 0
              : Words.clampedToLong(
                  callStack
                      .current()
                      .frame()
                      .getStackItem(currentOpCode.callMayNotTransferValue() ? 2 : 3));

      final long callDataSize =
          isDeployment
              ? 0
              : Words.clampedToLong(
                  callStack
                      .current()
                      .frame()
                      .getStackItem(currentOpCode.callMayNotTransferValue() ? 3 : 4));

      final long callDataContextNumber = this.callStack.current().contextNumber();

      this.callStack.enter(
          this.state.stamps().hub(),
          frame.getRecipientAddress(),
          frame.getContractAddress(),
          frame.getRecipientAddress(), // TODO: this is likely false
          new Bytecode(frame.getCode().getBytes()),
          frameType,
          frame.getValue(),
          frame.getRemainingGas(),
          frame.getInputData(),
          callDataOffset,
          callDataSize,
          callDataContextNumber,
          this.transients.conflation().deploymentInfo().deploymentNumber(codeAddress),
          codeDeploymentNumber,
          isDeployment);
      this.currentFrame().initializeFrame(frame); // TODO should be done in enter

      this.defers.resolveUponImmediateContextEntry(this);

      for (Module m : this.modules) {
        m.traceContextEnter(frame);
      }
    }
  }

  public void traceContextReEnter(MessageFrame frame) {
    // Note: the update of the current call frame is made during traceContextExit of the child frame
    this.currentFrame().initializeFrame(frame); // TODO: is it needed ?
    defers.resolveAtContextReEntry(this, this.currentFrame());
    this.unlatchStack(frame, this.currentFrame().childSpanningSection());
  }

  @Override
  public void traceContextExit(MessageFrame frame) {
    this.currentFrame().initializeFrame(frame); // TODO: is it needed ?

    exitDeploymentFromDeploymentInfoPointOfView();

    // TODO: why only do this at positive depth ?
    if (frame.getDepth() > 0) {

      DeploymentExceptions contextExceptions =
          DeploymentExceptions.fromFrame(this.currentFrame(), frame);
      this.currentTraceSection().setContextExceptions(contextExceptions);

      if (contextExceptions.any()) {
        this.callStack.revert(this.state.stamps().hub()); // TODO: Duplicate s?
      }
    }

    // We take a snapshot before exiting the transaction
    if (frame.getDepth() == 0) {
      final long leftOverGas = frame.getRemainingGas();
      final long gasRefund = frame.getGasRefund();
      final boolean coinbaseIsWarm = frame.isAddressWarm(txStack.current().getCoinbase());

      txStack
          .current()
          .setPreFinalisationValues(
              leftOverGas,
              gasRefund,
              coinbaseIsWarm,
              this.txStack.getAccumulativeGasUsedInBlockBeforeTxStart());

      if (this.state.getProcessingPhase() != TX_SKIP) {
        this.state.setProcessingPhase(TX_FINL);
        new TxFinalizationSection(this, frame.getWorldUpdater());
      }
    }

    this.defers.resolveUponExitingContext(this, this.currentFrame());
    // TODO: verify me please @Olivier
    if (this.currentFrame().opCode() == OpCode.REVERT || Exceptions.any(this.pch.exceptions())) {
      this.defers.resolvePostRollback(this, frame, this.currentFrame());
    }

    if (frame.getDepth() > 0) {
      this.callStack.exit();
    }
  }

  /**
   * If the current execution context is a deployment context the present method "exits" that
   * deployment in the sense that it updates the relevant deployment information.
   */
  private void exitDeploymentFromDeploymentInfoPointOfView() {

    // sanity check
    Preconditions.checkArgument(
        deploymentStatusOfBytecodeAddress()
            == (messageFrame().getType() == MessageFrame.Type.CONTRACT_CREATION));

    if (deploymentStatusOfBytecodeAddress()) {
      transients
          .conflation()
          .deploymentInfo()
          .markAsNotUnderDeployment(this.currentFrame().byteCodeAddress());
    }
  }

  public void tracePreExecution(final MessageFrame frame) {
    Preconditions.checkArgument(
        this.state().processingPhase == TX_EXEC,
        "There can't be any execution if the HUB is not in execution phase");

    this.processStateExec(frame);
  }

  public void tracePostExecution(MessageFrame frame, Operation.OperationResult operationResult) {
    Preconditions.checkArgument(
        this.state().processingPhase == TX_EXEC,
        "There can't be any execution if the HUB is not in execution phase");

    final long gasCost = operationResult.getGasCost();
    final TraceSection currentSection = this.state.currentTxTrace().currentSection();

    final short exceptions = this.pch().exceptions();

    final boolean memoryExpansionException = Exceptions.memoryExpansionException(exceptions);
    final boolean outOfGasException = Exceptions.outOfGasException(exceptions);
    final boolean unexceptional = Exceptions.none(exceptions);
    final boolean exceptional = Exceptions.any(exceptions);

    // NOTE: whenever there is an exception, a context row
    // is added at the end of the section; its purpose is
    // to update the caller / creator context with empty
    // return data.
    //////////////////////////////////////////////////////
    if (exceptional) {
      this.currentTraceSection()
          .addFragments(ContextFragment.executionProvidesEmptyReturnData(this));
      this.squashCurrentFrameOutputData();
      this.squashParentFrameReturnData();
    }

    // Setting gas cost IN MOST CASES
    // TODO:
    //  * complete this for CREATE's and CALL's
    //    + are we getting the correct cost (i.e. excluding the 63/64-th's) ?
    //  * make sure this aligns with exception handling of the zkevm
    //  * write a method `final boolean requiresGasCost()` (huge switch case)
    if ((!memoryExpansionException & outOfGasException) || unexceptional) {
      currentSection.commonValues.gasCost(gasCost);
      currentSection.commonValues.gasNext(
          unexceptional ? currentSection.commonValues.gasActual - gasCost : 0);
    } else {
      currentSection.commonValues.gasCost(
          0); // TODO: fill with correct values --- make sure this works in all cases
      currentSection.commonValues.gasNext(0);
    }

    if (this.currentFrame().opCode().isCreate() && operationResult.getHaltReason() == null) {
      this.handleCreate(Words.toAddress(frame.getStackItem(0)));
    }

    this.defers.resolvePostExecution(this, frame, operationResult);

    if (!this.currentFrame().opCode().isCall() && !this.currentFrame().opCode().isCreate()) {
      this.unlatchStack(frame);
    }

    switch (this.opCodeData().instructionFamily()) {
      case ADD -> {}
      case MOD -> {}
      case MUL -> {}
      case EXT -> {}
      case WCP -> {}
      case BIN -> {}
      case SHF -> {}
      case KEC -> {}
      case CONTEXT -> {}
      case ACCOUNT -> {}
      case COPY -> {}
      case TRANSACTION -> {}
      case BATCH -> this.blockhash.tracePostOpcode(frame);
      case STACK_RAM -> {}
      case STORAGE -> {}
      case JUMP -> {}
      case MACHINE_STATE -> {}
      case PUSH_POP -> {}
      case DUP -> {}
      case SWAP -> {}
      case LOG -> {}
      case CREATE -> this.romLex.tracePostOpcode(frame);
      case CALL -> {}
      case HALT -> {}
      case INVALID -> {}
      default -> {}
    }
  }

  private void handleCreate(Address target) {
    this.transients.conflation().deploymentInfo().newDeploymentAt(target);
  }

  public int getCfiByMetaData(
      final Address address, final int deploymentNumber, final boolean deploymentStatus) {
    return this.romLex()
        .getCodeFragmentIndexByMetadata(
            ContractMetadata.make(address, deploymentNumber, deploymentStatus));
  }

  public int newChildContextNumber() {
    return 1 + this.stamp();
  }

  public CallFrame currentFrame() {
    if (this.callStack().isEmpty()) {
      return CallFrame.EMPTY;
    }
    return this.callStack.current();
  }

  public MessageFrame messageFrame() {
    MessageFrame frame = this.callStack.current().frame();
    return frame;
  }

  private void handleStack(MessageFrame frame) {
    this.currentFrame()
        .stack()
        .processInstruction(this, frame, MULTIPLIER___STACK_HEIGHT * this.state.stamps().hub());
  }

  void triggerModules(MessageFrame frame) {
    if (this.pch.signals().add()) {
      this.add.tracePreOpcode(frame);
    }
    if (this.pch.signals().bin()) {
      this.bin.tracePreOpcode(frame);
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
    if (this.pch.signals().mxp()) {
      this.mxp.tracePreOpcode(frame);
    }
    if (this.pch.signals().oob()) {
      this.oob.tracePreOpcode(frame);
    }
    if (this.pch.signals().stp()) {
      this.stp.tracePreOpcode(frame);
    }
    if (this.pch.signals().exp()) {
      this.exp.tracePreOpcode(frame);
    }
    if (this.pch.signals().trm()) {
      this.trm.tracePreOpcode(frame);
    }
    if (this.pch.signals().hashInfo()) {
      // TODO: this.hashInfo.tracePreOpcode(frame);
    }
    if (this.pch.signals().blockhash()) {
      this.blockhash.tracePreOpcode(frame);
    }
  }

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

  public void addTraceSection(TraceSection section) {
    this.state.currentTxTrace().add(section);
  }

  private void unlatchStack(MessageFrame frame) {
    this.unlatchStack(frame, this.currentTraceSection());
  }

  public void unlatchStack(MessageFrame frame, TraceSection section) {
    if (this.currentFrame().pending() == null) {
      return;
    }

    final StackContext pending = this.currentFrame().pending();
    for (int i = 0; i < pending.lines().size(); i++) {
      final StackLine line = pending.lines().get(i);

      if (line.needsResult()) {
        Bytes result = Bytes.EMPTY;
        // Only pop from the stack if no exceptions have been encountered
        if (Exceptions.none(this.pch.exceptions())) {
          result = frame.getStackItem(0).copy();
        }

        // This works because we are certain that the stack chunks are the first.
        ((StackFragment) section.fragments().get(i))
            .stackOps()
            .get(line.resultColumn() - 1)
            .value(result);
      }
    }
  }

  void processStateExec(MessageFrame frame) {
    this.pch.setup(frame);

    this.handleStack(frame);
    this.triggerModules(frame);

    if (Exceptions.any(this.pch().exceptions()) || this.currentFrame().opCode() == OpCode.REVERT) {
      this.callStack.revert(this.state.stamps().hub());
    }

    if (this.currentFrame().stack().isOk()) {
      this.traceOpcode(frame);
    } else {

      this.squashCurrentFrameOutputData();
      this.squashParentFrameReturnData();

      new StackOnlySection(this);
    }
  }

  // TODO: how do these implementations of remainingGas()
  //  and expectedGas() behave with respect to resuming
  //  execution after a CALL / CREATE ? One of them is
  //  necessarily false ...
  public long remainingGas() {
    return this.state().getProcessingPhase() == TX_EXEC
        ? this.currentFrame().frame().getRemainingGas()
        : 0;
  }

  public long expectedGas() {
    return this.state().getProcessingPhase() == TX_EXEC
        ? this.currentFrame().frame().getRemainingGas()
        : 0;
  }

  public int cumulatedTxCount() {
    return this.state.txCount();
  }

  void traceOpcode(MessageFrame frame) {

    switch (this.opCodeData().instructionFamily()) {
      case ADD,
          MOD,
          SHF,
          BIN,
          WCP,
          EXT,
          BATCH,
          MACHINE_STATE,
          PUSH_POP,
          DUP,
          SWAP,
          INVALID -> new StackOnlySection(this);
      case MUL -> {
        switch (this.opCode()) {
          case OpCode.EXP -> new ExpSection(this);
          case OpCode.MUL -> new StackOnlySection(this);
          default -> throw new IllegalStateException(
              String.format("opcode %s not part of the MUL instruction family", this.opCode()));
        }
      }
      case HALT -> {
        final CallFrame parentFrame = this.callStack.parent();
        parentFrame.returnDataContextNumber(this.currentFrame().contextNumber());
        final Bytes outputData = this.transients.op().outputData();
        this.currentFrame().outputDataSpan(transients.op().outputDataSpan());
        this.currentFrame().outputData(outputData);

        // The output data always becomes return data of the caller when REVERT'ing
        // and in all other cases becomes return data of the caller iff the present
        // context is a message call context
        final boolean outputDataBecomesParentReturnData =
            (this.opCode() == OpCode.REVERT || this.currentFrame().isMessageCall());

        if (outputDataBecomesParentReturnData) {
          parentFrame.returnData(outputData);
          parentFrame.returnDataSpan(transients.op().outputDataSpan());
        } else {
          this.squashParentFrameReturnData();
        }

        switch (this.opCode()) {
          case RETURN -> new ReturnSection(this);
          case REVERT -> new RevertSection(this);
          case STOP -> new StopSection(this);
          case SELFDESTRUCT -> new SelfdestructSection(this);
        }
      }

      case KEC -> new KeccakSection(this);
      case CONTEXT -> new ContextSection(this);
      case LOG -> new LogSection(this);
      case ACCOUNT -> new AccountSection(this);
      case COPY -> {
        switch (this.opCode()) {
          case OpCode.CALLDATACOPY -> new CallDataCopySection(this);
          case OpCode.RETURNDATACOPY -> new ReturnDataCopySection(this);
          case OpCode.CODECOPY -> new CodeCopySection(this);
          case OpCode.EXTCODECOPY -> new ExtCodeCopySection(this);
          default -> throw new RuntimeException(
              "Invalid instruction: " + this.opCode().toString() + " not in the COPY family");
        }
      }

      case TRANSACTION -> new TransactionSection(this);

      case STACK_RAM -> {
        switch (this.currentFrame().opCode()) {
          case CALLDATALOAD -> new CallDataLoadSection(this);
          case MLOAD, MSTORE, MSTORE8 -> new StackRamSection(this);
          default -> throw new IllegalStateException("unexpected STACK_RAM opcode");
        }
      }

      case STORAGE -> {
        switch (this.currentFrame().opCode()) {
          case SSTORE -> new SstoreSection(this, frame.getWorldUpdater());
          case SLOAD -> new SloadSection(this, frame.getWorldUpdater());
          default -> throw new IllegalStateException("invalid operation in family STORAGE");
        }
      }

      case CREATE -> new CreateSection(this);

      case CALL -> new CallSection(this);

      case JUMP -> new JumpSection(this);
    }
  }

  public void squashCurrentFrameOutputData() {
    this.currentFrame().outputDataSpan(MemorySpan.empty());
    this.currentFrame().outputData(Bytes.EMPTY);
  }

  public void squashParentFrameReturnData() {
    final CallFrame parentFrame = this.callStack.parent();
    parentFrame.returnData(Bytes.EMPTY);
    parentFrame.returnDataSpan(MemorySpan.empty());
  }

  public CallFrame getLastChildCallFrame(final CallFrame parentFrame) {
    return this.callStack.getById(parentFrame.childFramesId().getLast());
  }

  public final int deploymentNumberOf(Address address) {
    return transients.conflation().deploymentInfo().deploymentNumber(address);
  }

  public final boolean deploymentStatusOf(Address address) {
    return transients.conflation().deploymentInfo().getDeploymentStatus(address);
  }

  // methods related to the byte code address
  // (c in the definition of \Theta in the EYP)
  public final Address bytecodeAddress() {
    return this.messageFrame().getContractAddress();
  }

  public final int deploymentNumberOfBytecodeAddress() {
    return deploymentNumberOf(bytecodeAddress());
  }

  public final boolean deploymentStatusOfBytecodeAddress() {
    return deploymentStatusOf(bytecodeAddress());
  }

  // methods related to the account address
  // (r in the definition of \Theta in the EYP)
  // (also I_a in the EYP)
  public final Address accountAddress() {
    return this.messageFrame().getRecipientAddress();
  }

  public final int deploymentNumberOfAccountAddress() {
    return deploymentNumberOf(this.accountAddress());
  }

  public final boolean deploymentStatusOfAccountAddress() {
    return deploymentStatusOf(this.accountAddress());
  }
}
