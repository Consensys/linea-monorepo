export type Config = {
  senderAddress: string;
  destinationAddress: string;
  threshold: string;
  blockchainRpcUrl: string;
  web3SignerUrl: string;
  web3SignerPublicKey: string;
  maxFeePerGas: bigint;
  gasEstimationPercentile: number;
  dryRun: boolean;
};
