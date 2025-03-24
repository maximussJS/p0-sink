package direction

import (
	"fmt"
	"p0-sink/internal/enums"
	pb "p0-sink/proto"
)

func DirectionToArrow(d *pb.Direction) enums.DirectionString {
	switch d.String() {
	case pb.Direction_FORWARD.String():
		return enums.DirectionStringForward
	case pb.Direction_BACKWARD.String():
		return enums.DirectionStringBackward
	default:
		panic(fmt.Sprintf("cannot convert direction %s to enum", d))
	}
}

func ArrowToDirection(d enums.DirectionString) pb.Direction {
	switch d {
	case enums.DirectionStringForward:
		return pb.Direction_FORWARD
	case enums.DirectionStringBackward:
		return pb.Direction_BACKWARD
	default:
		panic(fmt.Sprintf("cannot convert direction %s to enum", d))
	}
}
