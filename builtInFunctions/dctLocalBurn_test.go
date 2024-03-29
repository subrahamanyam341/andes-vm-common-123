package builtInFunctions

import (
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/subrahamanyam341/andes-core-16/core"
	"github.com/subrahamanyam341/andes-core-16/data/dct"
	vmcommon "github.com/subrahamanyam341/andes-vm-common-123"
	"github.com/subrahamanyam341/andes-vm-common-123/mock"
)

func TestNewDCTLocalBurnFunc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		argsFunc func() (c uint64, m vmcommon.Marshalizer, p vmcommon.ExtendedDCTGlobalSettingsHandler, r vmcommon.DCTRoleHandler, e vmcommon.EnableEpochsHandler)
		exError  error
	}{
		{
			name: "NilMarshalizer",
			argsFunc: func() (c uint64, m vmcommon.Marshalizer, p vmcommon.ExtendedDCTGlobalSettingsHandler, r vmcommon.DCTRoleHandler, e vmcommon.EnableEpochsHandler) {
				return 0, nil, &mock.GlobalSettingsHandlerStub{}, &mock.DCTRoleHandlerStub{}, &mock.EnableEpochsHandlerStub{}
			},
			exError: ErrNilMarshalizer,
		},
		{
			name: "NilGlobalSettingsHandler",
			argsFunc: func() (c uint64, m vmcommon.Marshalizer, p vmcommon.ExtendedDCTGlobalSettingsHandler, r vmcommon.DCTRoleHandler, e vmcommon.EnableEpochsHandler) {
				return 0, &mock.MarshalizerMock{}, nil, &mock.DCTRoleHandlerStub{}, &mock.EnableEpochsHandlerStub{}
			},
			exError: ErrNilGlobalSettingsHandler,
		},
		{
			name: "NilRolesHandler",
			argsFunc: func() (c uint64, m vmcommon.Marshalizer, p vmcommon.ExtendedDCTGlobalSettingsHandler, r vmcommon.DCTRoleHandler, e vmcommon.EnableEpochsHandler) {
				return 0, &mock.MarshalizerMock{}, &mock.GlobalSettingsHandlerStub{}, nil, &mock.EnableEpochsHandlerStub{}
			},
			exError: ErrNilRolesHandler,
		},
		{
			name: "NilEnableEpochsHandler",
			argsFunc: func() (c uint64, m vmcommon.Marshalizer, p vmcommon.ExtendedDCTGlobalSettingsHandler, r vmcommon.DCTRoleHandler, e vmcommon.EnableEpochsHandler) {
				return 0, &mock.MarshalizerMock{}, &mock.GlobalSettingsHandlerStub{}, &mock.DCTRoleHandlerStub{}, nil
			},
			exError: ErrNilEnableEpochsHandler,
		},
		{
			name: "Ok",
			argsFunc: func() (c uint64, m vmcommon.Marshalizer, p vmcommon.ExtendedDCTGlobalSettingsHandler, r vmcommon.DCTRoleHandler, e vmcommon.EnableEpochsHandler) {
				return 0, &mock.MarshalizerMock{}, &mock.GlobalSettingsHandlerStub{}, &mock.DCTRoleHandlerStub{}, &mock.EnableEpochsHandlerStub{}
			},
			exError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDCTLocalBurnFunc(tt.argsFunc())
			require.Equal(t, err, tt.exError)
		})
	}
}

func TestDctLocalBurn_ProcessBuiltinFunction_CalledWithValueShouldErr(t *testing.T) {
	t.Parallel()

	dctLocalBurnF, _ := NewDCTLocalBurnFunc(0, &mock.MarshalizerMock{}, &mock.GlobalSettingsHandlerStub{}, &mock.DCTRoleHandlerStub{}, &mock.EnableEpochsHandlerStub{})

	_, err := dctLocalBurnF.ProcessBuiltinFunction(&mock.AccountWrapMock{}, &mock.AccountWrapMock{}, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue: big.NewInt(1),
		},
	})
	require.Equal(t, ErrBuiltInFunctionCalledWithValue, err)
}

func TestDctLocalBurn_ProcessBuiltinFunction_CheckAllowToExecuteShouldErr(t *testing.T) {
	t.Parallel()

	localErr := errors.New("local err")
	dctLocalBurnF, _ := NewDCTLocalBurnFunc(0, &mock.MarshalizerMock{}, &mock.GlobalSettingsHandlerStub{}, &mock.DCTRoleHandlerStub{
		CheckAllowedToExecuteCalled: func(account vmcommon.UserAccountHandler, tokenID []byte, action []byte) error {
			return localErr
		},
	}, &mock.EnableEpochsHandlerStub{})

	_, err := dctLocalBurnF.ProcessBuiltinFunction(&mock.AccountWrapMock{}, &mock.AccountWrapMock{}, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue: big.NewInt(0),
			Arguments: [][]byte{[]byte("arg1"), []byte("arg2")},
		},
	})
	require.Equal(t, localErr, err)
}

func TestDctLocalBurn_ProcessBuiltinFunction_CannotAddToDctBalanceShouldErr(t *testing.T) {
	t.Parallel()

	dctLocalBurnF, _ := NewDCTLocalBurnFunc(0, &mock.MarshalizerMock{}, &mock.GlobalSettingsHandlerStub{}, &mock.DCTRoleHandlerStub{
		CheckAllowedToExecuteCalled: func(account vmcommon.UserAccountHandler, tokenID []byte, action []byte) error {
			return nil
		},
	}, &mock.EnableEpochsHandlerStub{})

	localErr := errors.New("local err")
	_, err := dctLocalBurnF.ProcessBuiltinFunction(&mock.UserAccountStub{
		AccountDataHandlerCalled: func() vmcommon.AccountDataHandler {
			return &mock.DataTrieTrackerStub{
				RetrieveValueCalled: func(_ []byte) ([]byte, uint32, error) {
					return nil, 0, localErr
				},
			}
		},
	}, &mock.AccountWrapMock{}, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue: big.NewInt(0),
			Arguments: [][]byte{[]byte("arg1"), []byte("arg2")},
		},
	})
	require.Equal(t, ErrInsufficientFunds, err)
}

func TestDctLocalBurn_ProcessBuiltinFunction_ValueTooLong(t *testing.T) {
	t.Parallel()

	marshaller := &mock.MarshalizerMock{}
	dctRoleHandler := &mock.DCTRoleHandlerStub{
		CheckAllowedToExecuteCalled: func(account vmcommon.UserAccountHandler, tokenID []byte, action []byte) error {
			assert.Equal(t, core.DCTRoleLocalBurn, string(action))
			return nil
		},
	}
	dctLocalBurnF, _ := NewDCTLocalBurnFunc(50, marshaller, &mock.GlobalSettingsHandlerStub{}, dctRoleHandler, &mock.EnableEpochsHandlerStub{})

	sndAccount := &mock.UserAccountStub{
		AccountDataHandlerCalled: func() vmcommon.AccountDataHandler {
			return &mock.DataTrieTrackerStub{
				RetrieveValueCalled: func(_ []byte) ([]byte, uint32, error) {
					dctData := &dct.DCToken{Value: big.NewInt(100)}
					serializedDctData, err := marshaller.Marshal(dctData)
					return serializedDctData, 0, err
				},
			}
		},
	}

	bigValueStr := "1" + strings.Repeat("0", 1000)
	bigValue, _ := big.NewInt(0).SetString(bigValueStr, 10)
	vmOutput, err := dctLocalBurnF.ProcessBuiltinFunction(sndAccount, &mock.AccountWrapMock{}, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:   big.NewInt(0),
			Arguments:   [][]byte{[]byte("arg1"), bigValue.Bytes()},
			GasProvided: 500,
		},
	})
	require.Equal(t, "insufficient funds", err.Error()) // before the activation of the flag
	require.Empty(t, vmOutput)

	// try again with the flag enabled
	dctLocalBurnF.enableEpochsHandler = &mock.EnableEpochsHandlerStub{
		IsConsistentTokensValuesLengthCheckEnabledField: true,
	}
	vmOutput, err = dctLocalBurnF.ProcessBuiltinFunction(sndAccount, &mock.AccountWrapMock{}, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:   big.NewInt(0),
			Arguments:   [][]byte{[]byte("arg1"), bigValue.Bytes()},
			GasProvided: 500,
		},
	})
	require.Equal(t, "invalid arguments to process built-in function: max length for dct local burn value is 100", err.Error())
	require.Empty(t, vmOutput)
}

func TestDctLocalBurn_ProcessBuiltinFunction_ShouldWork(t *testing.T) {
	t.Parallel()

	marshaller := &mock.MarshalizerMock{}
	dctRoleHandler := &mock.DCTRoleHandlerStub{
		CheckAllowedToExecuteCalled: func(account vmcommon.UserAccountHandler, tokenID []byte, action []byte) error {
			assert.Equal(t, core.DCTRoleLocalBurn, string(action))
			return nil
		},
	}
	dctLocalBurnF, _ := NewDCTLocalBurnFunc(50, marshaller, &mock.GlobalSettingsHandlerStub{}, dctRoleHandler, &mock.EnableEpochsHandlerStub{})

	sndAccout := &mock.UserAccountStub{
		AccountDataHandlerCalled: func() vmcommon.AccountDataHandler {
			return &mock.DataTrieTrackerStub{
				RetrieveValueCalled: func(_ []byte) ([]byte, uint32, error) {
					dctData := &dct.DCToken{Value: big.NewInt(100)}
					serializedDctData, err := marshaller.Marshal(dctData)
					return serializedDctData, 0, err
				},
				SaveKeyValueCalled: func(key []byte, value []byte) error {
					dctData := &dct.DCToken{}
					_ = marshaller.Unmarshal(dctData, value)
					require.Equal(t, big.NewInt(99), dctData.Value)
					return nil
				},
			}
		},
	}
	vmOutput, err := dctLocalBurnF.ProcessBuiltinFunction(sndAccout, &mock.AccountWrapMock{}, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:   big.NewInt(0),
			Arguments:   [][]byte{[]byte("arg1"), big.NewInt(1).Bytes()},
			GasProvided: 500,
		},
	})
	require.Equal(t, nil, err)

	expectedVMOutput := &vmcommon.VMOutput{
		ReturnCode:   vmcommon.Ok,
		GasRemaining: 450,
		Logs: []*vmcommon.LogEntry{
			{
				Identifier: []byte("DCTLocalBurn"),
				Address:    nil,
				Topics:     [][]byte{[]byte("arg1"), big.NewInt(0).Bytes(), big.NewInt(1).Bytes()},
				Data:       nil,
			},
		},
	}
	require.Equal(t, expectedVMOutput, vmOutput)
}

func TestDctLocalBurn_ProcessBuiltinFunction_WithGlobalBurn(t *testing.T) {
	t.Parallel()

	marshaller := &mock.MarshalizerMock{}
	dctLocalBurnF, _ := NewDCTLocalBurnFunc(50, marshaller, &mock.GlobalSettingsHandlerStub{
		IsBurnForAllCalled: func(token []byte) bool {
			return true
		},
	}, &mock.DCTRoleHandlerStub{
		CheckAllowedToExecuteCalled: func(account vmcommon.UserAccountHandler, tokenID []byte, action []byte) error {
			return errors.New("no role")
		},
	}, &mock.EnableEpochsHandlerStub{})

	sndAccout := &mock.UserAccountStub{
		AccountDataHandlerCalled: func() vmcommon.AccountDataHandler {
			return &mock.DataTrieTrackerStub{
				RetrieveValueCalled: func(_ []byte) ([]byte, uint32, error) {
					dctData := &dct.DCToken{Value: big.NewInt(100)}
					serializedDctData, err := marshaller.Marshal(dctData)
					return serializedDctData, 0, err
				},
				SaveKeyValueCalled: func(key []byte, value []byte) error {
					dctData := &dct.DCToken{}
					_ = marshaller.Unmarshal(dctData, value)
					require.Equal(t, big.NewInt(99), dctData.Value)
					return nil
				},
			}
		},
	}
	vmOutput, err := dctLocalBurnF.ProcessBuiltinFunction(sndAccout, &mock.AccountWrapMock{}, &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:   big.NewInt(0),
			Arguments:   [][]byte{[]byte("arg1"), big.NewInt(1).Bytes()},
			GasProvided: 500,
		},
	})
	require.Equal(t, nil, err)

	expectedVMOutput := &vmcommon.VMOutput{
		ReturnCode:   vmcommon.Ok,
		GasRemaining: 450,
		Logs: []*vmcommon.LogEntry{
			{
				Identifier: []byte("DCTLocalBurn"),
				Address:    nil,
				Topics:     [][]byte{[]byte("arg1"), big.NewInt(0).Bytes(), big.NewInt(1).Bytes()},
				Data:       nil,
			},
		},
	}
	require.Equal(t, expectedVMOutput, vmOutput)
}

func TestDctLocalBurn_SetNewGasConfig(t *testing.T) {
	t.Parallel()

	dctLocalBurnF, _ := NewDCTLocalBurnFunc(0, &mock.MarshalizerMock{}, &mock.GlobalSettingsHandlerStub{}, &mock.DCTRoleHandlerStub{}, &mock.EnableEpochsHandlerStub{})

	dctLocalBurnF.SetNewGasConfig(&vmcommon.GasCost{BuiltInCost: vmcommon.BuiltInCost{
		DCTLocalBurn: 500},
	})

	require.Equal(t, uint64(500), dctLocalBurnF.funcGasCost)
}

func TestCheckInputArgumentsForLocalAction_InvalidRecipientAddr(t *testing.T) {
	t.Parallel()

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:  big.NewInt(0),
			Arguments:  [][]byte{[]byte("arg1"), big.NewInt(1).Bytes()},
			CallerAddr: []byte("caller"),
		},
		RecipientAddr: []byte("rec"),
	}

	err := checkInputArgumentsForLocalAction(&mock.UserAccountStub{}, vmInput, 0)
	require.Equal(t, ErrInvalidRcvAddr, err)
}

func TestCheckInputArgumentsForLocalAction_NilUserAccount(t *testing.T) {
	t.Parallel()

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:  big.NewInt(0),
			Arguments:  [][]byte{[]byte("arg1"), big.NewInt(1).Bytes()},
			CallerAddr: []byte("caller"),
		},
		RecipientAddr: []byte("caller"),
	}

	err := checkInputArgumentsForLocalAction(nil, vmInput, 0)
	require.Equal(t, ErrNilUserAccount, err)
}

func TestCheckInputArgumentsForLocalAction_NotEnoughGas(t *testing.T) {
	t.Parallel()

	vmInput := &vmcommon.ContractCallInput{
		VMInput: vmcommon.VMInput{
			CallValue:   big.NewInt(0),
			Arguments:   [][]byte{[]byte("arg1"), big.NewInt(10).Bytes()},
			CallerAddr:  []byte("caller"),
			GasProvided: 1,
		},
		RecipientAddr: []byte("caller"),
	}

	err := checkInputArgumentsForLocalAction(&mock.UserAccountStub{}, vmInput, 500)
	require.Equal(t, ErrNotEnoughGas, err)
}
