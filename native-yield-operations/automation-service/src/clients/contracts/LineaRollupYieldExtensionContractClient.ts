import { IBlockchainClient, ILogger } from "@consensys/linea-shared-utils";
import {
  Address,
  encodeFunctionData,
  getContract,
  GetContractReturnType,
  PublicClient,
  TransactionReceipt,
} from "viem";
import { LineaRollupYieldExtensionABI } from "../../core/abis/LineaRollupYieldExtension.js";
import { ILineaRollupYieldExtension } from "../../core/clients/contracts/ILineaRollupYieldExtension.js";

export class LineaRollupYieldExtensionContractClient
  implements ILineaRollupYieldExtension<TransactionReceipt>
{
  private readonly contract: GetContractReturnType<typeof LineaRollupYieldExtensionABI, PublicClient, Address>;

  constructor(
    private readonly logger: ILogger,
    private readonly contractClientLibrary: IBlockchainClient<PublicClient, TransactionReceipt>,
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
    this.logger.debug(`transferFundsForNativeYield started, amount=${amount.toString()}`);
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "transferFundsForNativeYield",
      args: [amount],
    });

    const txReceipt = await this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
    this.logger.info(
      `transferFundsForNativeYield succeeded, amount=${amount.toString()}, txHash=${txReceipt.transactionHash}`,
    );
    return txReceipt;
  }
}
