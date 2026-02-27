package pgproto3

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgio"
)

const ProtocolVersionNumber = 196608 // 3.0

/*
StartupMessage (F):

	Int32
		消息长度，包括自己

	Int32(196610)
		协议版本号。最主要的 16 位是 major 版本号，次要的 16 位是 minor 版本号。
		当前版本是 3，所以版本号是 0x00030000，即 196608。

	协议版本号后一对或多对参数名称和值字符串。在name/value对之间使用零字节分隔，
	参数可以任何顺序出现，user 参数必须出现，其他参数可选。每个参数如下所示：

	String
		参数名称

		`user`
			要连接数据库的用户名，必填，没有默认值。

		`database`
			要连接的数据库，默认为用户名相同的数据库。

		`options`
			用于 backend 的命令行选项（不推荐在此处设置，推荐设置独立的运行时参数。）
			该字符串中的空格被视为分隔参数，除非使用反斜杠转义，写\\表示文字反斜杠。

		`replication`
			用于以流 replication 模式进行连接，可以发出一组 replication 命令而不是 SQL 语句。
			值可以是 `true`，`false`，或 `database`。默认为 `false`。

	String
		参数值

详细请查阅：https://www.postgresql.org/docs/current/protocol-flow.html#PROTOCOL-FLOW-SIMPLE-QUERY

后端可能返回的消息：

	ErrorResponse
		连接尝试被拒绝。然后服务器立即关闭连接。

	AuthenticationOk
		身份验证校验已完成。

	AuthenticationCleartextPassword
		Frontend 必须立即发送一个包含明文密码的 PasswordMessage。
		如果密码正确，服务器将返回 AuthenticationOk；
		否则，服务器将返回 ErrorResponse。

	AuthenticationMD5Password
		Frontend 必须立即发送一个包含使用 MD5 加密 password（包含用户名）
		然后再次使用 AuthenticationMD5Password 消息指定的 4 字节随机 salt
		再次加密的 PasswordMessage。
		如果密码正确，服务器将返回 AuthenticationOK；
		否则，服务器将返回 ErrorResponse。

		实际的 PasswordMessage 可以通过 SQL 计算得出，公式为 `concat('md5', md5(concat(md5(concat(password, username)), random-salt)))`。
		请注意：md5() 函数返回的结果是一个十六进制字符串。

	AuthenticationGSS
		Frontend 必须立即发起 GSSAPI 协商。前端将发送包含 GSSAPI 数据流的第一部分的
		GSSResponse。如果需要更多消息，服务器将使用 AuthenticationGSSContinue 进行响应。

	AuthenticationGSSContinue
		此消息包含来自 GSSAPI 或 SSPI 协商的上一步（AuthenticationGSS、AuthenticationSSPI 或先前的 AuthenticationGSSContinue）
		的响应数据。如果此消息中的 GSSAPI 或 SSPI 数据指示需要更多数据来完成身份验证，则 Frontend 必须将数据作为另一个
		GSSResponse 消息发送。如果通过此消息完成 GSSAPI 或 SSPI 身份验证，服务器接下来将发送 AuthenticationOk 来指示身份验证成功，
		或发送 ErrorResponse 来指示失败。

	AuthenticationSASL
		Frontend 必须立即使用消息中列出的 SASL 机制之一发起 SASL 协商。
		Frontend 将发送带有所选机制名称的 SASLInitialResponse 以及

如果 Frontend 不支持服务器请求的身份验证方法，则应立即关闭连接。

在收到 AuthenticationOk 的响应后，Frontend 必须等待来自 Server 的进一步消息。在这个阶段，
Backend 进程开始启动，Frontend 只是一个感兴趣的旁观者。启动尝试失败（ErrorResponse）或
服务器仍有可能拒绝对请求的 minor 版本协议（NegotiateProtocolVersion）的支持。在这种情况下，
Backend 将发送一些 ParameterStatus 消息、BackendKeyData，最后发送 ReadyForQuery。

在此阶段，后端将尝试应用启动消息中给出的任何其他运行时参数设置。如果成功，这些值将成为会话
默认值。错误会导致 ErrorResponse 并退出。

此阶段来自 Backend 的可能消息有：

	BackendKeyData
		该消息提供密钥数据，如果 Frontend 希望稍后能够发出取消请求，则必须保存该数据。
		Frontend 不应响应此消息，但应继续侦听 ReadyForQuery。

		PostgreSQL 服务器将始终发送此消息，但已知某些不支持查询取消的协议的第三方后端
		实现不会发送此消息。

	ParameterStatus

	ReadyForQuery
		启动完成，Frontend 现在可以发出命令。

	ErrorResponse

	NoticeResponse

*/

// StartupMessage 是一个客户端发送给 PostgreSQL 服务器的消息，用于初始化连接并提供必要的参数。
type StartupMessage struct {
	ProtocolVersion uint32            // 协议版本
	Parameters      map[string]string // 参数
}

// Frontend identifies this message as sendable by a PostgreSQL frontend.

// Frontend 识别出此消息可以被 PostgreSQL 前端发送。
func (*StartupMessage) Frontend() {}

// Decode decodes src into dst. src must contain the complete message with the exception of the initial 1 byte message
// type identifier and 4 byte message length.

// Decode 将 src 解码到 dst 中。src 必须包含完整的消息，除了最初的 1 字节消息类型标识符和 4 字节消息长度。
func (dst *StartupMessage) Decode(src []byte) error {
	// 长度至少5字节
	if len(src) < 4 {
		return errors.New("startup message too short")
	}

	// 取 src 前4字节作为协议版本
	dst.ProtocolVersion = binary.BigEndian.Uint32(src)
	rp := 4

	// 协议版本必须正确
	if dst.ProtocolVersion != ProtocolVersionNumber {
		return fmt.Errorf("Bad startup message version number. Expected %d, got %d", ProtocolVersionNumber, dst.ProtocolVersion)
	}

	// 解析参数
	dst.Parameters = make(map[string]string)
	for {
		// 读取 byte=0 的索引
		idx := bytes.IndexByte(src[rp:], 0)
		if idx < 0 {
			return &invalidMessageFormatErr{messageType: "StartupMesage"}
		}
		// 取出key
		key := string(src[rp : rp+idx])
		rp += idx + 1

		// 读取 byte=0 的索引
		idx = bytes.IndexByte(src[rp:], 0)
		if idx < 0 {
			return &invalidMessageFormatErr{messageType: "StartupMesage"}
		}
		// 取出value
		value := string(src[rp : rp+idx])
		rp += idx + 1

		// 存储到 dst.Parameters 中
		dst.Parameters[key] = value

		// 如果剩余的 src 只有一个字节，且该字节为0，则解析完成，否则报错。
		if len(src[rp:]) == 1 {
			if src[rp] != 0 {
				return fmt.Errorf("Bad startup message last byte. Expected 0, got %d", src[rp])
			}
			break
		}
	}

	return nil
}

// Encode encodes src into dst. dst will include the 1 byte message type identifier and the 4 byte message length.

// Encode 将 src 编码到 dst 中。dst 将包含 1 字节消息类型标识符和 4 字节消息长度。
func (src *StartupMessage) Encode(dst []byte) ([]byte, error) {
	// 获取 dst 长度
	sp := len(dst)
	dst = pgio.AppendInt32(dst, -1)

	// 写入协议版本
	dst = pgio.AppendUint32(dst, src.ProtocolVersion)

	// 写入参数，格式为 key\0value\0，最后以一个\0结尾
	for k, v := range src.Parameters {
		dst = append(dst, k...)
		dst = append(dst, 0)
		dst = append(dst, v...)
		dst = append(dst, 0)
	}
	// 消息末尾以\0结尾
	dst = append(dst, 0)

	// 最后在开头写入长度
	return finishMessage(dst, sp)
}

// MarshalJSON implements encoding/json.Marshaler.
func (src StartupMessage) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type            string
		ProtocolVersion uint32
		Parameters      map[string]string
	}{
		Type:            "StartupMessage",
		ProtocolVersion: src.ProtocolVersion,
		Parameters:      src.Parameters,
	})
}
