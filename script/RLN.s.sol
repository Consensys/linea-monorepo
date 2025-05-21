// SPDX-License-Identifier: Apache-2.0 OR MIT
pragma solidity ^0.8.26;

import "forge-std/Script.sol";
import "../src/rln/RLN.sol";
import "../src/rln/Verifier.sol";

contract RLNScript is Script {
    function run() public {
        uint256 minimalDeposit = vm.envUint("MINIMAL_DEPOSIT");
        uint256 maximalRate = vm.envUint("MAXIMAL_RATE");
        uint256 depth = vm.envUint("DEPTH");
        uint8 feePercentage = uint8(vm.envUint("FEE_PERCENTAGE"));
        address feeReceiver = vm.envAddress("FEE_RECEIVER");
        uint256 freezePeriod = vm.envUint("FREEZE_PERIOD");
        address token = vm.envAddress("ERC20TOKEN");

        vm.startBroadcast();

        Groth16Verifier verifier = new Groth16Verifier();
        RLN rln = new RLN(
            minimalDeposit, maximalRate, depth, feePercentage, feeReceiver, freezePeriod, token, address(verifier)
        );

        vm.stopBroadcast();
    }
}
