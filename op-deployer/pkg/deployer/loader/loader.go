package loader

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-service/sources/batching"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type ContractCaller interface {
	CallContract(context.Context, ethereum.CallMsg, *big.Int) ([]byte, error)
}

type ContractAPI struct {
	l1Client ContractCaller
	abi      *abi.ABI
	addr     common.Address
}

func NewContractAPI(l1Client ContractCaller, abi *abi.ABI, addr common.Address) *ContractAPI {
	return &ContractAPI{
		l1Client: l1Client,
		abi:      abi,
		addr:     addr,
	}
}

func (c *ContractAPI) GetAddress(ctx context.Context, method string, args ...interface{}) (common.Address, error) {
	result, err := CallContract(ctx, c.l1Client, c.abi, c.addr, method, args...)
	if err != nil {
		return common.Address{}, err
	}
	return result.GetAddress(0), nil
}

func (c *ContractAPI) GetBigInt(ctx context.Context, method string, args ...interface{}) (*big.Int, error) {
	result, err := CallContract(ctx, c.l1Client, c.abi, c.addr, method, args...)
	if err != nil {
		return nil, err
	}
	return result.GetBigInt(0), nil
}

func (c *ContractAPI) GetUint32(ctx context.Context, method string, args ...interface{}) (uint32, error) {
	result, err := CallContract(ctx, c.l1Client, c.abi, c.addr, method, args...)
	if err != nil {
		return 0, err
	}
	return result.GetUint32(0), nil
}

func (c *ContractAPI) GetUint64(ctx context.Context, method string, args ...interface{}) (uint64, error) {
	result, err := CallContract(ctx, c.l1Client, c.abi, c.addr, method, args...)
	if err != nil {
		return 0, err
	}
	return result.GetUint64(0), nil
}

func (c *ContractAPI) GetHash(ctx context.Context, method string, args ...interface{}) (common.Hash, error) {
	result, err := CallContract(ctx, c.l1Client, c.abi, c.addr, method, args...)
	if err != nil {
		return common.Hash{}, err
	}
	return result.GetHash(0), nil
}

func CallContract(ctx context.Context, l1Client ContractCaller, contractAbi *abi.ABI, to common.Address, method string, args ...interface{}) (*batching.CallResult, error) {
	call := batching.NewContractCall(contractAbi, to, method, args...)
	calldata, err := call.Pack()
	if err != nil {
		return nil, fmt.Errorf("failed to pack %s call: %w", method, err)
	}
	response, err := l1Client.CallContract(ctx, ethereum.CallMsg{To: &to, Data: calldata}, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call method %v: %w", method, err)
	}
	return call.Unpack(response)
}
