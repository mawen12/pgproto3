package pgproto3

import (
	"bytes"
	"encoding/json"
)

/*
CommandComplete (B)

	Byte1('C')
		识别该消息为 comamnd-completed 消息。

	Int32
		消息的字节长度，包含本身

	String
		命令标记。通常是单个单词用于标识哪个 SQL 名称完成了。
		CommandTag 是一个以 \0 结尾的字符串。

		对于 INSERT 命令，tag 是 INSERT oid rows，rows 是被插入的行数。
		如果 rows =1，且目标表具有 OID，这 oid 曾经是插入行的对象 ID。
		但不再支持 OID 系统列，因此 oid 始终为0。

		对于 DELETE 命令，tag 是 DELETE rows，rows 是被删除的行数。

		对于 UPDATE 命令，tag 是 UPDATE rows，rows 是被更新的行数。

		对于 MERGE 命令，tag 是 MERGE rows，rows 是被插入、被更新或被删除的行数。

		对于 SLEECT 或 CREATE TABLE AS 命令，tag 是 SELECT rows，rows 是被查找到的行数。

		对于 MOVE 命令，tag 是 MOVE rows，rows 是游标位置已发生更改的行数。

		对于 FETCH 命令，tag 是 FETCH rows，rows 是从游标检索的行数。

		对于 COPY 命令，tag 是 COPY rows，rows 是被复制的行数。
		请注意，row count 是从 PostgreSQL 8.2 版本开始出现的。
*/
type CommandComplete struct {
	CommandTag []byte
}

// Backend identifies this message as sendable by the PostgreSQL backend.
func (*CommandComplete) Backend() {}

// Decode decodes src into dst. src must contain the complete message with the exception of the initial 1 byte message
// type identifier and 4 byte message length.
func (dst *CommandComplete) Decode(src []byte) error {
	// 查找第一个 \0 的位置
	idx := bytes.IndexByte(src, 0)
	// 必须位于末尾才算合法
	if idx != len(src)-1 {
		return &invalidMessageFormatErr{messageType: "CommandComplete"}
	}

	dst.CommandTag = src[:idx]

	return nil
}

// Encode encodes src into dst. dst will include the 1 byte message type identifier and the 4 byte message length.
func (src *CommandComplete) Encode(dst []byte) ([]byte, error) {
	dst, sp := beginMessage(dst, 'C')
	dst = append(dst, src.CommandTag...)
	dst = append(dst, 0)
	return finishMessage(dst, sp)
}

// MarshalJSON implements encoding/json.Marshaler.
func (src CommandComplete) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type       string
		CommandTag string
	}{
		Type:       "CommandComplete",
		CommandTag: string(src.CommandTag),
	})
}

// UnmarshalJSON implements encoding/json.Unmarshaler.
func (dst *CommandComplete) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}

	var msg struct {
		CommandTag string
	}
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	dst.CommandTag = []byte(msg.CommandTag)
	return nil
}
