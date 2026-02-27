package pgproto3

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Frontend acts as a client for the PostgreSQL wire protocol version 3.

/*
Frontend 作为 PostgreSQL 版本 3 的客户端实现。

	ChunkReader

	io.Writer

	bodyLen
		从 Header 中读取的消息体的长度

	msgType
		消息类型

	partialMsg
		表示 header 已读，但是 body 未读的标志位

	authType
*/
type Frontend struct {
	cr ChunkReader
	w  io.Writer

	// Backend message flyweights
	authenticationOk                AuthenticationOk
	authenticationCleartextPassword AuthenticationCleartextPassword
	authenticationMD5Password       AuthenticationMD5Password
	authenticationGSS               AuthenticationGSS
	authenticationGSSContinue       AuthenticationGSSContinue
	authenticationSASL              AuthenticationSASL
	authenticationSASLContinue      AuthenticationSASLContinue
	authenticationSASLFinal         AuthenticationSASLFinal
	backendKeyData                  BackendKeyData
	bindComplete                    BindComplete
	closeComplete                   CloseComplete
	commandComplete                 CommandComplete
	copyBothResponse                CopyBothResponse
	copyData                        CopyData
	copyInResponse                  CopyInResponse
	copyOutResponse                 CopyOutResponse
	copyDone                        CopyDone
	dataRow                         DataRow
	emptyQueryResponse              EmptyQueryResponse
	errorResponse                   ErrorResponse
	functionCallResponse            FunctionCallResponse
	noData                          NoData
	noticeResponse                  NoticeResponse
	notificationResponse            NotificationResponse
	parameterDescription            ParameterDescription
	parameterStatus                 ParameterStatus
	parseComplete                   ParseComplete
	readyForQuery                   ReadyForQuery
	rowDescription                  RowDescription
	portalSuspended                 PortalSuspended

	bodyLen    int
	msgType    byte
	partialMsg bool // 表示 header 已读，但 body 未读的标志
	authType   uint32
}

// NewFrontend creates a new Frontend.
func NewFrontend(cr ChunkReader, w io.Writer) *Frontend {
	return &Frontend{cr: cr, w: w}
}

// Send sends a message to the backend.

// Send 向后端发送消息。
func (f *Frontend) Send(msg FrontendMessage) error {
	// 将 msg 编码到 buf 中
	buf, err := msg.Encode(nil)
	if err != nil {
		return err
	}
	// 写入后端
	_, err = f.w.Write(buf)
	return err
}

func translateEOFtoErrUnexpectedEOF(err error) error {
	if err == io.EOF {
		return io.ErrUnexpectedEOF
	}
	return err
}

// Receive receives a message from the backend. The returned message is only valid until the next call to Receive.

// Receive 从后端接收消息。返回的消息仅在下一次调用 Receive 之前有效。
func (f *Frontend) Receive() (BackendMessage, error) {
	if !f.partialMsg {
		// 读取消息头，包含一个字节消息类型和四个字节消息长度
		header, err := f.cr.Next(5)
		if err != nil {
			return nil, translateEOFtoErrUnexpectedEOF(err)
		}

		// 消息类型
		f.msgType = header[0]
		// 消息体长度
		f.bodyLen = int(binary.BigEndian.Uint32(header[1:])) - 4
		f.partialMsg = true
		// 如果消息体长度为负数，则说明消息格式错误，返回错误。
		if f.bodyLen < 0 {
			return nil, errors.New("invalid message with negative body length received")
		}
	}

	// 读取指定长度的消息体
	msgBody, err := f.cr.Next(f.bodyLen)
	if err != nil {
		return nil, translateEOFtoErrUnexpectedEOF(err)
	}

	// 读取完毕后重置为初始状态
	f.partialMsg = false

	var msg BackendMessage
	// 协议解析
	switch f.msgType {
	case '1': // 解析完成
		msg = &f.parseComplete
	case '2': // 绑定完成
		msg = &f.bindComplete
	case '3': // 关闭完成
		msg = &f.closeComplete
	case 'A': // 通知响应
		msg = &f.notificationResponse
	case 'c': // 复制完成
		msg = &f.copyDone
	case 'C': // 命令完成
		msg = &f.commandComplete
	case 'd': // 复制数据
		msg = &f.copyData
	case 'D': // 数据行
		msg = &f.dataRow
	case 'E': // 错误响应
		msg = &f.errorResponse
	case 'G': // 复制输入响应
		msg = &f.copyInResponse
	case 'H': // 复制输出响应
		msg = &f.copyOutResponse
	case 'I': // 空查询响应
		msg = &f.emptyQueryResponse
	case 'K': // 后端密钥数据
		msg = &f.backendKeyData
	case 'n': // 无数据
		msg = &f.noData
	case 'N': // 通知响应
		msg = &f.noticeResponse
	case 'R': // 认证响应
		var err error
		msg, err = f.findAuthenticationMessageType(msgBody)
		if err != nil {
			return nil, err
		}
	case 's': // 门户挂起
		msg = &f.portalSuspended
	case 'S': // 参数状态
		msg = &f.parameterStatus
	case 't': // 参数描述
		msg = &f.parameterDescription
	case 'T': // 行描述
		msg = &f.rowDescription
	case 'V': // 函数调用响应
		msg = &f.functionCallResponse
	case 'W': // 复制双向响应
		msg = &f.copyBothResponse
	case 'Z': // 准备就绪查询
		msg = &f.readyForQuery
	default:
		return nil, fmt.Errorf("unknown message type: %c", f.msgType)
	}

	// 将消息体解码到 msg 中
	err = msg.Decode(msgBody)
	return msg, err
}

// Authentication message type constants.
// See src/include/libpq/pqcomm.h for all
// constants.
const (
	AuthTypeOk                = 0
	AuthTypeCleartextPassword = 3
	AuthTypeMD5Password       = 5
	AuthTypeSCMCreds          = 6
	AuthTypeGSS               = 7
	AuthTypeGSSCont           = 8
	AuthTypeSSPI              = 9
	AuthTypeSASL              = 10
	AuthTypeSASLContinue      = 11
	AuthTypeSASLFinal         = 12
)

func (f *Frontend) findAuthenticationMessageType(src []byte) (BackendMessage, error) {
	if len(src) < 4 {
		return nil, errors.New("authentication message too short")
	}
	f.authType = binary.BigEndian.Uint32(src[:4])

	switch f.authType {
	case AuthTypeOk:
		return &f.authenticationOk, nil
	case AuthTypeCleartextPassword:
		return &f.authenticationCleartextPassword, nil
	case AuthTypeMD5Password:
		return &f.authenticationMD5Password, nil
	case AuthTypeSCMCreds:
		return nil, errors.New("AuthTypeSCMCreds is unimplemented")
	case AuthTypeGSS:
		return &f.authenticationGSS, nil
	case AuthTypeGSSCont:
		return &f.authenticationGSSContinue, nil
	case AuthTypeSSPI:
		return nil, errors.New("AuthTypeSSPI is unimplemented")
	case AuthTypeSASL:
		return &f.authenticationSASL, nil
	case AuthTypeSASLContinue:
		return &f.authenticationSASLContinue, nil
	case AuthTypeSASLFinal:
		return &f.authenticationSASLFinal, nil
	default:
		return nil, fmt.Errorf("unknown authentication type: %d", f.authType)
	}
}

// GetAuthType returns the authType used in the current state of the frontend.
// See SetAuthType for more information.
func (f *Frontend) GetAuthType() uint32 {
	return f.authType
}
