import { IBlockchainClient, ILogger } from "@consensys/linea-shared-utils";
import {
  Address,
  encodeFunctionData,
  getContract,
  GetContractReturnType,
  PublicClient,
  TransactionReceipt,
} from "viem";
import { LazyOracleABI } from "../../core/abis/LazyOracle.js";
import {
  ILazyOracle,
  UpdateVaultDataParams,
  WaitForVaultsReportDataUpdatedEventReturnType,
  LazyOracleReportData,
} from "../../core/clients/contracts/ILazyOracle.js";

export class LazyOracleContractClient implements ILazyOracle<TransactionReceipt> {
  private readonly contract: GetContractReturnType<typeof LazyOracleABI, PublicClient, Address>;
  constructor(
    private readonly logger: ILogger,
    private readonly contractClientLibrary: IBlockchainClient<PublicClient, TransactionReceipt>,
    private readonly contractAddress: Address,
    private readonly pollIntervalMs: number,
  ) {
    this.contract = getContract({
      abi: LazyOracleABI,
      address: contractAddress,
      client: contractClientLibrary.getBlockchainClient(),
    });
  }

  getAddress(): Address {
    return this.contractAddress;
  }

  getContract(): GetContractReturnType {
    return this.contract;
  }

  async latestReportData(): Promise<LazyOracleReportData> {
    const resp = await this.contract.read.latestReportData();
    const returnVal = {
      timestamp: resp[0],
      refSlot: resp[1],
      treeRoot: resp[2],
      reportCid: resp[3],
    }
    this.logger.debug("latestReportData", { returnVal });
    return returnVal;
  }

  async updateVaultData(params: UpdateVaultDataParams): Promise<TransactionReceipt> {
    const { vault, totalValue, cumulativeLidoFees, liabilityShares, maxLiabilityShares, slashingReserve, proof } =
      params;
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "updateVaultData",
      args: [vault, totalValue, cumulativeLidoFees, liabilityShares, maxLiabilityShares, slashingReserve, proof],
    });
    this.logger.debug(`updateVaultData started`, { params });
    const txReceipt = await this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
    this.logger.info(`updateVaultData succeeded, txHash=${txReceipt.transactionHash}`, { params });
    return txReceipt;
  }

  async simulateUpdateVaultData(params: UpdateVaultDataParams): Promise<void> {
    const { vault, totalValue, cumulativeLidoFees, liabilityShares, maxLiabilityShares, slashingReserve, proof } =
      params;
    await this.contract.simulate.updateVaultData([
      vault,
      totalValue,
      cumulativeLidoFees,
      liabilityShares,
      maxLiabilityShares,
      slashingReserve,
      proof,
    ]);
  }

  waitForVaultsReportDataUpdatedEvent(): WaitForVaultsReportDataUpdatedEventReturnType {
    // Create placeholder variable. Initialize to empty fn so it's callable before being reassigned (TS safety).
    let resolvePromise: () => void = () => {};
    // Create the waitForEvent Promise, create a 2nd reference to 'resolve' fn
    // Decouple Promise creation from resolve fn creation
    const waitForEvent = new Promise<void>((resolve) => {
      resolvePromise = resolve;
    });

    this.logger.info("waitForVaultsReportDataUpdatedEvent started...");
    const unwatch = this.contractClientLibrary.getBlockchainClient().watchContractEvent({
      address: this.contractAddress,
      abi: this.contract.abi,
      eventName: "VaultsReportDataUpdated",
      pollingInterval: this.pollIntervalMs,
      onLogs: (logs) => {
        // Filter out reorgs
        const valid = logs.find((l) => !l.removed);
        if (!valid) {
          this.logger.warn("waitForVaultsReportDataUpdatedEvent: Dropped VaultsReportDataUpdated event")
          return;
        }
        this.logger.info("waitForVaultsReportDataUpdatedEvent succeeded, VaultsReportDataUpdated event detected");
        // Call resolve through 2nd reference
        resolvePromise();
      },
      onError: (error) => {
        this.logger.error("waitForVaultsReportDataUpdatedEvent error", { error });
      },
    });

    return { unwatch, waitForEvent };
  }
}
