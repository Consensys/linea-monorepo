using ERC20A as staked;

methods {
    function totalStaked() external returns (uint256) envfree;
}

ghost mathint sumOfBalances {
	init_state axiom sumOfBalances == 0;
}

hook Sstore users[KEY address account].stakedBalance uint256 newValue (uint256 oldValue) {
    sumOfBalances = sumOfBalances - oldValue + newValue;
}

invariant sumOfBalancesIsTotalStaked()
  sumOfBalances == to_mathint(totalStaked());
