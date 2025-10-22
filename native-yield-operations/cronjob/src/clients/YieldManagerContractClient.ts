import { IContractClientLibrary } from "ts-libs/linea-shared-utils/src/core/client/IContractClientLibrary";
import {
  Address,
  encodeFunctionData,
  getContract,
  GetContractReturnType,
  PublicClient,
  TransactionReceipt,
} from "viem";
import {
  encodeLidoWithdrawalParams,
  LidoStakingVaultWithdrawalParams,
} from "../core/entities/LidoStakingVaultWithdrawalParams";
import { RebalanceRequirement, RebalanceDirection } from "../core/entities/RebalanceRequirement";

import { YieldManagerABI } from "../core/abis/YieldManager";
import { IYieldManager } from "../core/services/contracts/IYieldManager";
import { IBaseContractClient } from "../core/clients/IBaseContractClient";

export class YieldManagerContractClient implements IYieldManager<TransactionReceipt>, IBaseContractClient {
  private readonly contract: GetContractReturnType<typeof YieldManagerABI, PublicClient, Address>;

  constructor(
    private readonly contractClientLibrary: IContractClientLibrary<PublicClient, TransactionReceipt>,
    private readonly contractAddress: Address,
    private readonly rebalanceToleranceBps: number,
  ) {
    this.contractClientLibrary = contractClientLibrary;
    this.contractAddress = contractAddress;
    this.contract = getContract({
      abi: YieldManagerABI,
      address: contractAddress,
      client: this.contractClientLibrary.getBlockchainClient(),
    });
  }

  getAddress(): Address {
    return this.contractAddress;
  }

  getContract(): GetContractReturnType {
    return this.contract;
  }

  async L1_MESSAGE_SERVICE(): Promise<Address> {
    return this.contract.read.L1_MESSAGE_SERVICE();
  }

  async getTotalSystemBalance(): Promise<bigint> {
    return this.contract.read.getTotalSystemBalance();
  }

  async getEffectiveTargetWithdrawalReserve(): Promise<bigint> {
    return this.contract.read.getEffectiveTargetWithdrawalReserve();
  }

  async getTargetReserveDeficit(): Promise<bigint> {
    return this.contract.read.getTargetReserveDeficit();
  }

  async isStakingPaused(yieldProvider: Address): Promise<boolean> {
    return this.contract.read.isStakingPaused([yieldProvider]);
  }

  async isOssificationInitiated(yieldProvider: Address): Promise<boolean> {
    return this.contract.read.isOssificationInitiated([yieldProvider]);
  }

  async isOssified(yieldProvider: Address): Promise<boolean> {
    return this.contract.read.isOssified([yieldProvider]);
  }

  async withdrawableValue(yieldProvider: Address): Promise<bigint> {
    const { result } = await this.contract.simulate.withdrawableValue([yieldProvider]);
    return result;
  }

  async fundYieldProvider(yieldProvider: Address, amount: bigint): Promise<TransactionReceipt | null> {
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "fundYieldProvider",
      args: [yieldProvider, amount],
    });

    return this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
  }

  async transferFundsToReserve(amount: bigint): Promise<TransactionReceipt | null> {
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "transferFundsToReserve",
      args: [amount],
    });

    return this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
  }

  async reportYield(yieldProvider: Address, l2YieldRecipient: Address): Promise<TransactionReceipt | null> {
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "reportYield",
      args: [yieldProvider, l2YieldRecipient],
    });

    return this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
  }

  async unstake(
    yieldProvider: Address,
    withdrawalParams: LidoStakingVaultWithdrawalParams,
  ): Promise<TransactionReceipt | null> {
    const encodedWithdrawalParams = encodeLidoWithdrawalParams(withdrawalParams);
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "unstake",
      args: [yieldProvider, encodedWithdrawalParams],
    });

    return this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
  }

  async withdrawFromYieldProvider(yieldProvider: Address, amount: bigint): Promise<TransactionReceipt | null> {
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "withdrawFromYieldProvider",
      args: [yieldProvider, amount],
    });

    return this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
  }

  async addToWithdrawalReserve(yieldProvider: Address, amount: bigint): Promise<TransactionReceipt | null> {
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "addToWithdrawalReserve",
      args: [yieldProvider, amount],
    });

    return this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
  }

  async safeAddToWithdrawalReserve(yieldProvider: Address, amount: bigint): Promise<TransactionReceipt | null> {
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "safeAddToWithdrawalReserve",
      args: [yieldProvider, amount],
    });

    return this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
  }

  async pauseStaking(yieldProvider: Address): Promise<TransactionReceipt | null> {
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "pauseStaking",
      args: [yieldProvider],
    });

    return this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
  }

  async unpauseStaking(yieldProvider: Address): Promise<TransactionReceipt | null> {
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "unpauseStaking",
      args: [yieldProvider],
    });

    return this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
  }

  async progressPendingOssification(yieldProvider: Address): Promise<TransactionReceipt | null> {
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "progressPendingOssification",
      args: [yieldProvider],
    });

    return this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
  }

  async getRebalanceRequirements(): Promise<RebalanceRequirement> {
    const l1MessageServiceAddress = await this.L1_MESSAGE_SERVICE();
    const [l1MessageServiceBalance, totalSystemBalance, effectiveTargetWithdrawalReserve] = await Promise.all([
      this.contractClientLibrary.getBalance(l1MessageServiceAddress),
      this.getTotalSystemBalance(),
      this.getEffectiveTargetWithdrawalReserve(),
    ]);
    const isRebalanceRequired = this._isRebalanceRequired(
      totalSystemBalance,
      l1MessageServiceBalance,
      effectiveTargetWithdrawalReserve,
    );
    if (!isRebalanceRequired) {
      return {
        rebalanceDirection: RebalanceDirection.NONE,
        rebalanceAmount: 0n,
      };
    }
    // In deficit
    if (l1MessageServiceBalance < effectiveTargetWithdrawalReserve) {
      return {
        rebalanceDirection: RebalanceDirection.UNSTAKE,
        rebalanceAmount: effectiveTargetWithdrawalReserve - l1MessageServiceBalance,
      };
      // In surplus
    } else {
      return {
        rebalanceDirection: RebalanceDirection.STAKE,
        rebalanceAmount: l1MessageServiceBalance - effectiveTargetWithdrawalReserve,
      };
    }
  }

  private _isRebalanceRequired(
    totalSystemBalance: bigint,
    l1MessageServiceBalance: bigint,
    effectiveTargetWithdrawalReserve: bigint,
  ): boolean {
    const toleranceBand = (totalSystemBalance * BigInt(this.rebalanceToleranceBps)) / 10000n;
    // Below tolerance band
    if (l1MessageServiceBalance < effectiveTargetWithdrawalReserve - toleranceBand) {
      return true;
    }
    // Above tolerance band
    if (l1MessageServiceBalance < effectiveTargetWithdrawalReserve + toleranceBand) {
      return true;
    }
    return false;
  }
}
