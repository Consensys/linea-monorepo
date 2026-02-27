export const LINEA_ROLLUP_YIELD_EXTENSION_ABI = [
  {
    inputs: [
      {
        components: [
          { internalType: "bytes32[]", name: "proof", type: "bytes32[]" },
          { internalType: "uint256", name: "messageNumber", type: "uint256" },
          { internalType: "uint32", name: "leafIndex", type: "uint32" },
          { internalType: "address", name: "from", type: "address" },
          { internalType: "address", name: "to", type: "address" },
          { internalType: "uint256", name: "fee", type: "uint256" },
          { internalType: "uint256", name: "value", type: "uint256" },
          { internalType: "address payable", name: "feeRecipient", type: "address" },
          { internalType: "bytes32", name: "merkleRoot", type: "bytes32" },
          { internalType: "bytes", name: "data", type: "bytes" },
        ],
        internalType: "struct IL1MessageService.ClaimMessageWithProofParams",
        name: "_params",
        type: "tuple",
      },
      { internalType: "address", name: "_yieldProvider", type: "address" },
    ],
    name: "claimMessageWithProofAndWithdrawLST",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
] as const;
