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
    /// @custom:field _gameType The type ID of the game.
    /// @custom:field _absolutePrestate The absolute prestate of the instruction trace.
    /// @custom:field _maxGameDepth The maximum depth of bisection.
    /// @custom:field _splitDepth The final depth of the output bisection portion of the game.
    /// @custom:field _clockExtension The clock extension to perform when the remaining duration is less than the
    /// extension.
    /// @custom:field _maxClockDuration The maximum amount of time that may accumulate on a team's chess clock.
    /// @custom:field _vm An onchain VM that performs single instruction steps on an FPP trace.
    /// @custom:field _weth WETH contract for holding ETH.
    /// @custom:field _anchorStateRegistry The contract that stores the anchor state for each game type.
    /// @custom:field _l2ChainId Chain ID of the L2 network this contract argues about.
    /// @custom:field _proposer Address that is allowed to create instances of this contract.
    /// @custom:field _challenger Address that is allowed to challenge instances of this contract.
    struct PDGConstructorParams {
        GameType _gameType;
        Claim _absolutePrestate;
        uint256 _maxGameDepth;
        uint256 _splitDepth;
        Duration _clockExtension;
        Duration _maxClockDuration;
        IBigStepper _vm;
        IDelayedWETH _weth;
        IAnchorStateRegistry _anchorStateRegistry;
        uint256 _l2ChainId;
        address _proposer;
        address _challenger;
    }

    /// @notice The proposer role is allowed to create proposals and participate in the dispute game.
    address internal immutable PROPOSER;

    /// @notice The challenger role is allowed to participate in the dispute game.
    address internal immutable CHALLENGER;

    /// @notice Modifier that gates access to the `challenger` and `proposer` roles.
    modifier onlyAuthorized() {
        if (!(msg.sender == PROPOSER || msg.sender == CHALLENGER)) {
            revert BadAuth();
        }
        _;
    }

    constructor(PDGConstructorParams memory _params)
        FaultDisputeGame(
            _params._gameType,
            _params._absolutePrestate,
            _params._maxGameDepth,
            _params._splitDepth,
            _params._clockExtension,
            _params._maxClockDuration,
            _params._vm,
            _params._weth,
            _params._anchorStateRegistry,
            _params._l2ChainId
        )
    {
        PROPOSER = _params._proposer;
        CHALLENGER = _params._challenger;
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
        if (tx.origin != PROPOSER) revert BadAuth();

        // Fallthrough initialization.
        super.initialize();
    }

    ////////////////////////////////////////////////////////////////
    //                     IMMUTABLE GETTERS                      //
    ////////////////////////////////////////////////////////////////

    /// @notice Returns the proposer address.
    function proposer() external view returns (address proposer_) {
        proposer_ = PROPOSER;
    }

    /// @notice Returns the challenger address.
    function challenger() external view returns (address challenger_) {
        challenger_ = CHALLENGER;
    }
}
