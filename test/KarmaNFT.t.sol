// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Test } from "forge-std/Test.sol";
import { Base64 } from "@openzeppelin/contracts/utils/Base64.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
import { MockToken } from "./mocks/MockToken.sol";
import { KarmaNFT } from "../src/KarmaNFT.sol";
import { DeployKarmaNFTScript } from "../script/DeployKarmaNFT.s.sol";
import { INFTMetadataGenerator } from "../src/interfaces/INFTMetadataGenerator.sol";
import { MockMetadataGenerator } from "./mocks/MockMetadataGenerator.sol";

contract KarmaNFTTest is Test {
    MockToken public erc20Token;
    INFTMetadataGenerator public metadataGenerator;
    KarmaNFT public nft;

    address public alice = makeAddr("alice");

    function setUp() public {
        erc20Token = new MockToken("Test", "TEST");
        ((nft, metadataGenerator,)) = new DeployKarmaNFTScript().runForTest(address(erc20Token));
        nft = new KarmaNFT(address(erc20Token), address(metadataGenerator));

        address[1] memory users = [alice];
        for (uint256 i = 0; i < users.length; i++) {
            erc20Token.mint(users[i], 10e18);
        }
    }

    function addressToId(address addr) internal pure returns (uint256) {
        return uint256(uint160(addr));
    }

    function testTokenURI() public {
        INFTMetadataGenerator generator = new MockMetadataGenerator("https://test.local/");
        nft.setMetadataGenerator(address(generator));

        bytes memory expectedMetadata = abi.encodePacked(
            "{\"name\":\"KarmaNFT 0x328809bc894f92807417d2dad6b7c998c1afdac6\",",
            // solhint-disable-next-line
            "\"description\":\"This is a KarmaNFT for address 0x328809bc894f92807417d2dad6b7c998c1afdac6 with balance 10000000000000000000\",",
            "\"image\":\"https://test.local/0x328809bc894f92807417d2dad6b7c998c1afdac6\"}"
        );
        string memory metadata = nft.tokenURI(addressToId(alice));
        assertEq(metadata, string(abi.encodePacked("data:application/json;base64,", Base64.encode(expectedMetadata))));
    }

    function testSetMetadataGenerator() public {
        MockMetadataGenerator newMetadataGenerator = new MockMetadataGenerator("https://new-test.local/");

        nft.setMetadataGenerator(address(newMetadataGenerator));

        assertEq(address(nft.metadataGenerator()), address(newMetadataGenerator));
    }

    function testSetMetadataGeneratorRevert() public {
        MockMetadataGenerator newMetadataGenerator = new MockMetadataGenerator("https://new-test.local/");

        vm.prank(alice);
        vm.expectRevert("Ownable: caller is not the owner");
        nft.setMetadataGenerator(address(newMetadataGenerator));
    }

    function testTransferNotAllowed() public {
        vm.expectRevert(KarmaNFT.KarmaNFT__TransferNotAllowed.selector);
        nft.transferFrom(alice, address(0), addressToId(alice));
    }

    function testSafeTransferNotAllowed() public {
        vm.expectRevert(KarmaNFT.KarmaNFT__TransferNotAllowed.selector);
        nft.safeTransferFrom(alice, address(0), addressToId(alice));
    }

    function testSafeTransferWithDataNotAllowed() public {
        vm.expectRevert(KarmaNFT.KarmaNFT__TransferNotAllowed.selector);
        nft.safeTransferFrom(alice, address(0), addressToId(alice), "");
    }

    function testApproveNotAllowed() public {
        vm.expectRevert(KarmaNFT.KarmaNFT__TransferNotAllowed.selector);
        nft.approve(address(0), addressToId(alice));
    }

    function testSetApprovalForAllNotAllowed() public {
        vm.expectRevert(KarmaNFT.KarmaNFT__TransferNotAllowed.selector);
        nft.setApprovalForAll(address(0), true);
    }

    function testGetApproved() public view {
        address approved = nft.getApproved(addressToId(alice));
        assertEq(approved, address(0));
    }

    function testIsApprovedForAll() public view {
        bool isApproved = nft.isApprovedForAll(alice, address(0));
        assertFalse(isApproved);
    }
}
