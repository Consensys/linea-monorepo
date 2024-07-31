const MAX_GAS_LIMIT = process.env.TX_GAS_LIMIT ? parseInt(process.env.TX_GAS_LIMIT) : 500000000;

function getBlockchainNode(): string {
  return process.env.BLOCKCHAIN_NODE || "http://127.0.0.1:8545";
}

function getL2BlockchainNode(): string | undefined {
  return process.env.L2_BLOCKCHAIN_NODE;
}

export { MAX_GAS_LIMIT, getBlockchainNode, getL2BlockchainNode };
