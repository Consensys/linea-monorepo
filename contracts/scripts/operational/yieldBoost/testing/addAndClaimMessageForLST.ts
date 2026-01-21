import { task } from "hardhat/config";
import { getTaskCliOrEnvValue } from "../../../../common/helpers/environmentHelper";
import { delay } from "../../../../common/helpers/general";
import { encodeSendMessage, randomBytes32 } from "../../../../common/helpers/encoding";

/*
  *******************************************************************************************
  Setup and execute TestLineaRollup.claimMessageWithProofAndWithdrawLST.

  1) Signer must have DEFAULT_ADMIN_ROLE role for TestLineaRollup
  2) L1MessageService balance must be < `value` (LST withdrawal requires deficit)
  3) Caller must be the `to` address (LST withdrawal recipient)

  -------------------------------------------------------------------------------------------
  Example (Hoodi):
  -------------------------------------------------------------------------------------------
  CUSTOM_PRIVATE_KEY=<key> \
  CUSTOM_BLOCKCHAIN_URL=https://0xrpc.io/hoodi \
  npx hardhat addAndClaimMessageForLST \
    --linea-rollup-address <address> \
    --to <address> \
    --value <uint256> \
    --data <hex_string> \
    --yield-provider <address> \
    --network custom
  *******************************************************************************************
*/

// TASKS
task(
  "addAndClaimMessageForLST",
  "Setup and execute TestLineaRollup.claimMessageWithProofAndWithdrawLST by adding L2->L1 message merkle tree root",
)
  .addOptionalParam("lineaRollupAddress")
  .addOptionalParam("from")
  .addOptionalParam("to")
  .addOptionalParam("value")
  .addOptionalParam("data")
  .addOptionalParam("yieldProvider")
  .setAction(async (taskArgs, hre) => {
    const { ethers, getNamedAccounts } = hre;
    const { deployer } = await getNamedAccounts();
    const signer = await ethers.getSigner(deployer);

    // --- Resolve inputs from CLI or ENV (with sensible fallbacks to deployments) ---
    const lineaRollupAddress = getTaskCliOrEnvValue(taskArgs, "lineaRollupAddress", "LINEA_ROLLUP_ADDRESS");
    const fromAddress = getTaskCliOrEnvValue(taskArgs, "from", "FROM_ADDRESS") || signer.address;
    const toAddress = getTaskCliOrEnvValue(taskArgs, "to", "TO_ADDRESS");
    const valueRaw = getTaskCliOrEnvValue(taskArgs, "value", "VALUE");
    const data = getTaskCliOrEnvValue(taskArgs, "data", "DATA") || "0x";
    const yieldProvider = getTaskCliOrEnvValue(taskArgs, "yieldProvider", "YIELD_PROVIDER");

    // Validate required params
    const missing: string[] = [];
    if (!lineaRollupAddress) missing.push("lineaRollupAddress / LINEA_ROLLUP_ADDRESS");
    if (!toAddress) missing.push("to / TO_ADDRESS");
    if (!valueRaw) missing.push("value / VALUE");
    if (!data) missing.push("data / DATA");
    if (!yieldProvider) missing.push("yieldProvider / YIELD_PROVIDER");
    if (missing.length) {
      throw new Error(`Missing required params/envs: ${missing.join(", ")}`);
    }

    // --- Parse numeric params ---
    const value = valueRaw ? BigInt(valueRaw) : 0n;

    // Fixed values
    const fee = 0n;
    const leafIndex = 0n;
    // Generate random proof array first (depth determined by proof length)
    const proof = Array.from({ length: 32 }, () => randomBytes32());
    const proofDepth = proof.length;

    const lineaRollup = await ethers.getContractAt("TestLineaRollup", lineaRollupAddress!, signer);
    // Generate random messageNumber and check if it's already claimed
    let messageNumber: bigint;
    let attempts = 0;
    const maxAttempts = 100; // Prevent infinite loop
    do {
      messageNumber = ethers.toBigInt(ethers.randomBytes(32));
      attempts++;
      if (attempts > maxAttempts) {
        throw new Error(
          `Failed to find unclaimed message number after ${maxAttempts} attempts. This is highly unlikely.`,
        );
      }
      const isClaimed = await lineaRollup.isMessageClaimed(messageNumber);
      if (!isClaimed) {
        break;
      }
      console.log(`  Message number ${messageNumber.toString()} already claimed, generating new one...`);
      // eslint-disable-next-line no-constant-condition
    } while (true);

    // Log params
    console.log("Parameters:");
    console.log("  lineaRollupAddress:", lineaRollupAddress);
    console.log("  messageNumber:", messageNumber.toString(), "(auto-generated, verified unclaimed)");
    console.log("  from:", fromAddress);
    console.log("  to:", toAddress);
    console.log("  value:", value.toString());
    console.log("  fee:", fee.toString(), "(default: 0)");
    console.log("  data:", data);
    console.log("  yieldProvider:", yieldProvider);
    console.log("  leafIndex:", leafIndex.toString(), "(default: 0)");
    console.log("  proofDepth:", proofDepth, "(from generated proof)");

    // Encode message bytes
    const expectedBytes = encodeSendMessage(fromAddress, toAddress!, fee, value, messageNumber, data);
    const messageHash = ethers.keccak256(expectedBytes);
    console.log("  messageHash:", messageHash);

    // Get generatedRoot
    const generatedRoot = await lineaRollup.generateMerkleRoot.staticCall(messageHash, proof, leafIndex);
    console.log("generatedRoot=", generatedRoot);

    // Add merkle root to LineaRollup
    console.log("\nAdding merkle root to LineaRollup...");
    console.log("  NOTE: Signer must have DEFAULT_ADMIN_ROLE");
    {
      const tx = await lineaRollup.addL2MerkleRoots([generatedRoot], proofDepth);
      console.log("  Transaction hash:", tx.hash);
      const receipt = await tx.wait();
      console.log("  Transaction confirmed in block:", receipt?.blockNumber);
    }

    // Output claim parameters
    console.log("\n=== Claim Parameters ===");
    console.log("Claim parameters:");
    console.log(
      JSON.stringify(
        {
          proof: proof,
          messageNumber: messageNumber.toString(),
          leafIndex: leafIndex.toString(),
          from: fromAddress,
          to: toAddress,
          fee: fee.toString(),
          value: value.toString(),
          feeRecipient: ethers.ZeroAddress,
          merkleRoot: generatedRoot,
          data: data,
          yieldProvider: yieldProvider,
        },
        null,
        2,
      ),
    );
    {
      console.log("Waiting for 10 seconds...");
      await delay(10000);
      console.log("Claiming message with LST withdrawal...");
      const tx = await lineaRollup.claimMessageWithProofAndWithdrawLST(
        {
          proof,
          messageNumber,
          leafIndex: leafIndex,
          from: fromAddress,
          to: toAddress!,
          fee: fee.toString(),
          value,
          feeRecipient: ethers.ZeroAddress,
          merkleRoot: generatedRoot,
          data,
        },
        yieldProvider!,
      );
      console.log("  Transaction hash:", tx.hash);
      const receipt = await tx.wait();
      console.log("  Transaction confirmed in block:", receipt?.blockNumber);
    }
  });
