import { describe, it } from "@jest/globals";
import { ContractFunctionExecutionError, toHex } from "viem";

import { awaitUntil, getDockerImageTag, serialize } from "./common/utils";
import { createTestContext } from "./config/tests-config/setup";
import { LineaGetProofReturnType } from "./config/tests-config/setup/clients/extensions/linea-rpc/linea-get-proof";
import { L2RpcEndpoint } from "./config/tests-config/setup/clients/l2-client";

describe("Shomei Linea get proof test suite", () => {
  const context = createTestContext();
  const lineaRollupV6 = context.l1Contracts.lineaRollup(context.l1PublicClient());
  const lineaShomeiFrontendClient = context.l2PublicClient({ type: L2RpcEndpoint.ShomeiFrontend });
  const lineaShomeiClient = context.l2PublicClient({ type: L2RpcEndpoint.Shomei });

  it.concurrent(
    "Call linea_getProof to Shomei frontend node and get a valid proof",
    async () => {
      const shomeiImageTag = await getDockerImageTag("shomei-frontend", "consensys/linea-shomei");
      logger.debug(`shomeiImageTag=${shomeiImageTag}`);

      let targetL2BlockNumber = await awaitUntil(
        async () => {
          try {
            return await lineaRollupV6.read.currentL2BlockNumber({ blockTag: "finalized" });
          } catch (err) {
            if (err instanceof ContractFunctionExecutionError) {
              if (err.shortMessage.includes(`returned no data ("0x")`)) {
                // means the currentL2BlockNumber is not ready in the L1 rollup contract yet
                return -1n;
              }
            }
            throw err;
          }
        },
        (currentL2BlockNumber: bigint) => currentL2BlockNumber > 1n,
        2000,
        150000,
      );

      expect(targetL2BlockNumber).toBeGreaterThan(1n);

      const finalizedL2BlockNumbers = [targetL2BlockNumber!];
      const provingAddress = "0xfe3b557e8fb62b89f4916b721be55ceb828dbd73"; // from genesis file
      const getProofResponse = await awaitUntil(
        async () => {
          let getProofResponse: LineaGetProofReturnType | null = null;
          // Need to put all the latest currentL2BlockNumber in a list and traverse to get the proof
          // from one of them as we don't know on which finalized L2 block number the shomei frontend
          // was being notified
          for (const finalizedL2BlockNumber of finalizedL2BlockNumbers) {
            getProofResponse = await lineaShomeiFrontendClient.lineaGetProof({
              address: provingAddress,
              storageKeys: [],
              blockParameter: toHex(finalizedL2BlockNumber),
            });
            if (getProofResponse) {
              targetL2BlockNumber = finalizedL2BlockNumber;
              break;
            }
          }
          if (!getProofResponse) {
            const latestFinalizedL2BlockNumber = await lineaRollupV6.read.currentL2BlockNumber({
              blockTag: "finalized",
            });
            if (!finalizedL2BlockNumbers.includes(latestFinalizedL2BlockNumber)) {
              finalizedL2BlockNumbers.push(latestFinalizedL2BlockNumber);
              logger.debug(`finalizedL2BlockNumbers=${serialize(finalizedL2BlockNumbers.map((it) => Number(it)))}`);
            }
          }
          return getProofResponse;
        },
        (getProofResponse) => !!getProofResponse,
        2000,
        150000,
      );

      logger.debug(`targetL2BlockNumber=${targetL2BlockNumber}`);

      const { zkEndStateRootHash } = await lineaShomeiClient.rollupGetZkEVMStateMerkleProofV0({
        startBlockNumber: Number(targetL2BlockNumber),
        endBlockNumber: Number(targetL2BlockNumber),
        zkStateManagerVersion: shomeiImageTag,
      });

      logger.debug(`zkEndStateRootHash=${zkEndStateRootHash}`);
      expect(zkEndStateRootHash).toBeDefined();

      const l2SparseMerkleProofContract = context.l2Contracts.sparseMerkleProof(context.l2PublicClient());
      const isValid = await l2SparseMerkleProofContract.read.verifyProof([
        getProofResponse!.accountProof.proof.proofRelatedNodes,
        BigInt(getProofResponse!.accountProof.leafIndex),
        zkEndStateRootHash,
      ]);

      expect(isValid).toBeTruthy();

      // Modify the last hex character of the original state root hash should verify the same proof as invalid
      const modifiedStateRootHash =
        zkEndStateRootHash.slice(0, -1) + ((parseInt(zkEndStateRootHash.slice(-1), 16) + 1) % 16).toString(16);

      logger.debug(`originalStateRootHash=${zkEndStateRootHash}`);
      logger.debug(`modifiedStateRootHash=${modifiedStateRootHash}`);

      const isInvalid = !(await l2SparseMerkleProofContract.read.verifyProof([
        getProofResponse!.accountProof.proof.proofRelatedNodes,
        BigInt(getProofResponse!.accountProof.leafIndex),
        modifiedStateRootHash as `0x${string}`,
      ]));

      expect(isInvalid).toBeTruthy();
    },
    150_000,
  );
});
