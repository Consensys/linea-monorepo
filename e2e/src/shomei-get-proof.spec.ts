import { describe, it } from "@jest/globals";
import { config } from "./config/tests-config";
import { awaitUntil, getDockerImageTag, LineaShomeiClient, LineaShomeiFrontendClient } from "./common/utils";

describe("Shomei Linea get proof test suite", () => {
  const lineaRollupV6 = config.getLineaRollupContract();
  const shomeiFrontendEndpoint = config.getShomeiFrontendEndpoint();
  const shomeiEndpoint = config.getShomeiEndpoint();
  const lineaShomeiFrontenedClient = new LineaShomeiFrontendClient(shomeiFrontendEndpoint!);
  const lineaShomeiClient = new LineaShomeiClient(shomeiEndpoint!);

  it.concurrent(
    "Call linea_getProof to Shomei frontend node and get a valid proof",
    async () => {
      const shomeiImageTag = await getDockerImageTag("shomei-frontend", "consensys/linea-shomei");
      logger.debug(`shomeiImageTag=${shomeiImageTag}`);

      const currentL2BlockNumber = await awaitUntil(
        async () => {
          try {
            return await lineaRollupV6.currentL2BlockNumber({
              blockTag: "finalized",
            });
          } catch (err) {
            if (!(err as Error).message.includes("could not decode result data")) {
              throw err;
            } // else means the currentL2BlockNumber is not ready in the L1 rollup contract yet
            return -1n;
          }
        },
        (currentL2BlockNumber: bigint) => currentL2BlockNumber > 1n,
        2000,
        100000,
      );

      expect(currentL2BlockNumber).toBeGreaterThan(1n);

      logger.debug(`currentL2BlockNumber=${currentL2BlockNumber}`);

      const provingAddress = "0xfe3b557e8fb62b89f4916b721be55ceb828dbd73";
      const getProofResponse = await awaitUntil(
        async () =>
          lineaShomeiFrontenedClient.lineaGetProof(provingAddress, [], "0x" + currentL2BlockNumber!.toString(16)),
        (getProofResponse) => getProofResponse?.result,
        2000,
        100000,
      );

      const getZkEVMStateMerkleProofResponse = await lineaShomeiClient.rollupGetZkEVMStateMerkleProofV0(
        Number(currentL2BlockNumber),
        Number(currentL2BlockNumber),
        shomeiImageTag,
      );

      expect(getZkEVMStateMerkleProofResponse.result.zkEndStateRootHash).toBeDefined();

      const l2SparseMerkleProofContract = config.getL2SparseMerkleProofContract();
      const isValid = await l2SparseMerkleProofContract.verifyProof(
        getProofResponse.result.accountProof.proof.proofRelatedNodes,
        getProofResponse.result.accountProof.leafIndex,
        getZkEVMStateMerkleProofResponse.result.zkEndStateRootHash,
      );

      expect(isValid).toBeTruthy();

      const modifiedStateRootHash =
        getZkEVMStateMerkleProofResponse.result.zkEndStateRootHash.slice(0, -1) +
        ((parseInt(getZkEVMStateMerkleProofResponse.result.zkEndStateRootHash.slice(-1), 16) + 1) % 16).toString(16);

      logger.debug(`originalStateRootHash=${getZkEVMStateMerkleProofResponse.result.zkEndStateRootHash}`);
      logger.debug(`modifiedStateRootHash=${modifiedStateRootHash}`);

      const isInvalid = !(await l2SparseMerkleProofContract.verifyProof(
        getProofResponse.result.accountProof.proof.proofRelatedNodes,
        getProofResponse.result.accountProof.leafIndex,
        modifiedStateRootHash,
      ));

      expect(isInvalid).toBeTruthy();
    },
    100_000,
  );
});
