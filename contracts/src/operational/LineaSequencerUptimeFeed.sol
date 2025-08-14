// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.30;

import { AggregatorInterface } from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorInterface.sol";
import { AggregatorV3Interface } from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV3Interface.sol";
import { AggregatorV2V3Interface } from "@chainlink/contracts/src/v0.8/shared/interfaces/AggregatorV2V3Interface.sol";
import { ISequencerUptimeFeed } from "@chainlink/contracts/src/v0.8/l2ep/interfaces/ISequencerUptimeFeed.sol";
import { ITypeAndVersion } from "@chainlink/contracts/src/v0.8/shared/interfaces/ITypeAndVersion.sol";
import { SimpleReadAccessController } from "@chainlink/contracts/src/v0.8/shared/access/SimpleReadAccessController.sol";
import { AccessControl } from "@openzeppelin/contracts/access/AccessControl.sol";
import { ILineaSequencerUptimeFeed } from "./interfaces/ILineaSequencerUptimeFeed.sol";

/**
 * @title LineaSequencerUptimeFeed - L2 sequencer uptime status aggregator.
 * @notice L2 contract that receives status updates,
 *  and records a new answer if the status changed.
 */
contract LineaSequencerUptimeFeed is
  AggregatorV2V3Interface,
  ISequencerUptimeFeed,
  ILineaSequencerUptimeFeed,
  ITypeAndVersion,
  SimpleReadAccessController,
  AccessControl
{
  bytes32 public constant FEED_UPDATER_ROLE = keccak256("FEED_UPDATER_ROLE");

  // solhint-disable-next-line chainlink-solidity/all-caps-constant-storage-variables
  uint8 public constant override decimals = 0;
  // solhint-disable-next-line chainlink-solidity/all-caps-constant-storage-variables
  string public constant override description = "L2 Sequencer Uptime Status Feed";
  // solhint-disable-next-line chainlink-solidity/all-caps-constant-storage-variables
  uint256 public constant override version = 1;

  /// @dev s_latestRoundId == 0 means this contract is uninitialized.
  FeedState private s_feedState = FeedState({ latestRoundId: 0, latestStatus: false, startedAt: 0, updatedAt: 0 });

  mapping(uint80 roundId => Round round) private s_rounds;

  /**
   * @param _initialStatus The initial status of the feed.
   * @param _admin The address of the admin that can manage the feed.
   * @param _feedUpdater The address of the feed updater that can update the status.
   */
  constructor(bool _initialStatus, address _admin, address _feedUpdater) {
    require(_admin != address(0), ZeroAddressNotAllowed());
    require(_feedUpdater != address(0), ZeroAddressNotAllowed());

    uint64 timestamp = uint64(block.timestamp);

    /// @dev Initialise roundId == 1 as the first round
    _recordRound(1, _initialStatus, timestamp);

    _grantRole(DEFAULT_ADMIN_ROLE, _admin);
    _grantRole(FEED_UPDATER_ROLE, _feedUpdater);
  }

  /**
   * @notice Check if a roundId is valid in this current contract state.
   * @dev Mainly used for AggregatorV2V3Interface functions.
   * @param _roundId Round ID to check.
   */
  function _isValidRound(uint256 _roundId) private view returns (bool) {
    return _roundId > 0 && _roundId <= type(uint80).max && s_feedState.latestRoundId >= _roundId;
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
   * @notice Helper function to set the latest feed state.
   *
   * @param _status Sequencer status.
   * @param _timestamp The L2 block timestamp of status update.
   */
  function _recordRound(uint80 _roundId, bool _status, uint64 _timestamp) private {
    s_rounds[_roundId] = Round({ status: _status, startedAt: _timestamp, updatedAt: uint64(block.timestamp) });
    s_feedState = FeedState({
      latestRoundId: _roundId,
      latestStatus: _status,
      startedAt: _timestamp,
      updatedAt: uint64(block.timestamp)
    });

    emit NewRound(_roundId, msg.sender, _timestamp);
    emit AnswerUpdated(_getStatusAnswer(_status), _roundId, _timestamp);
  }

  /**
   * @notice Helper function to update when a round was last updated.
   * @param _roundId The round ID to update.
   * @param _status Sequencer status.
   */
  function _updateRound(uint80 _roundId, bool _status) private {
    s_rounds[_roundId].updatedAt = uint64(block.timestamp);
    s_feedState.updatedAt = uint64(block.timestamp);
    emit RoundUpdated(_getStatusAnswer(_status), uint64(block.timestamp));
  }

  /// @inheritdoc AggregatorInterface
  function latestAnswer() external view override checkAccess returns (int256) {
    return _getStatusAnswer(s_feedState.latestStatus);
  }

  /// @inheritdoc AggregatorInterface
  function latestTimestamp() external view override checkAccess returns (uint256) {
    return s_feedState.updatedAt;
  }

  /// @inheritdoc AggregatorInterface
  function latestRound() external view override checkAccess returns (uint256) {
    return s_feedState.latestRoundId;
  }

  /// @inheritdoc AggregatorInterface
  function getAnswer(uint256 _roundId) external view override checkAccess returns (int256) {
    if (!_isValidRound(_roundId)) {
      revert NoDataPresent();
    }
    return _getStatusAnswer(s_rounds[uint80(_roundId)].status);
  }

  /// @inheritdoc AggregatorInterface
  function getTimestamp(uint256 _roundId) external view override checkAccess returns (uint256) {
    if (!_isValidRound(_roundId)) {
      revert NoDataPresent();
    }
    return s_rounds[uint80(_roundId)].startedAt;
  }

  /**
   * @notice Record a new status and timestamp if it has changed since the last round.
   * @dev This function will revert if not called from an account with the `FEED_UPDATER_ROLE`.
   * @param _status Sequencer status.
   * @param _timestamp Block timestamp of status update.
   */
  function updateStatus(bool _status, uint64 _timestamp) external override {
    require(_isValidSender(msg.sender), InvalidSender());

    FeedState memory feedState = s_feedState;

    /// @dev Ignore if latest recorded timestamp is newer
    if (feedState.startedAt > _timestamp) {
      emit UpdateIgnored(feedState.latestStatus, feedState.startedAt, _status, _timestamp);
      return;
    }

    if (feedState.latestStatus == _status) {
      _updateRound(feedState.latestRoundId, _status);
    } else {
      feedState.latestRoundId += 1;
      _recordRound(feedState.latestRoundId, _status, _timestamp);
    }
  }

  /// @inheritdoc AggregatorV3Interface
  function getRoundData(
    uint80 _roundId
  )
    external
    view
    override
    checkAccess
    returns (uint80 roundId, int256 answer, uint256 startedAt, uint256 updatedAt, uint80 answeredInRound)
  {
    if (!_isValidRound(_roundId)) {
      revert NoDataPresent();
    }

    Round storage round = s_rounds[_roundId];
    return (_roundId, _getStatusAnswer(round.status), round.startedAt, round.updatedAt, _roundId);
  }

  /// @inheritdoc AggregatorV3Interface
  function latestRoundData()
    external
    view
    override
    checkAccess
    returns (uint80 roundId, int256 answer, uint256 startedAt, uint256 updatedAt, uint80 answeredInRound)
  {
    FeedState storage feedState = s_feedState;

    roundId = feedState.latestRoundId;
    answer = _getStatusAnswer(feedState.latestStatus);
    startedAt = feedState.startedAt;
    updatedAt = feedState.updatedAt;
    answeredInRound = roundId;
  }
}
