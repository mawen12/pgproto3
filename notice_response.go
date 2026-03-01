package pgproto3

/*
NoticeResponse (B)

	Byte1('N')
		标识消息是一个 notice。

	Int32
		消息的字节长度，包含本身。

消息体由一个或多个标识符字段组成，并跟随一个0字节作为终止符。
字段可能以任意顺序出现，每个字段如下：

	Byte1
		标识字段类型的代码，如果是0，该消息终止且没有后续字符串。

	String
		字段值。
*/
type NoticeResponse ErrorResponse

// Backend identifies this message as sendable by the PostgreSQL backend.
func (*NoticeResponse) Backend() {}

// Decode decodes src into dst. src must contain the complete message with the exception of the initial 1 byte message
// type identifier and 4 byte message length.
func (dst *NoticeResponse) Decode(src []byte) error {
	return (*ErrorResponse)(dst).Decode(src)
}

// Encode encodes src into dst. dst will include the 1 byte message type identifier and the 4 byte message length.
func (src *NoticeResponse) Encode(dst []byte) ([]byte, error) {
	dst, sp := beginMessage(dst, 'N')
	dst = (*ErrorResponse)(src).appendFields(dst)
	return finishMessage(dst, sp)
}
