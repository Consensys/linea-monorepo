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

import static net.consensys.linea.zktracer.types.AddressUtils.effectiveToAddress;
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
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.module.Module;
import net.consensys.linea.zktracer.module.add.Add;
import net.consensys.linea.zktracer.module.bin.Bin;
import net.consensys.linea.zktracer.module.blake2fmodexpdata.Blake2fModexpData;
import net.consensys.linea.zktracer.module.ec_data.EcData;
import net.consensys.linea.zktracer.module.euc.Euc;
import net.consensys.linea.zktracer.module.exp.Exp;
import net.consensys.linea.zktracer.module.ext.Ext;
import net.consensys.linea.zktracer.module.hub.defer.*;
import net.consensys.linea.zktracer.module.hub.fragment.*;
import net.consensys.linea.zktracer.module.hub.fragment.imc.ImcFragment;
import net.consensys.linea.zktracer.module.hub.fragment.scenario.ScenarioFragment;
import net.consensys.linea.zktracer.module.hub.precompiles.PrecompileInvocation;
import net.consensys.linea.zktracer.module.hub.section.*;
import net.consensys.linea.zktracer.module.hub.signals.PlatformController;
import net.consensys.linea.zktracer.module.hub.transients.Transients;
import net.consensys.linea.zktracer.module.limits.Keccak;
import net.consensys.linea.zktracer.module.limits.L2Block;
import net.consensys.linea.zktracer.module.limits.L2L1Logs;
import net.consensys.linea.zktracer.module.limits.precompiles.Blake2fRounds;
import net.consensys.linea.zktracer.module.limits.precompiles.EcAddEffectiveCall;
import net.consensys.linea.zktracer.module.limits.precompiles.EcMulEffectiveCall;
import net.consensys.linea.zktracer.module.limits.precompiles.EcPairingCallEffectiveCall;
import net.consensys.linea.zktracer.module.limits.precompiles.EcPairingMillerLoop;
import net.consensys.linea.zktracer.module.limits.precompiles.EcRecoverEffectiveCall;
import net.consensys.linea.zktracer.module.limits.precompiles.ModexpEffectiveCall;
import net.consensys.linea.zktracer.module.limits.precompiles.Rip160Blocks;
import net.consensys.linea.zktracer.module.limits.precompiles.Sha256Blocks;
import net.consensys.linea.zktracer.module.logData.LogData;
import net.consensys.linea.zktracer.module.logInfo.LogInfo;
import net.consensys.linea.zktracer.module.mmio.Mmio;
import net.consensys.linea.zktracer.module.mmu.Mmu;
import net.consensys.linea.zktracer.module.mod.Mod;
import net.consensys.linea.zktracer.module.mul.Mul;
import net.consensys.linea.zktracer.module.mxp.Mxp;
import net.consensys.linea.zktracer.module.rlp.addr.RlpAddr;
import net.consensys.linea.zktracer.module.rlp.txn.RlpTxn;
import net.consensys.linea.zktracer.module.rlp.txrcpt.RlpTxrcpt;
import net.consensys.linea.zktracer.module.rom.Rom;
import net.consensys.linea.zktracer.module.romLex.RomLex;
import net.consensys.linea.zktracer.module.shf.Shf;
import net.consensys.linea.zktracer.module.stp.Stp;
import net.consensys.linea.zktracer.module.tables.bin.BinRt;
import net.consensys.linea.zktracer.module.tables.instructionDecoder.InstructionDecoder;
import net.consensys.linea.zktracer.module.tables.shf.ShfRt;
import net.consensys.linea.zktracer.module.trm.Trm;
import net.consensys.linea.zktracer.module.txn_data.TxnData;
import net.consensys.linea.zktracer.module.wcp.Wcp;
import net.consensys.linea.zktracer.opcode.*;
import net.consensys.linea.zktracer.opcode.gas.projector.GasProjector;
import net.consensys.linea.zktracer.runtime.LogInvocation;
import net.consensys.linea.zktracer.runtime.callstack.CallFrame;
import net.consensys.linea.zktracer.runtime.callstack.CallFrameType;
import net.consensys.linea.zktracer.runtime.callstack.CallStack;
import net.consensys.linea.zktracer.runtime.stack.StackContext;
import net.consensys.linea.zktracer.runtime.stack.StackLine;
import net.consensys.linea.zktracer.types.Bytecode;
import net.consensys.linea.zktracer.types.EWord;
import net.consensys.linea.zktracer.types.Precompile;
import net.consensys.linea.zktracer.types.TxState;
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
import org.hyperledger.besu.evm.log.LogTopic;
import org.hyperledger.besu.evm.operation.Operation;
import org.hyperledger.besu.evm.worldstate.WorldView;
import org.hyperledger.besu.plugin.data.ProcessableBlockHeader;

@Slf4j
@Accessors(fluent = true)
public class Hub implements Module {
  private static final int TAU = 8;

  public static final GasProjector GAS_PROJECTOR = new GasProjector();

  /** accumulate the trace information for the Hub */
  @Getter private final State state = new State();
  /** contain the factories for trace segments that need complex initialization */
  @Getter private final Factories factories;

  /** provides phase-related volatile information */
  @Getter Transients transients;

  /**
   * Long-lived states, not used in tracing per se but keeping track of data of the associated
   * lifetime
   */
  @Getter CallStack callStack = new CallStack();

  @Getter TransactionStack txStack = new TransactionStack();

  /** Stores all the actions that must be deferred to a later time */
  @Getter private final DeferRegistry defers = new DeferRegistry();

  /** stores all data related to failure states & module activation */
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
  private final Module bin = new Bin(this);
  private final Blake2fModexpData blake2fModexpData = new Blake2fModexpData();
  private final EcData ecData;
  private final Euc euc;
  private final Ext ext = new Ext(this);
  private final Module mul = new Mul(this);
  private final Mod mod = new Mod();
  private final Module shf = new Shf();
  @Getter private final Wcp wcp = new Wcp(this);
  private final RlpTxn rlpTxn;
  private final Module mxp;
  private final Mmio mmio;

  @Getter private final Exp exp;
  @Getter private final Mmu mmu;
  private final RlpTxrcpt rlpTxrcpt = new RlpTxrcpt();
  private final LogInfo logInfo = new LogInfo(rlpTxrcpt);
  private final LogData logData = new LogData(rlpTxrcpt);
  private final RlpAddr rlpAddr = new RlpAddr(this);
  private final Rom rom;

  @Getter private final RomLex romLex;
  private final TxnData txnData;
  private final Trm trm = new Trm();
  private final ModexpEffectiveCall modexpEffectiveCall;
  private final Stp stp = new Stp(this, wcp, mod);
  private final L2Block l2Block;

  private final List<Module> modules;
  /* Those modules are not traced, we just compute the number of calls to those precompile to meet the prover limits */
  private final List<Module> precompileLimitModules;
  private final List<Module> refTableModules;

  public Hub(final Address l2l1ContractAddress, final Bytes l2l1Topic) {
    this.l2Block = new L2Block(l2l1ContractAddress, LogTopic.of(l2l1Topic));
    this.transients = new Transients(this);
    this.factories = new Factories(this);

    this.pch = new PlatformController(this);
    this.mxp = new Mxp(this);
    this.exp = new Exp(this, this.wcp);
    this.romLex = new RomLex(this);
    this.rom = new Rom(this.romLex);
    this.rlpTxn = new RlpTxn(this.romLex);
    this.txnData = new TxnData(this, this.romLex, this.wcp);
    this.ecData = new EcData(this, this.wcp, this.ext);
    this.euc = new Euc(this.wcp);
    this.mmu =
        new Mmu(
            this.euc,
            this.wcp,
            this.romLex,
            this.rlpTxn,
            this.rlpTxrcpt,
            this.ecData,
            this.blake2fModexpData,
            this.callStack);
    this.mmio = new Mmio(this.mmu);

    final EcRecoverEffectiveCall ecRec = new EcRecoverEffectiveCall(this);
    this.modexpEffectiveCall = new ModexpEffectiveCall(this, this.blake2fModexpData);
    final EcPairingCallEffectiveCall ecPairingCall = new EcPairingCallEffectiveCall(this);
    final L2Block l2Block = new L2Block(l2l1ContractAddress, LogTopic.of(l2l1Topic));

    this.precompileLimitModules =
        List.of(
            new Sha256Blocks(this),
            ecRec,
            new Rip160Blocks(this),
            this.modexpEffectiveCall,
            new EcAddEffectiveCall(this),
            new EcMulEffectiveCall(this),
            ecPairingCall,
            new EcPairingMillerLoop(ecPairingCall),
            new Blake2fRounds(this, this.blake2fModexpData),
            // Block level limits
            l2Block,
            new Keccak(this, ecRec, l2Block),
            new L2L1Logs(l2Block));

    this.refTableModules = List.of(new BinRt(), new InstructionDecoder(), new ShfRt());

    this.modules =
        Stream.concat(
                Stream.of(
                    this.add,
                    this.bin,
                    this.blake2fModexpData,
                    this.ecData,
                    this.euc,
                    this.ext,
                    this.logData,
                    this.logInfo,
                    this.mmio,
                    this.mmu,
                    this.mod,
                    this.mul,
                    this.mxp,
                    this.exp,
                    this.rlpAddr,
                    this.rlpTxn,
                    this.rlpTxrcpt,
                    this.rom,
                    this.romLex,
                    this.shf,
                    this.stp,
                    this.trm,
                    this.txnData,
                    this.wcp),
                this.precompileLimitModules.stream())
            .toList();
  }

  /**
   * @return a list of all modules for which to generate traces
   */
  public List<Module> getModulesToTrace() {
    return Stream.concat(
            this.refTableModules.stream(),
            // Modules
            Stream.of(
                this,
                this.add,
                this.bin,
                this.blake2fModexpData,
                //        this.ecData, // TODO: not yet
                this.ext,
                this.euc,
                this.exp,
                this.logData,
                this.logInfo,
                this.mmu, // WARN: must be called before the MMIO
                this.mmio,
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
                this.wcp))
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
                this.romLex,
                this.bin,
                this.add,
                this.ext,
                this.ecData,
                this.euc,
                this.mmu,
                this.mmio,
                this.logData,
                this.logInfo,
                this.mod,
                this.mul,
                this.mxp,
                this.exp,
                this.rlpAddr,
                this.rlpTxn,
                this.rlpTxrcpt,
                this.rom,
                this.shf,
                this.trm,
                this.txnData,
                this.wcp,
                this.l2Block),
            this.precompileLimitModules.stream())
        .toList();
  }

  /**
   * Traces a skipped transaction, i.e. a “pure” transaction without EVM execution.
   *
   * @param world a view onto the state
   */
  void processStateSkip(WorldView world) {
    this.state.stamps().stampHubIncrements();
    boolean isDeployment = this.transients.tx().besuTx().getTo().isEmpty();

    //
    // 3 sections -- account changes
    //
    // From account information
    Address fromAddress = this.transients.tx().besuTx().getSender();
    AccountSnapshot oldFromAccount =
        AccountSnapshot.fromAccount(
            world.get(fromAddress),
            false,
            this.transients.conflation().deploymentInfo().number(fromAddress),
            false);

    // To account information
    Address toAddress = effectiveToAddress(this.transients.tx().besuTx());
    if (isDeployment) {
      this.transients.conflation().deploymentInfo().deploy(toAddress);
    }
    boolean toIsWarm =
        (fromAddress == toAddress)
            || isPrecompile(toAddress); // should never happen – no TX to PC allowed
    AccountSnapshot oldToAccount =
        AccountSnapshot.fromAccount(
            world.get(toAddress),
            toIsWarm,
            this.transients.conflation().deploymentInfo().number(toAddress),
            false);

    // Miner account information
    boolean minerIsWarm =
        (this.transients.block().minerAddress() == fromAddress)
            || (this.transients.block().minerAddress() == toAddress)
            || isPrecompile(this.transients.block().minerAddress());
    AccountSnapshot oldMinerAccount =
        AccountSnapshot.fromAccount(
            world.get(this.transients.block().minerAddress()),
            minerIsWarm,
            this.transients
                .conflation()
                .deploymentInfo()
                .number(this.transients.block().minerAddress()),
            false);

    // Putatively updateCallerReturnData deployment number
    this.defers.postTx(
        new SkippedPostTransactionDefer(
            oldFromAccount,
            oldToAccount,
            oldMinerAccount,
            this.transients.tx().gasPrice(),
            this.transients.block().baseFee()));
  }

  /**
   * Traces the warm-up information of a transaction
   *
   * @param world a view onto the state
   */
  void processStateWarm(WorldView world) {
    this.transients
        .tx()
        .besuTx()
        .getAccessList()
        .ifPresent(
            preWarmed -> {
              if (!preWarmed.isEmpty()) {
                this.state.stamps().stampHubIncrements();

                Set<Address> seenAddresses = new HashSet<>();
                Map<Address, Set<Bytes32>> seenKeys = new HashMap<>();
                List<TraceFragment> fragments = new ArrayList<>();

                for (AccessListEntry entry : preWarmed) {
                  Address address = entry.address();
                  AccountSnapshot snapshot =
                      AccountSnapshot.fromAccount(
                          world.get(address), seenAddresses.contains(address), 0, false);
                  fragments.add(
                      this.factories.accountFragment().make(snapshot, snapshot, false, 0, false));
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
                            this.transients.conflation().deploymentInfo().number(address),
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
    this.transients.tx().state(TxState.TX_INIT);
  }

  /**
   * Trace the preamble of a transaction
   *
   * @param world a view onto the state
   */
  void processStateInit(WorldView world) {
    this.state.stamps().stampHubIncrements();
    final boolean isDeployment = this.transients.tx().besuTx().getTo().isEmpty();
    final Address toAddress = effectiveToAddress(this.transients.tx().besuTx());
    if (isDeployment) {
      this.transients.conflation().deploymentInfo().deploy(toAddress);
    }

    final Address fromAddress = this.transients.tx().besuTx().getSender();
    final Account fromAccount = world.get(fromAddress);
    final AccountSnapshot fromSnapshot =
        AccountSnapshot.fromAccount(
            fromAccount,
            true,
            this.transients.conflation().deploymentInfo().number(fromAddress),
            this.transients.conflation().deploymentInfo().isDeploying(fromAddress));

    final Account toAccount = world.get(toAddress);
    final AccountSnapshot toSnapshot =
        AccountSnapshot.fromAccount(
            toAccount,
            true,
            this.transients.conflation().deploymentInfo().number(toAddress),
            this.transients.conflation().deploymentInfo().isDeploying(toAddress));

    final Wei transactionGasPrice =
        ZkTracer.feeMarket
            .getTransactionPriceCalculator()
            .price(
                (org.hyperledger.besu.ethereum.core.Transaction) this.transients.tx().besuTx(),
                Optional.of(this.transients.block().baseFee()));
    final Wei value = (Wei) this.transients.tx().besuTx().getValue();
    final AccountSnapshot fromPostDebitSnapshot =
        fromSnapshot.debit(
            transactionGasPrice.multiply(this.transients.tx().besuTx().getGasLimit()).add(value));

    final boolean isSelfCredit = toAddress.equals(fromAddress);
    final TransactionFragment txFragment =
        TransactionFragment.prepare(
            this.transients.conflation().number(),
            this.transients.block().minerAddress(),
            this.transients.tx().besuTx(),
            true,
            ((org.hyperledger.besu.ethereum.core.Transaction) this.transients.tx().besuTx())
                .getEffectiveGasPrice(Optional.ofNullable(this.transients().block().baseFee())),
            this.transients.block().baseFee(),
            0 // TODO: find getInitialGas
            );
    this.defers.postTx(txFragment);

    this.addTraceSection(
        new TxInitSection(
            this,
            this.factories.accountFragment().make(fromSnapshot, fromPostDebitSnapshot),
            isDeployment
                ? this.factories.accountFragment().make(toSnapshot, toSnapshot.deploy(value))
                : (isSelfCredit
                    ? this.factories
                        .accountFragment()
                        .make(fromPostDebitSnapshot, fromPostDebitSnapshot.credit(value))
                    : this.factories.accountFragment().make(toSnapshot, toSnapshot.credit(value))),
            ImcFragment.forTxInit(this),
            ContextFragment.initializeExecutionContext(this),
            txFragment));

    this.transients.tx().state(TxState.TX_EXEC);
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
    this.currentFrame().stack().processInstruction(this, frame, TAU * this.state.stamps().hub());
  }

  void triggerModules(MessageFrame frame) {
    if (this.pch.exceptions().none() && this.pch.aborts().none()) {
      for (Module precompileLimit : this.precompileLimitModules) {
        precompileLimit.tracePreOpcode(frame);
      }
    }

    if (this.pch.signals().romLex()) {
      this.romLex.tracePreOpcode(frame);
    }
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
      this.exp.tracePreOpcode(frame);
      this.modexpEffectiveCall.tracePreOpcode(frame);
      // if (this.pch.exceptions().none() && this.pch.aborts().none())
    }
    if (this.pch.signals().trm()) {
      this.trm.tracePreOpcode(frame);
    }
    if (this.pch.signals().hashInfo()) {
      // TODO: this.hashInfo.tracePreOpcode(frame);
    }
    if (this.pch.signals().ecData()) {
      this.ecData.tracePreOpcode(frame);
    }
  }

  void processStateExec(MessageFrame frame) {
    this.currentFrame().frame(frame);
    this.state.stamps().stampHubIncrements();
    this.pch.setup(frame);
    this.state.stamps().stampSubmodules(this.pch());

    this.handleStack(frame);
    this.triggerModules(frame);
    if (this.pch().exceptions().any() || this.currentFrame().opCode() == OpCode.REVERT) {
      this.callStack.revert(this.state.stamps().hub());
    }

    if (this.currentFrame().stack().isOk()) {
      this.traceOperation(frame);
    } else {
      this.addTraceSection(new StackOnlySection(this));
      this.currentTraceSection()
          .addFragmentsWithoutStack(this, ContextFragment.executionEmptyReturnData(this.callStack));
    }
  }

  void processStateFinal(WorldView worldView, Transaction tx, boolean isSuccess) {
    this.transients().tx().state(TxState.TX_FINAL);
    this.state.stamps().stampHubIncrements();

    Address fromAddress = this.transients.tx().besuTx().getSender();
    Account fromAccount = worldView.get(fromAddress);
    AccountSnapshot fromSnapshot =
        AccountSnapshot.fromAccount(
            fromAccount,
            true,
            this.transients.conflation().deploymentInfo().number(fromAddress),
            this.transients.conflation().deploymentInfo().isDeploying(fromAddress));

    Account minerAccount = worldView.get(this.transients.block().minerAddress());
    AccountSnapshot minerSnapshot =
        AccountSnapshot.fromAccount(
            minerAccount,
            true,
            this.transients
                .conflation()
                .deploymentInfo()
                .number(this.transients.block().minerAddress()),
            this.transients
                .conflation()
                .deploymentInfo()
                .isDeploying(this.transients.block().minerAddress()));

    if (isSuccess) {
      // if no revert: 2 account rows (sender, coinbase) + 1 tx row
      this.addTraceSection(
          new EndTransaction(
              this,
              this.factories.accountFragment().make(fromSnapshot, fromSnapshot, false, 0, false),
              this.factories.accountFragment().make(minerSnapshot, minerSnapshot, false, 0, false),
              TransactionFragment.prepare(
                  this.transients.conflation().number(),
                  this.transients.block().minerAddress(),
                  tx,
                  true,
                  this.transients.tx().gasPrice(),
                  this.transients.block().baseFee(),
                  this.transients.tx().initialGas())));
    } else {
      // Trace the exceptions of a transaction that could not even start
      // TODO: integrate with PCH
      // if (this.exceptions == null) {
      // this.exceptions = Exceptions.fromOutOfGas();
      // }
      // otherwise 4 account rows (sender, coinbase, sender, recipient) + 1 tx row
      Address toAddress = this.transients.tx().besuTx().getSender();
      Account toAccount = worldView.get(toAddress);
      AccountSnapshot toSnapshot =
          AccountSnapshot.fromAccount(
              toAccount,
              true,
              this.transients.conflation().deploymentInfo().number(toAddress),
              this.transients.conflation().deploymentInfo().isDeploying(toAddress));
      this.addTraceSection(
          new EndTransaction(
              this,
              this.factories.accountFragment().make(fromSnapshot, fromSnapshot, false, 0, false),
              this.factories.accountFragment().make(minerSnapshot, minerSnapshot, false, 0, false),
              this.factories.accountFragment().make(fromSnapshot, fromSnapshot, false, 0, false),
              this.factories.accountFragment().make(toSnapshot, toSnapshot, false, 0, false)));
    }
  }

  @Override
  public void enterTransaction() {
    for (Module m : this.modules) {
      m.enterTransaction();
    }
  }

  @Override
  public void traceStartTx(final WorldView world, final Transaction tx) {
    this.pch.reset();
    this.state.enter();

    this.defers.postTx(this.state.currentTxTrace());

    this.txStack.enterTransaction(tx, requiresEvmExecution(world, tx));

    this.enterTransaction();

    if (this.transients.tx().shouldSkip(world)) {
      this.transients.tx().state(TxState.TX_SKIP);
      this.processStateSkip(world);
    } else {
      this.transients.tx().state(TxState.TX_WARM);
      this.processStateWarm(world);
      this.processStateInit(world);
    }

    for (Module m : this.modules) {
      m.traceStartTx(world, tx);
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

  @Override
  public void traceEndTx(
      WorldView world,
      Transaction tx,
      boolean isSuccessful,
      Bytes output,
      List<Log> logs,
      long gasUsed) {
    this.txStack.exitTransaction(this, isSuccessful);
    if (this.transients.tx().state() != TxState.TX_SKIP) {
      this.processStateFinal(world, tx, isSuccessful);
    }

    this.defers.runPostTx(this, world, tx, isSuccessful);

    for (Module m : this.modules) {
      m.traceEndTx(world, tx, isSuccessful, output, logs, gasUsed);
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
    for (int i = 0; i < pending.lines().size(); i++) {
      StackLine line = pending.lines().get(i);

      if (line.needsResult()) {
        Bytes result = Bytes.EMPTY;
        // Only pop from the stack if no exceptions have been encountered
        if (this.pch.exceptions().none()) {
          result = frame.getStackItem(0).copy();
        }

        // This works because we are certain that the stack chunks are the first.
        ((StackFragment) section.lines().get(i).specific())
            .stackOps()
            .get(line.resultColumn() - 1)
            .value(result);
      }
    }

    if (this.pch.exceptions().none()) {
      for (TraceSection.TraceLine line : section.lines()) {
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
      final TransactionStack.MetaTransaction currentTx = transients().tx();
      final Address toAddress = effectiveToAddress(currentTx.besuTx());
      final boolean isDeployment = this.transients.tx().besuTx().getTo().isEmpty();

      final boolean shouldCopyTxCallData =
          !isDeployment && !frame.getInputData().isEmpty() && currentTx.requiresEvmExecution();
      // TODO simplify this, the same bedRock context ( = root context ??)  seems to be generated in
      // both case
      if (shouldCopyTxCallData) {
        this.callStack.newMantleAndBedrock(
            this.state.stamps().hub(),
            this.transients.tx().besuTx().getSender(),
            toAddress,
            CallFrameType.MANTLE,
            new Bytecode(
                toAddress == null
                    ? this.transients.tx().besuTx().getData().orElse(Bytes.EMPTY)
                    : Optional.ofNullable(frame.getWorldUpdater().get(toAddress))
                        .map(AccountState::getCode)
                        .orElse(Bytes.EMPTY)),
            Wei.of(this.transients.tx().besuTx().getValue().getAsBigInteger()),
            this.transients.tx().besuTx().getGasLimit(),
            this.transients.tx().besuTx().getData().orElse(Bytes.EMPTY),
            this.transients.conflation().deploymentInfo().number(toAddress),
            toAddress.isEmpty()
                ? 0
                : this.transients.conflation().deploymentInfo().number(toAddress),
            this.transients.conflation().deploymentInfo().isDeploying(toAddress));
      } else {
        this.callStack.newBedrock(
            this.state.stamps().hub(),
            //            this.transients.tx().transaction().getSender(),
            toAddress,
            CallFrameType.BEDROCK,
            new Bytecode(
                toAddress == null
                    ? this.transients.tx().besuTx().getData().orElse(Bytes.EMPTY)
                    : Optional.ofNullable(frame.getWorldUpdater().get(toAddress))
                        .map(AccountState::getCode)
                        .orElse(Bytes.EMPTY)),
            Wei.of(this.transients.tx().besuTx().getValue().getAsBigInteger()),
            this.transients.tx().besuTx().getGasLimit(),
            this.transients.tx().besuTx().getData().orElse(Bytes.EMPTY),
            this.transients.conflation().deploymentInfo().number(toAddress),
            toAddress.isEmpty()
                ? 0
                : this.transients.conflation().deploymentInfo().number(toAddress),
            this.transients.conflation().deploymentInfo().isDeploying(toAddress));
      }
    } else {
      // ...or CALL
      final boolean isDeployment = frame.getType() == MessageFrame.Type.CONTRACT_CREATION;
      final Address codeAddress = frame.getContractAddress();
      final CallFrameType frameType =
          frame.isStatic() ? CallFrameType.STATIC : CallFrameType.STANDARD;
      if (isDeployment) {
        this.transients.conflation().deploymentInfo().markDeploying(codeAddress);
      }
      final int codeDeploymentNumber =
          this.transients.conflation().deploymentInfo().number(codeAddress);

      final int callDataOffsetStackArgument =
          callStack.current().opCode().callHasSixArgument() ? 2 : 3;

      final long callDataOffset =
          isDeployment
              ? 0
              : Words.clampedToLong(
                  callStack.current().frame().getStackItem(callDataOffsetStackArgument));

      final long callDataSize =
          isDeployment
              ? 0
              : Words.clampedToLong(
                  callStack.current().frame().getStackItem(callDataOffsetStackArgument + 1));

      final long callDataContextNumber = this.callStack.current().contextNumber();

      this.callStack.enter(
          this.state.stamps().hub(),
          frame.getRecipientAddress(),
          frame.getContractAddress(),
          new Bytecode(frame.getCode().getBytes()),
          frameType,
          frame.getValue(),
          frame.getRemainingGas(),
          frame.getInputData(),
          callDataOffset,
          callDataSize,
          callDataContextNumber,
          this.transients.conflation().deploymentInfo().number(codeAddress),
          codeDeploymentNumber,
          isDeployment);

      this.defers.runNextContext(this, frame);

      for (Module m : this.modules) {
        m.traceContextEnter(frame);
      }
    }
  }

  private boolean requiresEvmExecution(final WorldView worldView, final Transaction tx) {
    Optional<? extends Address> receiver = tx.getTo();

    if (receiver.isPresent()) {
      Optional<Account> receiverInWorld = Optional.ofNullable(worldView.get(receiver.get()));

      return receiverInWorld.map(AccountState::hasCode).orElse(false);
    }

    return !tx.getInit().get().isEmpty();
  }

  public void traceContextReEnter(MessageFrame frame) {
    this.defers.runReEntry(this, frame);
    if (this.currentFrame().needsUnlatchingAtReEntry() != null) {
      this.unlatchStack(frame, this.currentFrame().needsUnlatchingAtReEntry());
      this.currentFrame().needsUnlatchingAtReEntry(null);
    }
  }

  @Override
  public void traceContextExit(MessageFrame frame) {
    if (frame.getDepth() > 0) {
      this.transients
          .conflation()
          .deploymentInfo()
          .unmarkDeploying(this.currentFrame().codeAddress());

      DeploymentExceptions contextExceptions =
          DeploymentExceptions.fromFrame(this.currentFrame(), frame);
      this.currentTraceSection().setContextExceptions(contextExceptions);
      if (contextExceptions.any()) {
        this.callStack.revert(this.state.stamps().hub());
      }

      this.callStack.exit();

      for (Module m : this.modules) {
        m.traceContextExit(frame);
      }
    }
  }

  @Override
  public void tracePreOpcode(final MessageFrame frame) {
    if (this.transients.tx().state() == TxState.TX_SKIP) {
      return;
    }
    this.processStateExec(frame);
  }

  public void tracePostExecution(MessageFrame frame, Operation.OperationResult operationResult) {
    if (this.transients.tx().state() == TxState.TX_SKIP) {
      return;
    }

    if (this.currentFrame().opCode().isCreate() && operationResult.getHaltReason() == null) {
      this.handleCreate(Words.toAddress(frame.getStackItem(0)));
    }

    this.defers.runPostExec(this, frame, operationResult);
    this.romLex.tracePostOpcode(frame);

    if (this.currentFrame().needsUnlatchingAtReEntry() == null) {
      this.unlatchStack(frame);
    }

    switch (this.opCodeData().instructionFamily()) {
      case ADD -> {
        if (this.pch.exceptions().noStackException()) {
          this.add.tracePostOpcode(frame);
        }
      }
      case MOD -> {
        if (this.pch.exceptions().noStackException()) {
          this.mod.tracePostOpcode(frame);
        }
      }
      case MUL -> {
        if (this.pch.exceptions().noStackException()) {
          this.mul.tracePostOpcode(frame);
        }
      }
      case EXT -> {
        if (this.pch.exceptions().noStackException()) {
          this.ext.tracePostOpcode(frame);
        }
      }
      case WCP -> {
        if (this.pch.exceptions().noStackException()) {
          this.wcp.tracePostOpcode(frame);
        }
      }
      case BIN -> {}
      case SHF -> {
        if (this.pch.exceptions().noStackException()) {
          this.shf.tracePostOpcode(frame);
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
          this.mxp.tracePostOpcode(frame);
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
    this.transients.conflation().deploymentInfo().deploy(target);
  }

  @Override
  public void traceStartBlock(final ProcessableBlockHeader processableBlockHeader) {
    this.transients.block().update(processableBlockHeader);
    for (Module m : this.modules) {
      m.traceStartBlock(processableBlockHeader);
    }
  }

  @Override
  public String moduleKey() {
    return "HUB";
  }

  @Override
  public void traceStartConflation(long blockCount) {
    this.transients.conflation().update();
    for (Module m : this.modules) {
      m.traceStartConflation(blockCount);
    }
  }

  @Override
  public void traceEndConflation(final WorldView state) {
    this.defers.runPostConflation(this, state);

    for (Module m : this.modules) {
      m.traceEndConflation(state);
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

  public int cumulatedTxCount() {
    return this.state.txCount();
  }

  void traceOperation(MessageFrame frame) {
    switch (this.opCodeData().instructionFamily()) {
      case ADD, MOD, SHF, BIN, WCP, EXT, BATCH, MACHINE_STATE, PUSH_POP, DUP, SWAP, INVALID -> this
          .addTraceSection(new StackOnlySection(this));
      case MUL -> {
        if (this.opCode() == OpCode.EXP) {
          this.addTraceSection(
              new ExpSection(this, ImcFragment.forOpcode(this, this.messageFrame())));
        } else {
          this.addTraceSection(new StackOnlySection(this));
        }
      }
      case HALT -> {
        final CallFrame parentFrame = this.callStack.parent();

        switch (this.opCode()) {
          case RETURN -> {
            Bytes returnData = Bytes.EMPTY;
            // Trying to read memory with absurd arguments will throw an exception
            if (pch.exceptions().none()) {
              returnData = this.transients.op().returnData();
            }
            this.currentFrame().returnDataSource(transients.op().returnDataSegment());
            this.currentFrame().returnData(returnData);
            if (!this.pch.exceptions().any() && !this.currentFrame().underDeployment()) {
              parentFrame.latestReturnData(returnData);
            } else {
              parentFrame.latestReturnData(Bytes.EMPTY);
            }
            final ImcFragment imcFragment = ImcFragment.forOpcode(this, frame); // TODO finish it
          }
          case REVERT -> {
            final Bytes returnData = this.transients.op().returnData();
            this.currentFrame().returnDataSource(transients.op().returnDataSegment());
            this.currentFrame().returnData(returnData);
            if (!this.pch.exceptions().any()) {
              parentFrame.latestReturnData(returnData);
            } else {
              parentFrame.latestReturnData(Bytes.EMPTY);
            }
            final ImcFragment imcFragment = ImcFragment.forOpcode(this, frame); // TODO finish it
          }
          case STOP, SELFDESTRUCT -> parentFrame.latestReturnData(Bytes.EMPTY);
        }

        this.addTraceSection(new StackOnlySection(this));
      }
      case KEC -> this.addTraceSection(
          new KeccakSection(
              this, this.currentFrame(), ImcFragment.forOpcode(this, this.messageFrame())));
      case CONTEXT -> this.addTraceSection(
          new ContextLogSection(this, ContextFragment.readContextData(callStack)));
      case LOG -> {
        this.addTraceSection(
            new ContextLogSection(this, ContextFragment.readContextData(callStack)));
        LogInvocation.forOpcode(this);
      }
      case ACCOUNT -> {
        TraceSection accountSection = new AccountSection(this);
        if (this.opCodeData().stackSettings().flag1()) {
          accountSection.addFragment(
              this, this.currentFrame(), ContextFragment.readContextData(callStack));
        }

        final Bytes rawTargetAddress =
            switch (this.currentFrame().opCode()) {
              case BALANCE, EXTCODESIZE, EXTCODEHASH -> frame.getStackItem(0);
              default -> this.currentFrame().address();
            };
        final Address targetAddress = Words.toAddress(rawTargetAddress);
        final Account targetAccount = frame.getWorldUpdater().get(targetAddress);
        final AccountSnapshot accountSnapshot =
            AccountSnapshot.fromAccount(
                targetAccount,
                frame.isAddressWarm(targetAddress),
                this.transients.conflation().deploymentInfo().number(targetAddress),
                this.transients.conflation().deploymentInfo().isDeploying(targetAddress));
        accountSection.addFragment(
            this,
            this.currentFrame(),
            this.factories
                .accountFragment()
                .makeWithTrm(accountSnapshot, accountSnapshot, false, 0, false, rawTargetAddress));

        this.addTraceSection(accountSection);
      }
      case COPY -> {
        TraceSection copySection = new CopySection(this);
        if (this.opCodeData().stackSettings().flag1()) {

          final Bytes rawTargetAddress =
              switch (this.currentFrame().opCode()) {
                case CODECOPY -> this.currentFrame().codeAddress();
                case EXTCODECOPY -> frame.getStackItem(0);
                default -> throw new IllegalStateException("unexpected opcode");
              };
          final Address targetAddress = Words.toAddress(rawTargetAddress);
          final Account targetAccount = frame.getWorldUpdater().get(targetAddress);
          AccountSnapshot accountSnapshot =
              AccountSnapshot.fromAccount(
                  targetAccount,
                  frame.isAddressWarm(targetAddress),
                  this.transients.conflation().deploymentInfo().number(targetAddress),
                  this.transients.conflation().deploymentInfo().isDeploying(targetAddress));

          copySection.addFragment(
              this,
              this.currentFrame(),
              this.currentFrame().opCode() == OpCode.EXTCODECOPY
                  ? this.factories
                      .accountFragment()
                      .makeWithTrm(
                          accountSnapshot, accountSnapshot, false, 0, false, rawTargetAddress)
                  : this.factories
                      .accountFragment()
                      .make(accountSnapshot, accountSnapshot, false, 0, false));
        } else {
          copySection.addFragment(
              this, this.currentFrame(), ContextFragment.readContextData(callStack));
        }
        this.addTraceSection(copySection);
      }
      case TRANSACTION -> this.addTraceSection(
          new TransactionSection(
              this,
              TransactionFragment.prepare(
                  this.transients.conflation().number(),
                  frame.getMiningBeneficiary(),
                  this.transients.tx().besuTx(),
                  true,
                  frame.getGasPrice(),
                  frame.getBlockValues().getBaseFee().orElse(Wei.ZERO),
                  this.transients.tx().initialGas())));
      case STACK_RAM -> {
        switch (this.currentFrame().opCode()) {
          case CALLDATALOAD -> {
            final ImcFragment imcFragment = ImcFragment.forOpcode(this, frame);

            this.addTraceSection(
                new StackRam(this, imcFragment, ContextFragment.readContextData(callStack)));
          }
          case MLOAD, MSTORE, MSTORE8 -> this.addTraceSection(
              new StackRam(this, ImcFragment.forOpcode(this, frame)));
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
                    ContextFragment.readContextData(callStack),
                    new StorageFragment(
                        address,
                        this.currentFrame().accountDeploymentNumber(),
                        key,
                        this.transients
                            .tx()
                            .storage()
                            .getOriginalValueOrUpdate(address, key, valNext),
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
                    ContextFragment.readContextData(callStack),
                    new StorageFragment(
                        address,
                        this.currentFrame().accountDeploymentNumber(),
                        key,
                        this.transients.tx().storage().getOriginalValueOrUpdate(address, key),
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
                this.transients.conflation().deploymentInfo().number(myAddress),
                this.transients.conflation().deploymentInfo().isDeploying(myAddress));

        Address createdAddress = this.currentFrame().address();
        Account createdAccount = frame.getWorldUpdater().get(createdAddress);
        AccountSnapshot createdAccountSnapshot =
            AccountSnapshot.fromAccount(
                createdAccount,
                frame.isAddressWarm(createdAddress),
                this.transients.conflation().deploymentInfo().number(createdAddress),
                this.transients.conflation().deploymentInfo().isDeploying(createdAddress));

        CreateSection createSection =
            new CreateSection(this, myAccountSnapshot, createdAccountSnapshot);
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
                this.transients.conflation().deploymentInfo().number(myAddress),
                this.transients.conflation().deploymentInfo().isDeploying(myAddress));

        final Bytes rawCalledAddress = frame.getStackItem(1);
        final Address calledAddress = Words.toAddress(rawCalledAddress);
        final Optional<Account> calledAccount =
            Optional.ofNullable(frame.getWorldUpdater().get(calledAddress));
        final boolean hasCode = calledAccount.map(AccountState::hasCode).orElse(false);

        final AccountSnapshot calledAccountSnapshot =
            AccountSnapshot.fromAccount(
                calledAccount,
                frame.isAddressWarm(myAddress),
                this.transients.conflation().deploymentInfo().number(myAddress),
                this.transients.conflation().deploymentInfo().isDeploying(myAddress));

        Optional<Precompile> targetPrecompile = Precompile.maybeOf(calledAddress);

        if (this.pch().exceptions().any()) {
          //
          // THERE IS AN EXCEPTION
          //
          if (this.pch().exceptions().staticFault()) {
            this.addTraceSection(
                new FailedCallSection(
                    this,
                    ScenarioFragment.forCall(this, hasCode),
                    ImcFragment.forCall(this, myAccount, calledAccount),
                    ContextFragment.readContextData(callStack)));
          } else if (this.pch().exceptions().outOfMemoryExpansion()) {
            this.addTraceSection(
                new FailedCallSection(
                    this,
                    ScenarioFragment.forCall(this, hasCode),
                    ImcFragment.forCall(this, myAccount, calledAccount)));
          } else if (this.pch().exceptions().outOfGas()) {
            this.addTraceSection(
                new FailedCallSection(
                    this,
                    ScenarioFragment.forCall(this, hasCode),
                    ImcFragment.forCall(this, myAccount, calledAccount),
                    this.factories
                        .accountFragment()
                        .makeWithTrm(
                            calledAccountSnapshot, calledAccountSnapshot, rawCalledAddress)));
          }
        } else if (this.pch.aborts().any()) {
          //
          // THERE IS AN ABORT
          //
          TraceSection abortedSection =
              new FailedCallSection(
                  this,
                  ScenarioFragment.forCall(this, hasCode),
                  ImcFragment.forCall(this, myAccount, calledAccount),
                  ContextFragment.readContextData(callStack),
                  this.factories.accountFragment().make(myAccountSnapshot, myAccountSnapshot),
                  this.factories
                      .accountFragment()
                      .makeWithTrm(calledAccountSnapshot, calledAccountSnapshot, rawCalledAddress),
                  ContextFragment.nonExecutionEmptyReturnData(callStack));
          this.addTraceSection(abortedSection);
        } else {
          final ImcFragment imcFragment = ImcFragment.forOpcode(this, frame);

          if (hasCode) {
            final SmartContractCallSection section =
                new SmartContractCallSection(
                    this, myAccountSnapshot, calledAccountSnapshot, rawCalledAddress, imcFragment);
            this.addTraceSection(section);
            this.currentFrame().needsUnlatchingAtReEntry(section);
          } else {
            //
            // CALL EXECUTED
            //

            // TODO: fill the callee & requested return data for the current call frame
            // TODO: i.e. ensure that the precompile frame behaves as expected

            Optional<PrecompileInvocation> precompileInvocation =
                targetPrecompile.map(p -> PrecompileInvocation.of(this, p));

            // TODO: this is ugly, and surely not at the right place. It should provide the
            // precompile result (from the precompile module)
            // TODO useless (and potentially dangerous) if the precompile is a failure
            if (targetPrecompile.isPresent()) {
              this.callStack.newPrecompileResult(
                  this.stamp(), Bytes.EMPTY, 0, targetPrecompile.get().address);
            }

            final NoCodeCallSection section =
                new NoCodeCallSection(
                    this,
                    precompileInvocation,
                    myAccountSnapshot,
                    calledAccountSnapshot,
                    rawCalledAddress,
                    imcFragment);
            this.addTraceSection(section);
            this.currentFrame().needsUnlatchingAtReEntry(section);
          }
        }
      }

      case JUMP -> {
        AccountSnapshot codeAccountSnapshot =
            AccountSnapshot.fromAccount(
                frame.getWorldUpdater().get(this.currentFrame().codeAddress()),
                true,
                this.transients
                    .conflation()
                    .deploymentInfo()
                    .number(this.currentFrame().codeAddress()),
                this.currentFrame().underDeployment());

        JumpSection jumpSection =
            new JumpSection(
                this,
                ContextFragment.readContextData(callStack),
                this.factories
                    .accountFragment()
                    .make(codeAccountSnapshot, codeAccountSnapshot, false, 0, false),
                ImcFragment.forOpcode(this, frame));

        this.addTraceSection(jumpSection);
      }
    }

    // In all cases, add a context fragment if an exception occurred
    if (this.pch().exceptions().any()) {
      this.currentTraceSection()
          .addFragment(
              this, this.currentFrame(), ContextFragment.executionEmptyReturnData(callStack));
    }
  }
}
