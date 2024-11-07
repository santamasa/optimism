// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Contracts
import { FaultDisputeGame } from "src/dispute/FaultDisputeGame.sol";

// Libraries
import { GameType, Claim, Duration } from "src/dispute/lib/Types.sol";
import { BadAuth } from "src/dispute/lib/Errors.sol";

// Interfaces
import { IDelayedWETH } from "src/dispute/interfaces/IDelayedWETH.sol";
import { IAnchorStateRegistry } from "src/dispute/interfaces/IAnchorStateRegistry.sol";
import { IBigStepper } from "src/dispute/interfaces/IBigStepper.sol";

/// @title PermissionedDisputeGame
/// @notice PermissionedDisputeGame is a contract that inherits from `FaultDisputeGame`, and contains two roles:
///         - The `challenger` role, which is allowed to challenge a dispute.
///         - The `proposer` role, which is allowed to create proposals and participate in their game.
///         This contract exists as a way for networks to support the fault proof iteration of the OptimismPortal
///         contract without needing to support a fully permissionless system. Permissionless systems can introduce
///         costs that certain networks may not wish to support. This contract can also be used as a fallback mechanism
///         in case of a failure in the permissionless fault proof system in the stage one release.
contract PermissionedDisputeGame is FaultDisputeGame {
    /// @notice Modifier that gates access to the `challenger` and `proposer` roles.
    modifier onlyAuthorized() {
        if (!(msg.sender == proposer() || msg.sender == challenger())) {
            revert BadAuth();
        }
        _;
    }

    /// @inheritdoc FaultDisputeGame
    function step(
        uint256 _claimIndex,
        bool _isAttack,
        bytes calldata _stateData,
        bytes calldata _proof
    )
        public
        override
        onlyAuthorized
    {
        super.step(_claimIndex, _isAttack, _stateData, _proof);
    }

    /// @notice Generic move function, used for both `attack` and `defend` moves.
    /// @notice _disputed The disputed `Claim`.
    /// @param _challengeIndex The index of the claim being moved against. This must match the `_disputed` claim.
    /// @param _claim The claim at the next logical position in the game.
    /// @param _isAttack Whether or not the move is an attack or defense.
    function move(
        Claim _disputed,
        uint256 _challengeIndex,
        Claim _claim,
        bool _isAttack
    )
        public
        payable
        override
        onlyAuthorized
    {
        super.move(_disputed, _challengeIndex, _claim, _isAttack);
    }

    /// @notice Initializes the contract.
    function initialize() public payable override {
        // The creator of the dispute game must be the proposer EOA.
        if (tx.origin != proposer()) revert BadAuth();

        // Fallthrough initialization.
        super.initialize();
    }

    ////////////////////////////////////////////////////////////////
    //                     IMMUTABLE GETTERS                      //
    ////////////////////////////////////////////////////////////////

    /// @notice Getter for the proposer address.
    /// @dev `clones-with-immutable-args` argument #15
    /// @return proposer_ The address of the proposer.
    function proposer() public pure returns (address proposer_) {
        proposer_ = _getArgAddress(0x141);
    }

    /// @notice Getter for the challenger address.
    /// @dev `clones-with-immutable-args` argument #16
    /// @return challenger_ The address of the challenger.
    function challenger() public pure returns (address challenger_) {
        challenger_ = _getArgAddress(0x155);
    }
}
