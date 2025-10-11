// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

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

// EnforcedOptionParam is an auto generated low-level Go binding around an user-defined struct.
type EnforcedOptionParam struct {
	Eid     uint32
	MsgType uint16
	Options []byte
}

// MessagingFee is an auto generated low-level Go binding around an user-defined struct.
type MessagingFee struct {
	NativeFee  *big.Int
	LzTokenFee *big.Int
}

// Origin is an auto generated low-level Go binding around an user-defined struct.
type Origin struct {
	SrcEid uint32
	Sender [32]byte
	Nonce  uint64
}

// MyOAppMetaData contains all meta data concerning the MyOApp contract.
var MyOAppMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_endpoint\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_owner\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"requested\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"available\",\"type\":\"uint256\"}],\"name\":\"InsufficientLiquidity\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"provided\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"required\",\"type\":\"uint256\"}],\"name\":\"InsufficientMsgValue\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidDelegate\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidEndpointCall\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"options\",\"type\":\"bytes\"}],\"name\":\"InvalidOptions\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"LzTokenUnavailable\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"eid\",\"type\":\"uint32\"}],\"name\":\"NoPeer\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"msgValue\",\"type\":\"uint256\"}],\"name\":\"NotEnoughNative\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"OnlyEndpoint\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"eid\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"sender\",\"type\":\"bytes32\"}],\"name\":\"OnlyPeer\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"}],\"name\":\"SafeERC20FailedOperation\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"components\":[{\"internalType\":\"uint32\",\"name\":\"eid\",\"type\":\"uint32\"},{\"internalType\":\"uint16\",\"name\":\"msgType\",\"type\":\"uint16\"},{\"internalType\":\"bytes\",\"name\":\"options\",\"type\":\"bytes\"}],\"indexed\":false,\"internalType\":\"structEnforcedOptionParam[]\",\"name\":\"_enforcedOptions\",\"type\":\"tuple[]\"}],\"name\":\"EnforcedOptionSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"eid\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"peer\",\"type\":\"bytes32\"}],\"name\":\"PeerSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"merchant\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"token\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"TokenPayoutExecuted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"uint32\",\"name\":\"dstEid\",\"type\":\"uint32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"payer\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"merchant\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"srcToken\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"dstToken\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"grossAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"netAmount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"feeAmount\",\"type\":\"uint256\"}],\"name\":\"TokenPayoutRequested\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"FEE_BPS\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"PAYOUT\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SEND\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"TAG_TOKEN_PAYOUT\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint32\",\"name\":\"srcEid\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"sender\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"}],\"internalType\":\"structOrigin\",\"name\":\"origin\",\"type\":\"tuple\"}],\"name\":\"allowInitializePath\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"_eid\",\"type\":\"uint32\"},{\"internalType\":\"uint16\",\"name\":\"_msgType\",\"type\":\"uint16\"},{\"internalType\":\"bytes\",\"name\":\"_extraOptions\",\"type\":\"bytes\"}],\"name\":\"combineOptions\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"name\":\"dstTokenByDstEidAndSrcToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"endpoint\",\"outputs\":[{\"internalType\":\"contractILayerZeroEndpointV2\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"eid\",\"type\":\"uint32\"},{\"internalType\":\"uint16\",\"name\":\"msgType\",\"type\":\"uint16\"}],\"name\":\"enforcedOptions\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"enforcedOption\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint32\",\"name\":\"srcEid\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"sender\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"}],\"internalType\":\"structOrigin\",\"name\":\"\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"_sender\",\"type\":\"address\"}],\"name\":\"isComposeMsgSender\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"lastMessage\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint32\",\"name\":\"srcEid\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"sender\",\"type\":\"bytes32\"},{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"}],\"internalType\":\"structOrigin\",\"name\":\"_origin\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"_guid\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"_message\",\"type\":\"bytes\"},{\"internalType\":\"address\",\"name\":\"_executor\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"_extraData\",\"type\":\"bytes\"}],\"name\":\"lzReceive\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"nextNonce\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"nonce\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"oAppVersion\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"senderVersion\",\"type\":\"uint64\"},{\"internalType\":\"uint64\",\"name\":\"receiverVersion\",\"type\":\"uint64\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"ownerDepositToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_to\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"}],\"name\":\"ownerWithdrawToken\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"eid\",\"type\":\"uint32\"}],\"name\":\"peers\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"peer\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"_dstEid\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"_srcToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_merchant\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_options\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"_payInLzToken\",\"type\":\"bool\"}],\"name\":\"quotePayoutToken\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"nativeFee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lzTokenFee\",\"type\":\"uint256\"}],\"internalType\":\"structMessagingFee\",\"name\":\"fee\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"_dstEid\",\"type\":\"uint32\"},{\"internalType\":\"string\",\"name\":\"_string\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"_options\",\"type\":\"bytes\"},{\"internalType\":\"bool\",\"name\":\"_payInLzToken\",\"type\":\"bool\"}],\"name\":\"quoteSendString\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"nativeFee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lzTokenFee\",\"type\":\"uint256\"}],\"internalType\":\"structMessagingFee\",\"name\":\"fee\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"_dstEid\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"_srcToken\",\"type\":\"address\"}],\"name\":\"removeTokenRoute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"_dstEid\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"_srcToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_merchant\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_amount\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_options\",\"type\":\"bytes\"}],\"name\":\"requestPayoutToken\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"_dstEid\",\"type\":\"uint32\"},{\"internalType\":\"string\",\"name\":\"_string\",\"type\":\"string\"},{\"internalType\":\"bytes\",\"name\":\"_options\",\"type\":\"bytes\"}],\"name\":\"sendString\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_delegate\",\"type\":\"address\"}],\"name\":\"setDelegate\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint32\",\"name\":\"eid\",\"type\":\"uint32\"},{\"internalType\":\"uint16\",\"name\":\"msgType\",\"type\":\"uint16\"},{\"internalType\":\"bytes\",\"name\":\"options\",\"type\":\"bytes\"}],\"internalType\":\"structEnforcedOptionParam[]\",\"name\":\"_enforcedOptions\",\"type\":\"tuple[]\"}],\"name\":\"setEnforcedOptions\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"_eid\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"_peer\",\"type\":\"bytes32\"}],\"name\":\"setPeer\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"_dstEid\",\"type\":\"uint32\"},{\"internalType\":\"address\",\"name\":\"_srcToken\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_dstToken\",\"type\":\"address\"}],\"name\":\"setTokenRoute\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// MyOAppABI is the input ABI used to generate the binding from.
// Deprecated: Use MyOAppMetaData.ABI instead.
var MyOAppABI = MyOAppMetaData.ABI

// MyOApp is an auto generated Go binding around an Ethereum contract.
type MyOApp struct {
	MyOAppCaller     // Read-only binding to the contract
	MyOAppTransactor // Write-only binding to the contract
	MyOAppFilterer   // Log filterer for contract events
}

// MyOAppCaller is an auto generated read-only Go binding around an Ethereum contract.
type MyOAppCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MyOAppTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MyOAppTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MyOAppFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MyOAppFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MyOAppSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MyOAppSession struct {
	Contract     *MyOApp           // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MyOAppCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MyOAppCallerSession struct {
	Contract *MyOAppCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts // Call options to use throughout this session
}

// MyOAppTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MyOAppTransactorSession struct {
	Contract     *MyOAppTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MyOAppRaw is an auto generated low-level Go binding around an Ethereum contract.
type MyOAppRaw struct {
	Contract *MyOApp // Generic contract binding to access the raw methods on
}

// MyOAppCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MyOAppCallerRaw struct {
	Contract *MyOAppCaller // Generic read-only contract binding to access the raw methods on
}

// MyOAppTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MyOAppTransactorRaw struct {
	Contract *MyOAppTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMyOApp creates a new instance of MyOApp, bound to a specific deployed contract.
func NewMyOApp(address common.Address, backend bind.ContractBackend) (*MyOApp, error) {
	contract, err := bindMyOApp(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MyOApp{MyOAppCaller: MyOAppCaller{contract: contract}, MyOAppTransactor: MyOAppTransactor{contract: contract}, MyOAppFilterer: MyOAppFilterer{contract: contract}}, nil
}

// NewMyOAppCaller creates a new read-only instance of MyOApp, bound to a specific deployed contract.
func NewMyOAppCaller(address common.Address, caller bind.ContractCaller) (*MyOAppCaller, error) {
	contract, err := bindMyOApp(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MyOAppCaller{contract: contract}, nil
}

// NewMyOAppTransactor creates a new write-only instance of MyOApp, bound to a specific deployed contract.
func NewMyOAppTransactor(address common.Address, transactor bind.ContractTransactor) (*MyOAppTransactor, error) {
	contract, err := bindMyOApp(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MyOAppTransactor{contract: contract}, nil
}

// NewMyOAppFilterer creates a new log filterer instance of MyOApp, bound to a specific deployed contract.
func NewMyOAppFilterer(address common.Address, filterer bind.ContractFilterer) (*MyOAppFilterer, error) {
	contract, err := bindMyOApp(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MyOAppFilterer{contract: contract}, nil
}

// bindMyOApp binds a generic wrapper to an already deployed contract.
func bindMyOApp(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := MyOAppMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MyOApp *MyOAppRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MyOApp.Contract.MyOAppCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MyOApp *MyOAppRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MyOApp.Contract.MyOAppTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MyOApp *MyOAppRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MyOApp.Contract.MyOAppTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MyOApp *MyOAppCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MyOApp.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MyOApp *MyOAppTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MyOApp.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MyOApp *MyOAppTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MyOApp.Contract.contract.Transact(opts, method, params...)
}

// FEEBPS is a free data retrieval call binding the contract method 0xbf333f2c.
//
// Solidity: function FEE_BPS() view returns(uint16)
func (_MyOApp *MyOAppCaller) FEEBPS(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _MyOApp.contract.Call(opts, &out, "FEE_BPS")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// FEEBPS is a free data retrieval call binding the contract method 0xbf333f2c.
//
// Solidity: function FEE_BPS() view returns(uint16)
func (_MyOApp *MyOAppSession) FEEBPS() (uint16, error) {
	return _MyOApp.Contract.FEEBPS(&_MyOApp.CallOpts)
}

// FEEBPS is a free data retrieval call binding the contract method 0xbf333f2c.
//
// Solidity: function FEE_BPS() view returns(uint16)
func (_MyOApp *MyOAppCallerSession) FEEBPS() (uint16, error) {
	return _MyOApp.Contract.FEEBPS(&_MyOApp.CallOpts)
}

// PAYOUT is a free data retrieval call binding the contract method 0x32da3d10.
//
// Solidity: function PAYOUT() view returns(uint16)
func (_MyOApp *MyOAppCaller) PAYOUT(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _MyOApp.contract.Call(opts, &out, "PAYOUT")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// PAYOUT is a free data retrieval call binding the contract method 0x32da3d10.
//
// Solidity: function PAYOUT() view returns(uint16)
func (_MyOApp *MyOAppSession) PAYOUT() (uint16, error) {
	return _MyOApp.Contract.PAYOUT(&_MyOApp.CallOpts)
}

// PAYOUT is a free data retrieval call binding the contract method 0x32da3d10.
//
// Solidity: function PAYOUT() view returns(uint16)
func (_MyOApp *MyOAppCallerSession) PAYOUT() (uint16, error) {
	return _MyOApp.Contract.PAYOUT(&_MyOApp.CallOpts)
}

// SEND is a free data retrieval call binding the contract method 0x1f5e1334.
//
// Solidity: function SEND() view returns(uint16)
func (_MyOApp *MyOAppCaller) SEND(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _MyOApp.contract.Call(opts, &out, "SEND")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// SEND is a free data retrieval call binding the contract method 0x1f5e1334.
//
// Solidity: function SEND() view returns(uint16)
func (_MyOApp *MyOAppSession) SEND() (uint16, error) {
	return _MyOApp.Contract.SEND(&_MyOApp.CallOpts)
}

// SEND is a free data retrieval call binding the contract method 0x1f5e1334.
//
// Solidity: function SEND() view returns(uint16)
func (_MyOApp *MyOAppCallerSession) SEND() (uint16, error) {
	return _MyOApp.Contract.SEND(&_MyOApp.CallOpts)
}

// TAGTOKENPAYOUT is a free data retrieval call binding the contract method 0x38871250.
//
// Solidity: function TAG_TOKEN_PAYOUT() view returns(uint8)
func (_MyOApp *MyOAppCaller) TAGTOKENPAYOUT(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _MyOApp.contract.Call(opts, &out, "TAG_TOKEN_PAYOUT")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// TAGTOKENPAYOUT is a free data retrieval call binding the contract method 0x38871250.
//
// Solidity: function TAG_TOKEN_PAYOUT() view returns(uint8)
func (_MyOApp *MyOAppSession) TAGTOKENPAYOUT() (uint8, error) {
	return _MyOApp.Contract.TAGTOKENPAYOUT(&_MyOApp.CallOpts)
}

// TAGTOKENPAYOUT is a free data retrieval call binding the contract method 0x38871250.
//
// Solidity: function TAG_TOKEN_PAYOUT() view returns(uint8)
func (_MyOApp *MyOAppCallerSession) TAGTOKENPAYOUT() (uint8, error) {
	return _MyOApp.Contract.TAGTOKENPAYOUT(&_MyOApp.CallOpts)
}

// AllowInitializePath is a free data retrieval call binding the contract method 0xff7bd03d.
//
// Solidity: function allowInitializePath((uint32,bytes32,uint64) origin) view returns(bool)
func (_MyOApp *MyOAppCaller) AllowInitializePath(opts *bind.CallOpts, origin Origin) (bool, error) {
	var out []interface{}
	err := _MyOApp.contract.Call(opts, &out, "allowInitializePath", origin)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// AllowInitializePath is a free data retrieval call binding the contract method 0xff7bd03d.
//
// Solidity: function allowInitializePath((uint32,bytes32,uint64) origin) view returns(bool)
func (_MyOApp *MyOAppSession) AllowInitializePath(origin Origin) (bool, error) {
	return _MyOApp.Contract.AllowInitializePath(&_MyOApp.CallOpts, origin)
}

// AllowInitializePath is a free data retrieval call binding the contract method 0xff7bd03d.
//
// Solidity: function allowInitializePath((uint32,bytes32,uint64) origin) view returns(bool)
func (_MyOApp *MyOAppCallerSession) AllowInitializePath(origin Origin) (bool, error) {
	return _MyOApp.Contract.AllowInitializePath(&_MyOApp.CallOpts, origin)
}

// CombineOptions is a free data retrieval call binding the contract method 0xbc70b354.
//
// Solidity: function combineOptions(uint32 _eid, uint16 _msgType, bytes _extraOptions) view returns(bytes)
func (_MyOApp *MyOAppCaller) CombineOptions(opts *bind.CallOpts, _eid uint32, _msgType uint16, _extraOptions []byte) ([]byte, error) {
	var out []interface{}
	err := _MyOApp.contract.Call(opts, &out, "combineOptions", _eid, _msgType, _extraOptions)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// CombineOptions is a free data retrieval call binding the contract method 0xbc70b354.
//
// Solidity: function combineOptions(uint32 _eid, uint16 _msgType, bytes _extraOptions) view returns(bytes)
func (_MyOApp *MyOAppSession) CombineOptions(_eid uint32, _msgType uint16, _extraOptions []byte) ([]byte, error) {
	return _MyOApp.Contract.CombineOptions(&_MyOApp.CallOpts, _eid, _msgType, _extraOptions)
}

// CombineOptions is a free data retrieval call binding the contract method 0xbc70b354.
//
// Solidity: function combineOptions(uint32 _eid, uint16 _msgType, bytes _extraOptions) view returns(bytes)
func (_MyOApp *MyOAppCallerSession) CombineOptions(_eid uint32, _msgType uint16, _extraOptions []byte) ([]byte, error) {
	return _MyOApp.Contract.CombineOptions(&_MyOApp.CallOpts, _eid, _msgType, _extraOptions)
}

// DstTokenByDstEidAndSrcToken is a free data retrieval call binding the contract method 0x03dcb61e.
//
// Solidity: function dstTokenByDstEidAndSrcToken(uint32 , address ) view returns(address)
func (_MyOApp *MyOAppCaller) DstTokenByDstEidAndSrcToken(opts *bind.CallOpts, arg0 uint32, arg1 common.Address) (common.Address, error) {
	var out []interface{}
	err := _MyOApp.contract.Call(opts, &out, "dstTokenByDstEidAndSrcToken", arg0, arg1)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// DstTokenByDstEidAndSrcToken is a free data retrieval call binding the contract method 0x03dcb61e.
//
// Solidity: function dstTokenByDstEidAndSrcToken(uint32 , address ) view returns(address)
func (_MyOApp *MyOAppSession) DstTokenByDstEidAndSrcToken(arg0 uint32, arg1 common.Address) (common.Address, error) {
	return _MyOApp.Contract.DstTokenByDstEidAndSrcToken(&_MyOApp.CallOpts, arg0, arg1)
}

// DstTokenByDstEidAndSrcToken is a free data retrieval call binding the contract method 0x03dcb61e.
//
// Solidity: function dstTokenByDstEidAndSrcToken(uint32 , address ) view returns(address)
func (_MyOApp *MyOAppCallerSession) DstTokenByDstEidAndSrcToken(arg0 uint32, arg1 common.Address) (common.Address, error) {
	return _MyOApp.Contract.DstTokenByDstEidAndSrcToken(&_MyOApp.CallOpts, arg0, arg1)
}

// Endpoint is a free data retrieval call binding the contract method 0x5e280f11.
//
// Solidity: function endpoint() view returns(address)
func (_MyOApp *MyOAppCaller) Endpoint(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _MyOApp.contract.Call(opts, &out, "endpoint")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Endpoint is a free data retrieval call binding the contract method 0x5e280f11.
//
// Solidity: function endpoint() view returns(address)
func (_MyOApp *MyOAppSession) Endpoint() (common.Address, error) {
	return _MyOApp.Contract.Endpoint(&_MyOApp.CallOpts)
}

// Endpoint is a free data retrieval call binding the contract method 0x5e280f11.
//
// Solidity: function endpoint() view returns(address)
func (_MyOApp *MyOAppCallerSession) Endpoint() (common.Address, error) {
	return _MyOApp.Contract.Endpoint(&_MyOApp.CallOpts)
}

// EnforcedOptions is a free data retrieval call binding the contract method 0x5535d461.
//
// Solidity: function enforcedOptions(uint32 eid, uint16 msgType) view returns(bytes enforcedOption)
func (_MyOApp *MyOAppCaller) EnforcedOptions(opts *bind.CallOpts, eid uint32, msgType uint16) ([]byte, error) {
	var out []interface{}
	err := _MyOApp.contract.Call(opts, &out, "enforcedOptions", eid, msgType)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// EnforcedOptions is a free data retrieval call binding the contract method 0x5535d461.
//
// Solidity: function enforcedOptions(uint32 eid, uint16 msgType) view returns(bytes enforcedOption)
func (_MyOApp *MyOAppSession) EnforcedOptions(eid uint32, msgType uint16) ([]byte, error) {
	return _MyOApp.Contract.EnforcedOptions(&_MyOApp.CallOpts, eid, msgType)
}

// EnforcedOptions is a free data retrieval call binding the contract method 0x5535d461.
//
// Solidity: function enforcedOptions(uint32 eid, uint16 msgType) view returns(bytes enforcedOption)
func (_MyOApp *MyOAppCallerSession) EnforcedOptions(eid uint32, msgType uint16) ([]byte, error) {
	return _MyOApp.Contract.EnforcedOptions(&_MyOApp.CallOpts, eid, msgType)
}

// IsComposeMsgSender is a free data retrieval call binding the contract method 0x82413eac.
//
// Solidity: function isComposeMsgSender((uint32,bytes32,uint64) , bytes , address _sender) view returns(bool)
func (_MyOApp *MyOAppCaller) IsComposeMsgSender(opts *bind.CallOpts, arg0 Origin, arg1 []byte, _sender common.Address) (bool, error) {
	var out []interface{}
	err := _MyOApp.contract.Call(opts, &out, "isComposeMsgSender", arg0, arg1, _sender)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsComposeMsgSender is a free data retrieval call binding the contract method 0x82413eac.
//
// Solidity: function isComposeMsgSender((uint32,bytes32,uint64) , bytes , address _sender) view returns(bool)
func (_MyOApp *MyOAppSession) IsComposeMsgSender(arg0 Origin, arg1 []byte, _sender common.Address) (bool, error) {
	return _MyOApp.Contract.IsComposeMsgSender(&_MyOApp.CallOpts, arg0, arg1, _sender)
}

// IsComposeMsgSender is a free data retrieval call binding the contract method 0x82413eac.
//
// Solidity: function isComposeMsgSender((uint32,bytes32,uint64) , bytes , address _sender) view returns(bool)
func (_MyOApp *MyOAppCallerSession) IsComposeMsgSender(arg0 Origin, arg1 []byte, _sender common.Address) (bool, error) {
	return _MyOApp.Contract.IsComposeMsgSender(&_MyOApp.CallOpts, arg0, arg1, _sender)
}

// LastMessage is a free data retrieval call binding the contract method 0x32970710.
//
// Solidity: function lastMessage() view returns(string)
func (_MyOApp *MyOAppCaller) LastMessage(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _MyOApp.contract.Call(opts, &out, "lastMessage")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// LastMessage is a free data retrieval call binding the contract method 0x32970710.
//
// Solidity: function lastMessage() view returns(string)
func (_MyOApp *MyOAppSession) LastMessage() (string, error) {
	return _MyOApp.Contract.LastMessage(&_MyOApp.CallOpts)
}

// LastMessage is a free data retrieval call binding the contract method 0x32970710.
//
// Solidity: function lastMessage() view returns(string)
func (_MyOApp *MyOAppCallerSession) LastMessage() (string, error) {
	return _MyOApp.Contract.LastMessage(&_MyOApp.CallOpts)
}

// NextNonce is a free data retrieval call binding the contract method 0x7d25a05e.
//
// Solidity: function nextNonce(uint32 , bytes32 ) view returns(uint64 nonce)
func (_MyOApp *MyOAppCaller) NextNonce(opts *bind.CallOpts, arg0 uint32, arg1 [32]byte) (uint64, error) {
	var out []interface{}
	err := _MyOApp.contract.Call(opts, &out, "nextNonce", arg0, arg1)

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// NextNonce is a free data retrieval call binding the contract method 0x7d25a05e.
//
// Solidity: function nextNonce(uint32 , bytes32 ) view returns(uint64 nonce)
func (_MyOApp *MyOAppSession) NextNonce(arg0 uint32, arg1 [32]byte) (uint64, error) {
	return _MyOApp.Contract.NextNonce(&_MyOApp.CallOpts, arg0, arg1)
}

// NextNonce is a free data retrieval call binding the contract method 0x7d25a05e.
//
// Solidity: function nextNonce(uint32 , bytes32 ) view returns(uint64 nonce)
func (_MyOApp *MyOAppCallerSession) NextNonce(arg0 uint32, arg1 [32]byte) (uint64, error) {
	return _MyOApp.Contract.NextNonce(&_MyOApp.CallOpts, arg0, arg1)
}

// OAppVersion is a free data retrieval call binding the contract method 0x17442b70.
//
// Solidity: function oAppVersion() pure returns(uint64 senderVersion, uint64 receiverVersion)
func (_MyOApp *MyOAppCaller) OAppVersion(opts *bind.CallOpts) (struct {
	SenderVersion   uint64
	ReceiverVersion uint64
}, error) {
	var out []interface{}
	err := _MyOApp.contract.Call(opts, &out, "oAppVersion")

	outstruct := new(struct {
		SenderVersion   uint64
		ReceiverVersion uint64
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.SenderVersion = *abi.ConvertType(out[0], new(uint64)).(*uint64)
	outstruct.ReceiverVersion = *abi.ConvertType(out[1], new(uint64)).(*uint64)

	return *outstruct, err

}

// OAppVersion is a free data retrieval call binding the contract method 0x17442b70.
//
// Solidity: function oAppVersion() pure returns(uint64 senderVersion, uint64 receiverVersion)
func (_MyOApp *MyOAppSession) OAppVersion() (struct {
	SenderVersion   uint64
	ReceiverVersion uint64
}, error) {
	return _MyOApp.Contract.OAppVersion(&_MyOApp.CallOpts)
}

// OAppVersion is a free data retrieval call binding the contract method 0x17442b70.
//
// Solidity: function oAppVersion() pure returns(uint64 senderVersion, uint64 receiverVersion)
func (_MyOApp *MyOAppCallerSession) OAppVersion() (struct {
	SenderVersion   uint64
	ReceiverVersion uint64
}, error) {
	return _MyOApp.Contract.OAppVersion(&_MyOApp.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_MyOApp *MyOAppCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _MyOApp.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_MyOApp *MyOAppSession) Owner() (common.Address, error) {
	return _MyOApp.Contract.Owner(&_MyOApp.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_MyOApp *MyOAppCallerSession) Owner() (common.Address, error) {
	return _MyOApp.Contract.Owner(&_MyOApp.CallOpts)
}

// Peers is a free data retrieval call binding the contract method 0xbb0b6a53.
//
// Solidity: function peers(uint32 eid) view returns(bytes32 peer)
func (_MyOApp *MyOAppCaller) Peers(opts *bind.CallOpts, eid uint32) ([32]byte, error) {
	var out []interface{}
	err := _MyOApp.contract.Call(opts, &out, "peers", eid)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// Peers is a free data retrieval call binding the contract method 0xbb0b6a53.
//
// Solidity: function peers(uint32 eid) view returns(bytes32 peer)
func (_MyOApp *MyOAppSession) Peers(eid uint32) ([32]byte, error) {
	return _MyOApp.Contract.Peers(&_MyOApp.CallOpts, eid)
}

// Peers is a free data retrieval call binding the contract method 0xbb0b6a53.
//
// Solidity: function peers(uint32 eid) view returns(bytes32 peer)
func (_MyOApp *MyOAppCallerSession) Peers(eid uint32) ([32]byte, error) {
	return _MyOApp.Contract.Peers(&_MyOApp.CallOpts, eid)
}

// QuotePayoutToken is a free data retrieval call binding the contract method 0x3e67ffac.
//
// Solidity: function quotePayoutToken(uint32 _dstEid, address _srcToken, address _merchant, uint256 _amount, bytes _options, bool _payInLzToken) view returns((uint256,uint256) fee)
func (_MyOApp *MyOAppCaller) QuotePayoutToken(opts *bind.CallOpts, _dstEid uint32, _srcToken common.Address, _merchant common.Address, _amount *big.Int, _options []byte, _payInLzToken bool) (MessagingFee, error) {
	var out []interface{}
	err := _MyOApp.contract.Call(opts, &out, "quotePayoutToken", _dstEid, _srcToken, _merchant, _amount, _options, _payInLzToken)

	if err != nil {
		return *new(MessagingFee), err
	}

	out0 := *abi.ConvertType(out[0], new(MessagingFee)).(*MessagingFee)

	return out0, err

}

// QuotePayoutToken is a free data retrieval call binding the contract method 0x3e67ffac.
//
// Solidity: function quotePayoutToken(uint32 _dstEid, address _srcToken, address _merchant, uint256 _amount, bytes _options, bool _payInLzToken) view returns((uint256,uint256) fee)
func (_MyOApp *MyOAppSession) QuotePayoutToken(_dstEid uint32, _srcToken common.Address, _merchant common.Address, _amount *big.Int, _options []byte, _payInLzToken bool) (MessagingFee, error) {
	return _MyOApp.Contract.QuotePayoutToken(&_MyOApp.CallOpts, _dstEid, _srcToken, _merchant, _amount, _options, _payInLzToken)
}

// QuotePayoutToken is a free data retrieval call binding the contract method 0x3e67ffac.
//
// Solidity: function quotePayoutToken(uint32 _dstEid, address _srcToken, address _merchant, uint256 _amount, bytes _options, bool _payInLzToken) view returns((uint256,uint256) fee)
func (_MyOApp *MyOAppCallerSession) QuotePayoutToken(_dstEid uint32, _srcToken common.Address, _merchant common.Address, _amount *big.Int, _options []byte, _payInLzToken bool) (MessagingFee, error) {
	return _MyOApp.Contract.QuotePayoutToken(&_MyOApp.CallOpts, _dstEid, _srcToken, _merchant, _amount, _options, _payInLzToken)
}

// QuoteSendString is a free data retrieval call binding the contract method 0xd3b4866e.
//
// Solidity: function quoteSendString(uint32 _dstEid, string _string, bytes _options, bool _payInLzToken) view returns((uint256,uint256) fee)
func (_MyOApp *MyOAppCaller) QuoteSendString(opts *bind.CallOpts, _dstEid uint32, _string string, _options []byte, _payInLzToken bool) (MessagingFee, error) {
	var out []interface{}
	err := _MyOApp.contract.Call(opts, &out, "quoteSendString", _dstEid, _string, _options, _payInLzToken)

	if err != nil {
		return *new(MessagingFee), err
	}

	out0 := *abi.ConvertType(out[0], new(MessagingFee)).(*MessagingFee)

	return out0, err

}

// QuoteSendString is a free data retrieval call binding the contract method 0xd3b4866e.
//
// Solidity: function quoteSendString(uint32 _dstEid, string _string, bytes _options, bool _payInLzToken) view returns((uint256,uint256) fee)
func (_MyOApp *MyOAppSession) QuoteSendString(_dstEid uint32, _string string, _options []byte, _payInLzToken bool) (MessagingFee, error) {
	return _MyOApp.Contract.QuoteSendString(&_MyOApp.CallOpts, _dstEid, _string, _options, _payInLzToken)
}

// QuoteSendString is a free data retrieval call binding the contract method 0xd3b4866e.
//
// Solidity: function quoteSendString(uint32 _dstEid, string _string, bytes _options, bool _payInLzToken) view returns((uint256,uint256) fee)
func (_MyOApp *MyOAppCallerSession) QuoteSendString(_dstEid uint32, _string string, _options []byte, _payInLzToken bool) (MessagingFee, error) {
	return _MyOApp.Contract.QuoteSendString(&_MyOApp.CallOpts, _dstEid, _string, _options, _payInLzToken)
}

// LzReceive is a paid mutator transaction binding the contract method 0x13137d65.
//
// Solidity: function lzReceive((uint32,bytes32,uint64) _origin, bytes32 _guid, bytes _message, address _executor, bytes _extraData) payable returns()
func (_MyOApp *MyOAppTransactor) LzReceive(opts *bind.TransactOpts, _origin Origin, _guid [32]byte, _message []byte, _executor common.Address, _extraData []byte) (*types.Transaction, error) {
	return _MyOApp.contract.Transact(opts, "lzReceive", _origin, _guid, _message, _executor, _extraData)
}

// LzReceive is a paid mutator transaction binding the contract method 0x13137d65.
//
// Solidity: function lzReceive((uint32,bytes32,uint64) _origin, bytes32 _guid, bytes _message, address _executor, bytes _extraData) payable returns()
func (_MyOApp *MyOAppSession) LzReceive(_origin Origin, _guid [32]byte, _message []byte, _executor common.Address, _extraData []byte) (*types.Transaction, error) {
	return _MyOApp.Contract.LzReceive(&_MyOApp.TransactOpts, _origin, _guid, _message, _executor, _extraData)
}

// LzReceive is a paid mutator transaction binding the contract method 0x13137d65.
//
// Solidity: function lzReceive((uint32,bytes32,uint64) _origin, bytes32 _guid, bytes _message, address _executor, bytes _extraData) payable returns()
func (_MyOApp *MyOAppTransactorSession) LzReceive(_origin Origin, _guid [32]byte, _message []byte, _executor common.Address, _extraData []byte) (*types.Transaction, error) {
	return _MyOApp.Contract.LzReceive(&_MyOApp.TransactOpts, _origin, _guid, _message, _executor, _extraData)
}

// OwnerDepositToken is a paid mutator transaction binding the contract method 0x837c4020.
//
// Solidity: function ownerDepositToken(address _token, uint256 _amount) returns()
func (_MyOApp *MyOAppTransactor) OwnerDepositToken(opts *bind.TransactOpts, _token common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MyOApp.contract.Transact(opts, "ownerDepositToken", _token, _amount)
}

// OwnerDepositToken is a paid mutator transaction binding the contract method 0x837c4020.
//
// Solidity: function ownerDepositToken(address _token, uint256 _amount) returns()
func (_MyOApp *MyOAppSession) OwnerDepositToken(_token common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MyOApp.Contract.OwnerDepositToken(&_MyOApp.TransactOpts, _token, _amount)
}

// OwnerDepositToken is a paid mutator transaction binding the contract method 0x837c4020.
//
// Solidity: function ownerDepositToken(address _token, uint256 _amount) returns()
func (_MyOApp *MyOAppTransactorSession) OwnerDepositToken(_token common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MyOApp.Contract.OwnerDepositToken(&_MyOApp.TransactOpts, _token, _amount)
}

// OwnerWithdrawToken is a paid mutator transaction binding the contract method 0xf585b64f.
//
// Solidity: function ownerWithdrawToken(address _token, address _to, uint256 _amount) returns()
func (_MyOApp *MyOAppTransactor) OwnerWithdrawToken(opts *bind.TransactOpts, _token common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MyOApp.contract.Transact(opts, "ownerWithdrawToken", _token, _to, _amount)
}

// OwnerWithdrawToken is a paid mutator transaction binding the contract method 0xf585b64f.
//
// Solidity: function ownerWithdrawToken(address _token, address _to, uint256 _amount) returns()
func (_MyOApp *MyOAppSession) OwnerWithdrawToken(_token common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MyOApp.Contract.OwnerWithdrawToken(&_MyOApp.TransactOpts, _token, _to, _amount)
}

// OwnerWithdrawToken is a paid mutator transaction binding the contract method 0xf585b64f.
//
// Solidity: function ownerWithdrawToken(address _token, address _to, uint256 _amount) returns()
func (_MyOApp *MyOAppTransactorSession) OwnerWithdrawToken(_token common.Address, _to common.Address, _amount *big.Int) (*types.Transaction, error) {
	return _MyOApp.Contract.OwnerWithdrawToken(&_MyOApp.TransactOpts, _token, _to, _amount)
}

// RemoveTokenRoute is a paid mutator transaction binding the contract method 0xabd917a2.
//
// Solidity: function removeTokenRoute(uint32 _dstEid, address _srcToken) returns()
func (_MyOApp *MyOAppTransactor) RemoveTokenRoute(opts *bind.TransactOpts, _dstEid uint32, _srcToken common.Address) (*types.Transaction, error) {
	return _MyOApp.contract.Transact(opts, "removeTokenRoute", _dstEid, _srcToken)
}

// RemoveTokenRoute is a paid mutator transaction binding the contract method 0xabd917a2.
//
// Solidity: function removeTokenRoute(uint32 _dstEid, address _srcToken) returns()
func (_MyOApp *MyOAppSession) RemoveTokenRoute(_dstEid uint32, _srcToken common.Address) (*types.Transaction, error) {
	return _MyOApp.Contract.RemoveTokenRoute(&_MyOApp.TransactOpts, _dstEid, _srcToken)
}

// RemoveTokenRoute is a paid mutator transaction binding the contract method 0xabd917a2.
//
// Solidity: function removeTokenRoute(uint32 _dstEid, address _srcToken) returns()
func (_MyOApp *MyOAppTransactorSession) RemoveTokenRoute(_dstEid uint32, _srcToken common.Address) (*types.Transaction, error) {
	return _MyOApp.Contract.RemoveTokenRoute(&_MyOApp.TransactOpts, _dstEid, _srcToken)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_MyOApp *MyOAppTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MyOApp.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_MyOApp *MyOAppSession) RenounceOwnership() (*types.Transaction, error) {
	return _MyOApp.Contract.RenounceOwnership(&_MyOApp.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_MyOApp *MyOAppTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _MyOApp.Contract.RenounceOwnership(&_MyOApp.TransactOpts)
}

// RequestPayoutToken is a paid mutator transaction binding the contract method 0x975c860c.
//
// Solidity: function requestPayoutToken(uint32 _dstEid, address _srcToken, address _merchant, uint256 _amount, bytes _options) payable returns()
func (_MyOApp *MyOAppTransactor) RequestPayoutToken(opts *bind.TransactOpts, _dstEid uint32, _srcToken common.Address, _merchant common.Address, _amount *big.Int, _options []byte) (*types.Transaction, error) {
	return _MyOApp.contract.Transact(opts, "requestPayoutToken", _dstEid, _srcToken, _merchant, _amount, _options)
}

// RequestPayoutToken is a paid mutator transaction binding the contract method 0x975c860c.
//
// Solidity: function requestPayoutToken(uint32 _dstEid, address _srcToken, address _merchant, uint256 _amount, bytes _options) payable returns()
func (_MyOApp *MyOAppSession) RequestPayoutToken(_dstEid uint32, _srcToken common.Address, _merchant common.Address, _amount *big.Int, _options []byte) (*types.Transaction, error) {
	return _MyOApp.Contract.RequestPayoutToken(&_MyOApp.TransactOpts, _dstEid, _srcToken, _merchant, _amount, _options)
}

// RequestPayoutToken is a paid mutator transaction binding the contract method 0x975c860c.
//
// Solidity: function requestPayoutToken(uint32 _dstEid, address _srcToken, address _merchant, uint256 _amount, bytes _options) payable returns()
func (_MyOApp *MyOAppTransactorSession) RequestPayoutToken(_dstEid uint32, _srcToken common.Address, _merchant common.Address, _amount *big.Int, _options []byte) (*types.Transaction, error) {
	return _MyOApp.Contract.RequestPayoutToken(&_MyOApp.TransactOpts, _dstEid, _srcToken, _merchant, _amount, _options)
}

// SendString is a paid mutator transaction binding the contract method 0x36ef9620.
//
// Solidity: function sendString(uint32 _dstEid, string _string, bytes _options) payable returns()
func (_MyOApp *MyOAppTransactor) SendString(opts *bind.TransactOpts, _dstEid uint32, _string string, _options []byte) (*types.Transaction, error) {
	return _MyOApp.contract.Transact(opts, "sendString", _dstEid, _string, _options)
}

// SendString is a paid mutator transaction binding the contract method 0x36ef9620.
//
// Solidity: function sendString(uint32 _dstEid, string _string, bytes _options) payable returns()
func (_MyOApp *MyOAppSession) SendString(_dstEid uint32, _string string, _options []byte) (*types.Transaction, error) {
	return _MyOApp.Contract.SendString(&_MyOApp.TransactOpts, _dstEid, _string, _options)
}

// SendString is a paid mutator transaction binding the contract method 0x36ef9620.
//
// Solidity: function sendString(uint32 _dstEid, string _string, bytes _options) payable returns()
func (_MyOApp *MyOAppTransactorSession) SendString(_dstEid uint32, _string string, _options []byte) (*types.Transaction, error) {
	return _MyOApp.Contract.SendString(&_MyOApp.TransactOpts, _dstEid, _string, _options)
}

// SetDelegate is a paid mutator transaction binding the contract method 0xca5eb5e1.
//
// Solidity: function setDelegate(address _delegate) returns()
func (_MyOApp *MyOAppTransactor) SetDelegate(opts *bind.TransactOpts, _delegate common.Address) (*types.Transaction, error) {
	return _MyOApp.contract.Transact(opts, "setDelegate", _delegate)
}

// SetDelegate is a paid mutator transaction binding the contract method 0xca5eb5e1.
//
// Solidity: function setDelegate(address _delegate) returns()
func (_MyOApp *MyOAppSession) SetDelegate(_delegate common.Address) (*types.Transaction, error) {
	return _MyOApp.Contract.SetDelegate(&_MyOApp.TransactOpts, _delegate)
}

// SetDelegate is a paid mutator transaction binding the contract method 0xca5eb5e1.
//
// Solidity: function setDelegate(address _delegate) returns()
func (_MyOApp *MyOAppTransactorSession) SetDelegate(_delegate common.Address) (*types.Transaction, error) {
	return _MyOApp.Contract.SetDelegate(&_MyOApp.TransactOpts, _delegate)
}

// SetEnforcedOptions is a paid mutator transaction binding the contract method 0xb98bd070.
//
// Solidity: function setEnforcedOptions((uint32,uint16,bytes)[] _enforcedOptions) returns()
func (_MyOApp *MyOAppTransactor) SetEnforcedOptions(opts *bind.TransactOpts, _enforcedOptions []EnforcedOptionParam) (*types.Transaction, error) {
	return _MyOApp.contract.Transact(opts, "setEnforcedOptions", _enforcedOptions)
}

// SetEnforcedOptions is a paid mutator transaction binding the contract method 0xb98bd070.
//
// Solidity: function setEnforcedOptions((uint32,uint16,bytes)[] _enforcedOptions) returns()
func (_MyOApp *MyOAppSession) SetEnforcedOptions(_enforcedOptions []EnforcedOptionParam) (*types.Transaction, error) {
	return _MyOApp.Contract.SetEnforcedOptions(&_MyOApp.TransactOpts, _enforcedOptions)
}

// SetEnforcedOptions is a paid mutator transaction binding the contract method 0xb98bd070.
//
// Solidity: function setEnforcedOptions((uint32,uint16,bytes)[] _enforcedOptions) returns()
func (_MyOApp *MyOAppTransactorSession) SetEnforcedOptions(_enforcedOptions []EnforcedOptionParam) (*types.Transaction, error) {
	return _MyOApp.Contract.SetEnforcedOptions(&_MyOApp.TransactOpts, _enforcedOptions)
}

// SetPeer is a paid mutator transaction binding the contract method 0x3400288b.
//
// Solidity: function setPeer(uint32 _eid, bytes32 _peer) returns()
func (_MyOApp *MyOAppTransactor) SetPeer(opts *bind.TransactOpts, _eid uint32, _peer [32]byte) (*types.Transaction, error) {
	return _MyOApp.contract.Transact(opts, "setPeer", _eid, _peer)
}

// SetPeer is a paid mutator transaction binding the contract method 0x3400288b.
//
// Solidity: function setPeer(uint32 _eid, bytes32 _peer) returns()
func (_MyOApp *MyOAppSession) SetPeer(_eid uint32, _peer [32]byte) (*types.Transaction, error) {
	return _MyOApp.Contract.SetPeer(&_MyOApp.TransactOpts, _eid, _peer)
}

// SetPeer is a paid mutator transaction binding the contract method 0x3400288b.
//
// Solidity: function setPeer(uint32 _eid, bytes32 _peer) returns()
func (_MyOApp *MyOAppTransactorSession) SetPeer(_eid uint32, _peer [32]byte) (*types.Transaction, error) {
	return _MyOApp.Contract.SetPeer(&_MyOApp.TransactOpts, _eid, _peer)
}

// SetTokenRoute is a paid mutator transaction binding the contract method 0x50f2dde7.
//
// Solidity: function setTokenRoute(uint32 _dstEid, address _srcToken, address _dstToken) returns()
func (_MyOApp *MyOAppTransactor) SetTokenRoute(opts *bind.TransactOpts, _dstEid uint32, _srcToken common.Address, _dstToken common.Address) (*types.Transaction, error) {
	return _MyOApp.contract.Transact(opts, "setTokenRoute", _dstEid, _srcToken, _dstToken)
}

// SetTokenRoute is a paid mutator transaction binding the contract method 0x50f2dde7.
//
// Solidity: function setTokenRoute(uint32 _dstEid, address _srcToken, address _dstToken) returns()
func (_MyOApp *MyOAppSession) SetTokenRoute(_dstEid uint32, _srcToken common.Address, _dstToken common.Address) (*types.Transaction, error) {
	return _MyOApp.Contract.SetTokenRoute(&_MyOApp.TransactOpts, _dstEid, _srcToken, _dstToken)
}

// SetTokenRoute is a paid mutator transaction binding the contract method 0x50f2dde7.
//
// Solidity: function setTokenRoute(uint32 _dstEid, address _srcToken, address _dstToken) returns()
func (_MyOApp *MyOAppTransactorSession) SetTokenRoute(_dstEid uint32, _srcToken common.Address, _dstToken common.Address) (*types.Transaction, error) {
	return _MyOApp.Contract.SetTokenRoute(&_MyOApp.TransactOpts, _dstEid, _srcToken, _dstToken)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_MyOApp *MyOAppTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _MyOApp.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_MyOApp *MyOAppSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _MyOApp.Contract.TransferOwnership(&_MyOApp.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_MyOApp *MyOAppTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _MyOApp.Contract.TransferOwnership(&_MyOApp.TransactOpts, newOwner)
}

// MyOAppEnforcedOptionSetIterator is returned from FilterEnforcedOptionSet and is used to iterate over the raw logs and unpacked data for EnforcedOptionSet events raised by the MyOApp contract.
type MyOAppEnforcedOptionSetIterator struct {
	Event *MyOAppEnforcedOptionSet // Event containing the contract specifics and raw log

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
func (it *MyOAppEnforcedOptionSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MyOAppEnforcedOptionSet)
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
		it.Event = new(MyOAppEnforcedOptionSet)
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
func (it *MyOAppEnforcedOptionSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MyOAppEnforcedOptionSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MyOAppEnforcedOptionSet represents a EnforcedOptionSet event raised by the MyOApp contract.
type MyOAppEnforcedOptionSet struct {
	EnforcedOptions []EnforcedOptionParam
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterEnforcedOptionSet is a free log retrieval operation binding the contract event 0xbe4864a8e820971c0247f5992e2da559595f7bf076a21cb5928d443d2a13b674.
//
// Solidity: event EnforcedOptionSet((uint32,uint16,bytes)[] _enforcedOptions)
func (_MyOApp *MyOAppFilterer) FilterEnforcedOptionSet(opts *bind.FilterOpts) (*MyOAppEnforcedOptionSetIterator, error) {

	logs, sub, err := _MyOApp.contract.FilterLogs(opts, "EnforcedOptionSet")
	if err != nil {
		return nil, err
	}
	return &MyOAppEnforcedOptionSetIterator{contract: _MyOApp.contract, event: "EnforcedOptionSet", logs: logs, sub: sub}, nil
}

// WatchEnforcedOptionSet is a free log subscription operation binding the contract event 0xbe4864a8e820971c0247f5992e2da559595f7bf076a21cb5928d443d2a13b674.
//
// Solidity: event EnforcedOptionSet((uint32,uint16,bytes)[] _enforcedOptions)
func (_MyOApp *MyOAppFilterer) WatchEnforcedOptionSet(opts *bind.WatchOpts, sink chan<- *MyOAppEnforcedOptionSet) (event.Subscription, error) {

	logs, sub, err := _MyOApp.contract.WatchLogs(opts, "EnforcedOptionSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MyOAppEnforcedOptionSet)
				if err := _MyOApp.contract.UnpackLog(event, "EnforcedOptionSet", log); err != nil {
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

// ParseEnforcedOptionSet is a log parse operation binding the contract event 0xbe4864a8e820971c0247f5992e2da559595f7bf076a21cb5928d443d2a13b674.
//
// Solidity: event EnforcedOptionSet((uint32,uint16,bytes)[] _enforcedOptions)
func (_MyOApp *MyOAppFilterer) ParseEnforcedOptionSet(log types.Log) (*MyOAppEnforcedOptionSet, error) {
	event := new(MyOAppEnforcedOptionSet)
	if err := _MyOApp.contract.UnpackLog(event, "EnforcedOptionSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MyOAppOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the MyOApp contract.
type MyOAppOwnershipTransferredIterator struct {
	Event *MyOAppOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *MyOAppOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MyOAppOwnershipTransferred)
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
		it.Event = new(MyOAppOwnershipTransferred)
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
func (it *MyOAppOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MyOAppOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MyOAppOwnershipTransferred represents a OwnershipTransferred event raised by the MyOApp contract.
type MyOAppOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_MyOApp *MyOAppFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*MyOAppOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _MyOApp.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &MyOAppOwnershipTransferredIterator{contract: _MyOApp.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_MyOApp *MyOAppFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *MyOAppOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _MyOApp.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MyOAppOwnershipTransferred)
				if err := _MyOApp.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_MyOApp *MyOAppFilterer) ParseOwnershipTransferred(log types.Log) (*MyOAppOwnershipTransferred, error) {
	event := new(MyOAppOwnershipTransferred)
	if err := _MyOApp.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MyOAppPeerSetIterator is returned from FilterPeerSet and is used to iterate over the raw logs and unpacked data for PeerSet events raised by the MyOApp contract.
type MyOAppPeerSetIterator struct {
	Event *MyOAppPeerSet // Event containing the contract specifics and raw log

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
func (it *MyOAppPeerSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MyOAppPeerSet)
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
		it.Event = new(MyOAppPeerSet)
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
func (it *MyOAppPeerSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MyOAppPeerSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MyOAppPeerSet represents a PeerSet event raised by the MyOApp contract.
type MyOAppPeerSet struct {
	Eid  uint32
	Peer [32]byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterPeerSet is a free log retrieval operation binding the contract event 0x238399d427b947898edb290f5ff0f9109849b1c3ba196a42e35f00c50a54b98b.
//
// Solidity: event PeerSet(uint32 eid, bytes32 peer)
func (_MyOApp *MyOAppFilterer) FilterPeerSet(opts *bind.FilterOpts) (*MyOAppPeerSetIterator, error) {

	logs, sub, err := _MyOApp.contract.FilterLogs(opts, "PeerSet")
	if err != nil {
		return nil, err
	}
	return &MyOAppPeerSetIterator{contract: _MyOApp.contract, event: "PeerSet", logs: logs, sub: sub}, nil
}

// WatchPeerSet is a free log subscription operation binding the contract event 0x238399d427b947898edb290f5ff0f9109849b1c3ba196a42e35f00c50a54b98b.
//
// Solidity: event PeerSet(uint32 eid, bytes32 peer)
func (_MyOApp *MyOAppFilterer) WatchPeerSet(opts *bind.WatchOpts, sink chan<- *MyOAppPeerSet) (event.Subscription, error) {

	logs, sub, err := _MyOApp.contract.WatchLogs(opts, "PeerSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MyOAppPeerSet)
				if err := _MyOApp.contract.UnpackLog(event, "PeerSet", log); err != nil {
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

// ParsePeerSet is a log parse operation binding the contract event 0x238399d427b947898edb290f5ff0f9109849b1c3ba196a42e35f00c50a54b98b.
//
// Solidity: event PeerSet(uint32 eid, bytes32 peer)
func (_MyOApp *MyOAppFilterer) ParsePeerSet(log types.Log) (*MyOAppPeerSet, error) {
	event := new(MyOAppPeerSet)
	if err := _MyOApp.contract.UnpackLog(event, "PeerSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MyOAppTokenPayoutExecutedIterator is returned from FilterTokenPayoutExecuted and is used to iterate over the raw logs and unpacked data for TokenPayoutExecuted events raised by the MyOApp contract.
type MyOAppTokenPayoutExecutedIterator struct {
	Event *MyOAppTokenPayoutExecuted // Event containing the contract specifics and raw log

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
func (it *MyOAppTokenPayoutExecutedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MyOAppTokenPayoutExecuted)
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
		it.Event = new(MyOAppTokenPayoutExecuted)
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
func (it *MyOAppTokenPayoutExecutedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MyOAppTokenPayoutExecutedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MyOAppTokenPayoutExecuted represents a TokenPayoutExecuted event raised by the MyOApp contract.
type MyOAppTokenPayoutExecuted struct {
	Merchant common.Address
	Token    common.Address
	Amount   *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterTokenPayoutExecuted is a free log retrieval operation binding the contract event 0x2e6e04a8b2c1647c89193b83e490574a4ddedb4ce7a402f93453cd40b67a8402.
//
// Solidity: event TokenPayoutExecuted(address indexed merchant, address token, uint256 amount)
func (_MyOApp *MyOAppFilterer) FilterTokenPayoutExecuted(opts *bind.FilterOpts, merchant []common.Address) (*MyOAppTokenPayoutExecutedIterator, error) {

	var merchantRule []interface{}
	for _, merchantItem := range merchant {
		merchantRule = append(merchantRule, merchantItem)
	}

	logs, sub, err := _MyOApp.contract.FilterLogs(opts, "TokenPayoutExecuted", merchantRule)
	if err != nil {
		return nil, err
	}
	return &MyOAppTokenPayoutExecutedIterator{contract: _MyOApp.contract, event: "TokenPayoutExecuted", logs: logs, sub: sub}, nil
}

// WatchTokenPayoutExecuted is a free log subscription operation binding the contract event 0x2e6e04a8b2c1647c89193b83e490574a4ddedb4ce7a402f93453cd40b67a8402.
//
// Solidity: event TokenPayoutExecuted(address indexed merchant, address token, uint256 amount)
func (_MyOApp *MyOAppFilterer) WatchTokenPayoutExecuted(opts *bind.WatchOpts, sink chan<- *MyOAppTokenPayoutExecuted, merchant []common.Address) (event.Subscription, error) {

	var merchantRule []interface{}
	for _, merchantItem := range merchant {
		merchantRule = append(merchantRule, merchantItem)
	}

	logs, sub, err := _MyOApp.contract.WatchLogs(opts, "TokenPayoutExecuted", merchantRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MyOAppTokenPayoutExecuted)
				if err := _MyOApp.contract.UnpackLog(event, "TokenPayoutExecuted", log); err != nil {
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

// ParseTokenPayoutExecuted is a log parse operation binding the contract event 0x2e6e04a8b2c1647c89193b83e490574a4ddedb4ce7a402f93453cd40b67a8402.
//
// Solidity: event TokenPayoutExecuted(address indexed merchant, address token, uint256 amount)
func (_MyOApp *MyOAppFilterer) ParseTokenPayoutExecuted(log types.Log) (*MyOAppTokenPayoutExecuted, error) {
	event := new(MyOAppTokenPayoutExecuted)
	if err := _MyOApp.contract.UnpackLog(event, "TokenPayoutExecuted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MyOAppTokenPayoutRequestedIterator is returned from FilterTokenPayoutRequested and is used to iterate over the raw logs and unpacked data for TokenPayoutRequested events raised by the MyOApp contract.
type MyOAppTokenPayoutRequestedIterator struct {
	Event *MyOAppTokenPayoutRequested // Event containing the contract specifics and raw log

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
func (it *MyOAppTokenPayoutRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MyOAppTokenPayoutRequested)
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
		it.Event = new(MyOAppTokenPayoutRequested)
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
func (it *MyOAppTokenPayoutRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MyOAppTokenPayoutRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MyOAppTokenPayoutRequested represents a TokenPayoutRequested event raised by the MyOApp contract.
type MyOAppTokenPayoutRequested struct {
	DstEid      uint32
	Payer       common.Address
	Merchant    common.Address
	SrcToken    common.Address
	DstToken    common.Address
	GrossAmount *big.Int
	NetAmount   *big.Int
	FeeAmount   *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterTokenPayoutRequested is a free log retrieval operation binding the contract event 0xdd9e34114af31ed8b7896e826d4d77f69661c83c3fb0dfde856e2de117034601.
//
// Solidity: event TokenPayoutRequested(uint32 indexed dstEid, address indexed payer, address indexed merchant, address srcToken, address dstToken, uint256 grossAmount, uint256 netAmount, uint256 feeAmount)
func (_MyOApp *MyOAppFilterer) FilterTokenPayoutRequested(opts *bind.FilterOpts, dstEid []uint32, payer []common.Address, merchant []common.Address) (*MyOAppTokenPayoutRequestedIterator, error) {

	var dstEidRule []interface{}
	for _, dstEidItem := range dstEid {
		dstEidRule = append(dstEidRule, dstEidItem)
	}
	var payerRule []interface{}
	for _, payerItem := range payer {
		payerRule = append(payerRule, payerItem)
	}
	var merchantRule []interface{}
	for _, merchantItem := range merchant {
		merchantRule = append(merchantRule, merchantItem)
	}

	logs, sub, err := _MyOApp.contract.FilterLogs(opts, "TokenPayoutRequested", dstEidRule, payerRule, merchantRule)
	if err != nil {
		return nil, err
	}
	return &MyOAppTokenPayoutRequestedIterator{contract: _MyOApp.contract, event: "TokenPayoutRequested", logs: logs, sub: sub}, nil
}

// WatchTokenPayoutRequested is a free log subscription operation binding the contract event 0xdd9e34114af31ed8b7896e826d4d77f69661c83c3fb0dfde856e2de117034601.
//
// Solidity: event TokenPayoutRequested(uint32 indexed dstEid, address indexed payer, address indexed merchant, address srcToken, address dstToken, uint256 grossAmount, uint256 netAmount, uint256 feeAmount)
func (_MyOApp *MyOAppFilterer) WatchTokenPayoutRequested(opts *bind.WatchOpts, sink chan<- *MyOAppTokenPayoutRequested, dstEid []uint32, payer []common.Address, merchant []common.Address) (event.Subscription, error) {

	var dstEidRule []interface{}
	for _, dstEidItem := range dstEid {
		dstEidRule = append(dstEidRule, dstEidItem)
	}
	var payerRule []interface{}
	for _, payerItem := range payer {
		payerRule = append(payerRule, payerItem)
	}
	var merchantRule []interface{}
	for _, merchantItem := range merchant {
		merchantRule = append(merchantRule, merchantItem)
	}

	logs, sub, err := _MyOApp.contract.WatchLogs(opts, "TokenPayoutRequested", dstEidRule, payerRule, merchantRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MyOAppTokenPayoutRequested)
				if err := _MyOApp.contract.UnpackLog(event, "TokenPayoutRequested", log); err != nil {
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

// ParseTokenPayoutRequested is a log parse operation binding the contract event 0xdd9e34114af31ed8b7896e826d4d77f69661c83c3fb0dfde856e2de117034601.
//
// Solidity: event TokenPayoutRequested(uint32 indexed dstEid, address indexed payer, address indexed merchant, address srcToken, address dstToken, uint256 grossAmount, uint256 netAmount, uint256 feeAmount)
func (_MyOApp *MyOAppFilterer) ParseTokenPayoutRequested(log types.Log) (*MyOAppTokenPayoutRequested, error) {
	event := new(MyOAppTokenPayoutRequested)
	if err := _MyOApp.contract.UnpackLog(event, "TokenPayoutRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
