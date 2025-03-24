package lib

import (
	"fmt"
	"p0-sink/internal/errors"
	pb "p0-sink/proto"
	"strings"
)

type BlockOrderValidator struct {
	lastBlockNumber    *uint64
	lastBlockDirection *pb.Direction
	reorgAction        pb.ReorgAction
}

func NewBlockOrderValidator(
	reorgAction pb.ReorgAction,
) *BlockOrderValidator {
	return &BlockOrderValidator{
		lastBlockNumber:    nil,
		lastBlockDirection: nil,
		reorgAction:        reorgAction,
	}
}

func (v *BlockOrderValidator) Validate(block *pb.BlockWrapper) error {
	if v.isFirstBlock() {
		v.setLastBlock(block)
		return nil
	}

	err := v.validateByReorgAction(block)

	if err != nil {
		return err
	}

	v.setLastBlock(block)

	return nil
}

func (v *BlockOrderValidator) validateByReorgAction(block *pb.BlockWrapper) error {
	switch v.reorgAction {
	case pb.ReorgAction_RESEND:
		return v.validateForResendAction(block)
	case pb.ReorgAction_IGNORE:
		return v.validateForIgnoreAction(block)
	case pb.ReorgAction_ROLLBACK_AND_RESEND:
		return v.validateForRollbackAndResendAction(block)
	default:
		return errors.NewStreamTerminationError("invalid reorg action")
	}
}

func (v *BlockOrderValidator) validateForRollbackAndResendAction(block *pb.BlockWrapper) error {
	if block.Direction == pb.Direction_FORWARD && *v.lastBlockDirection == pb.Direction_FORWARD {
		if block.BlockNumber == *v.lastBlockNumber+1 {
			return nil
		}
	}

	if block.Direction == pb.Direction_BACKWARD && *v.lastBlockDirection == pb.Direction_FORWARD {
		if block.BlockNumber == *v.lastBlockNumber {
			return nil
		}
	}

	if block.Direction == pb.Direction_FORWARD && *v.lastBlockDirection == pb.Direction_BACKWARD {
		if block.BlockNumber == *v.lastBlockNumber {
			return nil
		}
	}

	if block.Direction == pb.Direction_BACKWARD && *v.lastBlockDirection == pb.Direction_BACKWARD {
		if block.BlockNumber == *v.lastBlockNumber-1 {
			return nil
		}
	}

	return v.constructError(block)
}

func (v *BlockOrderValidator) validateForIgnoreAction(block *pb.BlockWrapper) error {
	if block.Direction == pb.Direction_FORWARD && *v.lastBlockDirection == pb.Direction_FORWARD {
		if block.BlockNumber == *v.lastBlockNumber+1 {
			return nil
		}
	}

	return v.constructError(block)
}

func (v *BlockOrderValidator) validateForResendAction(block *pb.BlockWrapper) error {
	if block.Direction == pb.Direction_FORWARD && *v.lastBlockDirection == pb.Direction_FORWARD {
		return nil
	}

	return v.constructError(block)
}

func (v *BlockOrderValidator) isFirstBlock() bool {
	return v.lastBlockNumber == nil && v.lastBlockDirection == nil
}

func (v *BlockOrderValidator) setLastBlock(block *pb.BlockWrapper) {
	v.lastBlockNumber = &block.BlockNumber
	v.lastBlockDirection = &block.Direction
}

func (v *BlockOrderValidator) constructError(block *pb.BlockWrapper) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("invalid block order with reorg action: %s, ", v.reorgAction.String()))

	if !v.isFirstBlock() {
		sb.WriteString(fmt.Sprintf("last block number: %d, ", v.lastBlockNumber))
		sb.WriteString(fmt.Sprintf("last block direction: %s, ", v.lastBlockDirection.String()))
	}

	sb.WriteString(fmt.Sprintf("current block number: %d", block.BlockNumber))
	sb.WriteString(fmt.Sprintf("current block direction: %s", block.Direction.String()))

	return fmt.Errorf(sb.String())
}
