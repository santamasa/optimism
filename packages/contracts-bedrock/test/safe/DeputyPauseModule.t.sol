// SPDX-License-Identifier: MIT
pragma solidity 0.8.15;

// Testing
import { CommonTest } from "test/setup/CommonTest.sol";
import { ForgeArtifacts, Abi } from "scripts/libraries/ForgeArtifacts.sol";
import { GnosisSafe as Safe } from "safe-contracts/GnosisSafe.sol";
import "test/safe-tools/SafeTestTools.sol";

// Scripts
import { DeployUtils } from "scripts/libraries/DeployUtils.sol";

// Interfaces
import { IDeputyGuardianModule } from "src/safe/interfaces/IDeputyGuardianModule.sol";
import { IDeputyPauseModule } from "src/safe/interfaces/IDeputyPauseModule.sol";

contract DeputyPauseModule_TestInit is CommonTest, SafeTestTools {
    using SafeTestLib for SafeInstance;

    error Unauthorized();
    error ExecutionFailed(string);

    event ExecutionFromModuleSuccess(address indexed);

    IDeputyPauseModule deputyPauseModule;
    IDeputyGuardianModule deputyGuardianModule;
    SafeInstance securityCouncilSafeInstance;
    SafeInstance foundationSafeInstance;
    address deputy;

    /// @dev Sets up the test environment.
    function setUp() public virtual override {
        super.setUp();

        // Set up 20 keys.
        (, uint256[] memory keys) = SafeTestLib.makeAddrsAndKeys("DeputyPauseModule_test_", 20);

        // Split into two sets of 10 keys.
        uint256[] memory keys1 = new uint256[](10);
        uint256[] memory keys2 = new uint256[](10);
        for (uint256 i; i < 10; i++) {
            keys1[i] = keys[i];
            keys2[i] = keys[i + 10];
        }

        // Create a Security Council Safe with 10 owners.
        securityCouncilSafeInstance = _setupSafe(keys1, 10);

        // Create a Foundation Safe with 10 different owners.
        foundationSafeInstance = _setupSafe(keys2, 10);

        // Set the Security Council Safe as the Guardian of the SuperchainConfig.
        vm.store(
            address(superchainConfig),
            superchainConfig.GUARDIAN_SLOT(),
            bytes32(uint256(uint160(address(securityCouncilSafeInstance.safe))))
        );

        // Create a DeputyGuardianModule and set the Foundation Safe as the Deputy Guardian.
        deputyGuardianModule = IDeputyGuardianModule(
            DeployUtils.create1({
                _name: "DeputyGuardianModule",
                _args: DeployUtils.encodeConstructor(
                    abi.encodeCall(
                        IDeputyGuardianModule.__constructor__,
                        (securityCouncilSafeInstance.safe, superchainConfig, address(foundationSafeInstance.safe))
                    )
                )
            })
        );

        // Enable the DeputyGuardianModule on the Security Council Safe.
        securityCouncilSafeInstance.enableModule(address(deputyGuardianModule));

        // Create the deputy for the DeputyPauseModule.
        deputy = makeAddr("deputy");

        // Create the DeputyPauseModule.
        deputyPauseModule = IDeputyPauseModule(
            DeployUtils.create1({
                _name: "DeputyPauseModule",
                _args: DeployUtils.encodeConstructor(
                    abi.encodeCall(
                        IDeputyPauseModule.__constructor__,
                        (foundationSafeInstance.safe, deputyGuardianModule, deputy)
                    )
                )
            })
        );

        // Enable the DeputyPauseModule on the Foundation Safe.
        foundationSafeInstance.enableModule(address(deputyPauseModule));
    }
}

contract DeputyPauseModule_Getters_Test is DeputyPauseModule_TestInit {
    /// @dev Tests that the getters work.
    function test_getters_works() external view {
        assertEq(address(deputyPauseModule.foundationSafe()), address(foundationSafeInstance.safe));
        assertEq(address(deputyPauseModule.deputyGuardianModule()), address(deputyGuardianModule));
        assertEq(deputyPauseModule.deputy(), deputy);
    }
}

contract DeputyPauseModule_Pause_Test is DeputyPauseModule_TestInit {
    /// @dev Tests that `pause` successfully pauses when called by the deputy.
    function test_pause_succeeds() external {
        vm.expectEmit(address(superchainConfig));
        emit Paused("Deputy Guardian");

        vm.expectEmit(address(securityCouncilSafeInstance.safe));
        emit ExecutionFromModuleSuccess(address(deputyGuardianModule));

        vm.expectEmit(address(deputyGuardianModule));
        emit Paused("Deputy Guardian");

        vm.expectEmit(address(foundationSafeInstance.safe));
        emit ExecutionFromModuleSuccess(address(deputyPauseModule));

        vm.expectEmit(address(deputyPauseModule));
        emit Paused("Pause Deputy");

        vm.prank(address(deputy));
        deputyPauseModule.pause();
        assertEq(superchainConfig.paused(), true);
    }
}

contract DeputyPauseModule_Pause_TestFail is DeputyPauseModule_TestInit {
    /// @dev Tests that `pause` reverts when called by an address other than the deputy.
    function testFuzz_pause_notDeputy_reverts(address _sender) external {
        vm.assume(_sender != address(deputy));
        vm.expectRevert(abi.encodeWithSelector(Unauthorized.selector));
        vm.prank(_sender);
        deputyPauseModule.pause();
    }

    /// @dev Tests that the error message is returned when the call to the safe reverts.
    function test_pause_targetReverts_reverts() external {
        vm.mockCallRevert(
            address(superchainConfig),
            abi.encodePacked(superchainConfig.pause.selector),
            "SuperchainConfig: pause() reverted"
        );

        // Note that the error here will be somewhat awkwardly double-encoded because the
        // DeputyGuardianModule will encode the revert message as an ExecutionFailed error and then
        // the DeputyPauseModule will re-encode it as another ExecutionFailed error.
        vm.prank(address(deputy));
        vm.expectRevert(
            abi.encodeWithSelector(
                ExecutionFailed.selector,
                string(abi.encodeWithSelector(ExecutionFailed.selector, "SuperchainConfig: pause() reverted"))
            )
        );
        deputyPauseModule.pause();
    }
}
