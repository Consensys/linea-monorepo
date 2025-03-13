// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Test } from "forge-std/Test.sol";
import { Base64 } from "@openzeppelin/contracts/utils/Base64.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { MockToken } from "../mocks/MockToken.sol";
import { NFTMetadataGeneratorSVG } from "../../src/nft-metadata-generators/NFTMetadataGeneratorSVG.sol";

contract NFTMetadataGeneratorSVGTest is Test {
    MockToken private erc20Token;
    NFTMetadataGeneratorSVG private metadataGenerator;

    address private alice = makeAddr("alice");

    function setUp() public {
        erc20Token = new MockToken("Test", "TEST");
        metadataGenerator = new NFTMetadataGeneratorSVG("<svg>", "</svg>");

        erc20Token.mint(alice, 10e18);
    }

    function testGenerateMetadata() public view {
        string memory expectedName = "KarmaNFT 0x328809bc894f92807417d2dad6b7c998c1afdac6";
        string memory expectedDescription =
        // solhint-disable-next-line
            "This is a KarmaNFT for address 0x328809bc894f92807417d2dad6b7c998c1afdac6 with balance 10000000000000000000";
        string memory encodedImage = Base64.encode(abi.encodePacked("<svg>10</svg>"));
        string memory expectedImage = string(abi.encodePacked("data:image/svg+xml;base64,", encodedImage));

        bytes memory expectedMetadata = abi.encodePacked(
            "{\"name\":\"",
            expectedName,
            "\",",
            "\"description\":\"",
            expectedDescription,
            "\",",
            "\"image_data\":\"",
            expectedImage,
            "\"}"
        );

        string memory metadata = metadataGenerator.generate(alice, 10e18);
        assertEq(metadata, string(abi.encodePacked("data:application/json;base64,", Base64.encode(expectedMetadata))));
    }

    function testSetImageStrings() public {
        assertEq(metadataGenerator.imagePrefix(), "<svg>");
        assertEq(metadataGenerator.imageSuffix(), "</svg>");

        metadataGenerator.setImageStrings("<new-svg>", "</new-svg>");

        assertEq(metadataGenerator.imagePrefix(), "<new-svg>");
        assertEq(metadataGenerator.imageSuffix(), "</new-svg>");
    }

    function testSetImageStringsRevert() public {
        vm.prank(alice);
        vm.expectRevert("Ownable: caller is not the owner");
        metadataGenerator.setImageStrings("<new-svg>", "</new-svg>");
    }
}
