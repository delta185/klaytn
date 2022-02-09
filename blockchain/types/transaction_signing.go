// Modifications Copyright 2018 The klaytn Authors
// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from core/types/transaction_signing.go (2018/06/04).
// Modified and improved for the klaytn development.

package types

import (
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/klaytn/klaytn/blockchain/types/accountkey"
	"github.com/klaytn/klaytn/common"
	"github.com/klaytn/klaytn/crypto"
	"github.com/klaytn/klaytn/params"
)

var (
	ErrInvalidChainId        = errors.New("invalid chain id for signer")
	errNotTxInternalDataFrom = errors.New("not an TxInternalDataFrom")
)

// sigCache is used to cache the derived sender and contains
// the signer used to derive it.
type sigCache struct {
	signer Signer
	from   common.Address
}

// sigCachePubkey is used to cache the derived public key and contains
// the signer used to derive it.
type sigCachePubkey struct {
	signer Signer
	pubkey []*ecdsa.PublicKey
}

// TODO-Klaytn-RemoveLater Remove the second parameter blockNumber
// MakeSigner returns a Signer based on the given chain config and block number.
func MakeSigner(config *params.ChainConfig, blockNumber *big.Int) Signer {
	return NewEIP155Signer(config.ChainID)
}

// LatestSigner returns the 'most permissive' Signer available for the given chain
// configuration.
//
// Use this in transaction-handling code where the current block number is unknown. If you
// have the current block number available, use MakeSigner instead.
func LatestSigner(config *params.ChainConfig) Signer {
	return NewEIP155Signer(config.ChainID)
}

// LatestSignerForChainID returns the 'most permissive' Signer available.
//
// Use this in transaction-handling code where the current block number and fork
// configuration are unknown. If you have a ChainConfig, use LatestSigner instead.
// If you have a ChainConfig and know the current block number, use MakeSigner instead.
func LatestSignerForChainID(chainID *big.Int) Signer {
	return NewEIP155Signer(chainID)
}

// SignTx signs the transaction using the given signer and private key
func SignTx(tx *Transaction, s Signer, prv *ecdsa.PrivateKey) (*Transaction, error) {
	h := s.Hash(tx)
	sig, err := crypto.Sign(h[:], prv)
	if err != nil {
		return nil, err
	}
	return tx.WithSignature(s, sig)
}

// SignTxAsFeePayer signs the transaction as a fee payer using the given signer and private key
func SignTxAsFeePayer(tx *Transaction, s Signer, prv *ecdsa.PrivateKey) (*Transaction, error) {
	h, err := s.HashFeePayer(tx)
	if err != nil {
		return nil, err
	}
	sig, err := crypto.Sign(h[:], prv)
	if err != nil {
		return nil, err
	}
	return tx.WithFeePayerSignature(s, sig)
}

// AccountKeyPicker has a function GetKey() to retrieve an account key from statedb.
type AccountKeyPicker interface {
	GetKey(address common.Address) accountkey.AccountKey
	Exist(addr common.Address) bool
}

// Sender returns the address of the transaction.
// If a legacy transaction, it calls SenderFrom().
// Otherwise, it just returns tx.From() because the other transaction types have the field `from`.
// NOTE: this function should not be called if tx signature validation is required.
// In that situtation, you should call ValidateSender().
func Sender(signer Signer, tx *Transaction) (common.Address, error) {
	if tx.IsLegacyTransaction() {
		return SenderFrom(signer, tx)
	}

	return tx.From()
}

// SenderFeePayer returns the fee payer address of the transaction.
// If the transaction is not a fee-delegated transaction, the fee payer is set to
// the address of the `from` of the transaction.
func SenderFeePayer(signer Signer, tx *Transaction) (common.Address, error) {
	tf, ok := tx.data.(TxInternalDataFeePayer)
	if !ok {
		return Sender(signer, tx)
	}
	return tf.GetFeePayer(), nil
}

// SenderFeePayerPubkey returns the public key derived from the signature (V, R, S) using secp256k1
// elliptic curve and an error if it failed deriving or upon an incorrect
// signature.
//
// SenderFeePayerPubkey may cache the public key, allowing it to be used regardless of
// signing method. The cache is invalidated if the cached signer does
// not match the signer used in the current call.
func SenderFeePayerPubkey(signer Signer, tx *Transaction) ([]*ecdsa.PublicKey, error) {
	if sc := tx.feePayer.Load(); sc != nil {
		sigCache := sc.(sigCachePubkey)
		// If the signer used to derive from in a previous
		// call is not the same as used current, invalidate
		// the cache.
		if sigCache.signer.Equal(signer) {
			return sigCache.pubkey, nil
		}
	}

	pubkey, err := signer.SenderFeePayer(tx)
	if err != nil {
		return nil, err
	}

	tx.feePayer.Store(sigCachePubkey{signer: signer, pubkey: pubkey})
	return pubkey, nil
}

// SenderFrom returns the address derived from the signature (V, R, S) using secp256k1
// elliptic curve and an error if it failed deriving or upon an incorrect
// signature.
//
// SenderFrom may cache the address, allowing it to be used regardless of
// signing method. The cache is invalidated if the cached signer does
// not match the signer used in the current call.
func SenderFrom(signer Signer, tx *Transaction) (common.Address, error) {
	if sc := tx.from.Load(); sc != nil {
		sigCache := sc.(sigCache)
		// If the signer used to derive from in a previous
		// call is not the same as used current, invalidate
		// the cache.
		if sigCache.signer.Equal(signer) {
			return sigCache.from, nil
		}
	}

	addr, err := signer.Sender(tx)
	if err != nil {
		return common.Address{}, err
	}
	tx.from.Store(sigCache{signer: signer, from: addr})
	return addr, nil
}

// SenderPubkey returns the public key derived from the signature (V, R, S) using secp256k1
// elliptic curve and an error if it failed deriving or upon an incorrect
// signature.
//
// SenderPubkey may cache the public key, allowing it to be used regardless of
// signing method. The cache is invalidated if the cached signer does
// not match the signer used in the current call.
func SenderPubkey(signer Signer, tx *Transaction) ([]*ecdsa.PublicKey, error) {
	if sc := tx.from.Load(); sc != nil {
		sigCache := sc.(sigCachePubkey)
		// If the signer used to derive from in a previous
		// call is not the same as used current, invalidate
		// the cache.
		if sigCache.signer.Equal(signer) {
			return sigCache.pubkey, nil
		}
	}

	pubkey, err := signer.SenderPubkey(tx)
	if err != nil {
		return nil, err
	}
	tx.from.Store(sigCachePubkey{signer: signer, pubkey: pubkey})
	return pubkey, nil
}

// Signer encapsulates transaction signature handling. Note that this interface is not a
// stable API and may change at any time to accommodate new protocol rules.
type Signer interface {
	// Sender returns the sender address of the transaction.
	Sender(tx *Transaction) (common.Address, error)
	// SenderPubkey returns the public key derived from tx signature and txhash.
	SenderPubkey(tx *Transaction) ([]*ecdsa.PublicKey, error)
	// SenderFeePayer returns the public key derived from tx signature and txhash.
	SenderFeePayer(tx *Transaction) ([]*ecdsa.PublicKey, error)
	// SignatureValues returns the raw R, S, V values corresponding to the
	// given signature.
	SignatureValues(sig []byte) (r, s, v *big.Int, err error)
	// Hash returns the hash to be signed.
	Hash(tx *Transaction) common.Hash
	// HashFeePayer returns the hash with a fee payer's address to be signed by a fee payer.
	HashFeePayer(tx *Transaction) (common.Hash, error)
	// Equal returns true if the given signer is the same as the receiver.
	Equal(Signer) bool
}

// EIP155Transaction implements Signer using the EIP155 rules.
type EIP155Signer struct {
	chainId, chainIdMul *big.Int
}

func NewEIP155Signer(chainId *big.Int) EIP155Signer {
	if chainId == nil {
		chainId = new(big.Int)
	}
	return EIP155Signer{
		chainId:    chainId,
		chainIdMul: new(big.Int).Mul(chainId, common.Big2),
	}
}

func (s EIP155Signer) Equal(s2 Signer) bool {
	eip155, ok := s2.(EIP155Signer)
	return ok && eip155.chainId.Cmp(s.chainId) == 0
}

var big8 = big.NewInt(8)

func (s EIP155Signer) Sender(tx *Transaction) (common.Address, error) {
	if !tx.IsLegacyTransaction() {
		b, _ := json.Marshal(tx)
		logger.Warn("No need to execute Sender!", "tx", string(b))
	}

	if tx.ChainId().Cmp(s.chainId) != 0 {
		return common.Address{}, ErrInvalidChainId
	}
	return tx.data.RecoverAddress(s.Hash(tx), true, func(v *big.Int) *big.Int {
		V := new(big.Int).Sub(v, s.chainIdMul)
		return V.Sub(V, big8)
	})
}

func (s EIP155Signer) SenderPubkey(tx *Transaction) ([]*ecdsa.PublicKey, error) {
	if tx.IsLegacyTransaction() {
		b, _ := json.Marshal(tx)
		logger.Warn("No need to execute SenderPubkey!", "tx", string(b))
	}

	if tx.ChainId().Cmp(s.chainId) != 0 {
		return nil, ErrInvalidChainId
	}
	return tx.data.RecoverPubkey(s.Hash(tx), true, func(v *big.Int) *big.Int {
		V := new(big.Int).Sub(v, s.chainIdMul)
		return V.Sub(V, big8)
	})
}

func (s EIP155Signer) SenderFeePayer(tx *Transaction) ([]*ecdsa.PublicKey, error) {
	if tx.IsLegacyTransaction() {
		b, _ := json.Marshal(tx)
		logger.Warn("No need to execute SenderFeePayer!", "tx", string(b))
	}

	if tx.ChainId().Cmp(s.chainId) != 0 {
		return nil, ErrInvalidChainId
	}

	tf, ok := tx.data.(TxInternalDataFeePayer)
	if !ok {
		return nil, errNotFeeDelegationTransaction
	}

	hash, err := s.HashFeePayer(tx)
	if err != nil {
		return nil, err
	}

	return tf.RecoverFeePayerPubkey(hash, true, func(v *big.Int) *big.Int {
		V := new(big.Int).Sub(v, s.chainIdMul)
		return V.Sub(V, big8)
	})
}

// WithSignature returns a new transaction with the given signature. This signature
// needs to be in the [R || S || V] format where V is 0 or 1.
func (s EIP155Signer) SignatureValues(sig []byte) (R, S, V *big.Int, err error) {
	if len(sig) != crypto.SignatureLength {
		panic(fmt.Sprintf("wrong size for signature: got %d, want %d", len(sig), crypto.SignatureLength))
	}
	R = new(big.Int).SetBytes(sig[:32])
	S = new(big.Int).SetBytes(sig[32:64])
	V = big.NewInt(int64(sig[crypto.RecoveryIDOffset] + 35))
	V.Add(V, s.chainIdMul)
	return R, S, V, nil
}

// Hash returns the hash to be signed by the sender.
// It does not uniquely identify the transaction.
func (s EIP155Signer) Hash(tx *Transaction) common.Hash {
	// If the data object implements SerializeForSignToByte(), use it.
	if ser, ok := tx.data.(TxInternalDataSerializeForSignToByte); ok {
		return rlpHash(struct {
			Byte    []byte
			ChainId *big.Int
			R       uint
			S       uint
		}{
			ser.SerializeForSignToBytes(),
			s.chainId,
			uint(0),
			uint(0),
		})
	}

	infs := append(tx.data.SerializeForSign(),
		s.chainId, uint(0), uint(0))
	return rlpHash(infs)
}

// HashFeePayer returns the hash with a fee payer's address to be signed by a fee payer.
// It does not uniquely identify the transaction.
func (s EIP155Signer) HashFeePayer(tx *Transaction) (common.Hash, error) {
	tf, ok := tx.data.(TxInternalDataFeePayer)
	if !ok {
		return common.Hash{}, errNotFeeDelegationTransaction
	}

	// If the data object implements SerializeForSignToByte(), use it.
	if ser, ok := tx.data.(TxInternalDataSerializeForSignToByte); ok {
		return rlpHash(struct {
			Byte     []byte
			FeePayer common.Address
			ChainId  *big.Int
			R        uint
			S        uint
		}{
			ser.SerializeForSignToBytes(),
			tf.GetFeePayer(),
			s.chainId,
			uint(0),
			uint(0),
		}), nil
	}

	infs := append(tx.data.SerializeForSign(),
		tf.GetFeePayer(),
		s.chainId, uint(0), uint(0))
	return rlpHash(infs), nil
}

func recoverPlainCommon(sighash common.Hash, R, S, Vb *big.Int, homestead bool) ([]byte, error) {
	if Vb.BitLen() > 8 {
		return []byte{}, ErrInvalidSig
	}
	V := byte(Vb.Uint64() - 27)
	if !crypto.ValidateSignatureValues(V, R, S, homestead) {
		return []byte{}, ErrInvalidSig
	}
	// encode the snature in uncompressed format
	r, s := R.Bytes(), S.Bytes()
	sig := make([]byte, crypto.SignatureLength)
	copy(sig[32-len(r):32], r)
	copy(sig[64-len(s):64], s)
	sig[crypto.RecoveryIDOffset] = V
	// recover the public key from the snature
	pub, err := crypto.Ecrecover(sighash[:], sig)
	if err != nil {
		return []byte{}, err
	}
	if len(pub) == 0 || pub[0] != 4 {
		return []byte{}, errors.New("invalid public key")
	}
	return pub, nil
}

func recoverPlain(sighash common.Hash, R, S, Vb *big.Int, homestead bool) (common.Address, error) {
	pub, err := recoverPlainCommon(sighash, R, S, Vb, homestead)
	if err != nil {
		return common.Address{}, err
	}

	var addr common.Address
	copy(addr[:], crypto.Keccak256(pub[1:])[12:])
	return addr, nil
}

func recoverPlainPubkey(sighash common.Hash, R, S, Vb *big.Int, homestead bool) (*ecdsa.PublicKey, error) {
	pub, err := recoverPlainCommon(sighash, R, S, Vb, homestead)
	if err != nil {
		return nil, err
	}

	pubkey, err := crypto.UnmarshalPubkey(pub)
	if err != nil {
		return nil, err
	}

	return pubkey, nil
}

// deriveChainId derives the chain id from the given v parameter
func deriveChainId(v *big.Int) *big.Int {
	if v.BitLen() <= 64 {
		v := v.Uint64()
		if v == 27 || v == 28 {
			return new(big.Int)
		}
		return new(big.Int).SetUint64((v - 35) / 2)
	}
	v = new(big.Int).Sub(v, big.NewInt(35))
	return v.Div(v, common.Big2)
}
