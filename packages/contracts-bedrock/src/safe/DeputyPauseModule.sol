// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Safe
import { GnosisSafe as Safe } from "safe-contracts/GnosisSafe.sol";
import { Enum } from "safe-contracts/common/Enum.sol";

// Libraries
import { Unauthorized } from "src/libraries/PortalErrors.sol";

// Interfaces
import { ISemver } from "src/universal/interfaces/ISemver.sol";
import { IDeputyGuardianModule } from "src/safe/interfaces/IDeputyGuardianModule.sol";

/// @title DeputyPauseModule
/// @notice Safe Module designed to be installed in the Foundation Safe which allows a specific
///         deputy address to act as the Foundation Safe for the sake of triggering the
///         Superchain-wide pause functionality. Significantly simplifies the process of triggering
///         a Superchain-wide pause without changing the existing security model.
contract DeputyPauseModule is ISemver {
    /// @notice Error message for failed transaction execution.
    error ExecutionFailed(string);

    /// @notice Emitted when the SuperchainConfig is paused.
    event Paused(string identifier);

    /// @notice Foundation Safe.
    Safe internal immutable FOUNDATION_SAFE;

    /// @notice DeputyGuardianModule used by the Security Council Safe.
    IDeputyGuardianModule internal immutable DEPUTY_GUARDIAN_MODULE;

    /// @notice Address of the deputy account.
    address internal immutable DEPUTY;

    /// @notice Semantic version.
    /// @custom:semver 2.0.1-beta.4
    string public constant version = "2.0.1-beta.4";

    /// @param _foundationSafe Address of the Foundation Safe.
    /// @param _deputyGuardianModule Address of the DeputyGuardianModule used by the Security Council Safe.
    /// @param _deputy Address of the deputy account.
    constructor(Safe _foundationSafe, IDeputyGuardianModule _deputyGuardianModule, address _deputy) {
        FOUNDATION_SAFE = _foundationSafe;
        DEPUTY_GUARDIAN_MODULE = _deputyGuardianModule;
        DEPUTY = _deputy;
    }

    /// @notice Getter function for the Foundation Safe address.
    /// @return foundationSafe_ Foundation Safe address.
    function foundationSafe() public view returns (Safe foundationSafe_) {
        foundationSafe_ = FOUNDATION_SAFE;
    }

    /// @notice Getter function for the DeputyGuardianModule address.
    /// @return deputyGuardianModule_ DeputyGuardianModule address.
    function deputyGuardianModule() public view returns (IDeputyGuardianModule deputyGuardianModule_) {
        deputyGuardianModule_ = DEPUTY_GUARDIAN_MODULE;
    }

    /// @notice Getter function for the deputy address.
    /// @return deputy_ Deputy address.
    function deputy() public view returns (address deputy_) {
        deputy_ = DEPUTY;
    }

    /// @notice Calls the Foundation Safe's `execTransactionFromModuleReturnData()` function with
    ///         the arguments necessary to call `pause()` on the Security Council Safe, which will
    ///         then cause the Security Council Safe to trigger SuperchainConfig pause.
    function pause() external {
        // Only the deputy can call this function.
        if (msg.sender != DEPUTY) {
            revert Unauthorized();
        }

        // Attempt to trigger the call.
        (bool success, bytes memory returnData) = FOUNDATION_SAFE.execTransactionFromModuleReturnData(
            address(DEPUTY_GUARDIAN_MODULE), 0, abi.encodeCall(IDeputyGuardianModule.pause, ()), Enum.Operation.Call
        );

        // If the call fails, revert.
        if (!success) {
            revert ExecutionFailed(string(returnData));
        }

        // Emit the Paused event and note who triggered it.
        emit Paused("Pause Deputy");
    }
}
