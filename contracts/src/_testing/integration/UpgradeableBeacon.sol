// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import { UpgradeableBeacon as OZUpgradeableBeacon } from "@openzeppelin/contracts/proxy/beacon/UpgradeableBeacon.sol";
import { BeaconProxy as OZBeaconProxy } from "@openzeppelin/contracts/proxy/beacon/BeaconProxy.sol";

contract UpgradeableBeacon is OZUpgradeableBeacon {
  constructor(address implementation_) OZUpgradeableBeacon(implementation_) {}
}

contract BeaconProxy is OZBeaconProxy {
  constructor(address beacon_, bytes memory data_) OZBeaconProxy(beacon_, data_) {}
}
