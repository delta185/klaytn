// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/klaytn/klaytn/governance (interfaces: Engine)

// Package governance is a generated GoMock package.
package governance

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	types "github.com/klaytn/klaytn/blockchain/types"
	common "github.com/klaytn/klaytn/common"
	istanbul "github.com/klaytn/klaytn/consensus/istanbul"
	params "github.com/klaytn/klaytn/params"
	database "github.com/klaytn/klaytn/storage/database"
)

// MockEngine is a mock of Engine interface.
type MockEngine struct {
	ctrl     *gomock.Controller
	recorder *MockEngineMockRecorder
}

// MockEngineMockRecorder is the mock recorder for MockEngine.
type MockEngineMockRecorder struct {
	mock *MockEngine
}

// NewMockEngine creates a new mock instance.
func NewMockEngine(ctrl *gomock.Controller) *MockEngine {
	mock := &MockEngine{ctrl: ctrl}
	mock.recorder = &MockEngineMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEngine) EXPECT() *MockEngineMockRecorder {
	return m.recorder
}

// AddVote mocks base method.
func (m *MockEngine) AddVote(arg0 string, arg1 interface{}) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddVote", arg0, arg1)
	ret0, _ := ret[0].(bool)
	return ret0
}

// AddVote indicates an expected call of AddVote.
func (mr *MockEngineMockRecorder) AddVote(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddVote", reflect.TypeOf((*MockEngine)(nil).AddVote), arg0, arg1)
}

// BlockChain mocks base method.
func (m *MockEngine) BlockChain() blockChain {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BlockChain")
	ret0, _ := ret[0].(blockChain)
	return ret0
}

// BlockChain indicates an expected call of BlockChain.
func (mr *MockEngineMockRecorder) BlockChain() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BlockChain", reflect.TypeOf((*MockEngine)(nil).BlockChain))
}

// CanWriteGovernanceState mocks base method.
func (m *MockEngine) CanWriteGovernanceState(arg0 uint64) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CanWriteGovernanceState", arg0)
	ret0, _ := ret[0].(bool)
	return ret0
}

// CanWriteGovernanceState indicates an expected call of CanWriteGovernanceState.
func (mr *MockEngineMockRecorder) CanWriteGovernanceState(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CanWriteGovernanceState", reflect.TypeOf((*MockEngine)(nil).CanWriteGovernanceState), arg0)
}

// ClearVotes mocks base method.
func (m *MockEngine) ClearVotes(arg0 uint64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "ClearVotes", arg0)
}

// ClearVotes indicates an expected call of ClearVotes.
func (mr *MockEngineMockRecorder) ClearVotes(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ClearVotes", reflect.TypeOf((*MockEngine)(nil).ClearVotes), arg0)
}

// ContractGov mocks base method.
func (m *MockEngine) ContractGov() ReaderEngine {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ContractGov")
	ret0, _ := ret[0].(ReaderEngine)
	return ret0
}

// ContractGov indicates an expected call of ContractGov.
func (mr *MockEngineMockRecorder) ContractGov() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContractGov", reflect.TypeOf((*MockEngine)(nil).ContractGov))
}

// CurrentParams mocks base method.
func (m *MockEngine) CurrentParams() *params.GovParamSet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CurrentParams")
	ret0, _ := ret[0].(*params.GovParamSet)
	return ret0
}

// CurrentParams indicates an expected call of CurrentParams.
func (mr *MockEngineMockRecorder) CurrentParams() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CurrentParams", reflect.TypeOf((*MockEngine)(nil).CurrentParams))
}

// CurrentSetCopy mocks base method.
func (m *MockEngine) CurrentSetCopy() map[string]interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CurrentSetCopy")
	ret0, _ := ret[0].(map[string]interface{})
	return ret0
}

// CurrentSetCopy indicates an expected call of CurrentSetCopy.
func (mr *MockEngineMockRecorder) CurrentSetCopy() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CurrentSetCopy", reflect.TypeOf((*MockEngine)(nil).CurrentSetCopy))
}

// DB mocks base method.
func (m *MockEngine) DB() database.DBManager {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DB")
	ret0, _ := ret[0].(database.DBManager)
	return ret0
}

// DB indicates an expected call of DB.
func (mr *MockEngineMockRecorder) DB() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DB", reflect.TypeOf((*MockEngine)(nil).DB))
}

// EffectiveParams mocks base method.
func (m *MockEngine) EffectiveParams(arg0 uint64) (*params.GovParamSet, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EffectiveParams", arg0)
	ret0, _ := ret[0].(*params.GovParamSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EffectiveParams indicates an expected call of EffectiveParams.
func (mr *MockEngineMockRecorder) EffectiveParams(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EffectiveParams", reflect.TypeOf((*MockEngine)(nil).EffectiveParams), arg0)
}

// GetEncodedVote mocks base method.
func (m *MockEngine) GetEncodedVote(arg0 common.Address, arg1 uint64) []byte {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEncodedVote", arg0, arg1)
	ret0, _ := ret[0].([]byte)
	return ret0
}

// GetEncodedVote indicates an expected call of GetEncodedVote.
func (mr *MockEngineMockRecorder) GetEncodedVote(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEncodedVote", reflect.TypeOf((*MockEngine)(nil).GetEncodedVote), arg0, arg1)
}

// GetGovernanceChange mocks base method.
func (m *MockEngine) GetGovernanceChange() map[string]interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGovernanceChange")
	ret0, _ := ret[0].(map[string]interface{})
	return ret0
}

// GetGovernanceChange indicates an expected call of GetGovernanceChange.
func (mr *MockEngineMockRecorder) GetGovernanceChange() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGovernanceChange", reflect.TypeOf((*MockEngine)(nil).GetGovernanceChange))
}

// GetGovernanceTalliesCopy mocks base method.
func (m *MockEngine) GetGovernanceTalliesCopy() []GovernanceTallyItem {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetGovernanceTalliesCopy")
	ret0, _ := ret[0].([]GovernanceTallyItem)
	return ret0
}

// GetGovernanceTalliesCopy indicates an expected call of GetGovernanceTalliesCopy.
func (mr *MockEngineMockRecorder) GetGovernanceTalliesCopy() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetGovernanceTalliesCopy", reflect.TypeOf((*MockEngine)(nil).GetGovernanceTalliesCopy))
}

// GetTxPool mocks base method.
func (m *MockEngine) GetTxPool() txPool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTxPool")
	ret0, _ := ret[0].(txPool)
	return ret0
}

// GetTxPool indicates an expected call of GetTxPool.
func (mr *MockEngineMockRecorder) GetTxPool() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTxPool", reflect.TypeOf((*MockEngine)(nil).GetTxPool))
}

// GetVoteMapCopy mocks base method.
func (m *MockEngine) GetVoteMapCopy() map[string]VoteStatus {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVoteMapCopy")
	ret0, _ := ret[0].(map[string]VoteStatus)
	return ret0
}

// GetVoteMapCopy indicates an expected call of GetVoteMapCopy.
func (mr *MockEngineMockRecorder) GetVoteMapCopy() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVoteMapCopy", reflect.TypeOf((*MockEngine)(nil).GetVoteMapCopy))
}

// HandleGovernanceVote mocks base method.
func (m *MockEngine) HandleGovernanceVote(arg0 istanbul.ValidatorSet, arg1 []GovernanceVote, arg2 []GovernanceTallyItem, arg3 *types.Header, arg4, arg5 common.Address, arg6 bool) (istanbul.ValidatorSet, []GovernanceVote, []GovernanceTallyItem) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HandleGovernanceVote", arg0, arg1, arg2, arg3, arg4, arg5, arg6)
	ret0, _ := ret[0].(istanbul.ValidatorSet)
	ret1, _ := ret[1].([]GovernanceVote)
	ret2, _ := ret[2].([]GovernanceTallyItem)
	return ret0, ret1, ret2
}

// HandleGovernanceVote indicates an expected call of HandleGovernanceVote.
func (mr *MockEngineMockRecorder) HandleGovernanceVote(arg0, arg1, arg2, arg3, arg4, arg5, arg6 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HandleGovernanceVote", reflect.TypeOf((*MockEngine)(nil).HandleGovernanceVote), arg0, arg1, arg2, arg3, arg4, arg5, arg6)
}

// HeaderGov mocks base method.
func (m *MockEngine) HeaderGov() HeaderEngine {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HeaderGov")
	ret0, _ := ret[0].(HeaderEngine)
	return ret0
}

// HeaderGov indicates an expected call of HeaderGov.
func (mr *MockEngineMockRecorder) HeaderGov() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HeaderGov", reflect.TypeOf((*MockEngine)(nil).HeaderGov))
}

// IdxCache mocks base method.
func (m *MockEngine) IdxCache() []uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IdxCache")
	ret0, _ := ret[0].([]uint64)
	return ret0
}

// IdxCache indicates an expected call of IdxCache.
func (mr *MockEngineMockRecorder) IdxCache() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IdxCache", reflect.TypeOf((*MockEngine)(nil).IdxCache))
}

// IdxCacheFromDb mocks base method.
func (m *MockEngine) IdxCacheFromDb() []uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IdxCacheFromDb")
	ret0, _ := ret[0].([]uint64)
	return ret0
}

// IdxCacheFromDb indicates an expected call of IdxCacheFromDb.
func (mr *MockEngineMockRecorder) IdxCacheFromDb() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IdxCacheFromDb", reflect.TypeOf((*MockEngine)(nil).IdxCacheFromDb))
}

// InitGovCache mocks base method.
func (m *MockEngine) InitGovCache() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "InitGovCache")
}

// InitGovCache indicates an expected call of InitGovCache.
func (mr *MockEngineMockRecorder) InitGovCache() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitGovCache", reflect.TypeOf((*MockEngine)(nil).InitGovCache))
}

// InitLastGovStateBlkNum mocks base method.
func (m *MockEngine) InitLastGovStateBlkNum() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "InitLastGovStateBlkNum")
}

// InitLastGovStateBlkNum indicates an expected call of InitLastGovStateBlkNum.
func (mr *MockEngineMockRecorder) InitLastGovStateBlkNum() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitLastGovStateBlkNum", reflect.TypeOf((*MockEngine)(nil).InitLastGovStateBlkNum))
}

// MyVotingPower mocks base method.
func (m *MockEngine) MyVotingPower() uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MyVotingPower")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// MyVotingPower indicates an expected call of MyVotingPower.
func (mr *MockEngineMockRecorder) MyVotingPower() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MyVotingPower", reflect.TypeOf((*MockEngine)(nil).MyVotingPower))
}

// NodeAddress mocks base method.
func (m *MockEngine) NodeAddress() common.Address {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NodeAddress")
	ret0, _ := ret[0].(common.Address)
	return ret0
}

// NodeAddress indicates an expected call of NodeAddress.
func (mr *MockEngineMockRecorder) NodeAddress() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NodeAddress", reflect.TypeOf((*MockEngine)(nil).NodeAddress))
}

// PendingChanges mocks base method.
func (m *MockEngine) PendingChanges() map[string]interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PendingChanges")
	ret0, _ := ret[0].(map[string]interface{})
	return ret0
}

// PendingChanges indicates an expected call of PendingChanges.
func (mr *MockEngineMockRecorder) PendingChanges() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PendingChanges", reflect.TypeOf((*MockEngine)(nil).PendingChanges))
}

// ReadGovernance mocks base method.
func (m *MockEngine) ReadGovernance(arg0 uint64) (uint64, map[string]interface{}, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadGovernance", arg0)
	ret0, _ := ret[0].(uint64)
	ret1, _ := ret[1].(map[string]interface{})
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ReadGovernance indicates an expected call of ReadGovernance.
func (mr *MockEngineMockRecorder) ReadGovernance(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadGovernance", reflect.TypeOf((*MockEngine)(nil).ReadGovernance), arg0)
}

// SetBlockchain mocks base method.
func (m *MockEngine) SetBlockchain(arg0 blockChain) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetBlockchain", arg0)
}

// SetBlockchain indicates an expected call of SetBlockchain.
func (mr *MockEngineMockRecorder) SetBlockchain(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetBlockchain", reflect.TypeOf((*MockEngine)(nil).SetBlockchain), arg0)
}

// SetMyVotingPower mocks base method.
func (m *MockEngine) SetMyVotingPower(arg0 uint64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetMyVotingPower", arg0)
}

// SetMyVotingPower indicates an expected call of SetMyVotingPower.
func (mr *MockEngineMockRecorder) SetMyVotingPower(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetMyVotingPower", reflect.TypeOf((*MockEngine)(nil).SetMyVotingPower), arg0)
}

// SetNodeAddress mocks base method.
func (m *MockEngine) SetNodeAddress(arg0 common.Address) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetNodeAddress", arg0)
}

// SetNodeAddress indicates an expected call of SetNodeAddress.
func (mr *MockEngineMockRecorder) SetNodeAddress(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetNodeAddress", reflect.TypeOf((*MockEngine)(nil).SetNodeAddress), arg0)
}

// SetTotalVotingPower mocks base method.
func (m *MockEngine) SetTotalVotingPower(arg0 uint64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTotalVotingPower", arg0)
}

// SetTotalVotingPower indicates an expected call of SetTotalVotingPower.
func (mr *MockEngineMockRecorder) SetTotalVotingPower(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTotalVotingPower", reflect.TypeOf((*MockEngine)(nil).SetTotalVotingPower), arg0)
}

// SetTxPool mocks base method.
func (m *MockEngine) SetTxPool(arg0 txPool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetTxPool", arg0)
}

// SetTxPool indicates an expected call of SetTxPool.
func (mr *MockEngineMockRecorder) SetTxPool(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetTxPool", reflect.TypeOf((*MockEngine)(nil).SetTxPool), arg0)
}

// TotalVotingPower mocks base method.
func (m *MockEngine) TotalVotingPower() uint64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TotalVotingPower")
	ret0, _ := ret[0].(uint64)
	return ret0
}

// TotalVotingPower indicates an expected call of TotalVotingPower.
func (mr *MockEngineMockRecorder) TotalVotingPower() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TotalVotingPower", reflect.TypeOf((*MockEngine)(nil).TotalVotingPower))
}

// UpdateCurrentSet mocks base method.
func (m *MockEngine) UpdateCurrentSet(arg0 uint64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "UpdateCurrentSet", arg0)
}

// UpdateCurrentSet indicates an expected call of UpdateCurrentSet.
func (mr *MockEngineMockRecorder) UpdateCurrentSet(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCurrentSet", reflect.TypeOf((*MockEngine)(nil).UpdateCurrentSet), arg0)
}

// UpdateParams mocks base method.
func (m *MockEngine) UpdateParams(arg0 uint64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateParams", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateParams indicates an expected call of UpdateParams.
func (mr *MockEngineMockRecorder) UpdateParams(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateParams", reflect.TypeOf((*MockEngine)(nil).UpdateParams), arg0)
}

// ValidateVote mocks base method.
func (m *MockEngine) ValidateVote(arg0 *GovernanceVote) (*GovernanceVote, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateVote", arg0)
	ret0, _ := ret[0].(*GovernanceVote)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// ValidateVote indicates an expected call of ValidateVote.
func (mr *MockEngineMockRecorder) ValidateVote(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateVote", reflect.TypeOf((*MockEngine)(nil).ValidateVote), arg0)
}

// VerifyGovernance mocks base method.
func (m *MockEngine) VerifyGovernance(arg0 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyGovernance", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// VerifyGovernance indicates an expected call of VerifyGovernance.
func (mr *MockEngineMockRecorder) VerifyGovernance(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyGovernance", reflect.TypeOf((*MockEngine)(nil).VerifyGovernance), arg0)
}

// Votes mocks base method.
func (m *MockEngine) Votes() []GovernanceVote {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Votes")
	ret0, _ := ret[0].([]GovernanceVote)
	return ret0
}

// Votes indicates an expected call of Votes.
func (mr *MockEngineMockRecorder) Votes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Votes", reflect.TypeOf((*MockEngine)(nil).Votes))
}

// WriteGovernance mocks base method.
func (m *MockEngine) WriteGovernance(arg0 uint64, arg1, arg2 GovernanceSet) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteGovernance", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteGovernance indicates an expected call of WriteGovernance.
func (mr *MockEngineMockRecorder) WriteGovernance(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteGovernance", reflect.TypeOf((*MockEngine)(nil).WriteGovernance), arg0, arg1, arg2)
}

// WriteGovernanceForNextEpoch mocks base method.
func (m *MockEngine) WriteGovernanceForNextEpoch(arg0 uint64, arg1 []byte) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "WriteGovernanceForNextEpoch", arg0, arg1)
}

// WriteGovernanceForNextEpoch indicates an expected call of WriteGovernanceForNextEpoch.
func (mr *MockEngineMockRecorder) WriteGovernanceForNextEpoch(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteGovernanceForNextEpoch", reflect.TypeOf((*MockEngine)(nil).WriteGovernanceForNextEpoch), arg0, arg1)
}

// WriteGovernanceState mocks base method.
func (m *MockEngine) WriteGovernanceState(arg0 uint64, arg1 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteGovernanceState", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteGovernanceState indicates an expected call of WriteGovernanceState.
func (mr *MockEngineMockRecorder) WriteGovernanceState(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteGovernanceState", reflect.TypeOf((*MockEngine)(nil).WriteGovernanceState), arg0, arg1)
}
