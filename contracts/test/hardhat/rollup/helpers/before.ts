import { ethers } from "hardhat";
import { OPERATOR_ROLE } from "../../common/constants";
import { LINEA_ROLLUP_ROLES } from "contracts/common/constants";
import { generateRoleAssignments } from "contracts/common/helpers";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";

export async function getAccountsFixture() {
  const [admin, securityCouncil, operator, nonAuthorizedAccount] = await ethers.getSigners();
  return { admin, securityCouncil, operator, nonAuthorizedAccount };
}

export async function getRoleAddressesFixture() {
  const { securityCouncil, operator } = await loadFixture(getAccountsFixture);
  const roleAddresses = generateRoleAssignments(LINEA_ROLLUP_ROLES, await securityCouncil.getAddress(), [
    {
      role: OPERATOR_ROLE,
      addresses: [operator.address],
    },
  ]);
  return roleAddresses;
}
