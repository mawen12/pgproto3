package pgproto3

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jackc/pgio"
)

// maxMessageBodyLen is the maximum length of a message body in bytes. See PG_LARGE_MESSAGE_LIMIT in the PostgreSQL
// source. It is defined as (MaxAllocSize - 1). MaxAllocSize is defined as 0x3fffffff.

// maxMessageBodyLen 是消息体的最大长度（以字节为单位）。
// 查看 PostgreSQL 源代码中的 PG_LARGE_MESSAGE_LIMIT。
// 它被定义为 (MaxAllocSize - 1)。MaxAllocSize 定义为 0x3fffffff。
const maxMessageBodyLen = (0x3fffffff - 1)

// Message is the interface implemented by an object that can decode and encode
// a particular PostgreSQL message.

// Message 是一个接口，由能够解码和编码特定 PostgreSQL 消息的对象实现。
type Message interface {
	// Decode is allowed and expected to retain a reference to data after
	// returning (unlike encoding.BinaryUnmarshaler).

	// Decode 允许并且期望在返回后保留对 data 的引用（与 encoding.BinaryUnmarshaler 不同）。
	Decode(data []byte) error

	// Encode appends itself to dst and returns the new buffer.

	// Encode 将自身追加到 dst 中并返回新的缓冲区。
	Encode(dst []byte) ([]byte, error)
}

type FrontendMessage interface {
	Message
	Frontend() // no-op method to distinguish frontend from backend methods 区分前后端方法的空操作方法
}

type BackendMessage interface {
	Message
	Backend() // no-op method to distinguish frontend from backend methods 区分前后端方法的空操作方法
}

type AuthenticationResponseMessage interface {
	BackendMessage
	AuthenticationResponse() // no-op method to distinguish authentication responses // 区分认证响应的空操作方法
}

type invalidMessageLenErr struct {
	messageType string
	expectedLen int
	actualLen   int
}

func (e *invalidMessageLenErr) Error() string {
	return fmt.Sprintf("%s body must have length of %d, but it is %d", e.messageType, e.expectedLen, e.actualLen)
}

type invalidMessageFormatErr struct {
	messageType string
}

func (e *invalidMessageFormatErr) Error() string {
	return fmt.Sprintf("%s body is invalid", e.messageType)
}

// getValueFromJSON gets the value from a protocol message representation in JSON.
func getValueFromJSON(v map[string]string) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	if text, ok := v["text"]; ok {
		return []byte(text), nil
	}
	if binary, ok := v["binary"]; ok {
		return hex.DecodeString(binary)
	}
	return nil, errors.New("unknown protocol representation")
}

// beginMessage begines a new message of type t. It appends the message type and a placeholder for the message length to
// dst. It returns the new buffer and the position of the message length placeholder.

// beginMessage 开始一个新的消息类型 t。它将消息类型和消息长度的占位符追加到 dst 中。
// 它返回新的缓冲区和消息长度占位符的位置
func beginMessage(dst []byte, t byte) ([]byte, int) {
	dst = append(dst, t)
	sp := len(dst)
	dst = pgio.AppendInt32(dst, -1)
	return dst, sp
}

// finishMessage finishes a message that was started with beginMessage. It computes the message length and writes it to
// dst[sp]. If the message length is too large it returns an error. Otherwise it returns the final message buffer.

// finishMessage 完成一个用 beginMessage 开始的消息。
// 它计算消息长度并将其写入 dst[sp]。如果消息长度太大，它会返回一个错误。否则它返回最终的消息缓冲区。
func finishMessage(dst []byte, sp int) ([]byte, error) {
	messageBodyLen := len(dst[sp:])
	if messageBodyLen > maxMessageBodyLen {
		return nil, errors.New("message body too large")
	}
	pgio.SetInt32(dst[sp:], int32(messageBodyLen))
	return dst, nil
}
