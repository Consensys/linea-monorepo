// SPDX-License-Identifier: AGPL-3.0
pragma solidity ^0.8.30;

interface ILineaSequencerUptimeFeed {
  /**
   * @dev Round info for uptime history.
   * @param startedAt The timestamp at which the round started.
   * @param updatedAt The timestamp at which the round was updated.
   * @param status The sequencer status for the round.
   */
  struct Round {
    uint64 startedAt;
    uint64 updatedAt;
    bool status;
  }

  /**
   * @notice Current sequencer uptime status.
   * @dev Packed state struct to save sloads.
   * @param latestRoundId The ID of the latest round.
   * @param latestStatus The latest sequencer status.
   * @dev false: The sequencer is up.
   * @dev true: The sequencer is down.
   * @param startedAt The timestamp when the feed was started.
   * @param updatedAt The timestamp when the feed was last updated.
   */
  struct FeedState {
    uint80 latestRoundId;
    bool latestStatus;
    uint64 startedAt;
    uint64 updatedAt;
  }

  /**
   * @notice Emitted when an `updateStatus` call is ignored due to unchanged status or stale timestamp.
   * @param latestStatus The current status of the sequencer.
   * @param latestTimestamp The timestamp of the latest status update.
   * @param incomingStatus The new status of the sequencer.
   * @param incomingTimestamp The timestamp of the new status update.
   */
  event UpdateIgnored(bool latestStatus, uint64 latestTimestamp, bool incomingStatus, uint64 incomingTimestamp);

  /**
   * @notice Emitted when a updateStatus is called without the status changing.
   * @param status The current status of the sequencer.
   * @param updatedAt The timestamp of the status update.
   */
  event RoundUpdated(int256 status, uint64 updatedAt);

  /**
   * @dev Thrown when sender is not allowed to update the status.
   */
  error InvalidSender();

  /**
   * @dev Thrown when a parameter is the zero address.
   */
  error ZeroAddressNotAllowed();

  /**
   * @dev Thrown when no data is present for a given roundId.
   */
  error NoDataPresent();
}
