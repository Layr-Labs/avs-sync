// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contractContractsRegistry

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// ContractContractsRegistryMetaData contains all meta data concerning the ContractContractsRegistry contract.
var ContractContractsRegistryMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"contractCount\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"contractNames\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"contracts\",\"inputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"registerContract\",\"inputs\":[{\"name\":\"name\",\"type\":\"string\",\"internalType\":\"string\"},{\"name\":\"_contract\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"}]",
	Bin: "0x608060405234801561001057600080fd5b5061050e806100206000396000f3fe608060405234801561001057600080fd5b506004361061004c5760003560e01c80633ca6bb92146100515780637f3c2c281461007a5780638736381a1461008f5780638c5b8385146100a6575b600080fd5b61006461005f36600461029e565b6100f2565b60405161007191906102e7565b60405180910390f35b61008d6100883660046103bd565b61018c565b005b61009860025481565b604051908152602001610071565b6100da6100b436600461041b565b80516020818301810180516000825292820191909301209152546001600160a01b031681565b6040516001600160a01b039091168152602001610071565b6001602052600090815260409020805461010b90610458565b80601f016020809104026020016040519081016040528092919081815260200182805461013790610458565b80156101845780601f1061015957610100808354040283529160200191610184565b820191906000526020600020905b81548152906001019060200180831161016757829003601f168201915b505050505081565b8060008360405161019d9190610493565b908152604080516020928190038301902080546001600160a01b0319166001600160a01b0394909416939093179092556002546000908152600182529190912083516101eb92850190610205565b50600280549060006101fc836104af565b91905055505050565b82805461021190610458565b90600052602060002090601f0160209004810192826102335760008555610279565b82601f1061024c57805160ff1916838001178555610279565b82800160010185558215610279579182015b8281111561027957825182559160200191906001019061025e565b50610285929150610289565b5090565b5b80821115610285576000815560010161028a565b6000602082840312156102b057600080fd5b5035919050565b60005b838110156102d25781810151838201526020016102ba565b838111156102e1576000848401525b50505050565b60208152600082518060208401526103068160408501602087016102b7565b601f01601f19169190910160400192915050565b634e487b7160e01b600052604160045260246000fd5b600082601f83011261034157600080fd5b813567ffffffffffffffff8082111561035c5761035c61031a565b604051601f8301601f19908116603f011681019082821181831017156103845761038461031a565b8160405283815286602085880101111561039d57600080fd5b836020870160208301376000602085830101528094505050505092915050565b600080604083850312156103d057600080fd5b823567ffffffffffffffff8111156103e757600080fd5b6103f385828601610330565b92505060208301356001600160a01b038116811461041057600080fd5b809150509250929050565b60006020828403121561042d57600080fd5b813567ffffffffffffffff81111561044457600080fd5b61045084828501610330565b949350505050565b600181811c9082168061046c57607f821691505b6020821081141561048d57634e487b7160e01b600052602260045260246000fd5b50919050565b600082516104a58184602087016102b7565b9190910192915050565b60006000198214156104d157634e487b7160e01b600052601160045260246000fd5b506001019056fea26469706673582212206e9799c7a52c6ba7717de687c66ced1f21e72d3f41d6c13cc30dafa903d473b664736f6c634300080c0033",
}

// ContractContractsRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractContractsRegistryMetaData.ABI instead.
var ContractContractsRegistryABI = ContractContractsRegistryMetaData.ABI

// ContractContractsRegistryBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ContractContractsRegistryMetaData.Bin instead.
var ContractContractsRegistryBin = ContractContractsRegistryMetaData.Bin

// DeployContractContractsRegistry deploys a new Ethereum contract, binding an instance of ContractContractsRegistry to it.
func DeployContractContractsRegistry(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ContractContractsRegistry, error) {
	parsed, err := ContractContractsRegistryMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractContractsRegistryBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ContractContractsRegistry{ContractContractsRegistryCaller: ContractContractsRegistryCaller{contract: contract}, ContractContractsRegistryTransactor: ContractContractsRegistryTransactor{contract: contract}, ContractContractsRegistryFilterer: ContractContractsRegistryFilterer{contract: contract}}, nil
}

// ContractContractsRegistry is an auto generated Go binding around an Ethereum contract.
type ContractContractsRegistry struct {
	ContractContractsRegistryCaller     // Read-only binding to the contract
	ContractContractsRegistryTransactor // Write-only binding to the contract
	ContractContractsRegistryFilterer   // Log filterer for contract events
}

// ContractContractsRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractContractsRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractContractsRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractContractsRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractContractsRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractContractsRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractContractsRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractContractsRegistrySession struct {
	Contract     *ContractContractsRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts              // Call options to use throughout this session
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// ContractContractsRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractContractsRegistryCallerSession struct {
	Contract *ContractContractsRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                    // Call options to use throughout this session
}

// ContractContractsRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractContractsRegistryTransactorSession struct {
	Contract     *ContractContractsRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                    // Transaction auth options to use throughout this session
}

// ContractContractsRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractContractsRegistryRaw struct {
	Contract *ContractContractsRegistry // Generic contract binding to access the raw methods on
}

// ContractContractsRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractContractsRegistryCallerRaw struct {
	Contract *ContractContractsRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// ContractContractsRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractContractsRegistryTransactorRaw struct {
	Contract *ContractContractsRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContractContractsRegistry creates a new instance of ContractContractsRegistry, bound to a specific deployed contract.
func NewContractContractsRegistry(address common.Address, backend bind.ContractBackend) (*ContractContractsRegistry, error) {
	contract, err := bindContractContractsRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ContractContractsRegistry{ContractContractsRegistryCaller: ContractContractsRegistryCaller{contract: contract}, ContractContractsRegistryTransactor: ContractContractsRegistryTransactor{contract: contract}, ContractContractsRegistryFilterer: ContractContractsRegistryFilterer{contract: contract}}, nil
}

// NewContractContractsRegistryCaller creates a new read-only instance of ContractContractsRegistry, bound to a specific deployed contract.
func NewContractContractsRegistryCaller(address common.Address, caller bind.ContractCaller) (*ContractContractsRegistryCaller, error) {
	contract, err := bindContractContractsRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractContractsRegistryCaller{contract: contract}, nil
}

// NewContractContractsRegistryTransactor creates a new write-only instance of ContractContractsRegistry, bound to a specific deployed contract.
func NewContractContractsRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractContractsRegistryTransactor, error) {
	contract, err := bindContractContractsRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractContractsRegistryTransactor{contract: contract}, nil
}

// NewContractContractsRegistryFilterer creates a new log filterer instance of ContractContractsRegistry, bound to a specific deployed contract.
func NewContractContractsRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractContractsRegistryFilterer, error) {
	contract, err := bindContractContractsRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractContractsRegistryFilterer{contract: contract}, nil
}

// bindContractContractsRegistry binds a generic wrapper to an already deployed contract.
func bindContractContractsRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ContractContractsRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractContractsRegistry *ContractContractsRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractContractsRegistry.Contract.ContractContractsRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractContractsRegistry *ContractContractsRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractContractsRegistry.Contract.ContractContractsRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractContractsRegistry *ContractContractsRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractContractsRegistry.Contract.ContractContractsRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractContractsRegistry *ContractContractsRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractContractsRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractContractsRegistry *ContractContractsRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractContractsRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractContractsRegistry *ContractContractsRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractContractsRegistry.Contract.contract.Transact(opts, method, params...)
}

// ContractCount is a free data retrieval call binding the contract method 0x8736381a.
//
// Solidity: function contractCount() view returns(uint256)
func (_ContractContractsRegistry *ContractContractsRegistryCaller) ContractCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractContractsRegistry.contract.Call(opts, &out, "contractCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ContractCount is a free data retrieval call binding the contract method 0x8736381a.
//
// Solidity: function contractCount() view returns(uint256)
func (_ContractContractsRegistry *ContractContractsRegistrySession) ContractCount() (*big.Int, error) {
	return _ContractContractsRegistry.Contract.ContractCount(&_ContractContractsRegistry.CallOpts)
}

// ContractCount is a free data retrieval call binding the contract method 0x8736381a.
//
// Solidity: function contractCount() view returns(uint256)
func (_ContractContractsRegistry *ContractContractsRegistryCallerSession) ContractCount() (*big.Int, error) {
	return _ContractContractsRegistry.Contract.ContractCount(&_ContractContractsRegistry.CallOpts)
}

// ContractNames is a free data retrieval call binding the contract method 0x3ca6bb92.
//
// Solidity: function contractNames(uint256 ) view returns(string)
func (_ContractContractsRegistry *ContractContractsRegistryCaller) ContractNames(opts *bind.CallOpts, arg0 *big.Int) (string, error) {
	var out []interface{}
	err := _ContractContractsRegistry.contract.Call(opts, &out, "contractNames", arg0)

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// ContractNames is a free data retrieval call binding the contract method 0x3ca6bb92.
//
// Solidity: function contractNames(uint256 ) view returns(string)
func (_ContractContractsRegistry *ContractContractsRegistrySession) ContractNames(arg0 *big.Int) (string, error) {
	return _ContractContractsRegistry.Contract.ContractNames(&_ContractContractsRegistry.CallOpts, arg0)
}

// ContractNames is a free data retrieval call binding the contract method 0x3ca6bb92.
//
// Solidity: function contractNames(uint256 ) view returns(string)
func (_ContractContractsRegistry *ContractContractsRegistryCallerSession) ContractNames(arg0 *big.Int) (string, error) {
	return _ContractContractsRegistry.Contract.ContractNames(&_ContractContractsRegistry.CallOpts, arg0)
}

// Contracts is a free data retrieval call binding the contract method 0x8c5b8385.
//
// Solidity: function contracts(string ) view returns(address)
func (_ContractContractsRegistry *ContractContractsRegistryCaller) Contracts(opts *bind.CallOpts, arg0 string) (common.Address, error) {
	var out []interface{}
	err := _ContractContractsRegistry.contract.Call(opts, &out, "contracts", arg0)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Contracts is a free data retrieval call binding the contract method 0x8c5b8385.
//
// Solidity: function contracts(string ) view returns(address)
func (_ContractContractsRegistry *ContractContractsRegistrySession) Contracts(arg0 string) (common.Address, error) {
	return _ContractContractsRegistry.Contract.Contracts(&_ContractContractsRegistry.CallOpts, arg0)
}

// Contracts is a free data retrieval call binding the contract method 0x8c5b8385.
//
// Solidity: function contracts(string ) view returns(address)
func (_ContractContractsRegistry *ContractContractsRegistryCallerSession) Contracts(arg0 string) (common.Address, error) {
	return _ContractContractsRegistry.Contract.Contracts(&_ContractContractsRegistry.CallOpts, arg0)
}

// RegisterContract is a paid mutator transaction binding the contract method 0x7f3c2c28.
//
// Solidity: function registerContract(string name, address _contract) returns()
func (_ContractContractsRegistry *ContractContractsRegistryTransactor) RegisterContract(opts *bind.TransactOpts, name string, _contract common.Address) (*types.Transaction, error) {
	return _ContractContractsRegistry.contract.Transact(opts, "registerContract", name, _contract)
}

// RegisterContract is a paid mutator transaction binding the contract method 0x7f3c2c28.
//
// Solidity: function registerContract(string name, address _contract) returns()
func (_ContractContractsRegistry *ContractContractsRegistrySession) RegisterContract(name string, _contract common.Address) (*types.Transaction, error) {
	return _ContractContractsRegistry.Contract.RegisterContract(&_ContractContractsRegistry.TransactOpts, name, _contract)
}

// RegisterContract is a paid mutator transaction binding the contract method 0x7f3c2c28.
//
// Solidity: function registerContract(string name, address _contract) returns()
func (_ContractContractsRegistry *ContractContractsRegistryTransactorSession) RegisterContract(name string, _contract common.Address) (*types.Transaction, error) {
	return _ContractContractsRegistry.Contract.RegisterContract(&_ContractContractsRegistry.TransactOpts, name, _contract)
}
