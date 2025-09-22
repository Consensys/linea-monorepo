// SPDX-FileCopyrightText: 2025 Lido <info@lido.fi>
// SPDX-License-Identifier: GPL-3.0

// See contracts/COMPILERS.md
pragma solidity >=0.8.0;

import { IVaultHub } from "./IVaultHub.sol";
import { IStakingVault } from "./IStakingVault.sol";
import { IPredepositGuarantee } from "./IPredepositGuarantee.sol";


/**
 * @title Dashboard
 * @notice This contract is a UX-layer for StakingVault and meant to be used as its owner.
 * This contract improves the vault UX by bundling all functions from the StakingVault and VaultHub
 * in this single contract. It provides administrative functions for managing the StakingVault,
 * including funding, withdrawing, minting, burning, and rebalancing operations.
 */
interface IDashboard {
    /**
     * @notice The stETH token contract
     */
    function STETH() external view returns(address);

    /**
     * @notice The wstETH token contract
     */
    function WSTETH() external view returns(address);

    /**
     * @notice ETH address convention per EIP-7528
     */
    function ETH() external view returns(address);

    // ==================== View Functions ====================

    /**
     * @notice Returns the vault connection data for the staking vault.
     * @return VaultConnection struct containing vault data
     */
    function vaultConnection() external view returns (IVaultHub.VaultConnection memory);

    /**
     * @notice Returns the stETH share limit of the vault
     */
    function shareLimit() external view returns (uint256);

    /**
     * @notice Returns the number of stETH shares minted
     */
    function liabilityShares() external view returns (uint256);

    /**
     * @notice Returns the reserve ratio of the vault in basis points
     */
    function reserveRatioBP() external view returns (uint16);

    /**
     * @notice Returns the rebalance threshold of the vault in basis points.
     */
    function forcedRebalanceThresholdBP() external view returns (uint16);

    /**
     * @notice Returns the infra fee basis points.
     */
    function infraFeeBP() external view returns (uint16);

    /**
     * @notice Returns the liquidity fee basis points.
     */
    function liquidityFeeBP() external view returns (uint16);

    /**
     * @notice Returns the reservation fee basis points.
     */
    function reservationFeeBP() external view returns (uint16);

    /**
     * @notice Returns the total value of the vault in ether.
     */
    function totalValue() external view returns (uint256);

    /**
     * @notice Returns the overall unsettled obligations of the vault in ether
     * @dev includes the node operator fee
     */
    function unsettledObligations() external view returns (uint256);

    /**
     * @notice Returns the locked amount of ether for the vault
     */
    function locked() external view returns (uint256);

    /**
     * @notice Returns the max total lockable amount of ether for the vault (excluding the Lido and node operator fees)
     */
    function maxLockableValue() external view returns (uint256);

    /**
     * @notice Returns the overall capacity for stETH shares that can be minted by the vault
     */
    function totalMintingCapacityShares() external view returns (uint256);

    /**
     * @notice Returns the remaining capacity for stETH shares that can be minted
     *         by the vault if additional ether is funded
     * @param _etherToFund the amount of ether to be funded, can be zero
     * @return the number of shares that can be minted using additional ether
     */
    function remainingMintingCapacityShares(uint256 _etherToFund) external view returns (uint256);

    /**
     * @notice Returns the amount of ether that can be instantly withdrawn from the staking vault.
     * @dev This is the amount of ether that is not locked in the StakingVault and not reserved for fees and obligations.
     */
    function withdrawableValue() external view returns (uint256);

    // ==================== Vault Management Functions ====================

    /**
     * @notice Transfers the ownership of the underlying StakingVault from this contract to a new owner
     *         without disconnecting it from the hub
     * @param _newOwner Address of the new owner.
     */
    function transferVaultOwnership(address _newOwner) external;

    /**
     * @notice Disconnects the underlying StakingVault from the hub and passing its ownership to Dashboard.
     *         After receiving the final report, one can call reconnectToVaultHub() to reconnect to the hub
     *         or abandonDashboard() to transfer the ownership to a new owner.
     */
    function voluntaryDisconnect() external;

    /**
     * @notice Accepts the ownership over the StakingVault transferred from VaultHub on disconnect
     * and immediately transfers it to a new pending owner. This new owner will have to accept the ownership
     * on the StakingVault contract.
     * @param _newOwner The address to transfer the StakingVault ownership to.
     */
    function abandonDashboard(address _newOwner) external;

    /**
     * @notice Accepts the ownership over the StakingVault and connects to VaultHub. Can be called to reconnect
     *         to the hub after voluntaryDisconnect()
     */
    function reconnectToVaultHub() external;

    /**
     * @notice Connects to VaultHub, transferring ownership to VaultHub.
     */
    function connectToVaultHub() external payable;

    /**
     * @notice Changes the tier of the vault and connects to VaultHub
     * @param _tierId The tier to change to
     * @param _requestedShareLimit The requested share limit
     */
    function connectAndAcceptTier(uint256 _tierId, uint256 _requestedShareLimit) external payable;

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
     * @notice Mints stETH shares backed by the vault to the recipient.
     * @param _recipient Address of the recipient
     * @param _amountOfShares Amount of stETH shares to mint
     */
    function mintShares(address _recipient, uint256 _amountOfShares) external payable;

    /**
     * @notice Mints stETH tokens backed by the vault to the recipient.
     * !NB: this will revert with`VaultHub.ZeroArgument("_amountOfShares")` if the amount of stETH is less than 1 share
     * @param _recipient Address of the recipient
     * @param _amountOfStETH Amount of stETH to mint
     */
    function mintStETH(address _recipient, uint256 _amountOfStETH) external payable;

    /**
     * @notice Mints wstETH tokens backed by the vault to a recipient.
     * @param _recipient Address of the recipient
     * @param _amountOfWstETH Amount of tokens to mint
     */
    function mintWstETH(address _recipient, uint256 _amountOfWstETH) external payable;

    /**
     * @notice Burns stETH shares from the sender backed by the vault.
     *         Expects corresponding amount of stETH approved to this contract.
     * @param _amountOfShares Amount of stETH shares to burn
     */
    function burnShares(uint256 _amountOfShares) external;

    /**
     * @notice Burns stETH tokens from the sender backed by the vault. Expects stETH amount approved to this contract.
     * !NB: this will revert with `VaultHub.ZeroArgument("_amountOfShares")` if the amount of stETH is less than 1 share
     * @param _amountOfStETH Amount of stETH tokens to burn
     */
    function burnStETH(uint256 _amountOfStETH) external;

    /**
     * @notice Burns wstETH tokens from the sender backed by the vault. Expects wstETH amount approved to this contract.
     * !NB: this will revert with `VaultHub.ZeroArgument("_amountOfShares")` on 1 wei of wstETH due to rounding inside wstETH unwrap method
     * @param _amountOfWstETH Amount of wstETH tokens to burn

     */
    function burnWstETH(uint256 _amountOfWstETH) external;

    /**
     * @notice Rebalances StakingVault by withdrawing ether to VaultHub corresponding to shares amount provided
     * @param _shares amount of shares to rebalance
     */
    function rebalanceVaultWithShares(uint256 _shares) external;

    /**
     * @notice Rebalances the vault by transferring ether given the shares amount
     * @param _ether amount of ether to rebalance
     */
    function rebalanceVaultWithEther(uint256 _ether) external payable;

    /**
     * @notice Withdraws ether from vault and deposits directly to provided validators bypassing the default PDG process,
     *          allowing validators to be proven post-factum via `proveUnknownValidatorsToPDG`
     *          clearing them for future deposits via `PDG.depositToBeaconChain`
     * @param _deposits array of IStakingVault.Deposit structs containing deposit data
     * @return totalAmount total amount of ether deposited to beacon chain
     * @dev requires the caller to have the `UNGUARANTEED_BEACON_CHAIN_DEPOSIT_ROLE`
     * @dev can be used as PDG shortcut if the node operator is trusted to not frontrun provided deposits
     */
    function unguaranteedDepositToBeaconChain(
        IStakingVault.Deposit[] calldata _deposits
    ) external returns (uint256 totalAmount);

    /**
     * @notice Proves validators with correct vault WC if they are unknown to PDG
     * @param _witnesses array of IPredepositGuarantee.ValidatorWitness structs containing proof data for validators
     * @dev requires the caller to have the `PDG_PROVE_VALIDATOR_ROLE`
     */
    function proveUnknownValidatorsToPDG(IPredepositGuarantee.ValidatorWitness[] calldata _witnesses) external;

    /**
     * @notice Compensates ether of disproven validator's predeposit from PDG to the recipient.
     *         Can be called if validator which was predeposited via `PDG.predeposit` with vault funds
     *         was frontrun by NO's with non-vault WC (effectively NO's stealing the predeposit) and then
     *         proof of the validator's invalidity has been provided via `PDG.proveInvalidValidatorWC`.
     * @param _pubkey of validator that was proven invalid in PDG
     * @param _recipient address to receive the `PDG.PREDEPOSIT_AMOUNT`
     * @dev PDG will revert if _recipient is vault address, use fund() instead to return ether to vault
     * @dev requires the caller to have the `PDG_COMPENSATE_PREDEPOSIT_ROLE`
     */
    function compensateDisprovenPredepositFromPDG(bytes calldata _pubkey, address _recipient) external;

    /**
     * @notice Recovers ERC20 tokens or ether from the dashboard contract to sender
     * @param _token Address of the token to recover or 0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee for ether
     * @param _recipient Address of the recovery recipient
     */
    function recoverERC20(
        address _token,
        address _recipient,
        uint256 _amount
    ) external;

    /**
     * @notice Transfers a given token_id of an ERC721-compatible NFT (defined by the token contract address)
     * from the dashboard contract to sender
     *
     * @param _token an ERC721-compatible token
     * @param _tokenId token id to recover
     * @param _recipient Address of the recovery recipient
     */
    function recoverERC721(
        address _token,
        uint256 _tokenId,
        address _recipient
    ) external;

    /**
     * @notice Pauses beacon chain deposits on the StakingVault.
     */
    function pauseBeaconChainDeposits() external;

    /**
     * @notice Resumes beacon chain deposits on the StakingVault.
     */
    function resumeBeaconChainDeposits() external;

    /**
     * @notice Signals to node operators that specific validators should exit from the beacon chain. It DOES NOT
     *         directly trigger the exit - node operators must monitor for request events and handle the exits.
     * @param _pubkeys Concatenated validator external keys (48 bytes each).
     * @dev    Emits `ValidatorExitRequested` event for each validator external key through the `StakingVault`.
     *         This is a voluntary exit request - node operators can choose whether to act on it or not.
     */
    function requestValidatorExit(bytes calldata _pubkeys) external;

    /**
     * @notice Initiates a withdrawal from validator(s) on the beacon chain using EIP-7002 triggerable withdrawals
     *         Both partial withdrawals (disabled for if vault is unhealthy) and full validator exits are supported.
     * @param _pubkeys Concatenated validator external keys (48 bytes each).
     * @param _amounts Withdrawal amounts in wei for each validator key and must match _pubkeys length.
     *         Set amount to 0 for a full validator exit.
     *         For partial withdrawals, amounts will be trimmed to keep MIN_ACTIVATION_BALANCE on the validator to avoid deactivation
     * @param _refundRecipient Address to receive any fee refunds, if zero, refunds go to msg.sender.
     * @dev    A withdrawal fee must be paid via msg.value.
     *         Use `StakingVault.calculateValidatorWithdrawalFee()` to determine the required fee for the current block.
     */
    function triggerValidatorWithdrawals(
        bytes calldata _pubkeys,
        uint64[] calldata _amounts,
        address _refundRecipient
    ) external payable;
}