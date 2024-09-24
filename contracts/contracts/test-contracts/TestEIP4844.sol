// SPDX-License-Identifier: AGPL-3.0
pragma solidity 0.8.26;

contract TestEIP4844 {
  event BlobHashEvent(bytes32 blobHash);

  function submitData(
    bytes32 _snarkHash,
    uint256 _y,
    bytes1[48] memory _kzgCommitment,
    bytes1[48] memory _kzgProof
  ) external {
    bytes32 h = blobhash(0);

    bytes32 compressedDataComputedX = keccak256(abi.encode(_snarkHash, h));

    (bool success, ) = address(0x0a).staticcall(
      abi.encodePacked(h, compressedDataComputedX, _y, _kzgCommitment, _kzgProof)
    );
    if (!success) {
      revert("PointEvaluationFailed");
    }

    emit BlobHashEvent(h);
  }
}
