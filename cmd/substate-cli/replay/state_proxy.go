// Copyright 2014 The go-ethereum Authors
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

package replay

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/substate"
)

type StateProxyDB struct {
	db       StateDB // state db
}

func NewStateProxyDB(db StateDB) StateDB {
	p := new(StateProxyDB)
	p.db = db
	return p
}

func (s *StateProxyDB) CreateAccount(addr common.Address) {
	s.db.CreateAccount(addr)
}

func (s *StateProxyDB) SubBalance(addr common.Address, amount *big.Int) {
	s.db.SubBalance(addr, amount)
}

func (s *StateProxyDB) AddBalance(addr common.Address, amount *big.Int) {
	s.db.AddBalance(addr, amount)
}

func (s *StateProxyDB) GetBalance(addr common.Address) *big.Int {
	balance:= s.db.GetBalance(addr)
	return balance
}

func (s *StateProxyDB) GetNonce(addr common.Address) uint64 {
	nonce:= s.db.GetNonce(addr)
	return nonce
}

func (s *StateProxyDB) SetNonce(addr common.Address, nonce uint64) {
	s.db.SetNonce(addr, nonce)
}

func (s *StateProxyDB) GetCodeHash(addr common.Address) common.Hash {
	hash := s.db.GetCodeHash(addr) 
	return hash
}

func (s *StateProxyDB) GetCode(addr common.Address) []byte {
	code := s.db.GetCode(addr)
	return code
}

func (s *StateProxyDB) SetCode(addr common.Address, code []byte) {
	s.db.SetCode(addr,code)
}

func (s *StateProxyDB) GetCodeSize(addr common.Address) int {
	size := s.db.GetCodeSize(addr)
	return size
}

func (s *StateProxyDB) AddRefund(gas uint64) {
	s.db.AddRefund(gas)
}

func (s *StateProxyDB) SubRefund(gas uint64) {
	s.db.SubRefund(gas)
}

func (s *StateProxyDB) GetRefund() uint64 {
	gas := s.db.GetRefund() 
	return gas
}

func (s *StateProxyDB) GetCommittedState(addr common.Address, key common.Hash) common.Hash {
	value := s.db.GetCommittedState(addr, key)
	return value 
}

func (s *StateProxyDB) GetState(addr common.Address, key common.Hash) common.Hash {
	value := s.db.GetState(addr,key) 
	return value
}

func (s *StateProxyDB) SetState(addr common.Address, key common.Hash, value common.Hash) {
	s.db.SetState(addr, key, value)
}

func (s *StateProxyDB) Suicide(addr common.Address) bool {
	ok := s.db.Suicide(addr)
	return ok
}

func (s *StateProxyDB) HasSuicided(addr common.Address) bool {
	hasSuicided := s.db.HasSuicided(addr)
	return hasSuicided
}

func (s *StateProxyDB) Exist(addr common.Address) bool {
	exists := s.db.Exist(addr)
	return exists
}

func (s *StateProxyDB) Empty(addr common.Address) bool {
	empty := s.db.Empty(addr)
	return empty
}

func (s *StateProxyDB) PrepareAccessList(sender common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	s.db.PrepareAccessList(sender, dest, precompiles, txAccesses)
}

func (s *StateProxyDB) AddressInAccessList(addr common.Address) bool {
	ok := s.db.AddressInAccessList(addr)
	return ok
}

func (s *StateProxyDB) SlotInAccessList(addr common.Address, slot common.Hash) (bool, bool) {
	addressOk, slotOk := s.db.SlotInAccessList(addr, slot) 
	return addressOk, slotOk
}

func (s *StateProxyDB) AddAddressToAccessList(addr common.Address) {
	s.db.AddAddressToAccessList(addr)
}

func (s *StateProxyDB) AddSlotToAccessList(addr common.Address, slot common.Hash) {
	s.db.AddSlotToAccessList(addr, slot)
}

func (s *StateProxyDB) RevertToSnapshot(snapshot int) {
	s.db.RevertToSnapshot(snapshot)
}

func (s *StateProxyDB) Snapshot() int {
	snapshot := s.db.Snapshot()
	return snapshot
}

func (s *StateProxyDB) AddLog(log *types.Log) {
	s.db.AddLog(log)
}

func (s *StateProxyDB) AddPreimage(addr common.Hash, image []byte) {
	s.db.AddPreimage(addr, image)
}

func (s *StateProxyDB) ForEachStorage(addr common.Address, fn func(common.Hash, common.Hash) bool) error {
	err:= s.db.ForEachStorage(addr, fn) 
	return err
}

func (s *StateProxyDB) Prepare(thash common.Hash, ti int) {
	s.db.Prepare(thash, ti)
}

func (s *StateProxyDB) Finalise(deleteEmptyObjects bool) {
	s.db.Finalise(deleteEmptyObjects)
}

func (s *StateProxyDB) IntermediateRoot(deleteEmptyObjects bool) common.Hash{
	return s.db.IntermediateRoot(deleteEmptyObjects)
}

func (s *StateProxyDB) GetLogs(hash common.Hash, blockHash common.Hash) []*types.Log{
	return s.db.GetLogs(hash, blockHash)
}

func (s *StateProxyDB) GetSubstatePostAlloc() substate.SubstateAlloc{
	return s.db.GetSubstatePostAlloc()
}
