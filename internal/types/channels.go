package types

import pb "p0-sink/proto"

type ErrorChannel = chan error

type BlockWrapperChannel = chan *pb.BlockWrapper

type BlockWrapperReadChannel = <-chan *pb.BlockWrapper
