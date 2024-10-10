using RewardsStreamerMP as streamer;
using ERC20A as staked;

methods {
    function totalStaked() external returns (uint256) envfree;
    function users(address) external returns (uint256, uint256, uint256, uint256, uint256, uint256) envfree;
}

ghost mathint sumOfBalances {
	init_state axiom sumOfBalances == 0;
}

hook Sstore users[KEY address account].stakedBalance uint256 newValue (uint256 oldValue) {
    sumOfBalances = sumOfBalances - oldValue + newValue;
}

function getAccountMaxMP(address account) returns uint256 {
    uint256 maxMP;
    _, _, _, maxMP, _, _ = streamer.users(account);
    return maxMP;
}

function getAccountMP(address account) returns uint256 {
    uint256 accountMP;
    _, _, accountMP, _, _, _ = streamer.users(account);
    return accountMP;
}

function getAccountStakedBalance(address account) returns uint256 {
    uint256 stakedBalance;
    stakedBalance, _, _, _, _, _ = streamer.users(account);
    return stakedBalance;
}

invariant sumOfBalancesIsTotalStaked()
  sumOfBalances == to_mathint(totalStaked());

invariant accountMPLessEqualAccountMaxMP(address account)
  to_mathint(getAccountMP(account)) <= to_mathint(getAccountMaxMP(account));

