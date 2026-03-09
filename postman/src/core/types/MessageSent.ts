export type MessageSent = {
  messageHash: string;
  messageSender: string;
  destination: string;
  fee: bigint;
  value: bigint;
  messageNonce: bigint;
  calldata: string;
  contractAddress: string;
  blockNumber: number;
  transactionHash: string;
  logIndex: number;
};
