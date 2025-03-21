import { CctpAttestationApiResponse, CctpV2ReattestationApiResponse, CctpFeeApiResponse } from "@/types/cctp";

export async function fetchCctpAttestationByTxHash(
  cctpDomain: number,
  transactionHash: string,
  isTestnet: boolean,
): Promise<CctpAttestationApiResponse> {
  const response = await fetch(
    `https://iris-api${isTestnet ? "-sandbox" : ""}.circle.com/v2/messages/${cctpDomain}?transactionHash=${transactionHash}`,
    {
      headers: {
        "Content-Type": "application/json",
      },
    },
  );
  if (!response.ok) {
    throw new Error(
      `Error in fetchCctpAttestationByTxHash: cctpDomain=${cctpDomain} transactionHash=${transactionHash}`,
    );
  }
  const data: CctpAttestationApiResponse = await response.json();
  return data;
}

export async function fetchCctpAttestationByNonce(
  cctpDomain: number,
  nonce: string,
  isTestnet: boolean,
): Promise<CctpAttestationApiResponse> {
  const response = await fetch(
    `https://iris-api${isTestnet ? "-sandbox" : ""}.circle.com/v2/messages/${cctpDomain}?nonce=${nonce}`,
    {
      headers: {
        "Content-Type": "application/json",
      },
    },
  );
  if (!response.ok) {
    throw new Error(`Error in fetchCctpAttestationByNonce: cctpDomain=${cctpDomain} nonce=${nonce}`);
  }
  const data: CctpAttestationApiResponse = await response.json();
  return data;
}

// https://developers.circle.com/api-reference/stablecoins/common/reattest-message
export async function reattestCctpV2PreFinalityMessage(
  nonce: string,
  isTestnet: boolean,
): Promise<CctpV2ReattestationApiResponse> {
  const response = await fetch(`https://iris-api${isTestnet ? "-sandbox" : ""}.circle.com/v2/reattest/${nonce}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!response.ok) {
    throw new Error(`Error in reattestCctpV2PreFinalityMessage: nonce=${nonce}`);
  }
  const data: CctpV2ReattestationApiResponse = await response.json();
  return data;
}

export async function getCctpFee(
  srcDomain: number,
  dstDomain: number,
  isTestnet: boolean,
): Promise<CctpFeeApiResponse> {
  const response = await fetch(
    `https://iris-api${isTestnet ? "-sandbox" : ""}.circle.com/v2/fastBurn/USDC/fees/${srcDomain}/${dstDomain}`,
    {
      headers: {
        "Content-Type": "application/json",
      },
    },
  );
  if (!response.ok) {
    throw new Error(`Error in getCctpFee: srcDomain=${srcDomain} dstDomain=${dstDomain}`);
  }
  const data: CctpFeeApiResponse = await response.json();
  return data;
}
