package types

import pb "p0-sink/proto"

type DoneChannel = chan struct{}

type ErrorChannel = chan error

type BlockChannel = chan *pb.BlockWrapper

type BlockReadonlyChannel = <-chan *pb.BlockWrapper

type BatchChannel = chan *Batch

type BatchReadonlyChannel = <-chan *Batch

type ProcessedBatchChannel = chan *ProcessedBatch

type ProcessedBatchReadonlyChannel = <-chan *ProcessedBatch
