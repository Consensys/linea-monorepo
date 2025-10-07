// import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
// import { getAccountsFixture } from "../../common/helpers";
// import { deployLidoStVaultYieldProviderFactory } from "../helpers";
// import { LidoStVaultYieldProvider } from "contracts/typechain-types";
// import { expect } from "chai";

// describe("YieldManager contract - control operations", () => {
//   let lidoStVaultYieldProviderFactory: LidoStVaultYieldProvider;

//   let nativeYieldOperator: SignerWithAddress;

//   before(async () => {
//     ({ nativeYieldOperator } = await loadFixture(getAccountsFixture));
//   });

//   beforeEach(async () => {
//     ({ beacon, lidoStVaultYieldProviderFactory } = await loadFixture(deployLidoStVaultYieldProviderFactory));
//   });

//   describe("Deployment", () => {
//     it("Should deploy with correct beacon", async () => {
//       expect(await lidoStVaultYieldProviderFactory.BEACON()).eq(beacon);
//     });
//   });
// });
