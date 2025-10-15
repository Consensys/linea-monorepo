pragma solidity ^0.8.26;

import { Test } from "forge-std/Test.sol";
import { StakeManagerTest } from "../../StakeManagerBase.t.sol";

contract StakeMathTest is StakeManagerTest {
    function test_CalcInitialMP() public pure {
        assertEq(_initialMP(1), 1, "wrong initial MP");
        assertEq(_initialMP(10e18), 10e18, "wrong initial MP");
        assertEq(_initialMP(20e18), 20e18, "wrong initial MP");
        assertEq(_initialMP(30e18), 30e18, "wrong initial MP");
    }

    function test_CalcAccrueMP() public pure {
        assertEq(_accrueMP(10e18, 0), 0, "wrong accrued MP");
        assertEq(_accrueMP(10e18, 365 days / 2), 5e18, "wrong accrued MP");
        assertEq(_accrueMP(10e18, 365 days), 10e18, "wrong accrued MP");
        assertEq(_accrueMP(10e18, 365 days * 2), 20e18, "wrong accrued MP");
        assertEq(_accrueMP(10e18, 365 days * 3), 30e18, "wrong accrued MP");
    }

    function test_CalcBonusMP() public view {
        assertEq(_bonusMP(10e18, 0), 0, "wrong bonus MP");
        assertEq(_bonusMP(10e18, streamer.MIN_LOCKUP_PERIOD()), 2_465_753_424_657_534_246, "wrong bonus MP");
        assertEq(_bonusMP(10e18, streamer.MIN_LOCKUP_PERIOD() + 13 days), 2_821_917_808_219_178_082, "wrong bonus MP");
        assertEq(_bonusMP(100e18, 0), 0, "wrong bonus MP");
    }

    function test_CalcMaxTotalMP() public view {
        assertEq(_maxTotalMP(10e18, 0), 50e18, "wrong max total MP");
        assertEq(_maxTotalMP(10e18, streamer.MIN_LOCKUP_PERIOD()), 52_465_753_424_657_534_246, "wrong max total MP");
        assertEq(
            _maxTotalMP(10e18, streamer.MIN_LOCKUP_PERIOD() + 13 days),
            52_821_917_808_219_178_082,
            "wrong max total MP"
        );
        assertEq(_maxTotalMP(100e18, 0), 500e18, "wrong max total MP");
    }

    function test_CalcAbsoluteMaxTotalMP() public pure {
        assertEq(_maxAbsoluteTotalMP(10e18), 90e18, "wrong absolute max total MP");
        assertEq(_maxAbsoluteTotalMP(100e18), 900e18, "wrong absolute max total MP");
    }

    function test_CalcMaxAccruedMP() public pure {
        assertEq(_maxAccrueMP(10e18), 40e18, "wrong max accrued MP");
        assertEq(_maxAccrueMP(100e18), 400e18, "wrong max accrued MP");
    }
}