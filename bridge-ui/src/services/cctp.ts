import { CctpAttestationApiResponse, CctpV2ReattestationApiResponse, CCTPFeeApiResponse } from "@/types/cctp";

export async function fetchCctpAttestationByTxHash(
  cctpDomain: number,
  transactionHash: string,
): Promise<CctpAttestationApiResponse> {
  const response = await fetch(
    `https://iris-api-sandbox.circle.com/v2/messages/${cctpDomain}?transactionHash=${transactionHash}`,
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
): Promise<CctpAttestationApiResponse> {
  const response = await fetch(`https://iris-api-sandbox.circle.com/v2/messages/${cctpDomain}?nonce=${nonce}`, {
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!response.ok) {
    throw new Error(`Error in fetchCctpAttestationByNonce: cctpDomain=${cctpDomain} nonce=${nonce}`);
  }
  const data: CctpAttestationApiResponse = await response.json();
  return data;
}

// https://developers.circle.com/api-reference/stablecoins/common/reattest-message
export async function reattestCCTPV2PreFinalityMessage(nonce: string): Promise<CctpV2ReattestationApiResponse> {
  const response = await fetch(`https://iris-api-sandbox.circle.com/v2/reattest/${nonce}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!response.ok) {
    throw new Error(`Error in reattestCCTPV2PreFinalityMessage: nonce=${nonce}`);
  }
  const data: CctpV2ReattestationApiResponse = await response.json();
  return data;
}

export async function getCCTPFee(srcDomain: number, dstDomain: number): Promise<CCTPFeeApiResponse> {
  const response = await fetch(`https://iris-api-sandbox.circle.com/v2/fastBurn/USDC/fees/${srcDomain}/${dstDomain}`, {
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!response.ok) {
    throw new Error(`Error in getCCTPFee: srcDomain=${srcDomain} dstDomain=${dstDomain}`);
  }
  const data: CCTPFeeApiResponse = await response.json();
  return data;
}
