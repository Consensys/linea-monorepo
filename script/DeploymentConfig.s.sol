//// SPDX-License-Identifier: UNLICENSED

pragma solidity >=0.8.26 <=0.9.0;

import { Script } from "forge-std/Script.sol";
import { MockToken } from "../test/mocks/MockToken.sol";

contract DeploymentConfig is Script {
    error DeploymentConfig_InvalidDeployerAddress();
    error DeploymentConfig_NoConfigForChain(uint256);

    struct NetworkConfig {
        address deployer;
        address stakingToken;
    }

    NetworkConfig public activeNetworkConfig;

    address private deployer;

    // solhint-disable-next-line var-name-mixedcase
    address internal SNT_ADDRESS_SEPOLIA = 0xE452027cdEF746c7Cd3DB31CB700428b16cD8E51;

    // solhint-disable-next-line var-name-mixedcase
    address internal SNT_ADDRESS_STATUS_NETWORK_SEPOLIA = 0x1C3Ac2a186c6149Ae7Cb4D716eBbD0766E4f898a;

    // solhint-disable-next-line var-name-mixedcase
    address internal STAKING_MANAGER_PROXY_ADDRESS_SEPOLIA = 0xD302Bd9F60c5192e46258028a2F3b4B2B846F61F;

    constructor(address _broadcaster) {
        if (_broadcaster == address(0)) revert DeploymentConfig_InvalidDeployerAddress();
        deployer = _broadcaster;
        if (block.chainid == 31_337) {
            activeNetworkConfig = getOrCreateAnvilEthConfig();
        } else if (block.chainid == 11_155_111) {
            activeNetworkConfig = getSepoliaConfig();
        } else if (block.chainid == 1_660_990_954) {
            activeNetworkConfig = getStatusNetworkSepoliaConfig();
        } else {
            revert DeploymentConfig_NoConfigForChain(block.chainid);
        }
    }

    function getOrCreateAnvilEthConfig() public returns (NetworkConfig memory) {
        MockToken stakingToken = new MockToken("Staking Token", "ST");
        return NetworkConfig({ deployer: deployer, stakingToken: address(stakingToken) });
    }

    function getSepoliaConfig() public view returns (NetworkConfig memory) {
        return NetworkConfig({ deployer: deployer, stakingToken: SNT_ADDRESS_SEPOLIA });
    }

    function getStatusNetworkSepoliaConfig() public view returns (NetworkConfig memory) {
        return NetworkConfig({ deployer: deployer, stakingToken: SNT_ADDRESS_STATUS_NETWORK_SEPOLIA });
    }

    // This function is a hack to have it excluded by `forge coverage` until
    // https://github.com/foundry-rs/foundry/issues/2988 is fixed.
    // See: https://github.com/foundry-rs/foundry/issues/2988#issuecomment-1437784542
    // for more info.
    // solhint-disable-next-line
    function test() public { }
}
