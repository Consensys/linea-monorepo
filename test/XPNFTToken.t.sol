// SPDX-License-Identifier: MIT
pragma solidity ^0.8.26;

import { Test, console } from "forge-std/Test.sol";
import { Base64 } from "@openzeppelin/contracts/utils/Base64.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";
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

    function addressToId(address addr) internal pure returns (uint256) {
        return uint256(uint160(addr));
    }

    function testTokenURI() public view {
        bytes memory expectedMetadata = abi.encodePacked(
            "{\"name\":\"XPNFT Token 0x328809bc894f92807417d2dad6b7c998c1afdac6\",",
            // solhint-disable-next-line
            "\"description\":\"This is a XPNFT token for address 0x328809bc894f92807417d2dad6b7c998c1afdac6 with balance 10000000000000000000\",",
            "\"image\":\"https://test.local/0x328809bc894f92807417d2dad6b7c998c1afdac6\"}"
        );
        string memory metadata = nft.tokenURI(addressToId(alice));
        assertEq(metadata, string(abi.encodePacked("data:application/json;base64,", Base64.encode(expectedMetadata))));
    }

    function testSetMetadataGenerator() public {
        MockMetadataGenerator newMetadataGenerator =
            new MockMetadataGenerator(address(erc20Token), "https://new-test.local/");

        nft.setMetadataGenerator(address(newMetadataGenerator));

        assertEq(address(nft.metadataGenerator()), address(newMetadataGenerator));
    }

    function testSetMetadataGeneratorRevert() public {
        MockMetadataGenerator newMetadataGenerator =
            new MockMetadataGenerator(address(erc20Token), "https://new-test.local/");

        vm.prank(alice);
        vm.expectPartialRevert(Ownable.OwnableUnauthorizedAccount.selector);
        nft.setMetadataGenerator(address(newMetadataGenerator));
    }

    function testTransferNotAllowed() public {
        vm.expectRevert(XPNFTToken.XPNFT__TransferNotAllowed.selector);
        nft.transferFrom(alice, address(0), addressToId(alice));
    }

    function testSafeTransferNotAllowed() public {
        vm.expectRevert(XPNFTToken.XPNFT__TransferNotAllowed.selector);
        nft.safeTransferFrom(alice, address(0), addressToId(alice));
    }

    function testSafeTransferWithDataNotAllowed() public {
        vm.expectRevert(XPNFTToken.XPNFT__TransferNotAllowed.selector);
        nft.safeTransferFrom(alice, address(0), addressToId(alice), "");
    }

    function testApproveNotAllowed() public {
        vm.expectRevert(XPNFTToken.XPNFT__TransferNotAllowed.selector);
        nft.approve(address(0), addressToId(alice));
    }

    function testSetApprovalForAllNotAllowed() public {
        vm.expectRevert(XPNFTToken.XPNFT__TransferNotAllowed.selector);
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
