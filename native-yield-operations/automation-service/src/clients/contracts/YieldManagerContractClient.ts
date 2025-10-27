import { IBlockchainClient, ILogger } from "@consensys/linea-shared-utils";
import {
  Address,
  decodeEventLog,
  encodeAbiParameters,
  encodeFunctionData,
  getContract,
  GetContractReturnType,
  Hex,
  PublicClient,
  TransactionReceipt,
} from "viem";
import {
  LidoStakingVaultWithdrawalParams,
  WithdrawalRequests,
} from "../../core/entities/LidoStakingVaultWithdrawalParams.js";
import { RebalanceRequirement, RebalanceDirection } from "../../core/entities/RebalanceRequirement.js";

import { YieldManagerABI } from "../../core/abis/YieldManager.js";
import { IYieldManager, YieldProviderData } from "../../core/clients/contracts/IYieldManager.js";
import { ONE_ETHER } from "@consensys/linea-shared-utils";

export class YieldManagerContractClient implements IYieldManager<TransactionReceipt> {
  private readonly contract: GetContractReturnType<typeof YieldManagerABI, PublicClient, Address>;

  constructor(
    private readonly logger: ILogger,
    private readonly contractClientLibrary: IBlockchainClient<PublicClient, TransactionReceipt>,
    private readonly contractAddress: Address,
    private readonly rebalanceToleranceBps: number,
    private readonly minWithdrawalThresholdEth: bigint,
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

  async getYieldProviderData(yieldProvider: Address): Promise<YieldProviderData> {
    return this.contract.read.getYieldProviderData([yieldProvider]);
  }

  async fundYieldProvider(yieldProvider: Address, amount: bigint): Promise<TransactionReceipt> {
    this.logger.debug(`fundYieldProvider started, yieldProvider=${yieldProvider}, amount=${amount.toString()}`);
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "fundYieldProvider",
      args: [yieldProvider, amount],
    });

    const txReceipt = await this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
    this.logger.info(
      `fundYieldProvider succeeded, yieldProvider=${yieldProvider}, amount=${amount.toString()}, txHash=${txReceipt.transactionHash}`,
    );
    return txReceipt;
  }

  async transferFundsToReserve(amount: bigint): Promise<TransactionReceipt> {
    this.logger.debug(`transferFundsToReserve started, amount=${amount.toString()}`);
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "transferFundsToReserve",
      args: [amount],
    });

    const txReceipt = await this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
    this.logger.info(
      `transferFundsToReserve succeeded, amount=${amount.toString()}, txHash=${txReceipt.transactionHash}`,
    );
    return txReceipt;
  }

  async reportYield(yieldProvider: Address, l2YieldRecipient: Address): Promise<TransactionReceipt> {
    this.logger.debug(`reportYield started, yieldProvider=${yieldProvider}, l2YieldRecipient=${l2YieldRecipient}`);
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "reportYield",
      args: [yieldProvider, l2YieldRecipient],
    });

    const txReceipt = await this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
    this.logger.info(
      `reportYield succeeded, yieldProvider=${yieldProvider}, l2YieldRecipient=${l2YieldRecipient}, txHash=${txReceipt.transactionHash}`,
    );
    return txReceipt;
  }

  async unstake(yieldProvider: Address, withdrawalParams: WithdrawalRequests): Promise<TransactionReceipt> {
    this.logger.debug(`unstake started, yieldProvider=${yieldProvider}`, { withdrawalParams });
    const encodedWithdrawalParams = this._encodeLidoWithdrawalParams({
      ...withdrawalParams,
      refundRecipient: this.contractAddress,
    });
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "unstake",
      args: [yieldProvider, encodedWithdrawalParams],
    });

    const txReceipt = await this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
    this.logger.info(`unstake succeeded, yieldProvider=${yieldProvider}, txHash=${txReceipt.transactionHash}`, {
      withdrawalParams,
    });
    return txReceipt;
  }

  async withdrawFromYieldProvider(yieldProvider: Address, amount: bigint): Promise<TransactionReceipt> {
    this.logger.debug(`withdrawFromYieldProvider started, yieldProvider=${yieldProvider}, amount=${amount.toString()}`);
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "withdrawFromYieldProvider",
      args: [yieldProvider, amount],
    });

    const txReceipt = await this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
    this.logger.info(
      `withdrawFromYieldProvider succeeded, yieldProvider=${yieldProvider}, amount=${amount.toString()}, txHash=${txReceipt.transactionHash}`,
    );
    return txReceipt;
  }

  async addToWithdrawalReserve(yieldProvider: Address, amount: bigint): Promise<TransactionReceipt> {
    this.logger.debug(`addToWithdrawalReserve started, yieldProvider=${yieldProvider}, amount=${amount.toString()}`);
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "addToWithdrawalReserve",
      args: [yieldProvider, amount],
    });

    const txReceipt = await this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
    this.logger.info(
      `addToWithdrawalReserve succeeded, yieldProvider=${yieldProvider}, amount=${amount.toString()}, txHash=${txReceipt.transactionHash}`,
    );
    return txReceipt;
  }

  async safeAddToWithdrawalReserve(yieldProvider: Address, amount: bigint): Promise<TransactionReceipt> {
    this.logger.debug(
      `safeAddToWithdrawalReserve started, yieldProvider=${yieldProvider}, amount=${amount.toString()}`,
    );
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "safeAddToWithdrawalReserve",
      args: [yieldProvider, amount],
    });

    const txReceipt = await this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
    this.logger.info(
      `safeAddToWithdrawalReserve succeeded, yieldProvider=${yieldProvider}, amount=${amount.toString()}, txHash=${txReceipt.transactionHash}`,
    );
    return txReceipt;
  }

  async pauseStaking(yieldProvider: Address): Promise<TransactionReceipt> {
    this.logger.debug(`pauseStaking started, yieldProvider=${yieldProvider}`);
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "pauseStaking",
      args: [yieldProvider],
    });

    const txReceipt = await this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
    this.logger.info(`pauseStaking succeeded, yieldProvider=${yieldProvider}, txHash=${txReceipt.transactionHash}`);
    return txReceipt;
  }

  async unpauseStaking(yieldProvider: Address): Promise<TransactionReceipt> {
    this.logger.debug(`unpauseStaking started, yieldProvider=${yieldProvider}`);
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "unpauseStaking",
      args: [yieldProvider],
    });

    const txReceipt = await this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
    this.logger.info(`unpauseStaking succeeded, yieldProvider=${yieldProvider}, txHash=${txReceipt.transactionHash}`);
    return txReceipt;
  }

  async progressPendingOssification(yieldProvider: Address): Promise<TransactionReceipt> {
    this.logger.debug(`progressPendingOssification started, yieldProvider=${yieldProvider}`);
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "progressPendingOssification",
      args: [yieldProvider],
    });

    const txReceipt = await this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
    this.logger.info(
      `progressPendingOssification succeeded, yieldProvider=${yieldProvider}, txHash=${txReceipt.transactionHash}`,
    );
    return txReceipt;
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
    if (l1MessageServiceBalance > effectiveTargetWithdrawalReserve + toleranceBand) {
      return true;
    }
    return false;
  }

  private _encodeLidoWithdrawalParams(params: LidoStakingVaultWithdrawalParams): Hex {
    return encodeAbiParameters(
      [
        {
          type: "tuple",
          components: [
            { name: "pubkeys", type: "bytes[]" },
            { name: "amounts", type: "uint64[]" },
            { name: "refundRecipient", type: "address" },
          ],
        },
      ],
      [
        {
          pubkeys: params.pubkeys,
          amounts: params.amountsGwei,
          refundRecipient: params.refundRecipient,
        },
      ],
    );
  }

  async getLidoStakingVaultAddress(yieldProvider: Address): Promise<Address> {
    const yieldProviderData = await this.getYieldProviderData(yieldProvider);
    return yieldProviderData.ossifiedEntrypoint;
  }

  async pauseStakingIfNotAlready(yieldProvider: Address): Promise<TransactionReceipt | null> {
    if (!(await this.isStakingPaused(yieldProvider))) {
      const txReceipt = await this.pauseStaking(yieldProvider);
      return txReceipt;
    }
    this.logger.info(`Already paused staking for yieldProvider=${yieldProvider}`);
    return null;
  }

  async unpauseStakingIfNotAlready(yieldProvider: Address): Promise<TransactionReceipt | null> {
    if (await this.isStakingPaused(yieldProvider)) {
      const txReceipt = await this.unpauseStaking(yieldProvider);
      return txReceipt;
    }
    this.logger.info(`Already resumed staking for yieldProvider=${yieldProvider}`);
    return null;
  }

  async getAvailableUnstakingRebalanceBalance(yieldProvider: Address): Promise<bigint> {
    const [yieldManagerBalance, yieldProviderWithdrawableBalance] = await Promise.all([
      this.contractClientLibrary.getBalance(this.contractAddress),
      this.withdrawableValue(yieldProvider),
    ]);
    return yieldManagerBalance + yieldProviderWithdrawableBalance;
  }

  async safeAddToWithdrawalReserveIfAboveThreshold(
    yieldProvider: Address,
    amount: bigint,
  ): Promise<TransactionReceipt | null> {
    const availableWithdrawalBalance = await this.getAvailableUnstakingRebalanceBalance(yieldProvider);
    if (availableWithdrawalBalance < this.minWithdrawalThresholdEth * ONE_ETHER) return null;
    return await this.safeAddToWithdrawalReserve(yieldProvider, amount);
  }

  async safeMaxAddToWithdrawalReserve(yieldProvider: Address): Promise<TransactionReceipt | null> {
    const availableWithdrawalBalance = await this.getAvailableUnstakingRebalanceBalance(yieldProvider);
    if (availableWithdrawalBalance < this.minWithdrawalThresholdEth * ONE_ETHER) return null;
    return await this.safeAddToWithdrawalReserve(yieldProvider, availableWithdrawalBalance);
  }

  getWithdrawalAmountFromTxReceipt(txReceipt: TransactionReceipt): bigint {
    for (const log of txReceipt.logs) {
      // Only decode logs emitted by this contract
      if (log.address.toLowerCase() !== this.contractAddress.toLowerCase()) continue;

      try {
        const decoded = decodeEventLog({
          abi: this.contract.abi,
          data: log.data,
          topics: log.topics,
        });

        if (decoded.eventName === "WithdrawalReserveAugmented") {
          const { reserveIncrementAmount } = decoded.args;
          return reserveIncrementAmount as bigint;
        }
      } catch {
        // skip unrelated logs (from the same contract or different ABIs)
      }
    }

    // If event not found
    return 0n;
  }
}
