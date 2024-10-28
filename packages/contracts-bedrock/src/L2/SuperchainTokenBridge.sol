// SPDX-License-Identifier: MIT
pragma solidity 0.8.25;

// Libraries
import { Predeploys } from "src/libraries/Predeploys.sol";
import { ZeroAddress, Unauthorized } from "src/libraries/errors/CommonErrors.sol";
import { SafeSend } from "src/universal/SafeSend.sol";

// Interfaces
import { ISuperchainERC20 } from "src/L2/interfaces/ISuperchainERC20.sol";
import { ISuperchainWETH } from "src/L2/interfaces/ISuperchainWETH.sol";
import { IL2ToL2CrossDomainMessenger } from "src/L2/interfaces/IL2ToL2CrossDomainMessenger.sol";

/// @custom:proxied true
/// @custom:predeploy 0x4200000000000000000000000000000000000028
/// @title SuperchainTokenBridge
/// @notice The SuperchainTokenBridge allows for the bridging of ERC20 tokens to make them fungible across the
///         Superchain. It builds on top of the L2ToL2CrossDomainMessenger for both replay protection and domain
///         binding.
contract SuperchainTokenBridge {
    /// @notice Thrown when attempting to relay a message and the cross domain message sender is not the
    /// SuperchainTokenBridge.
    error InvalidCrossDomainSender();

    /// @notice Emitted when tokens are sent from one chain to another.
    /// @param token         Address of the token sent.
    /// @param from          Address of the sender.
    /// @param to            Address of the recipient.
    /// @param amount        Number of tokens sent.
    /// @param destination   Chain ID of the destination chain.
    event SendERC20(
        address indexed token, address indexed from, address indexed to, uint256 amount, uint256 destination
    );

    /// @notice Emitted whenever tokens are successfully relayed on this chain.
    /// @param token         Address of the token relayed.
    /// @param from          Address of the msg.sender of sendERC20 on the source chain.
    /// @param to            Address of the recipient.
    /// @param amount        Amount of tokens relayed.
    /// @param source        Chain ID of the source chain.
    event RelayERC20(address indexed token, address indexed from, address indexed to, uint256 amount, uint256 source);

    /// @notice Address of the L2ToL2CrossDomainMessenger Predeploy.
    address internal constant MESSENGER = Predeploys.L2_TO_L2_CROSS_DOMAIN_MESSENGER;

    /// @notice Semantic version.
    /// @custom:semver 1.0.0-beta.2
    string public constant version = "1.0.0-beta.2";

    // Fallback function to receive ETH
    receive() external payable { }

    /// @notice Sends tokens to a target address on another chain.
    /// @dev Tokens are burned on the source chain.
    /// @param _token    Token to send.
    /// @param _to       Address to send tokens to.
    /// @param _amount   Amount of tokens to send.
    /// @param _chainId  Chain ID of the destination chain.
    /// @return msgHash_ Hash of the message sent.
    function sendERC20(
        address _token,
        address _to,
        uint256 _amount,
        uint256 _chainId
    )
        external
        returns (bytes32 msgHash_)
    {
        if (_to == address(0)) revert ZeroAddress();

        ISuperchainERC20(_token).crosschainBurn(msg.sender, _amount);

        bytes memory message = abi.encodeCall(this.relayERC20, (_token, msg.sender, _to, _amount));
        msgHash_ = IL2ToL2CrossDomainMessenger(MESSENGER).sendMessage(_chainId, address(this), message);

        emit SendERC20(_token, msg.sender, _to, _amount, _chainId);
    }

    /// @notice Relays tokens received from another chain.
    /// @dev Tokens are minted on the destination chain.
    /// @param _token   Token to relay.
    /// @param _from    Address of the msg.sender of sendERC20 on the source chain.
    /// @param _to      Address to relay tokens to.
    /// @param _amount  Amount of tokens to relay.
    function relayERC20(address _token, address _from, address _to, uint256 _amount) external {
        if (msg.sender != MESSENGER) revert Unauthorized();

        (address crossDomainMessageSender, uint256 source) =
            IL2ToL2CrossDomainMessenger(MESSENGER).crossDomainMessageContext();
        if (crossDomainMessageSender != address(this)) revert InvalidCrossDomainSender();

        ISuperchainERC20(_token).crosschainMint(_to, _amount);

        emit RelayERC20(_token, _from, _to, _amount, source);
    }

    /// @notice Sends ETH to a target address on another chain.
    /// @param _to       Address to send ETH to.
    /// @param _amount   Amount of ETH to send.
    /// @param _chainId  Chain ID of the destination chain.
    /// @return msgHash_ Hash of the message sent.
    function sendETH(address _to, uint256 _amount, uint256 _chainId) external payable returns (bytes32 msgHash_) {
        if (_to == address(0)) revert ZeroAddress();

        ISuperchainWETH(payable(Predeploys.SUPERCHAIN_WETH)).deposit{ value: msg.value }();
        ISuperchainWETH(payable(Predeploys.SUPERCHAIN_WETH)).crosschainBurn(msg.sender, msg.value);

        bytes memory message = abi.encodeCall(this.relayETH, (msg.sender, _to, _amount));
        msgHash_ = IL2ToL2CrossDomainMessenger(MESSENGER).sendMessage(_chainId, address(this), message);

        emit SendERC20(Predeploys.SUPERCHAIN_WETH, msg.sender, _to, _amount, _chainId);
    }

    /// @notice Relays ETH received from another chain.
    /// @param _from    Address of the msg.sender of sendETH on the source chain.
    /// @param _to      Address to relay ETH to.
    /// @param _amount  Amount of ETH to relay.
    function relayETH(address _from, address _to, uint256 _amount) external {
        if (msg.sender != MESSENGER) revert Unauthorized();

        (address crossDomainMessageSender, uint256 source) =
            IL2ToL2CrossDomainMessenger(MESSENGER).crossDomainMessageContext();
        if (crossDomainMessageSender != address(this)) revert InvalidCrossDomainSender();

        ISuperchainWETH(payable(Predeploys.SUPERCHAIN_WETH)).crosschainMint(address(this), _amount);
        ISuperchainWETH(payable(Predeploys.SUPERCHAIN_WETH)).withdraw(_amount);

        new SafeSend{ value: _amount }(payable(_to));

        emit RelayERC20(Predeploys.SUPERCHAIN_WETH, _from, _to, _amount, source);
    }
}
