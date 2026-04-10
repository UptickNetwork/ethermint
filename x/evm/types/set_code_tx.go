package types

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/evmos/ethermint/types"
	"github.com/holiman/uint256"
)

func newSetCodeTx(tx *ethtypes.Transaction) (*SetCodeTx, error) {
	txData := &SetCodeTx{
		Nonce:    tx.Nonce(),
		Data:     tx.Data(),
		GasLimit: tx.Gas(),
	}

	v, r, s := tx.RawSignatureValues()
	if to := tx.To(); to != nil {
		txData.To = to.Hex()
	}
	if tx.Value() != nil {
		amountInt, err := types.SafeNewIntFromBigInt(tx.Value())
		if err != nil {
			return nil, err
		}
		txData.Amount = &amountInt
	}
	if tx.GasFeeCap() != nil {
		gasFeeCapInt, err := types.SafeNewIntFromBigInt(tx.GasFeeCap())
		if err != nil {
			return nil, err
		}
		txData.GasFeeCap = &gasFeeCapInt
	}
	if tx.GasTipCap() != nil {
		gasTipCapInt, err := types.SafeNewIntFromBigInt(tx.GasTipCap())
		if err != nil {
			return nil, err
		}
		txData.GasTipCap = &gasTipCapInt
	}
	if tx.AccessList() != nil {
		al := tx.AccessList()
		txData.Accesses = NewAccessList(&al)
	}
	if authList := tx.SetCodeAuthorizations(); len(authList) > 0 {
		txData.AuthList = make([]SetCodeAuthorization, 0, len(authList))
		for _, auth := range authList {
			chainID := sdkmath.NewIntFromBigInt(auth.ChainID.ToBig())
			txData.AuthList = append(txData.AuthList, SetCodeAuthorization{
				ChainID: &chainID,
				Address: auth.Address.Hex(),
				Nonce:   auth.Nonce,
				V:       uint64(auth.V),
				R:       auth.R.Bytes(),
				S:       auth.S.Bytes(),
			})
		}
	}

	txData.SetSignatureValues(tx.ChainId(), v, r, s)
	return txData, nil
}

func (tx *SetCodeTx) TxType() uint8 { return ethtypes.SetCodeTxType }

func (tx *SetCodeTx) Copy() TxData {
	dataCpy := common.CopyBytes(tx.Data)
	vCpy := common.CopyBytes(tx.V)
	rCpy := common.CopyBytes(tx.R)
	sCpy := common.CopyBytes(tx.S)
	authCpy := make([]SetCodeAuthorization, len(tx.AuthList))
	copy(authCpy, tx.AuthList)
	return &SetCodeTx{
		ChainID:   tx.ChainID,
		Nonce:     tx.Nonce,
		GasTipCap: tx.GasTipCap,
		GasFeeCap: tx.GasFeeCap,
		GasLimit:  tx.GasLimit,
		To:        tx.To,
		Amount:    tx.Amount,
		Data:      dataCpy,
		Accesses:  tx.Accesses,
		AuthList:  authCpy,
		V:         vCpy,
		R:         rCpy,
		S:         sCpy,
	}
}

func (tx *SetCodeTx) GetChainID() *big.Int {
	if tx.ChainID == nil {
		return nil
	}
	return tx.ChainID.BigInt()
}

func (tx *SetCodeTx) GetAccessList() ethtypes.AccessList {
	if tx.Accesses == nil {
		return nil
	}
	return *tx.Accesses.ToEthAccessList()
}

func (tx *SetCodeTx) GetData() []byte  { return common.CopyBytes(tx.Data) }
func (tx *SetCodeTx) GetGas() uint64   { return tx.GasLimit }
func (tx *SetCodeTx) GetNonce() uint64 { return tx.Nonce }

func (tx *SetCodeTx) GetGasPrice() *big.Int { return tx.GetGasFeeCap() }

func (tx *SetCodeTx) GetGasTipCap() *big.Int {
	if tx.GasTipCap == nil {
		return nil
	}
	return tx.GasTipCap.BigInt()
}

func (tx *SetCodeTx) GetGasFeeCap() *big.Int {
	if tx.GasFeeCap == nil {
		return nil
	}
	return tx.GasFeeCap.BigInt()
}

func (tx *SetCodeTx) GetBlobFeeCap() *big.Int { return nil }
func (tx *SetCodeTx) GetBlobHashes() []common.Hash {
	return nil
}

func (tx *SetCodeTx) GetSetCodeAuthorizations() []ethtypes.SetCodeAuthorization {
	if len(tx.AuthList) == 0 {
		return nil
	}
	authList := make([]ethtypes.SetCodeAuthorization, 0, len(tx.AuthList))
	for _, auth := range tx.AuthList {
		authList = append(authList, auth.AsEthereumAuthorization())
	}
	return authList
}

func (tx *SetCodeTx) GetValue() *big.Int {
	if tx.Amount == nil {
		return nil
	}
	return tx.Amount.BigInt()
}

func (tx *SetCodeTx) GetTo() *common.Address {
	if tx.To == "" {
		return nil
	}
	to := common.HexToAddress(tx.To)
	return &to
}

func (tx *SetCodeTx) AsEthereumData() ethtypes.TxData {
	v, r, s := tx.GetRawSignatureValues()
	to := common.Address{}
	if tx.GetTo() != nil {
		to = *tx.GetTo()
	}
	return &ethtypes.SetCodeTx{
		ChainID:    uint256FromBig(tx.GetChainID()),
		Nonce:      tx.GetNonce(),
		GasTipCap:  uint256FromBig(tx.GetGasTipCap()),
		GasFeeCap:  uint256FromBig(tx.GetGasFeeCap()),
		Gas:        tx.GetGas(),
		To:         to,
		Value:      uint256FromBig(tx.GetValue()),
		Data:       tx.GetData(),
		AccessList: tx.GetAccessList(),
		AuthList:   tx.GetSetCodeAuthorizations(),
		V:          uint256FromBig(v),
		R:          uint256FromBig(r),
		S:          uint256FromBig(s),
	}
}

func (tx *SetCodeTx) GetRawSignatureValues() (v, r, s *big.Int) {
	return rawSignatureValues(tx.V, tx.R, tx.S)
}

func (tx *SetCodeTx) SetSignatureValues(chainID, v, r, s *big.Int) {
	if v != nil {
		tx.V = v.Bytes()
	}
	if r != nil {
		tx.R = r.Bytes()
	}
	if s != nil {
		tx.S = s.Bytes()
	}
	if chainID != nil {
		chainIDInt := sdkmath.NewIntFromBigInt(chainID)
		tx.ChainID = &chainIDInt
	}
}

func (tx SetCodeTx) Validate() error {
	if tx.GasTipCap == nil {
		return errorsmod.Wrap(ErrInvalidGasCap, "gas tip cap cannot nil")
	}
	if tx.GasFeeCap == nil {
		return errorsmod.Wrap(ErrInvalidGasCap, "gas fee cap cannot nil")
	}
	if tx.GasTipCap.IsNegative() {
		return errorsmod.Wrapf(ErrInvalidGasCap, "gas tip cap cannot be negative %s", tx.GasTipCap)
	}
	if tx.GasFeeCap.IsNegative() {
		return errorsmod.Wrapf(ErrInvalidGasCap, "gas fee cap cannot be negative %s", tx.GasFeeCap)
	}
	if tx.GasFeeCap.LT(*tx.GasTipCap) {
		return errorsmod.Wrapf(
			ErrInvalidGasCap,
			"max priority fee per gas higher than max fee per gas (%s > %s)",
			tx.GasTipCap, tx.GasFeeCap,
		)
	}
	if !types.IsValidInt256(tx.GetGasTipCap()) {
		return errorsmod.Wrap(ErrInvalidGasCap, "out of bound")
	}
	if !types.IsValidInt256(tx.GetGasFeeCap()) {
		return errorsmod.Wrap(ErrInvalidGasCap, "out of bound")
	}
	if !types.IsValidInt256(tx.Fee()) {
		return errorsmod.Wrap(ErrInvalidGasFee, "out of bound")
	}
	amount := tx.GetValue()
	if amount != nil && amount.Sign() == -1 {
		return errorsmod.Wrapf(ErrInvalidAmount, "amount cannot be negative %s", amount)
	}
	if !types.IsValidInt256(amount) {
		return errorsmod.Wrap(ErrInvalidAmount, "out of bound")
	}
	if tx.To != "" {
		if err := types.ValidateAddress(tx.To); err != nil {
			return errorsmod.Wrap(err, "invalid to address")
		}
	}
	if tx.GetChainID() == nil {
		return errorsmod.Wrap(errortypes.ErrInvalidChainID, "chain ID must be present on setcode txs")
	}
	for i, auth := range tx.AuthList {
		if err := auth.Validate(); err != nil {
			return errorsmod.Wrapf(err, "invalid auth list entry at index %d", i)
		}
	}
	return nil
}

func (tx SetCodeTx) Fee() *big.Int  { return fee(tx.GetGasFeeCap(), tx.GasLimit) }
func (tx SetCodeTx) Cost() *big.Int { return cost(tx.Fee(), tx.GetValue()) }

func (tx *SetCodeTx) EffectiveGasPrice(baseFee *big.Int) *big.Int {
	return EffectiveGasPrice(baseFee, tx.GasFeeCap.BigInt(), tx.GasTipCap.BigInt())
}
func (tx SetCodeTx) EffectiveFee(baseFee *big.Int) *big.Int {
	return fee(tx.EffectiveGasPrice(baseFee), tx.GasLimit)
}
func (tx SetCodeTx) EffectiveCost(baseFee *big.Int) *big.Int {
	return cost(tx.EffectiveFee(baseFee), tx.GetValue())
}

func (auth SetCodeAuthorization) AsEthereumAuthorization() ethtypes.SetCodeAuthorization {
	chainID := big.NewInt(0)
	if auth.ChainID != nil {
		chainID = auth.ChainID.BigInt()
	}
	ethAuth := ethtypes.SetCodeAuthorization{
		ChainID: *uint256.MustFromBig(chainID),
		Address: common.HexToAddress(auth.Address),
		Nonce:   auth.Nonce,
		V:       uint8(auth.V),
	}
	ethAuth.R.SetBytes(auth.R)
	ethAuth.S.SetBytes(auth.S)
	return ethAuth
}

func (auth SetCodeAuthorization) Validate() error {
	if auth.ChainID == nil {
		return errorsmod.Wrap(errortypes.ErrInvalidChainID, "chain ID must be present on setcode authorization")
	}
	if !types.IsValidInt256(auth.ChainID.BigInt()) {
		return errorsmod.Wrap(errortypes.ErrInvalidChainID, "chain ID out of bound")
	}
	if err := types.ValidateAddress(auth.Address); err != nil {
		return errorsmod.Wrap(err, "invalid authorization address")
	}
	if auth.V > 1 {
		return errorsmod.Wrapf(errortypes.ErrInvalidRequest, "invalid y parity %d", auth.V)
	}
	var r, s uint256.Int
	r.SetBytes(auth.R)
	s.SetBytes(auth.S)
	if !r.IsZero() && !s.IsZero() {
		ethAuth := auth.AsEthereumAuthorization()
		if _, err := ethAuth.Authority(); err != nil {
			return errorsmod.Wrap(errortypes.ErrInvalidRequest, err.Error())
		}
	}
	return nil
}
