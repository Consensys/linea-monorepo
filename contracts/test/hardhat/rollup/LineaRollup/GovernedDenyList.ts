import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";

import { getAccountsFixture, deployGovernedDenyListFixture, deployGovernedDenyList } from "../helpers";
import {
  buildAccessErrorMessage,
  expectEvent,
  expectRevertWithCustomError,
  expectRevertWithReason,
} from "../../common/helpers";
import { ADDRESS_ZERO } from "../../common/constants";
import { GovernedDenyList } from "contracts/typechain-types";
import { DEFAULT_ADMIN_ROLE } from "contracts/common/constants";
import { expect } from "chai";

describe("GovernedDenyList: Forced Transactions", () => {
  let governedDenyList: GovernedDenyList;
  let securityCouncil: SignerWithAddress;
  let operator: SignerWithAddress;
  let nonAuthorizedAccount: SignerWithAddress;
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  // let securityCouncil: SignerWithAddress;

  before(async () => {
    ({ nonAuthorizedAccount, securityCouncil, operator } = await loadFixture(getAccountsFixture));
  });

  beforeEach(async () => {
    ({ governedDenyList } = await loadFixture(deployGovernedDenyListFixture));
  });

  describe("Contract Construction", () => {
    it("Should fail if the _defaultAdminis set as address(0)", async () => {
      await expectRevertWithCustomError(
        governedDenyList,
        deployGovernedDenyList(ADDRESS_ZERO, [nonAuthorizedAccount.address]),
        "ZeroAddressNotAllowed",
      );
    });

    it("Should set the _defaultAdmin role", async () => {
      await deployGovernedDenyList(securityCouncil.address, [nonAuthorizedAccount.address]);

      const hasRole = await governedDenyList.hasRole(DEFAULT_ADMIN_ROLE, securityCouncil.address);
      expect(hasRole).to.be.true;
    });

    it("Should set the default denied addresses", async () => {
      ({ governedDenyList } = await deployGovernedDenyList(securityCouncil.address, [
        nonAuthorizedAccount.address,
        operator.address,
      ]));

      let isDenied = await governedDenyList.addressIsDenied(nonAuthorizedAccount.address);
      expect(isDenied).to.be.true;

      isDenied = await governedDenyList.addressIsDenied(operator.address);
      expect(isDenied).to.be.true;

      // double check it isn't just returning true
      isDenied = await governedDenyList.addressIsDenied(securityCouncil.address);
      expect(isDenied).to.be.false;
    });
  });

  describe("Setting denied addresses statuses", () => {
    it("Should set the denied addresses to true", async () => {
      let isDenied = await governedDenyList.addressIsDenied(operator.address);
      expect(isDenied).to.be.false;

      await governedDenyList.connect(securityCouncil).setAddressesDeniedStatus([operator.address], true);

      // double check it isn't just returning true
      isDenied = await governedDenyList.addressIsDenied(operator.address);
      expect(isDenied).to.be.true;
    });

    it("Should set the denied addresses to false", async () => {
      let isDenied = await governedDenyList.addressIsDenied(nonAuthorizedAccount.address);
      expect(isDenied).to.be.true;

      await governedDenyList.connect(securityCouncil).setAddressesDeniedStatus([nonAuthorizedAccount.address], true);

      // double check it isn't just returning true
      isDenied = await governedDenyList.addressIsDenied(operator.address);
      expect(isDenied).to.be.false;
    });

    it("Should emit DeniedStatusesSet when setting status to true", async () => {
      await expectEvent(
        governedDenyList,
        governedDenyList.connect(securityCouncil).setAddressesDeniedStatus([operator.address], true),
        "DeniedStatusesSet",
        [[operator.address], true],
      );
    });

    it("Should emit DeniedStatusesSet when setting status to false", async () => {
      await expectEvent(
        governedDenyList,
        governedDenyList.connect(securityCouncil).setAddressesDeniedStatus([nonAuthorizedAccount.address], false),
        "DeniedStatusesSet",
        [[nonAuthorizedAccount.address], false],
      );
    });
  });

  describe("Access Control", () => {
    it("Should fail to set statuses if not authorized", async () => {
      await expectRevertWithReason(
        governedDenyList.connect(nonAuthorizedAccount).setAddressesDeniedStatus([operator.address], true),
        buildAccessErrorMessage(nonAuthorizedAccount, DEFAULT_ADMIN_ROLE),
      );
    });
  });
});
