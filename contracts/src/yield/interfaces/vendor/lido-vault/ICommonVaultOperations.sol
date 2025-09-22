pragma solidity >=0.8.0;

// Common interface between IDashboard and IStakingVault

interface ICommonVaultOperations {
    /**
     * @notice Funds the staking vault with ether
     */
    function fund() external payable;

    /**
     * @notice Withdraws ether from the staking vault to a recipient
     * @param _recipient Address of the recipient
     * @param _ether Amount of ether to withdraw
     */
    function withdraw(address _recipient, uint256 _ether) external;

    /**
     * @notice Pauses beacon chain deposits on the StakingVault.
     */
    function pauseBeaconChainDeposits() external;

    /**
     * @notice Resumes beacon chain deposits on the StakingVault.
     */
    function resumeBeaconChainDeposits() external;
}