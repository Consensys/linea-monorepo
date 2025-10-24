import { IBlockchainClientAdapter, ILogger } from "@consensys/linea-shared-utils";
import {
  Address,
  encodeFunctionData,
  getContract,
  GetContractReturnType,
  PublicClient,
  TransactionReceipt,
} from "viem";
import { LineaRollupYieldExtensionABI } from "../core/abis/LineaRollupYieldExtension.js";
import { ILineaRollupYieldExtension } from "../core/services/contracts/ILineaRollupYieldExtension.js";
import { IBaseContractClient } from "../core/clients/IBaseContractClient.js";

export class LineaRollupYieldExtensionContractClient
  implements ILineaRollupYieldExtension<TransactionReceipt>, IBaseContractClient
{
  private readonly contract: GetContractReturnType<typeof LineaRollupYieldExtensionABI, PublicClient, Address>;

  constructor(
    private readonly logger: ILogger,
    private readonly contractClientLibrary: IBlockchainClientAdapter<PublicClient, TransactionReceipt>,
    private readonly contractAddress: Address,
  ) {
    this.contract = getContract({
      abi: LineaRollupYieldExtensionABI,
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

  async transferFundsForNativeYield(amount: bigint): Promise<TransactionReceipt> {
    this.logger.info(`transferFundsForNativeYield: amount=${amount.toString()}`);
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "transferFundsForNativeYield",
      args: [amount],
    });

    return this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
  }
}
