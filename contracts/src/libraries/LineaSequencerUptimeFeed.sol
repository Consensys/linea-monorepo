// SPDX-License-Identifier: MIT
pragma solidity ^0.8.4;

import { AggregatorInterface } from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorInterface.sol";
import { AggregatorV3Interface } from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";
import { AggregatorV2V3Interface } from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV2V3Interface.sol";
import { ITypeAndVersion } from "@chainlink/contracts/src/v0.8/shared/interfaces/ITypeAndVersion.sol";
import { LineaSequencerUptimeFeedInterface } from "./LineaSequencerUptimeFeedInterface.sol";
import { SimpleReadAccessController } from "@chainlink/contracts/src/v0.8/shared/access/SimpleReadAccessController.sol";
import { AccessControl } from "@openzeppelin/contracts/access/AccessControl.sol";

/**
 * @title LineaSequencerUptimeFeed - L2 sequencer uptime status aggregator.
 * @notice L2 contract that receives status updates,
 *  and records a new answer if the status changed.
 */
contract LineaSequencerUptimeFeed is
  AggregatorV2V3Interface,
  LineaSequencerUptimeFeedInterface,
  ITypeAndVersion,
  SimpleReadAccessController,
  AccessControl
{
  /// @dev Packed state struct to save sloads.
  struct FeedState {
    /// @dev Dummy roundId for backward compatibility.
    uint80 latestRoundId;
    /**
     * @dev  A variable with a value of either true or false.
     * @dev  false: The sequencer is up.
     * @dev  true: The sequencer is down.
     */
    bool latestStatus;
    uint64 startedAt;
    uint64 updatedAt;
  }

  /// @notice Sender is not allowed to update the status.
  error InvalidSender();
  /// @notice The address is zero.
  error ZeroAddressNotAllowed();

  /// @dev Emitted when an `updateStatus` call is ignored due to unchanged status or stale timestamp.
  event UpdateIgnored(bool latestStatus, uint64 latestTimestamp, bool incomingStatus, uint64 incomingTimestamp);
  /// @dev Emitted when a updateStatus is called without the status changing.
  event RoundUpdated(int256 status, uint64 updatedAt);

  bytes32 public constant FEED_UPDATER_ROLE = keccak256("FEED_UPDATER_ROLE");

  // solhint-disable-next-line chainlink-solidity/all-caps-constant-storage-variables
  uint8 public constant override decimals = 0;
  // solhint-disable-next-line chainlink-solidity/all-caps-constant-storage-variables
  string public constant override description = "L2 Sequencer Uptime Status Feed";
  // solhint-disable-next-line chainlink-solidity/all-caps-constant-storage-variables
  uint256 public constant override version = 1;

  /// @dev s_latestRoundId == 0 means this contract is uninitialized.
  FeedState private s_feedState = FeedState({ latestRoundId: 0, latestStatus: false, startedAt: 0, updatedAt: 0 });

  /**
   * @param _initialStatus The initial status of the feed.
   * @param _admin The address of the admin that can manage the feed.
   * @param _l2Sender The address of the L2 sender that can update the status.
   */
  constructor(bool _initialStatus, address _admin, address _l2Sender) {
    require(_admin != address(0), ZeroAddressNotAllowed());
    require(_l2Sender != address(0), ZeroAddressNotAllowed());

    uint64 timestamp = uint64(block.timestamp);

    /// @dev Initialise dummy roundId == 0.
    _recordRound(_initialStatus, timestamp);

    _grantRole(DEFAULT_ADMIN_ROLE, _admin);
    _grantRole(FEED_UPDATER_ROLE, _l2Sender);
  }

  /**
   * @notice versions:
   *
   * - LineaSequencerUptimeFeed 1.0.0: initial release.
   *
   * @inheritdoc ITypeAndVersion
   */
  function typeAndVersion() external pure virtual override returns (string memory) {
    return "LineaSequencerUptimeFeed 1.0.0";
  }

  /**
   * @dev Returns an AggregatorV2V3Interface compatible answer from status flag.
   * @param _status The status flag to convert to an aggregator-compatible answer.
   */
  function _getStatusAnswer(bool _status) private pure returns (int256) {
    return _status ? int256(1) : int256(0);
  }

  /**
   * @notice Helper function to check if the sender is allowed to update the status.
   * @param _sender Sender address.
   */
  function _isValidSender(address _sender) private view returns (bool) {
    return hasRole(FEED_UPDATER_ROLE, _sender);
  }

  /**
   * @notice Helper function to record a round and set the latest feed state.
   *
   * @param _status Sequencer status.
   * @param _timestamp The L2 block timestamp of status update.
   */
  function _recordRound(bool _status, uint64 _timestamp) private {
    uint64 updatedAt = uint64(block.timestamp);
    FeedState memory feedState = FeedState(0, _status, _timestamp, updatedAt);
    s_feedState = feedState;
    emit AnswerUpdated(_getStatusAnswer(_status), 0, _timestamp);
  }

  /**
   * @notice Helper function to update when a round was last updated.
   * @param _status Sequencer status.
   */
  function _updateRound(bool _status) private {
    uint64 updatedAt = uint64(block.timestamp);
    s_feedState.updatedAt = updatedAt;
    emit RoundUpdated(_getStatusAnswer(_status), updatedAt);
  }

  /**
   * @notice Record a new status and timestamp if it has changed since the last round.
   * @dev This function will revert if not called from `l1Sender` via the L1->L2 messenger.
   * @param _status Sequencer status.
   * @param _timestamp Block timestamp of status update.
   */
  function updateStatus(bool _status, uint64 _timestamp) external override {
    FeedState memory feedState = s_feedState;

    require(_isValidSender(msg.sender), InvalidSender());

    /// @dev Ignore if latest recorded timestamp is newer
    if (feedState.startedAt > _timestamp) {
      emit UpdateIgnored(feedState.latestStatus, feedState.startedAt, _status, _timestamp);
      return;
    }

    if (feedState.latestStatus == _status) {
      _updateRound(_status);
    } else {
      _recordRound(_status, _timestamp);
    }
  }

  /// @inheritdoc AggregatorInterface
  function latestAnswer() external view override checkAccess returns (int256) {
    FeedState memory feedState = s_feedState;
    return _getStatusAnswer(feedState.latestStatus);
  }

  /// @inheritdoc AggregatorV3Interface
  function latestRoundData()
    external
    view
    override
    checkAccess
    returns (uint80 roundId, int256 answer, uint256 startedAt, uint256 updatedAt, uint80 answeredInRound)
  {
    FeedState memory feedState = s_feedState;

    roundId = feedState.latestRoundId;
    answer = _getStatusAnswer(feedState.latestStatus);
    startedAt = feedState.startedAt;
    updatedAt = feedState.updatedAt;
    answeredInRound = roundId;
  }

  function getAnswer(uint256 _roundId) external view override returns (int256) {
    return _getStatusAnswer(s_feedState.latestStatus);
  }

  function getRoundData(
    uint80 _roundId
  )
    external
    view
    override
    returns (uint80 roundId, int256 answer, uint256 startedAt, uint256 updatedAt, uint80 answeredInRound)
  {
    return (0, _getStatusAnswer(s_feedState.latestStatus), s_feedState.startedAt, s_feedState.updatedAt, 0);
  }

  function getTimestamp(uint256 _roundId) external view override returns (uint256) {
    return s_feedState.updatedAt;
  }

  function latestTimestamp() external view override returns (uint256) {
    return s_feedState.updatedAt;
  }

  function latestRound() external view override returns (uint256) {
    return s_feedState.latestRoundId;
  }
}
