export const SUBMIT_INVOICE_ABI = [
  {
    inputs: [
      {
        internalType: "uint256",
        name: "_startTimestamp",
        type: "uint256",
      },
      {
        internalType: "uint256",
        name: "_endTimestamp",
        type: "uint256",
      },
      {
        internalType: "uint256",
        name: "_invoiceAmount",
        type: "uint256",
      },
    ],
    name: "submitInvoice",
    outputs: [],
    stateMutability: "payable",
    type: "function",
  },
] as const;

export const INVOICE_PROCESSED_EVENT_ABI = [
  {
    anonymous: false,
    inputs: [
      {
        indexed: true,
        internalType: "address",
        name: "receiver",
        type: "address",
      },
      {
        indexed: true,
        internalType: "uint256",
        name: "startTimestamp",
        type: "uint256",
      },
      {
        indexed: true,
        internalType: "uint256",
        name: "endTimestamp",
        type: "uint256",
      },
      {
        indexed: false,
        internalType: "uint256",
        name: "amountPaid",
        type: "uint256",
      },
      {
        indexed: false,
        internalType: "uint256",
        name: "amountRequested",
        type: "uint256",
      },
    ],
    name: "InvoiceProcessed",
    type: "event",
  },
] as const;

export const ROLLUP_REVENUE_VAULT_ERRORS_ABI = [
  {
    inputs: [],
    name: "DexSwapFailed",
    type: "error",
  },
  {
    inputs: [],
    name: "EndTimestampMustBeGreaterThanStartTimestamp",
    type: "error",
  },
  {
    inputs: [],
    name: "EthBurnFailed",
    type: "error",
  },
  {
    inputs: [],
    name: "ExistingAddressTheSame",
    type: "error",
  },
  {
    inputs: [],
    name: "InsufficientBalance",
    type: "error",
  },
  {
    inputs: [],
    name: "InvoiceDateTooOld",
    type: "error",
  },
  {
    inputs: [],
    name: "InvoiceInArrears",
    type: "error",
  },
  {
    inputs: [],
    name: "InvoiceTransferFailed",
    type: "error",
  },
  {
    inputs: [],
    name: "TimestampsNotInSequence",
    type: "error",
  },
  {
    inputs: [],
    name: "ZeroAddressNotAllowed",
    type: "error",
  },
  {
    inputs: [],
    name: "ZeroInvoiceAmount",
    type: "error",
  },
  {
    inputs: [],
    name: "ZeroLineaTokensReceived",
    type: "error",
  },
  {
    inputs: [],
    name: "ZeroTimestampNotAllowed",
    type: "error",
  },
] as const;
