import { Address, BaseError, Client, encodeFunctionData, Hex } from "viem";
import { readContract } from "viem/actions";
import { SUBMIT_INVOICE_ABI } from "./constants.js";

export async function getLastInvoiceDate(client: Client, contractAddress: Address): Promise<bigint | null> {
  try {
    const lastInvoiceDate = await readContract(client, {
      address: contractAddress,
      abi: [
        {
          inputs: [],
          name: "lastInvoiceDate",
          outputs: [
            {
              internalType: "uint256",
              name: "",
              type: "uint256",
            },
          ],
          stateMutability: "view",
          type: "function",
        },
      ],
      functionName: "lastInvoiceDate",
    });
    return lastInvoiceDate;
  } catch (error) {
    if (error instanceof BaseError) {
      const err = error.walk();
      console.log("Get last invoice date failed with the following error:", err.message);
    }
    return null;
  }
}

export function computeSubmitInvoiceCalldata(startDate: bigint, endDate: bigint, invoiceAmount: bigint): Hex {
  return encodeFunctionData({
    abi: SUBMIT_INVOICE_ABI,
    functionName: "submitInvoice",
    args: [startDate, endDate, invoiceAmount],
  });
}
