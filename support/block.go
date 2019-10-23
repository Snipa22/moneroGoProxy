package support

import (
	"encoding/binary"
	"encoding/hex"
	"github.com/snipa22/monerocnutils"
	"github.com/snipa22/monerocnutils/serialization"
	"math/rand"
)

type BlockTemplateWire struct {
	BlockTemplateBlob string `json:"blocktemplate_blob"`  // The auctual block template, this is where things get spicy
	Difficulty        uint64 `json:"difficulty"`          // Current block diff
	ExpectedReward    uint64 `json:"expected_reward"`     // Expected reward for the block
	Height            uint32 `json:"height"`              // Mining height
	PreviousHash      string `json:"prev_hash"`           // Hash of the last found block
	ReservedOffset    uint8  `json:"reserved_offset"`     // Offset in the data structure where the requested bytes are
	WorkerOffset      uint8  `json:"client_nonce_offset"` // Offset at which the worker's ID should be slammed in
	PoolOffset        uint8  `json:"client_pool_offset"`  // Offset at which the pool's ID should be slammed in
	TargetDiff        uint64 `json:"target_diff"`         // Target difficulty for the job
	TargetHex         string `json:"target_hex"`          // Target difficulty in hex
	JobID             string `json:"job_id"`              // JobID to send to the server
}

type BlockTemplate struct {
	serialization.Block
	Difficulty     uint64
	Height         uint32
	ReservedOffset uint8
	WorkerOffset   uint8
	PoolOffset     uint8
	TargetDiff     uint64
	TargetHex      string
	PreviousHash   []byte
	JobID          string
	WorkerNonce    uint32
	PoolNonce      uint32
	Solo           bool
}

func (b *BlockTemplate) GetWorkerBlob() string {
	b.PoolNonce += 1
	binary.BigEndian.PutUint32(b.Block.MinerTxn.Extra[8:12], b.PoolNonce)
	return hex.EncodeToString(b.Serialize())
}

func (b *BlockTemplate) NextBlob() string {
	b.WorkerNonce += 1
	if b.Solo {
		binary.BigEndian.PutUint32(b.Block.MinerTxn.Extra[0:4], b.WorkerNonce)
	} else {
		binary.BigEndian.PutUint32(b.Block.MinerTxn.Extra[12:16], b.WorkerNonce)
	}
	return hex.EncodeToString(b.Serialize())
}

func (b *BlockTemplate) GetJob(m *Miner, bc bool) MinerJob {
	if !bc && m.NewDiff > 0 && len(m.CachedJob.ID) > 3 {
		return m.CachedJob
	}
	token := make([]byte, 21)
	rand.Read(token)

	if m.NewDiff > 0 {
		m.Difficulty = m.NewDiff
		m.NewDiff = 0
	}

	_, st := getTarget(m.Difficulty)

	var j MinerJob = MinerJob{
		ID:            hex.EncodeToString(token),
		BlockTemplate: *b,
		HashBlob:      b.GetWorkerBlob(),
		Target:        st,
		JobNonce:      b.WorkerNonce,
		Difficulty:    m.Difficulty,
		Submissions:   []string{},
	}

	m.CachedJob = j
	return j
}

func NewBlockTemplate(btw BlockTemplateWire, instanceID uint32) (BlockTemplate, error) {
	var err error
	bt := BlockTemplate{
		Difficulty:     btw.Difficulty,
		Height:         btw.Height,
		ReservedOffset: btw.ReservedOffset,
		WorkerOffset:   btw.WorkerOffset,
		PoolOffset:     btw.PoolOffset,
		TargetDiff:     btw.TargetDiff,
		TargetHex:      btw.TargetHex,
		JobID:          btw.JobID,
		WorkerNonce:    0,
		PoolNonce:      0,
	}
	bt.PreviousHash, err = hex.DecodeString(btw.PreviousHash)
	if err != nil {
		return bt, err
	}
	t, err := monerocnutils.ParseBlockFromTemplateBlob(btw.BlockTemplateBlob)
	if err != nil {
		return bt, err
	}
	bt.Block = t
	if bt.WorkerOffset == 0 {
		bt.Solo = true
		binary.LittleEndian.PutUint32(bt.Block.MinerTxn.Extra[4:8], instanceID)
		copy(bt.Block.PreviousID[:], bt.PreviousHash)
	}
	return bt, nil
}
