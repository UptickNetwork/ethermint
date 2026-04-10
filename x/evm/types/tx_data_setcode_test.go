package types

import (
	"math/big"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"
)

func TestNewTxDataFromTx_SetCodeTxSupported(t *testing.T) {
	to := ethtypes.SetCodeAuthorization{}
	auth := ethtypes.SetCodeAuthorization{
		ChainID: *uint256.MustFromBig(big.NewInt(1)),
		Address: to.Address,
		Nonce:   7,
		V:       1,
		R:       *uint256.MustFromBig(big.NewInt(2)),
		S:       *uint256.MustFromBig(big.NewInt(3)),
	}
	tx := ethtypes.NewTx(&ethtypes.SetCodeTx{
		ChainID:   uint256.MustFromBig(big.NewInt(1)),
		Nonce:     1,
		GasTipCap: uint256.MustFromBig(big.NewInt(1)),
		GasFeeCap: uint256.MustFromBig(big.NewInt(1)),
		Gas:       21000,
		To:        to.Address,
		Value:     uint256.MustFromBig(big.NewInt(0)),
		AuthList:  []ethtypes.SetCodeAuthorization{auth},
	})

	got, err := NewTxDataFromTx(tx)
	require.NoError(t, err)
	require.Equal(t, uint8(ethtypes.SetCodeTxType), got.TxType())
	require.Len(t, got.GetSetCodeAuthorizations(), 1)
	require.Equal(t, auth.Nonce, got.GetSetCodeAuthorizations()[0].Nonce)
}

func TestSetCodeTx_AsEthereumData_RoundTripAuthList(t *testing.T) {
	to := common.HexToAddress("0x0000000000000000000000000000000000000011")
	chainID := sdkmath.NewInt(1)
	gasCap := sdkmath.NewInt(20)
	tipCap := sdkmath.NewInt(5)
	amount := sdkmath.NewInt(9)

	txData := &SetCodeTx{
		ChainID:   &chainID,
		Nonce:     3,
		GasTipCap: &tipCap,
		GasFeeCap: &gasCap,
		GasLimit:  50000,
		To:        to.Hex(),
		Amount:    &amount,
		Data:      []byte{0x01, 0x02},
		AuthList: []SetCodeAuthorization{
			{
				ChainID: &chainID,
				Address: to.Hex(),
				Nonce:   8,
				V:       1,
				R:       big.NewInt(2).Bytes(),
				S:       big.NewInt(3).Bytes(),
			},
		},
		V: big.NewInt(27).Bytes(),
		R: big.NewInt(4).Bytes(),
		S: big.NewInt(5).Bytes(),
	}

	ethData := txData.AsEthereumData()
	ethTx, ok := ethData.(*ethtypes.SetCodeTx)
	require.True(t, ok)
	require.Equal(t, txData.Nonce, ethTx.Nonce)
	require.Equal(t, txData.To, ethTx.To.Hex())
	require.Len(t, ethTx.AuthList, 1)
	require.Equal(t, txData.AuthList[0].Nonce, ethTx.AuthList[0].Nonce)
	require.Equal(t, txData.AuthList[0].Address, ethTx.AuthList[0].Address.Hex())
}

func TestSetCodeTx_Validate_FailsOnInvalidYParity(t *testing.T) {
	chainID := sdkmath.NewInt(1)
	gasCap := sdkmath.NewInt(10)
	tipCap := sdkmath.NewInt(1)
	txData := SetCodeTx{
		ChainID:   &chainID,
		Nonce:     1,
		GasTipCap: &tipCap,
		GasFeeCap: &gasCap,
		GasLimit:  21000,
		To:        common.HexToAddress("0x0000000000000000000000000000000000000022").Hex(),
		AuthList: []SetCodeAuthorization{
			{
				ChainID: &chainID,
				Address: common.HexToAddress("0x0000000000000000000000000000000000000033").Hex(),
				Nonce:   1,
				V:       2,
			},
		},
	}

	err := txData.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid y parity")
}

func TestSetCodeTx_Validate_FailsOnInvalidAuthAddress(t *testing.T) {
	chainID := sdkmath.NewInt(1)
	gasCap := sdkmath.NewInt(10)
	tipCap := sdkmath.NewInt(1)
	txData := SetCodeTx{
		ChainID:   &chainID,
		Nonce:     1,
		GasTipCap: &tipCap,
		GasFeeCap: &gasCap,
		GasLimit:  21000,
		To:        common.HexToAddress("0x0000000000000000000000000000000000000022").Hex(),
		AuthList: []SetCodeAuthorization{
			{
				ChainID: &chainID,
				Address: "invalid-hex-address",
				Nonce:   1,
				V:       1,
			},
		},
	}

	err := txData.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid authorization address")
}

func TestSetCodeTx_Validate_FailsOnGasCapOrdering(t *testing.T) {
	chainID := sdkmath.NewInt(1)
	gasCap := sdkmath.NewInt(1)
	tipCap := sdkmath.NewInt(2)
	txData := SetCodeTx{
		ChainID:   &chainID,
		Nonce:     1,
		GasTipCap: &tipCap,
		GasFeeCap: &gasCap,
		GasLimit:  21000,
		To:        common.HexToAddress("0x0000000000000000000000000000000000000022").Hex(),
	}

	err := txData.Validate()
	require.Error(t, err)
	require.Contains(t, err.Error(), "max priority fee per gas higher than max fee per gas")
}
