package pgproto3

import (
	"encoding/binary"
	"encoding/json"
	"errors"

	"github.com/jackc/pgio"
)

const sslRequestNumber = 80877103

/*
SSLRequest (F):

	Int32(8)
		消息长度，包括自己

	Int32(80877103)
		SSL 请求代码。该值被选择为在最高有效16位中包含1234,
		在最低有效16位中包含5679。（为避免混淆，此代码不得与任何协议版本号相同）
*/
type SSLRequest struct {
}

// Frontend identifies this message as sendable by a PostgreSQL frontend.
func (*SSLRequest) Frontend() {}

func (dst *SSLRequest) Decode(src []byte) error {
	if len(src) < 4 {
		return errors.New("ssl request too short")
	}

	requestCode := binary.BigEndian.Uint32(src)

	if requestCode != sslRequestNumber {
		return errors.New("bad ssl request code")
	}

	return nil
}

// Encode encodes src into dst. dst will include the 4 byte message length.
func (src *SSLRequest) Encode(dst []byte) ([]byte, error) {
	dst = pgio.AppendInt32(dst, 8)
	dst = pgio.AppendInt32(dst, sslRequestNumber)
	return dst, nil
}

// MarshalJSON implements encoding/json.Marshaler.
func (src SSLRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type            string
		ProtocolVersion uint32
		Parameters      map[string]string
	}{
		Type: "SSLRequest",
	})
}
