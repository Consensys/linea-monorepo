export const L2MessageServiceAbi = [
  // claimMessage(address _from, address _to, uint256 _fee, uint256 _value, address _feeRecipient, bytes _calldata, uint256 _nonce)
  {
    inputs: [
      { name: "_from", type: "address" },
      { name: "_to", type: "address" },
      { name: "_fee", type: "uint256" },
      { name: "_value", type: "uint256" },
      { name: "_feeRecipient", type: "address" },
      { name: "_calldata", type: "bytes" },
      { name: "_nonce", type: "uint256" },
    ],
    name: "claimMessage",
    outputs: [],
    stateMutability: "nonpayable",
    type: "function",
  },
  // inboxL1L2MessageStatus
  {
    inputs: [{ name: "_messageHash", type: "bytes32" }],
    name: "inboxL1L2MessageStatus",
    outputs: [{ name: "", type: "uint256" }],
    stateMutability: "view",
    type: "function",
  },
  // limitInWei
  {
    inputs: [],
    name: "limitInWei",
    outputs: [{ name: "", type: "uint256" }],
    stateMutability: "view",
    type: "function",
  },
  // currentPeriodAmountInWei
  {
    inputs: [],
    name: "currentPeriodAmountInWei",
    outputs: [{ name: "", type: "uint256" }],
    stateMutability: "view",
    type: "function",
  },
  // Events
  {
    anonymous: false,
    inputs: [
      { indexed: true, name: "_from", type: "address" },
      { indexed: true, name: "_to", type: "address" },
      { indexed: false, name: "_fee", type: "uint256" },
      { indexed: false, name: "_value", type: "uint256" },
      { indexed: false, name: "_nonce", type: "uint256" },
      { indexed: false, name: "_calldata", type: "bytes" },
      { indexed: true, name: "_messageHash", type: "bytes32" },
    ],
    name: "MessageSent",
    type: "event",
  },
  {
    anonymous: false,
    inputs: [{ indexed: true, name: "_messageHash", type: "bytes32" }],
    name: "MessageClaimed",
    type: "event",
  },
  {
    anonymous: false,
    inputs: [{ indexed: true, name: "version", type: "uint256" }],
    name: "ServiceVersionMigrated",
    type: "event",
  },
  // RateLimitExceeded error
  {
    inputs: [],
    name: "RateLimitExceeded",
    type: "error",
  },
] as const;
