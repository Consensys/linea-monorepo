// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Test } from "forge-std/Test.sol";
import { Base64 } from "@openzeppelin/contracts/utils/Base64.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { MockToken } from "../mocks/MockToken.sol";
import { NFTMetadataGeneratorURL } from "../../src/nft-metadata-generators/NFTMetadataGeneratorURL.sol";

contract NFTMetadataGeneratorURLTest is Test {
    MockToken private erc20Token;
    NFTMetadataGeneratorURL private metadataGenerator;

    address private alice = makeAddr("alice");

    function setUp() public {
        erc20Token = new MockToken("Test", "TEST");
        metadataGenerator = new NFTMetadataGeneratorURL("http://test.local/images/", ".jpg");

        erc20Token.mint(alice, 10e18);
    }

    function testGenerateMetadata() public view {
        string memory expectedMetadata = "data:application/json;base64,";
        bytes memory json = abi.encodePacked(
            "{\"name\":\"KarmaNFT 0x328809bc894f92807417d2dad6b7c998c1afdac6\",",
            // solhint-disable-next-line
            "\"description\":\"This is a KarmaNFT for address 0x328809bc894f92807417d2dad6b7c998c1afdac6 with balance 10000000000000000000\",",
            "\"image\":\"http://test.local/images/0x328809bc894f92807417d2dad6b7c998c1afdac6.jpg\"}"
        );
        expectedMetadata = string(abi.encodePacked(expectedMetadata, Base64.encode(json)));

        string memory metadata = metadataGenerator.generate(alice, erc20Token.balanceOf(alice));
        assertEq(metadata, expectedMetadata);
    }

    function testSetBaseURL() public {
        string memory newURLPrefix = "http://new-test.local/images/";
        string memory newURLSuffix = ".png";

        metadataGenerator.setURLStrings(newURLPrefix, newURLSuffix);

        assertEq(metadataGenerator.urlPrefix(), newURLPrefix);
        assertEq(metadataGenerator.urlSuffix(), newURLSuffix);
    }

    function testSetBaseURLRevert() public {
        vm.prank(alice);
        vm.expectRevert("Ownable: caller is not the owner");
        metadataGenerator.setURLStrings("http://new-test.local/images/", ".png");
    }
}
