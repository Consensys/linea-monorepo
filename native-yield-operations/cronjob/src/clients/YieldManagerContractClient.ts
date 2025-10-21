import { IContractClientLibrary } from "ts-libs/linea-shared-utils/src/core/client/IContractClientLibrary";
import { IYieldManager } from "../core/services/contracts/IYieldManager";
import { Address, getContract, GetContractReturnType, PublicClient, TransactionReceipt } from "viem";
import { YieldManagerABI } from "../core/abis/YieldManager";

export class YieldManagerContractClient implements IYieldManager<TransactionReceipt> {
  private readonly blockchainClient: PublicClient;
  private readonly contract: GetContractReturnType<typeof YieldManagerABI, PublicClient, Address>;

  constructor(
    contractClientLibrary: IContractClientLibrary<PublicClient, TransactionReceipt>,
    contractAddress: Address,
  ) {
    this.blockchainClient = contractClientLibrary.getBlockchainClient();
    this.contract = getContract({ abi: YieldManagerABI, address: contractAddress, client: this.blockchainClient });
  }

  async getTargetReserveDeficit(): Promise<bigint> {
    return await this.contract.read.getTargetReserveDeficit();
  }

  async isStakingPaused(yieldProvider: Address): Promise<boolean> {
    return await this.contract.read.isStakingPaused([yieldProvider]);
  }

  async isOssificationInitiated(yieldProvider: Address): Promise<boolean> {
    return await this.contract.read.isOssificationInitiated([yieldProvider]);
  }

  async isOssified(yieldProvider: Address): Promise<boolean> {
    return await this.contract.read.isOssified([yieldProvider]);
  }

  async withdrawableValue(yieldProvider: Address): Promise<bigint> {
    const { result } = await this.contract.simulate.withdrawableValue([yieldProvider]);
    return result;
  }
}
