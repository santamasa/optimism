#!/bin/bash

set -eu

#Deployer: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
#Deployed to: 0x5FbDB2315678afecb367f032d93F642f64180aa3
#Transaction hash: 0x0c0f888d59aa00726f0b5ba93ec079910a0239a8fd0fe4b36fcf1a60b1abced1

# emit example data on chain A

export OP_INTEROP_DEVKEY_CHAINID=900200
export OP_INTEROP_DEVKEY_DOMAIN=user
export OP_INTEROP_DEVKEY_NAME=0
export OP_INTEROP_MNEMONIC="test test test test test test test test test test test junk"
export ETH_RPC_URL="http://yolo:9145" # chain A
export RAW_PRIVATE_KEY=0x$(go run ./op-node/cmd interop devkey secret)
export EMITTER_ADDRESS="0x5FbDB2315678afecb367f032d93F642f64180aa3"
cast send --private-key=$RAW_PRIVATE_KEY $EMITTER_ADDRESS 'emitData(bytes)' "$(cast from-utf8 'hello world')"


#blockHash               0xd8ff0bc5e1454113dc141d3bb223e36acd7b150717d545919e8af99e795c4aee
#blockNumber             4509
#contractAddress
#cumulativeGasUsed       115744
#effectiveGasPrice       253
#from                    0xa0eFcF89188eE2dd36113aeb03B7a68c82da9923
#gasUsed                 23273
#logs                    [{"address":"0x5fbdb2315678afecb367f032d93f642f64180aa3","topics":["0xe00bbfe6f6f8f1bbed2da38e3f5a139c6f9da594ab248a3cf8b44fc73627772c","0x47173285a8d7341e5e972fc677286384f802f8ef42a5ec5f03bbfa254cb01fad"],"data":"0x","blockHash":"0xd8ff0bc5e1454113dc141d3bb223e36acd7b150717d545919e8af99e795c4aee","blockNumber":"0x119d","transactionHash":"0x6ac4b1b292469a6347c956c2fd80ea531331c295491285fa5d6eb084d99e7638","transactionIndex":"0x2","logIndex":"0x0","removed":false}]
#logsBloom               0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000040000020000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000200000040000000000000000000000000000000000220000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000
#root
#status                  1 (success)
#transactionHash         0x6ac4b1b292469a6347c956c2fd80ea531331c295491285fa5d6eb084d99e7638
#transactionIndex        2
#type                    2
#blobGasPrice
#blobGasUsed
#authorizationList
#to                      0x5FbDB2315678afecb367f032d93F642f64180aa3
#l1BaseFeeScalar             "0x558"
#l1BlobBaseFee             "0x1"
#l1BlobBaseFeeScalar             "0xc5fc5"
#l1Fee             "0x60"
#l1GasPrice             "0x7"
#l1GasUsed             "0x640"

#
# struct Identifier {
#     address origin;
#     uint256 blockNumber;
#     uint256 logIndex;
#     uint256 timestamp;
#     uint256 chainId;
# }
# function executeMessage(Identifier calldata _id, address _target, bytes calldata _message) external payable;

cast block -f timestamp 4509


export OP_INTEROP_DEVKEY_CHAINID=900201
export OP_INTEROP_DEVKEY_DOMAIN=user
export OP_INTEROP_DEVKEY_NAME=0
export ETH_RPC_URL="http://yolo:9245" # chain B
export RAW_PRIVATE_KEY=0x$(go run ./op-node/cmd interop devkey secret)

export CROSS_L2_INBOX_ADDRESS="0x4200000000000000000000000000000000000022"

cast send --private-key=$RAW_PRIVATE_KEY $CROSS_L2_INBOX_ADDRESS \
  'executeMessage((address,uint256,uint256,uint256,uint256),address,bytes)' \
    '(0x5fbdb2315678afecb367f032d93f642f64180aa3, 4509, 0, 1730467171, 900200)' \
    '0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa' \
    "$(cast concat-hex "0xe00bbfe6f6f8f1bbed2da38e3f5a139c6f9da594ab248a3cf8b44fc73627772c" "0x47173285a8d7341e5e972fc677286384f802f8ef42a5ec5f03bbfa254cb01fad")"

