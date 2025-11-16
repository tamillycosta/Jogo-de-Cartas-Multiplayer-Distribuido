// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package blockchain

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

// PackageRegistryPackage is an auto generated low-level Go binding around an user-defined struct.
type PackageRegistryPackage struct {
	Id        [32]byte
	CardIds   [][32]byte
	Opened    bool
	OpenedBy  [32]byte
	CreatedAt *big.Int
}

// PackageRegistryMetaData contains all meta data concerning the PackageRegistry contract.
var PackageRegistryMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"packageId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"bytes32[]\",\"name\":\"cardIds\",\"type\":\"bytes32[]\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"PackageCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"packageId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"playerId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"PackageOpened\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"packageIds\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"packages\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"bool\",\"name\":\"opened\",\"type\":\"bool\"},{\"internalType\":\"bytes32\",\"name\":\"openedBy\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_packageId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32[]\",\"name\":\"_cardIds\",\"type\":\"bytes32[]\"}],\"name\":\"createPackage\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_packageId\",\"type\":\"bytes32\"}],\"name\":\"packageExists\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"_packageId\",\"type\":\"bytes32\"}],\"name\":\"getPackage\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32[]\",\"name\":\"cardIds\",\"type\":\"bytes32[]\"},{\"internalType\":\"bool\",\"name\":\"opened\",\"type\":\"bool\"},{\"internalType\":\"bytes32\",\"name\":\"openedBy\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"}],\"internalType\":\"structPackageRegistry.Package\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[],\"name\":\"getTotalPackages\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getPackageByIndex\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32[]\",\"name\":\"cardIds\",\"type\":\"bytes32[]\"},{\"internalType\":\"bool\",\"name\":\"opened\",\"type\":\"bool\"},{\"internalType\":\"bytes32\",\"name\":\"openedBy\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"}],\"internalType\":\"structPackageRegistry.Package\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"count\",\"type\":\"uint256\"}],\"name\":\"getRecentPackages\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32[]\",\"name\":\"cardIds\",\"type\":\"bytes32[]\"},{\"internalType\":\"bool\",\"name\":\"opened\",\"type\":\"bool\"},{\"internalType\":\"bytes32\",\"name\":\"openedBy\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"}],\"internalType\":\"structPackageRegistry.Package[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\",\"constant\":true}]",
}

// PackageRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use PackageRegistryMetaData.ABI instead.
var PackageRegistryABI = PackageRegistryMetaData.ABI

// PackageRegistry is an auto generated Go binding around an Ethereum contract.
type PackageRegistry struct {
	PackageRegistryCaller     // Read-only binding to the contract
	PackageRegistryTransactor // Write-only binding to the contract
	PackageRegistryFilterer   // Log filterer for contract events
}

// PackageRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type PackageRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PackageRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type PackageRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PackageRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type PackageRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// PackageRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type PackageRegistrySession struct {
	Contract     *PackageRegistry  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// PackageRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type PackageRegistryCallerSession struct {
	Contract *PackageRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// PackageRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type PackageRegistryTransactorSession struct {
	Contract     *PackageRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// PackageRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type PackageRegistryRaw struct {
	Contract *PackageRegistry // Generic contract binding to access the raw methods on
}

// PackageRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type PackageRegistryCallerRaw struct {
	Contract *PackageRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// PackageRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type PackageRegistryTransactorRaw struct {
	Contract *PackageRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewPackageRegistry creates a new instance of PackageRegistry, bound to a specific deployed contract.
func NewPackageRegistry(address common.Address, backend bind.ContractBackend) (*PackageRegistry, error) {
	contract, err := bindPackageRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &PackageRegistry{PackageRegistryCaller: PackageRegistryCaller{contract: contract}, PackageRegistryTransactor: PackageRegistryTransactor{contract: contract}, PackageRegistryFilterer: PackageRegistryFilterer{contract: contract}}, nil
}

// NewPackageRegistryCaller creates a new read-only instance of PackageRegistry, bound to a specific deployed contract.
func NewPackageRegistryCaller(address common.Address, caller bind.ContractCaller) (*PackageRegistryCaller, error) {
	contract, err := bindPackageRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &PackageRegistryCaller{contract: contract}, nil
}

// NewPackageRegistryTransactor creates a new write-only instance of PackageRegistry, bound to a specific deployed contract.
func NewPackageRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*PackageRegistryTransactor, error) {
	contract, err := bindPackageRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &PackageRegistryTransactor{contract: contract}, nil
}

// NewPackageRegistryFilterer creates a new log filterer instance of PackageRegistry, bound to a specific deployed contract.
func NewPackageRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*PackageRegistryFilterer, error) {
	contract, err := bindPackageRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &PackageRegistryFilterer{contract: contract}, nil
}

// bindPackageRegistry binds a generic wrapper to an already deployed contract.
func bindPackageRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := PackageRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PackageRegistry *PackageRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PackageRegistry.Contract.PackageRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PackageRegistry *PackageRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PackageRegistry.Contract.PackageRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PackageRegistry *PackageRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PackageRegistry.Contract.PackageRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_PackageRegistry *PackageRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _PackageRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_PackageRegistry *PackageRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _PackageRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_PackageRegistry *PackageRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _PackageRegistry.Contract.contract.Transact(opts, method, params...)
}

// GetPackage is a free data retrieval call binding the contract method 0xa0f38f06.
//
// Solidity: function getPackage(bytes32 _packageId) view returns((bytes32,bytes32[],bool,bytes32,uint256))
func (_PackageRegistry *PackageRegistryCaller) GetPackage(opts *bind.CallOpts, _packageId [32]byte) (PackageRegistryPackage, error) {
	var out []interface{}
	err := _PackageRegistry.contract.Call(opts, &out, "getPackage", _packageId)

	if err != nil {
		return *new(PackageRegistryPackage), err
	}

	out0 := *abi.ConvertType(out[0], new(PackageRegistryPackage)).(*PackageRegistryPackage)

	return out0, err

}

// GetPackage is a free data retrieval call binding the contract method 0xa0f38f06.
//
// Solidity: function getPackage(bytes32 _packageId) view returns((bytes32,bytes32[],bool,bytes32,uint256))
func (_PackageRegistry *PackageRegistrySession) GetPackage(_packageId [32]byte) (PackageRegistryPackage, error) {
	return _PackageRegistry.Contract.GetPackage(&_PackageRegistry.CallOpts, _packageId)
}

// GetPackage is a free data retrieval call binding the contract method 0xa0f38f06.
//
// Solidity: function getPackage(bytes32 _packageId) view returns((bytes32,bytes32[],bool,bytes32,uint256))
func (_PackageRegistry *PackageRegistryCallerSession) GetPackage(_packageId [32]byte) (PackageRegistryPackage, error) {
	return _PackageRegistry.Contract.GetPackage(&_PackageRegistry.CallOpts, _packageId)
}

// GetPackageByIndex is a free data retrieval call binding the contract method 0x527f30bc.
//
// Solidity: function getPackageByIndex(uint256 index) view returns((bytes32,bytes32[],bool,bytes32,uint256))
func (_PackageRegistry *PackageRegistryCaller) GetPackageByIndex(opts *bind.CallOpts, index *big.Int) (PackageRegistryPackage, error) {
	var out []interface{}
	err := _PackageRegistry.contract.Call(opts, &out, "getPackageByIndex", index)

	if err != nil {
		return *new(PackageRegistryPackage), err
	}

	out0 := *abi.ConvertType(out[0], new(PackageRegistryPackage)).(*PackageRegistryPackage)

	return out0, err

}

// GetPackageByIndex is a free data retrieval call binding the contract method 0x527f30bc.
//
// Solidity: function getPackageByIndex(uint256 index) view returns((bytes32,bytes32[],bool,bytes32,uint256))
func (_PackageRegistry *PackageRegistrySession) GetPackageByIndex(index *big.Int) (PackageRegistryPackage, error) {
	return _PackageRegistry.Contract.GetPackageByIndex(&_PackageRegistry.CallOpts, index)
}

// GetPackageByIndex is a free data retrieval call binding the contract method 0x527f30bc.
//
// Solidity: function getPackageByIndex(uint256 index) view returns((bytes32,bytes32[],bool,bytes32,uint256))
func (_PackageRegistry *PackageRegistryCallerSession) GetPackageByIndex(index *big.Int) (PackageRegistryPackage, error) {
	return _PackageRegistry.Contract.GetPackageByIndex(&_PackageRegistry.CallOpts, index)
}

// GetRecentPackages is a free data retrieval call binding the contract method 0xbd6c4285.
//
// Solidity: function getRecentPackages(uint256 count) view returns((bytes32,bytes32[],bool,bytes32,uint256)[])
func (_PackageRegistry *PackageRegistryCaller) GetRecentPackages(opts *bind.CallOpts, count *big.Int) ([]PackageRegistryPackage, error) {
	var out []interface{}
	err := _PackageRegistry.contract.Call(opts, &out, "getRecentPackages", count)

	if err != nil {
		return *new([]PackageRegistryPackage), err
	}

	out0 := *abi.ConvertType(out[0], new([]PackageRegistryPackage)).(*[]PackageRegistryPackage)

	return out0, err

}

// GetRecentPackages is a free data retrieval call binding the contract method 0xbd6c4285.
//
// Solidity: function getRecentPackages(uint256 count) view returns((bytes32,bytes32[],bool,bytes32,uint256)[])
func (_PackageRegistry *PackageRegistrySession) GetRecentPackages(count *big.Int) ([]PackageRegistryPackage, error) {
	return _PackageRegistry.Contract.GetRecentPackages(&_PackageRegistry.CallOpts, count)
}

// GetRecentPackages is a free data retrieval call binding the contract method 0xbd6c4285.
//
// Solidity: function getRecentPackages(uint256 count) view returns((bytes32,bytes32[],bool,bytes32,uint256)[])
func (_PackageRegistry *PackageRegistryCallerSession) GetRecentPackages(count *big.Int) ([]PackageRegistryPackage, error) {
	return _PackageRegistry.Contract.GetRecentPackages(&_PackageRegistry.CallOpts, count)
}

// GetTotalPackages is a free data retrieval call binding the contract method 0x61cd6636.
//
// Solidity: function getTotalPackages() view returns(uint256)
func (_PackageRegistry *PackageRegistryCaller) GetTotalPackages(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _PackageRegistry.contract.Call(opts, &out, "getTotalPackages")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalPackages is a free data retrieval call binding the contract method 0x61cd6636.
//
// Solidity: function getTotalPackages() view returns(uint256)
func (_PackageRegistry *PackageRegistrySession) GetTotalPackages() (*big.Int, error) {
	return _PackageRegistry.Contract.GetTotalPackages(&_PackageRegistry.CallOpts)
}

// GetTotalPackages is a free data retrieval call binding the contract method 0x61cd6636.
//
// Solidity: function getTotalPackages() view returns(uint256)
func (_PackageRegistry *PackageRegistryCallerSession) GetTotalPackages() (*big.Int, error) {
	return _PackageRegistry.Contract.GetTotalPackages(&_PackageRegistry.CallOpts)
}

// PackageExists is a free data retrieval call binding the contract method 0xa9b35240.
//
// Solidity: function packageExists(bytes32 _packageId) view returns(bool)
func (_PackageRegistry *PackageRegistryCaller) PackageExists(opts *bind.CallOpts, _packageId [32]byte) (bool, error) {
	var out []interface{}
	err := _PackageRegistry.contract.Call(opts, &out, "packageExists", _packageId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// PackageExists is a free data retrieval call binding the contract method 0xa9b35240.
//
// Solidity: function packageExists(bytes32 _packageId) view returns(bool)
func (_PackageRegistry *PackageRegistrySession) PackageExists(_packageId [32]byte) (bool, error) {
	return _PackageRegistry.Contract.PackageExists(&_PackageRegistry.CallOpts, _packageId)
}

// PackageExists is a free data retrieval call binding the contract method 0xa9b35240.
//
// Solidity: function packageExists(bytes32 _packageId) view returns(bool)
func (_PackageRegistry *PackageRegistryCallerSession) PackageExists(_packageId [32]byte) (bool, error) {
	return _PackageRegistry.Contract.PackageExists(&_PackageRegistry.CallOpts, _packageId)
}

// PackageIds is a free data retrieval call binding the contract method 0xa20c8d52.
//
// Solidity: function packageIds(uint256 ) view returns(bytes32)
func (_PackageRegistry *PackageRegistryCaller) PackageIds(opts *bind.CallOpts, arg0 *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _PackageRegistry.contract.Call(opts, &out, "packageIds", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PackageIds is a free data retrieval call binding the contract method 0xa20c8d52.
//
// Solidity: function packageIds(uint256 ) view returns(bytes32)
func (_PackageRegistry *PackageRegistrySession) PackageIds(arg0 *big.Int) ([32]byte, error) {
	return _PackageRegistry.Contract.PackageIds(&_PackageRegistry.CallOpts, arg0)
}

// PackageIds is a free data retrieval call binding the contract method 0xa20c8d52.
//
// Solidity: function packageIds(uint256 ) view returns(bytes32)
func (_PackageRegistry *PackageRegistryCallerSession) PackageIds(arg0 *big.Int) ([32]byte, error) {
	return _PackageRegistry.Contract.PackageIds(&_PackageRegistry.CallOpts, arg0)
}

// Packages is a free data retrieval call binding the contract method 0x71102819.
//
// Solidity: function packages(bytes32 ) view returns(bytes32 id, bool opened, bytes32 openedBy, uint256 createdAt)
func (_PackageRegistry *PackageRegistryCaller) Packages(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Id        [32]byte
	Opened    bool
	OpenedBy  [32]byte
	CreatedAt *big.Int
}, error) {
	var out []interface{}
	err := _PackageRegistry.contract.Call(opts, &out, "packages", arg0)

	outstruct := new(struct {
		Id        [32]byte
		Opened    bool
		OpenedBy  [32]byte
		CreatedAt *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Id = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.Opened = *abi.ConvertType(out[1], new(bool)).(*bool)
	outstruct.OpenedBy = *abi.ConvertType(out[2], new([32]byte)).(*[32]byte)
	outstruct.CreatedAt = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Packages is a free data retrieval call binding the contract method 0x71102819.
//
// Solidity: function packages(bytes32 ) view returns(bytes32 id, bool opened, bytes32 openedBy, uint256 createdAt)
func (_PackageRegistry *PackageRegistrySession) Packages(arg0 [32]byte) (struct {
	Id        [32]byte
	Opened    bool
	OpenedBy  [32]byte
	CreatedAt *big.Int
}, error) {
	return _PackageRegistry.Contract.Packages(&_PackageRegistry.CallOpts, arg0)
}

// Packages is a free data retrieval call binding the contract method 0x71102819.
//
// Solidity: function packages(bytes32 ) view returns(bytes32 id, bool opened, bytes32 openedBy, uint256 createdAt)
func (_PackageRegistry *PackageRegistryCallerSession) Packages(arg0 [32]byte) (struct {
	Id        [32]byte
	Opened    bool
	OpenedBy  [32]byte
	CreatedAt *big.Int
}, error) {
	return _PackageRegistry.Contract.Packages(&_PackageRegistry.CallOpts, arg0)
}

// CreatePackage is a paid mutator transaction binding the contract method 0xb4f07ff8.
//
// Solidity: function createPackage(bytes32 _packageId, bytes32[] _cardIds) returns()
func (_PackageRegistry *PackageRegistryTransactor) CreatePackage(opts *bind.TransactOpts, _packageId [32]byte, _cardIds [][32]byte) (*types.Transaction, error) {
	return _PackageRegistry.contract.Transact(opts, "createPackage", _packageId, _cardIds)
}

// CreatePackage is a paid mutator transaction binding the contract method 0xb4f07ff8.
//
// Solidity: function createPackage(bytes32 _packageId, bytes32[] _cardIds) returns()
func (_PackageRegistry *PackageRegistrySession) CreatePackage(_packageId [32]byte, _cardIds [][32]byte) (*types.Transaction, error) {
	return _PackageRegistry.Contract.CreatePackage(&_PackageRegistry.TransactOpts, _packageId, _cardIds)
}

// CreatePackage is a paid mutator transaction binding the contract method 0xb4f07ff8.
//
// Solidity: function createPackage(bytes32 _packageId, bytes32[] _cardIds) returns()
func (_PackageRegistry *PackageRegistryTransactorSession) CreatePackage(_packageId [32]byte, _cardIds [][32]byte) (*types.Transaction, error) {
	return _PackageRegistry.Contract.CreatePackage(&_PackageRegistry.TransactOpts, _packageId, _cardIds)
}

// PackageRegistryPackageCreatedIterator is returned from FilterPackageCreated and is used to iterate over the raw logs and unpacked data for PackageCreated events raised by the PackageRegistry contract.
type PackageRegistryPackageCreatedIterator struct {
	Event *PackageRegistryPackageCreated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PackageRegistryPackageCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PackageRegistryPackageCreated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PackageRegistryPackageCreated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PackageRegistryPackageCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PackageRegistryPackageCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PackageRegistryPackageCreated represents a PackageCreated event raised by the PackageRegistry contract.
type PackageRegistryPackageCreated struct {
	PackageId [32]byte
	CardIds   [][32]byte
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterPackageCreated is a free log retrieval operation binding the contract event 0xfde0e0c080024f506c4a2d841a9d00029b84154a74fb3507475f7f2b00bf5276.
//
// Solidity: event PackageCreated(bytes32 indexed packageId, bytes32[] cardIds, uint256 timestamp)
func (_PackageRegistry *PackageRegistryFilterer) FilterPackageCreated(opts *bind.FilterOpts, packageId [][32]byte) (*PackageRegistryPackageCreatedIterator, error) {

	var packageIdRule []interface{}
	for _, packageIdItem := range packageId {
		packageIdRule = append(packageIdRule, packageIdItem)
	}

	logs, sub, err := _PackageRegistry.contract.FilterLogs(opts, "PackageCreated", packageIdRule)
	if err != nil {
		return nil, err
	}
	return &PackageRegistryPackageCreatedIterator{contract: _PackageRegistry.contract, event: "PackageCreated", logs: logs, sub: sub}, nil
}

// WatchPackageCreated is a free log subscription operation binding the contract event 0xfde0e0c080024f506c4a2d841a9d00029b84154a74fb3507475f7f2b00bf5276.
//
// Solidity: event PackageCreated(bytes32 indexed packageId, bytes32[] cardIds, uint256 timestamp)
func (_PackageRegistry *PackageRegistryFilterer) WatchPackageCreated(opts *bind.WatchOpts, sink chan<- *PackageRegistryPackageCreated, packageId [][32]byte) (event.Subscription, error) {

	var packageIdRule []interface{}
	for _, packageIdItem := range packageId {
		packageIdRule = append(packageIdRule, packageIdItem)
	}

	logs, sub, err := _PackageRegistry.contract.WatchLogs(opts, "PackageCreated", packageIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PackageRegistryPackageCreated)
				if err := _PackageRegistry.contract.UnpackLog(event, "PackageCreated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePackageCreated is a log parse operation binding the contract event 0xfde0e0c080024f506c4a2d841a9d00029b84154a74fb3507475f7f2b00bf5276.
//
// Solidity: event PackageCreated(bytes32 indexed packageId, bytes32[] cardIds, uint256 timestamp)
func (_PackageRegistry *PackageRegistryFilterer) ParsePackageCreated(log types.Log) (*PackageRegistryPackageCreated, error) {
	event := new(PackageRegistryPackageCreated)
	if err := _PackageRegistry.contract.UnpackLog(event, "PackageCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// PackageRegistryPackageOpenedIterator is returned from FilterPackageOpened and is used to iterate over the raw logs and unpacked data for PackageOpened events raised by the PackageRegistry contract.
type PackageRegistryPackageOpenedIterator struct {
	Event *PackageRegistryPackageOpened // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *PackageRegistryPackageOpenedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(PackageRegistryPackageOpened)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(PackageRegistryPackageOpened)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *PackageRegistryPackageOpenedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *PackageRegistryPackageOpenedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// PackageRegistryPackageOpened represents a PackageOpened event raised by the PackageRegistry contract.
type PackageRegistryPackageOpened struct {
	PackageId [32]byte
	PlayerId  [32]byte
	Timestamp *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterPackageOpened is a free log retrieval operation binding the contract event 0xf179ebd7ed2d4fa311fb743d08bb78d1b90705a72e92f19dc5a3ee33fda1bcf0.
//
// Solidity: event PackageOpened(bytes32 indexed packageId, bytes32 indexed playerId, uint256 timestamp)
func (_PackageRegistry *PackageRegistryFilterer) FilterPackageOpened(opts *bind.FilterOpts, packageId [][32]byte, playerId [][32]byte) (*PackageRegistryPackageOpenedIterator, error) {

	var packageIdRule []interface{}
	for _, packageIdItem := range packageId {
		packageIdRule = append(packageIdRule, packageIdItem)
	}
	var playerIdRule []interface{}
	for _, playerIdItem := range playerId {
		playerIdRule = append(playerIdRule, playerIdItem)
	}

	logs, sub, err := _PackageRegistry.contract.FilterLogs(opts, "PackageOpened", packageIdRule, playerIdRule)
	if err != nil {
		return nil, err
	}
	return &PackageRegistryPackageOpenedIterator{contract: _PackageRegistry.contract, event: "PackageOpened", logs: logs, sub: sub}, nil
}

// WatchPackageOpened is a free log subscription operation binding the contract event 0xf179ebd7ed2d4fa311fb743d08bb78d1b90705a72e92f19dc5a3ee33fda1bcf0.
//
// Solidity: event PackageOpened(bytes32 indexed packageId, bytes32 indexed playerId, uint256 timestamp)
func (_PackageRegistry *PackageRegistryFilterer) WatchPackageOpened(opts *bind.WatchOpts, sink chan<- *PackageRegistryPackageOpened, packageId [][32]byte, playerId [][32]byte) (event.Subscription, error) {

	var packageIdRule []interface{}
	for _, packageIdItem := range packageId {
		packageIdRule = append(packageIdRule, packageIdItem)
	}
	var playerIdRule []interface{}
	for _, playerIdItem := range playerId {
		playerIdRule = append(playerIdRule, playerIdItem)
	}

	logs, sub, err := _PackageRegistry.contract.WatchLogs(opts, "PackageOpened", packageIdRule, playerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(PackageRegistryPackageOpened)
				if err := _PackageRegistry.contract.UnpackLog(event, "PackageOpened", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePackageOpened is a log parse operation binding the contract event 0xf179ebd7ed2d4fa311fb743d08bb78d1b90705a72e92f19dc5a3ee33fda1bcf0.
//
// Solidity: event PackageOpened(bytes32 indexed packageId, bytes32 indexed playerId, uint256 timestamp)
func (_PackageRegistry *PackageRegistryFilterer) ParsePackageOpened(log types.Log) (*PackageRegistryPackageOpened, error) {
	event := new(PackageRegistryPackageOpened)
	if err := _PackageRegistry.contract.UnpackLog(event, "PackageOpened", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
