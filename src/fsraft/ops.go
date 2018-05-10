package fsraft

import (
	"ad"
	"crypto/sha1"
	"filesystem"
	"fmt"
	"reflect"
)

// This file contains various structs to encapsulate the filesystem operations, their arguments, and their replies.

// OpType ==============================================================================================================

type OpType int

const (
	MkdirOp OpType = iota
	OpenOp
	CloseOp
	SeekOp
	ReadOp
	WriteOp
	DeleteOp
)

var opTypesToStrings = map[OpType]string{
	MkdirOp:  "Mkdir",
	OpenOp:   "Open",
	CloseOp:  "Close",
	SeekOp:   "Seek",
	ReadOp:   "Read",
	WriteOp:  "Write",
	DeleteOp: "Delete",
}

func (o OpType) String() string {
	return opTypesToStrings[o]
}

// AbstractOperation ===================================================================================================

// An operation to be performed on the filesystem.
type AbstractOperation struct {
	OpType OpType
	// Depending on the optype, some of the below args might be ignored.
	Path           string
	FileDescriptor int
	OpenMode       filesystem.OpenMode
	OpenFlags      filesystem.OpenFlags
	Offset         int
	Base           filesystem.SeekMode
	NumBytes       int
	Data           []byte
}

func (ab *AbstractOperation) String() string {
	var args string
	switch ab.OpType {
	case MkdirOp:
		args = ab.Path
	case OpenOp:
		args = fmt.Sprintf("%v, %v, %v", ab.Path, ab.OpenMode, ab.OpenFlags)
	case CloseOp:
		args = fmt.Sprintf("%d", ab.FileDescriptor)
	case SeekOp:
		args = fmt.Sprintf("%v, %v, %v", ab.FileDescriptor, ab.Offset, ab.Base)
	case ReadOp:
		args = fmt.Sprintf("%v, %v", ab.FileDescriptor, ab.NumBytes)
	case WriteOp:
		args = fmt.Sprintf("%v, %v, %+v", ab.FileDescriptor, ab.NumBytes, ab.Data)
	case DeleteOp:
		args = ab.Path
	}
	return fmt.Sprintf("%v(%v)", ab.OpType.String(), args)
}

// Asserts that the length and types of reply are valid.
func assertReplyTypesValid(opType OpType, reply interface{}) {
	arr, isArray := reply.([]interface{})
	if !isArray {
		panic(fmt.Sprintf("Called assertReplyValid(%v, %+v) expecting reply to be an arr, but it wasn't!", opType, reply))
	}
	switch opType {
	case MkdirOp:
		ad.AssertEquals(2, len(arr))
		// This is the least bad way to check that arr[0] is of type int.
		_ = arr[0].(bool) // success
		ad.AssertIsErrorOrNil(arr[1])
	case OpenOp:
	case CloseOp:
		ad.AssertEquals(2, len(arr))
		_ = arr[0].(bool) // success
		ad.AssertIsErrorOrNil(arr[1])
	case SeekOp:
		ad.AssertEquals(2, len(arr))
		_ = arr[0].(int) // newPosition
		ad.AssertIsErrorOrNil(arr[1])
	case ReadOp:
		ad.AssertEquals(3, len(arr))
		_ = arr[0].(int)    // bytesRead
		_ = arr[1].([]byte) // data
		ad.AssertIsErrorOrNil(arr[2])
	case WriteOp:
		ad.AssertEquals(2, len(arr))
		_ = arr[0].(int) // bytesWritten
		ad.AssertIsErrorOrNil(arr[1])
	case DeleteOp:
		ad.AssertEquals(2, len(arr))
		_ = arr[0].(bool) // success
		ad.AssertIsErrorOrNil(arr[1])
	}
}

// Cast a reply structure to the appropriate return type for Mkdir, panicking if the reply is malformed.
func castMkdirReply(reply interface{}) (success bool, err error) {
	arr := reply.([]interface{})
	ad.AssertEquals(2, len(arr))
	success = arr[0].(bool)
	err = ad.AssertIsErrorOrNil(arr[1])
	return success, err
}

// Cast a reply structure to the appropriate return type for Open, panicking if the reply is malformed.
func castOpenReply(reply interface{}) (fileDescriptor int, err error) {
	arr := reply.([]interface{})
	ad.AssertEquals(2, len(arr))
	fileDescriptor = arr[0].(int)
	err = ad.AssertIsErrorOrNil(arr[1])
	return fileDescriptor, err
}

// Cast a reply structure to the appropriate return type for Close, panicking if the reply is malformed.
func castCloseReply(reply interface{}) (success bool, err error) {
	arr := reply.([]interface{})
	ad.AssertEquals(2, len(arr))
	success = arr[0].(bool)
	err = ad.AssertIsErrorOrNil(arr[1])
	return success, err
}

// Cast a reply structure to the appropriate return type for Seek, panicking if the reply is malformed.
func castSeekReply(reply interface{}) (newPosition int, err error) {
	arr := reply.([]interface{})
	ad.AssertEquals(2, len(arr))
	newPosition = arr[0].(int)
	err = ad.AssertIsErrorOrNil(arr[1])
	return newPosition, err
}

// Cast a reply structure to the appropriate return type for Read, panicking if the reply is malformed.
func castReadReply(reply interface{}) (bytesRead int, data []byte, err error) {
	arr := reply.([]interface{})
	ad.AssertEquals(3, len(arr))
	bytesRead = arr[0].(int)
	data = arr[1].([]byte)
	err = ad.AssertIsErrorOrNil(arr[2])
	return bytesRead, data, err
}

// Cast a reply structure to the appropriate return type for Write, panicking if the reply is malformed.
func castWriteReply(reply interface{}) (bytesWritten int, err error) {
	arr := reply.([]interface{})
	ad.AssertEquals(2, len(arr))
	bytesWritten = arr[0].(int)
	err = ad.AssertIsErrorOrNil(arr[1])
	return bytesWritten, err
}

// Cast a reply structure to the appropriate return type for Delete, panicking if the reply is malformed.
func castDeleteReply(reply interface{}) (success bool, err error) {
	arr := reply.([]interface{})
	ad.AssertEquals(2, len(arr))
	success = arr[0].(bool)
	err = ad.AssertIsErrorOrNil(arr[1])
	return success, err
}

// OperationArgs =======================================================================================================

type OperationArgs struct {
	// Field names must start with capital letters, or else RPC will break.
	AbstractOperation AbstractOperation // the operation to be performed
	ClerkId           int64
	ClerkIndex        int   // this is the ClerkIndex-th operation submitted by this clerk (1-indexed)
	Birthday          int64 // The number of ms between the epoch and the creation time of this object. Used to ensure no hash collisions.
}

func OpArgsEquals(o1, o2 OperationArgs) bool {
	return reflect.DeepEqual(o1, o2)
}

type OpArgsHash [20]byte // 20 because a SHA256 hash is 20 bytes.

// Hash using the SHA256 hash.
func HashOpArgs(args OperationArgs) OpArgsHash {
	return sha1.Sum([]byte(fmt.Sprintf("%+v", args)))

}

// ReplyStatus =========================================================================================================

type ReplyStatus int

const (
	Unset ReplyStatus = iota
	OK
	NotLeader
	Killed
)

func (rs ReplyStatus) String() string {
	switch rs {
	case Unset:
		return "Unset"
	case OK:
		return "OK"
	case NotLeader:
		return "NotLeader"
	case Killed:
		return "Killed"
	default:
		panic(fmt.Sprintf("Unrecognized ReplyStatus %v!\n", rs))
	}
}

// OperationReply ======================================================================================================

type OperationReply struct {
	// DO NOT construct an OperationReply where the ReturnValue types do not line up with the appropriate OpType!
	ReturnValue []interface{}
	Status      ReplyStatus
}

// OperationInProgress =================================================================================================

type OperationInProgress struct {
	operationArgs OperationArgs       // The operation submitted by the clerk
	expectedIndex int                 // the index this should appear at in the log
	resultChannel chan OperationReply // The channel on which the reply will be sent
}
