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

package net.consensys.linea.testing;

import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LINEA_BASE_FEE;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LINEA_BLOCK_GAS_LIMIT;
import static net.consensys.linea.zktracer.module.constants.GlobalConstants.LINEA_DIFFICULTY;
import static net.consensys.linea.zktracer.runtime.stack.Stack.MAX_STACK_SIZE;

import java.math.BigInteger;
import java.util.*;
import java.util.function.Supplier;
import java.util.stream.Collectors;

import lombok.Builder;
import lombok.Singular;
import lombok.extern.slf4j.Slf4j;
import net.consensys.linea.zktracer.ZkTracer;
import net.consensys.linea.zktracer.module.constants.GlobalConstants;
import org.hyperledger.besu.datatypes.*;
import org.hyperledger.besu.ethereum.core.*;
import org.hyperledger.besu.ethereum.core.Transaction;
import org.hyperledger.besu.ethereum.core.feemarket.CoinbaseFeePriceCalculator;
import org.hyperledger.besu.ethereum.mainnet.LondonTargetingGasLimitCalculator;
import org.hyperledger.besu.ethereum.mainnet.MainnetTransactionProcessor;
import org.hyperledger.besu.ethereum.mainnet.TransactionValidatorFactory;
import org.hyperledger.besu.ethereum.mainnet.feemarket.FeeMarket;
import org.hyperledger.besu.ethereum.mainnet.feemarket.LondonFeeMarket;
import org.hyperledger.besu.ethereum.referencetests.GeneralStateTestCaseEipSpec;
import org.hyperledger.besu.ethereum.referencetests.ReferenceTestWorldState;
import org.hyperledger.besu.evm.EVM;
import org.hyperledger.besu.evm.MainnetEVMs;
import org.hyperledger.besu.evm.gascalculator.GasCalculator;
import org.hyperledger.besu.evm.internal.EvmConfiguration;
import org.hyperledger.besu.evm.precompile.MainnetPrecompiledContracts;
import org.hyperledger.besu.evm.precompile.PrecompileContractRegistry;
import org.hyperledger.besu.evm.processor.ContractCreationProcessor;
import org.hyperledger.besu.evm.processor.MessageCallProcessor;

@Builder
@Slf4j
public class ToyExecutionEnvironmentV2 {
  public static final BigInteger CHAIN_ID = BigInteger.valueOf(1337);

  private static final Wei DEFAULT_BASE_FEE = Wei.of(LINEA_BASE_FEE);

  private static final GasCalculator gasCalculator = ZkTracer.gasCalculator;
  private static final Address minerAddress = Address.fromHexString("0x1234532342");
  private static final long DEFAULT_BLOCK_NUMBER = 6678980;
  private static final long DEFAULT_TIME_STAMP = 1347310;
  private static final Hash DEFAULT_HASH =
      Hash.fromHexStringLenient("0xdeadbeef123123666dead666dead666");
  private static final FeeMarket feeMarket = FeeMarket.london(-1);

  @Builder.Default private BigInteger chainId = CHAIN_ID;
  private final ToyWorld toyWorld;
  @Singular private final List<Transaction> transactions;

  public void run() {
    final EVM evm = MainnetEVMs.london(this.chainId, EvmConfiguration.DEFAULT);
    GeneralStateReferenceTestTools.executeTest(
        this.buildGeneralStateTestCaseSpec(evm), getMainnetTransactionProcessor(evm), feeMarket);
  }

  public GeneralStateTestCaseEipSpec buildGeneralStateTestCaseSpec(EVM evm) {
    Map<String, ReferenceTestWorldState.AccountMock> accountMockMap =
        toyWorld.getAddressAccountMap().entrySet().stream()
            .collect(
                Collectors.toMap(
                    entry -> entry.getKey().toHexString(),
                    entry -> entry.getValue().toAccountMock()));
    ReferenceTestWorldState referenceTestWorldState =
        ReferenceTestWorldState.create(accountMockMap, evm.getEvmConfiguration());
    BlockHeader blockHeader =
        BlockHeaderBuilder.createDefault()
            .baseFee(DEFAULT_BASE_FEE)
            .gasLimit(LINEA_BLOCK_GAS_LIMIT)
            .difficulty(Difficulty.of(LINEA_DIFFICULTY))
            .number(DEFAULT_BLOCK_NUMBER)
            .coinbase(minerAddress)
            .timestamp(DEFAULT_TIME_STAMP)
            .parentHash(DEFAULT_HASH)
            .buildBlockHeader();

    List<Supplier<Transaction>> txSuppliers = new ArrayList<>();
    for (Transaction tx : transactions) {
      txSuppliers.add(() -> tx);
    }

    return new GeneralStateTestCaseEipSpec(
        /*fork*/ evm.getEvmVersion().getName().toLowerCase(),
        txSuppliers,
        referenceTestWorldState,
        /*expectedRootHash*/ null,
        /*expectedLogsHash*/ null,
        blockHeader,
        /*dataIndex*/ -1,
        /*gasIndex*/ -1,
        /*valueIndex*/ -1,
        /*expectException*/ null);
  }

  private MainnetTransactionProcessor getMainnetTransactionProcessor(EVM evm) {
    PrecompileContractRegistry precompileContractRegistry = new PrecompileContractRegistry();

    MainnetPrecompiledContracts.populateForIstanbul(
        precompileContractRegistry, evm.getGasCalculator());

    final MessageCallProcessor messageCallProcessor =
        new MessageCallProcessor(evm, precompileContractRegistry);

    final ContractCreationProcessor contractCreationProcessor =
        new ContractCreationProcessor(evm, false, List.of(), 0);

    return new MainnetTransactionProcessor(
        gasCalculator,
        new TransactionValidatorFactory(
            gasCalculator,
            new LondonTargetingGasLimitCalculator(0L, new LondonFeeMarket(0)),
            new LondonFeeMarket(0L),
            false,
            Optional.of(this.chainId),
            Set.of(TransactionType.FRONTIER, TransactionType.ACCESS_LIST, TransactionType.EIP1559),
            GlobalConstants.MAX_CODE_SIZE),
        contractCreationProcessor,
        messageCallProcessor,
        true,
        true,
        MAX_STACK_SIZE,
        feeMarket,
        CoinbaseFeePriceCalculator.eip1559());
  }
}
