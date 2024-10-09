// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Test, console } from "forge-std/Test.sol";
import { Base64 } from "@openzeppelin/contracts/utils/Base64.sol";
import { MockToken } from "./mocks/MockToken.sol";
import { XPNFTToken } from "../src/XPNFTToken.sol";
import { MockMetadataGenerator } from "./mocks/MockMetadataGenerator.sol";

contract XPNFTTokenTest is Test {
    MockToken erc20Token;
    MockMetadataGenerator metadataGenerator;
    XPNFTToken nft;

    address alice = makeAddr("alice");

    function setUp() public {
        erc20Token = new MockToken("Test", "TEST");
        metadataGenerator = new MockMetadataGenerator(address(erc20Token), "https://test.local/");
        nft = new XPNFTToken(address(erc20Token), address(metadataGenerator));

        address[1] memory users = [alice];
        for (uint256 i = 0; i < users.length; i++) {
            erc20Token.mint(users[i], 10e18);
        }
    }

    function test() public view {
        bytes memory expectedMetadata = abi.encodePacked(
            '{"name":"XPNFT Token 0x328809bc894f92807417d2dad6b7c998c1afdac6",',
            '"description":"This is a XPNFT token for address 0x328809bc894f92807417d2dad6b7c998c1afdac6 with balance 10000000000000000000",',
            '"image":"https://test.local/0x328809bc894f92807417d2dad6b7c998c1afdac6"}'
        );
        string memory metadata = nft.tokenURI(uint256(uint160(alice)));
        assertEq(metadata, string(abi.encodePacked("data:application/json;base64,", Base64.encode(expectedMetadata))));
    }
}
