package codec

import (
	"encoding/json"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/encoding"
)

func init() {
	encoding.RegisterCodec(JSON{})
}

type JSON struct {
}

func (_ JSON) Name() string {
	return "json"
}

func (j JSON) Marshal(v interface{}) ([]byte, error) {

	switch val := v.(type) {
	case proto.Message:
		return json.Marshal(v)
	case []byte:
		// 已经是tyte数组了，不需要转换了
		return val, nil
	default:
		return json.Marshal(v)
	}

}

func (j JSON) Unmarshal(data []byte, v interface{}) error {

	switch val := v.(type) {
	case proto.Message:
		return json.Unmarshal(data, v)
	case *[]byte:
		*val = data
		return nil
	default:

		return json.Unmarshal(data, v)
	}
}
