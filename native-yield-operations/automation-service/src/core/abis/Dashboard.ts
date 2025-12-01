export const DashboardABI = [
  {
    anonymous: false,
    inputs: [
      { indexed: true, internalType: "address", name: "sender", type: "address" },
      { indexed: false, internalType: "uint256", name: "fee", type: "uint256" },
    ],
    name: "FeeDisbursed",
    type: "event",
  },
  {
    inputs: [],
    name: "withdrawableValue",
    outputs: [{ internalType: "uint256", name: "", type: "uint256" }],
    stateMutability: "view",
    type: "function",
  },
  { inputs: [], name: "resumeBeaconChainDeposits", outputs: [], stateMutability: "nonpayable", type: "function" },
  {
    inputs: [],
    name: "stakingVault",
    outputs: [{ internalType: "contract IStakingVault", name: "", type: "address" }],
    stateMutability: "view",
    type: "function",
  },
  {
    inputs: [],
    name: "totalValue",
    outputs: [{ internalType: "uint256", name: "", type: "uint256" }],
    stateMutability: "view",
    type: "function",
  },
] as const;
