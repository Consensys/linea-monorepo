// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Test, console } from "forge-std/Test.sol";
import { MockToken } from "./mocks/MockToken.sol";
import { XPNFTToken } from "../src/XPNFTToken.sol";

contract XPNFTTokenTest is Test {
    MockToken erc20Token;
    XPNFTToken nft;

    address alice = makeAddr("alice");

    string imagePrefix =
    // solhint-disable-next-line
        '<svg xmlns="http://www.w3.org/2000/svg" height="200" width="200"><rect width="100%" height="100%" fill="blue"/><text x="50%" y="50%" fill="white" font-size="20" text-anchor="middle">';
    string imageSuffix = "</text></svg>";

    function setUp() public {
        erc20Token = new MockToken("Test", "TEST");
        nft = new XPNFTToken(address(erc20Token), imagePrefix, imageSuffix);

        address[1] memory users = [alice];
        for (uint256 i = 0; i < users.length; i++) {
            erc20Token.mint(users[i], 10e18);
        }
    }

    function test() public {
        string memory metadata = nft.tokenURI(uint256(uint160(alice)));
        console.log(metadata);
    }
}
