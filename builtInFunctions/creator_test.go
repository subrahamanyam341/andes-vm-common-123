package builtInFunctions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/subrahamanyam341/andes-core-16/core"
	"github.com/subrahamanyam341/andes-core-16/core/check"
	"github.com/subrahamanyam341/andes-vm-common-123/mock"
)

func createMockArguments() ArgsCreateBuiltInFunctionContainer {
	gasMap := make(map[string]map[string]uint64)
	fillGasMapInternal(gasMap, 1)

	args := ArgsCreateBuiltInFunctionContainer{
		GasMap:                           gasMap,
		MapDNSAddresses:                  make(map[string]struct{}),
		MapDNSV2Addresses:                make(map[string]struct{}),
		EnableUserNameChange:             false,
		Marshalizer:                      &mock.MarshalizerMock{},
		Accounts:                         &mock.AccountsStub{},
		ShardCoordinator:                 mock.NewMultiShardsCoordinatorMock(1),
		EnableEpochsHandler:              &mock.EnableEpochsHandlerStub{},
		GuardedAccountHandler:            &mock.GuardedAccountHandlerStub{},
		MaxNumOfAddressesForTransferRole: 100,
	}

	return args
}

func fillGasMapInternal(gasMap map[string]map[string]uint64, value uint64) map[string]map[string]uint64 {
	gasMap[core.BaseOperationCostString] = fillGasMapBaseOperationCosts(value)
	gasMap[core.BuiltInCostString] = fillGasMapBuiltInCosts(value)

	return gasMap
}

func fillGasMapBaseOperationCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["StorePerByte"] = value
	gasMap["DataCopyPerByte"] = value
	gasMap["ReleasePerByte"] = value
	gasMap["PersistPerByte"] = value
	gasMap["CompilePerByte"] = value
	gasMap["AoTPreparePerByte"] = value
	gasMap["GetCode"] = value
	return gasMap
}

func fillGasMapBuiltInCosts(value uint64) map[string]uint64 {
	gasMap := make(map[string]uint64)
	gasMap["ClaimDeveloperRewards"] = value
	gasMap["ChangeOwnerAddress"] = value
	gasMap["SaveUserName"] = value
	gasMap["SaveKeyValue"] = value
	gasMap["DCTTransfer"] = value
	gasMap["DCTBurn"] = value
	gasMap["ChangeOwnerAddress"] = value
	gasMap["ClaimDeveloperRewards"] = value
	gasMap["SaveUserName"] = value
	gasMap["SaveKeyValue"] = value
	gasMap["DCTTransfer"] = value
	gasMap["DCTBurn"] = value
	gasMap["DCTLocalMint"] = value
	gasMap["DCTLocalBurn"] = value
	gasMap["DCTNFTCreate"] = value
	gasMap["DCTNFTAddQuantity"] = value
	gasMap["DCTNFTBurn"] = value
	gasMap["DCTNFTTransfer"] = value
	gasMap["DCTNFTChangeCreateOwner"] = value
	gasMap["DCTNFTAddUri"] = value
	gasMap["DCTNFTUpdateAttributes"] = value
	gasMap["DCTNFTMultiTransfer"] = value
	gasMap["SetGuardian"] = value
	gasMap["GuardAccount"] = value
	gasMap["UnGuardAccount"] = value
	gasMap["TrieLoadPerNode"] = value
	gasMap["TrieStorePerNode"] = value

	return gasMap
}

func TestCreateBuiltInFunctionContainer_Errors(t *testing.T) {
	args := createMockArguments()
	args.GasMap[core.BuiltInCostString]["ClaimDeveloperRewards"] = 0

	f, err := NewBuiltInFunctionsCreator(args)
	assert.Nil(t, f)
	assert.NotNil(t, err)

	args = createMockArguments()
	args.ShardCoordinator = nil
	_, err = NewBuiltInFunctionsCreator(args)
	assert.Equal(t, err, ErrNilShardCoordinator)

	args = createMockArguments()
	args.EnableEpochsHandler = nil
	_, err = NewBuiltInFunctionsCreator(args)
	assert.Equal(t, err, ErrNilEnableEpochsHandler)

	args = createMockArguments()
	args.Marshalizer = nil
	_, err = NewBuiltInFunctionsCreator(args)
	assert.Equal(t, err, ErrNilMarshalizer)

	args = createMockArguments()
	args.Accounts = nil
	_, err = NewBuiltInFunctionsCreator(args)
	assert.Equal(t, err, ErrNilAccountsAdapter)

	args = createMockArguments()
	args.GuardedAccountHandler = nil
	_, err = NewBuiltInFunctionsCreator(args)
	assert.Equal(t, err, ErrNilGuardedAccountHandler)

	args = createMockArguments()
	f, err = NewBuiltInFunctionsCreator(args)
	assert.Nil(t, err)
	assert.NotNil(t, f)
}

func TestBuiltInFuncCreator_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	var instance *builtInFuncCreator
	require.True(t, instance.IsInterfaceNil())

	instance, _ = NewBuiltInFunctionsCreator(createMockArguments())
	require.False(t, instance.IsInterfaceNil())
}

func TestCreateBuiltInContainer_GasScheduleChange(t *testing.T) {
	args := createMockArguments()
	f, _ := NewBuiltInFunctionsCreator(args)

	fillGasMapInternal(args.GasMap, 5)
	args.GasMap[core.BuiltInCostString]["ClaimDeveloperRewards"] = 0
	f.GasScheduleChange(args.GasMap)
	assert.Equal(t, f.gasConfig.BuiltInCost.ClaimDeveloperRewards, uint64(1))

	args.GasMap[core.BuiltInCostString]["ClaimDeveloperRewards"] = 5
	f.GasScheduleChange(args.GasMap)
	assert.Equal(t, f.gasConfig.BuiltInCost.ClaimDeveloperRewards, uint64(5))
}

func TestCreateBuiltInContainer_Create(t *testing.T) {
	args := createMockArguments()
	f, _ := NewBuiltInFunctionsCreator(args)

	err := f.CreateBuiltInFunctionContainer()
	assert.Nil(t, err)
	assert.Equal(t, 36, f.BuiltInFunctionContainer().Len())

	err = f.SetPayableHandler(nil)
	assert.NotNil(t, err)

	err = f.SetPayableHandler(&mock.PayableHandlerStub{})
	assert.Nil(t, err)

	fillGasMapInternal(args.GasMap, 5)
	f.GasScheduleChange(args.GasMap)
	assert.Equal(t, f.gasConfig.BuiltInCost.ClaimDeveloperRewards, uint64(5))

	nftStorageHandler := f.NFTStorageHandler()
	assert.False(t, check.IfNil(nftStorageHandler))
}
