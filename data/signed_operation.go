package data

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/lacker/coinkit/util"
)

type SignedOperation struct {
	Operation

	// The type of the operation
	Type string

	// The signature to prove that the sender has signed this
	// Nil if the operation has not been signed
	Signature string
}

func NewSignedOperation(op Operation, kp *util.KeyPair) *SignedOperation {
	if op == nil || reflect.ValueOf(op).IsNil() {
		util.Logger.Fatal("cannot sign nil operation")
	}

	if kp.PublicKey().String() != op.GetSigner() {
		util.Logger.Fatal("you can only sign your own operations")
	}

	bytes, err := json.Marshal(op)
	if err != nil {
		util.Logger.Fatal("failed to sign operation because json encoding failed")
	}
	sig := kp.Sign(op.OperationType() + string(bytes))

	return &SignedOperation{
		Operation: op,
		Type:      op.OperationType(),
		Signature: sig,
	}
}

type partiallyUnmarshaledSignedOperation struct {
	Operation json.RawMessage
	Type      string
	Signature string
}

func (s *SignedOperation) UnmarshalJSON(data []byte) error {
	var partial partiallyUnmarshaledSignedOperation
	err := json.Unmarshal(data, &partial)
	if err != nil {
		return err
	}
	opType, ok := OperationTypeMap[partial.Type]
	if !ok {
		return fmt.Errorf("unregistered op type: %s", partial.Type)
	}
	op := reflect.New(opType).Interface().(Operation)
	err = json.Unmarshal(partial.Operation, &op)
	if err != nil {
		return err
	}
	if op == nil {
		return fmt.Errorf("decoding a nil operation is not valid")
	}
	if !op.Verify() {
		return fmt.Errorf("the decoded operation was invalid")
	}

	pk, err := util.ReadPublicKey(op.GetSigner())
	if err != nil {
		return err
	}
	if !util.VerifySignature(pk, partial.Type+string(partial.Operation),
		partial.Signature) {
		return fmt.Errorf("invalid signature on SignedOperation")
	}

	// It's valid
	s.Operation = op
	s.Type = partial.Type
	s.Signature = partial.Signature
	return nil
}

// HighestFeeFirst is a comparator in the emirpasic/gods comparator style.
// Negative return indicates a < b
// Positive return indicates a > b
// Comparison indicates overall "priority" putting the highest priority first.
// This means that when a has a higher fee than b, a < b.
func HighestFeeFirst(a, b interface{}) int {
	s1 := a.(*SignedOperation)
	s2 := b.(*SignedOperation)

	switch {
	case s1.Operation.GetFee() > s2.Operation.GetFee():
		// s1 is higher priority. so a < b
		return -1
	case s1.Operation.GetFee() < s2.Operation.GetFee():
		return 1
	case s1.Signature < s2.Signature:
		// s1 is higher priority
		return -1
	case s1.Signature > s2.Signature:
		return 1
	default:
		return 0
	}
}
