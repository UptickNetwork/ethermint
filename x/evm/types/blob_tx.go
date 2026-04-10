package types

import (
	"math/big"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/evmos/ethermint/types"
	"github.com/holiman/uint256"
)

func newBlobTx(tx *ethtypes.Transaction) (*BlobTx, error) {
	txData := &BlobTx{
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
	if tx.BlobGasFeeCap() != nil {
		blobFeeCapInt, err := types.SafeNewIntFromBigInt(tx.BlobGasFeeCap())
		if err != nil {
			return nil, err
		}
		txData.MaxFeePerBlobGas = &blobFeeCapInt
	}
	if tx.AccessList() != nil {
		al := tx.AccessList()
		txData.Accesses = NewAccessList(&al)
	}
	if blobs := tx.BlobHashes(); len(blobs) > 0 {
		hashes := make([]string, 0, len(blobs))
		for _, h := range blobs {
			hashes = append(hashes, h.Hex())
		}
		txData.BlobVersionedHashes = hashes
	}

	txData.SetSignatureValues(tx.ChainId(), v, r, s)
	return txData, nil
}

func (tx *BlobTx) TxType() uint8 { return ethtypes.BlobTxType }

func (tx *BlobTx) Copy() TxData {
	hashes := make([]string, len(tx.BlobVersionedHashes))
	copy(hashes, tx.BlobVersionedHashes)
	return &BlobTx{
		ChainID:             tx.ChainID,
		Nonce:               tx.Nonce,
		GasTipCap:           tx.GasTipCap,
		GasFeeCap:           tx.GasFeeCap,
		GasLimit:            tx.GasLimit,
		To:                  tx.To,
		Amount:              tx.Amount,
		Data:                common.CopyBytes(tx.Data),
		Accesses:            tx.Accesses,
		MaxFeePerBlobGas:    tx.MaxFeePerBlobGas,
		BlobVersionedHashes: hashes,
		V:                   common.CopyBytes(tx.V),
		R:                   common.CopyBytes(tx.R),
		S:                   common.CopyBytes(tx.S),
	}
}

func (tx *BlobTx) GetChainID() *big.Int {
	if tx.ChainID == nil {
		return nil
	}
	return tx.ChainID.BigInt()
}

func (tx *BlobTx) GetAccessList() ethtypes.AccessList {
	if tx.Accesses == nil {
		return nil
	}
	return *tx.Accesses.ToEthAccessList()
}

func (tx *BlobTx) GetData() []byte  { return common.CopyBytes(tx.Data) }
func (tx *BlobTx) GetGas() uint64   { return tx.GasLimit }
func (tx *BlobTx) GetNonce() uint64 { return tx.Nonce }

func (tx *BlobTx) GetGasPrice() *big.Int { return tx.GetGasFeeCap() }

func (tx *BlobTx) GetGasTipCap() *big.Int {
	if tx.GasTipCap == nil {
		return nil
	}
	return tx.GasTipCap.BigInt()
}

func (tx *BlobTx) GetGasFeeCap() *big.Int {
	if tx.GasFeeCap == nil {
		return nil
	}
	return tx.GasFeeCap.BigInt()
}

func (tx *BlobTx) GetBlobFeeCap() *big.Int {
	if tx.MaxFeePerBlobGas == nil {
		return nil
	}
	return tx.MaxFeePerBlobGas.BigInt()
}

func (tx *BlobTx) GetBlobHashes() []common.Hash {
	if len(tx.BlobVersionedHashes) == 0 {
		return nil
	}
	out := make([]common.Hash, 0, len(tx.BlobVersionedHashes))
	for _, h := range tx.BlobVersionedHashes {
		out = append(out, common.HexToHash(h))
	}
	return out
}

// GetSetCodeAuthorizations returns nil for non-setcode transactions.
func (tx *BlobTx) GetSetCodeAuthorizations() []ethtypes.SetCodeAuthorization {
	return nil
}

func (tx *BlobTx) GetValue() *big.Int {
	if tx.Amount == nil {
		return nil
	}
	return tx.Amount.BigInt()
}

func (tx *BlobTx) GetTo() *common.Address {
	if tx.To == "" {
		return nil
	}
	to := common.HexToAddress(tx.To)
	return &to
}

func (tx *BlobTx) AsEthereumData() ethtypes.TxData {
	v, r, s := tx.GetRawSignatureValues()
	return &ethtypes.BlobTx{
		ChainID:    uint256FromBig(tx.GetChainID()),
		Nonce:      tx.GetNonce(),
		GasTipCap:  uint256FromBig(tx.GetGasTipCap()),
		GasFeeCap:  uint256FromBig(tx.GetGasFeeCap()),
		Gas:        tx.GetGas(),
		To:         *tx.GetTo(),
		Value:      uint256FromBig(tx.GetValue()),
		Data:       tx.GetData(),
		AccessList: tx.GetAccessList(),
		BlobFeeCap: uint256FromBig(tx.GetBlobFeeCap()),
		BlobHashes: tx.GetBlobHashes(),
		V:          uint256FromBig(v),
		R:          uint256FromBig(r),
		S:          uint256FromBig(s),
	}
}

func (tx *BlobTx) GetRawSignatureValues() (v, r, s *big.Int) {
	return rawSignatureValues(tx.V, tx.R, tx.S)
}

func (tx *BlobTx) SetSignatureValues(chainID, v, r, s *big.Int) {
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

func (tx BlobTx) Validate() error {
	if tx.GasTipCap == nil || tx.GasFeeCap == nil || tx.MaxFeePerBlobGas == nil {
		return errorsmod.Wrap(ErrInvalidGasCap, "blob tx gas caps cannot be nil")
	}
	if tx.GasTipCap.IsNegative() || tx.GasFeeCap.IsNegative() || tx.MaxFeePerBlobGas.IsNegative() {
		return errorsmod.Wrap(ErrInvalidGasCap, "blob tx gas caps cannot be negative")
	}
	if tx.GetChainID() == nil {
		return errorsmod.Wrap(errortypes.ErrInvalidChainID, "chain ID must be present on blob txs")
	}
	if tx.GetBlobFeeCap().Sign() == 0 {
		return errorsmod.Wrap(ErrInvalidGasCap, "max fee per blob gas must be greater than 0")
	}
	if len(tx.GetBlobHashes()) == 0 {
		return errorsmod.Wrap(ErrInvalidGasCap, "blob tx must include blob hashes")
	}
	if tx.GasFeeCap.LT(*tx.GasTipCap) {
		return errorsmod.Wrap(ErrInvalidGasCap, "max priority fee per gas higher than max fee per gas")
	}
	return nil
}

func (tx BlobTx) Fee() *big.Int { return fee(tx.GetGasFeeCap(), tx.GasLimit) }

func (tx BlobTx) Cost() *big.Int {
	total := cost(tx.Fee(), tx.GetValue())
	blobFeeCap := tx.GetBlobFeeCap()
	if blobFeeCap == nil {
		return total
	}
	blobGas := tx.blobGas()
	if blobGas.Sign() == 0 {
		return total
	}
	blobCost := new(big.Int).Mul(blobFeeCap, blobGas)
	return new(big.Int).Add(total, blobCost)
}
func (tx *BlobTx) EffectiveGasPrice(baseFee *big.Int) *big.Int {
	return EffectiveGasPrice(baseFee, tx.GasFeeCap.BigInt(), tx.GasTipCap.BigInt())
}
func (tx BlobTx) EffectiveFee(baseFee *big.Int) *big.Int {
	return fee(tx.EffectiveGasPrice(baseFee), tx.GasLimit)
}
func (tx BlobTx) EffectiveCost(baseFee *big.Int) *big.Int {
	return cost(tx.EffectiveFee(baseFee), tx.GetValue())
}

func (tx BlobTx) blobGas() *big.Int {
	if len(tx.GetBlobHashes()) == 0 {
		return big.NewInt(0)
	}
	blobCount := uint64(len(tx.GetBlobHashes()))
	return new(big.Int).SetUint64(blobCount * params.BlobTxBlobGasPerBlob)
}

func uint256FromBig(v *big.Int) *uint256.Int {
	if v == nil {
		return new(uint256.Int)
	}
	out, _ := uint256.FromBig(v)
	return out
}
