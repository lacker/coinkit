package util

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type SignedOperation struct {
	Operation

	// The type of the operation
	Type string

	// The signature to prove that the sender has signed this
	// Nil if the transaction has not been signed
	Signature string
}

func NewSignedOperation(op Operation, kp *KeyPair) *SignedOperation {
	if op == nil || reflect.ValueOf(op).IsNil() {
		Logger.Fatal("cannot sign nil operation")
	}

	if kp.PublicKey().String() != op.GetSigner() {
		Logger.Fatal("you can only sign your own operations")
	}

	bytes, err := json.Marshal(op)
	if err != nil {
		Logger.Fatal("failed to sign operation because json encoding failed")
	}
	sig := kp.Sign(op.OperationType() + string(bytes))

	return &SignedOperation{
		Operation: op,
		T:         op.OperationType(),
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

	pk, err := ReadPublicKey(op.GetSigner())
	if err != nil {
		return err
	}
	if !VerifySignature(pk, partial.Type+string(partial.Operation), partial.Signature) {
		return fmt.Errorf("invalid signature on SignedOperation")
	}

	// It's valid
	s.Operation = op
	s.Type = partial.Type
	s.Signature = partial.Signature
	return nil
}

// TODO: can we get rid of this because verification happens on decode now
func (s *SignedOperation) Verify() bool {
	if s.Operation == nil || reflect.ValueOf(s.Operation).IsNil() {
		return false
	}
	pk, err := ReadPublicKey(s.Operation.GetSigner())
	if err != nil {
		return false
	}
	bytes, err := json.Marshal(s.Operation)
	if err != nil {
		return false
	}
	if !VerifySignature(pk, s.T+string(bytes), s.Signature) {
		return false
	}
	if !s.Operation.Verify() {
		return false
	}

	return true
}

// HighestPriorityFirst is a comparator in the emirpasic/gods comparator style.
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
