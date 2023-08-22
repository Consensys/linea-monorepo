import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { BigNumber } from "ethers";
import { TestPlonkVerifierFullLarge } from "../../typechain-types";
import { deployFromFactory } from "../utils/deployment";
import { getProverTestData } from "./../utils/helpers";

describe("test plonk", () => {
  let plonkVerifier: TestPlonkVerifierFullLarge;

  const PROOF_MODE = "FullLarge";
  const { proof } = getProverTestData(PROOF_MODE, "output-file.json");

  async function deployPlonkVerifierFixture() {
    return deployFromFactory("TestPlonkVerifierFullLarge") as Promise<TestPlonkVerifierFullLarge>;
  }

  beforeEach(async () => {
    plonkVerifier = await loadFixture(deployPlonkVerifierFixture);
  });

  describe("testVerifier_go", () => {
    it("Should verify proof successfully", async () => {
      expect(await plonkVerifier.testVerifier(
        proof,
        [
        BigNumber.from("21826973039313591084461008381720124636263968477099612249120776239336034572329"),
      ]),
    ).to.not.be.reverted;
    });
  });
});
