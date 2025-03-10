package types

import pb "p0-sink/proto"

type ErrorChannel = chan error

type BlockWrapperChannel = chan *pb.BlockWrapper

type BlockWrapperReadChannel = <-chan *pb.BlockWrapper

type DownloadedBlockChannel = chan *DownloadedBlock

type DownloadedBlockReadChannel = <-chan *DownloadedBlock

type BatchChannel = chan *Batch

type BatchReadChannel = <-chan *Batch

type SerializedBatchChannel = chan *SerializedBatch

type SerializedBatchReadChannel = <-chan *SerializedBatch
