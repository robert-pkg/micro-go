package codec

import (
	"encoding/json"
	"errors"

	"github.com/golang/protobuf/proto"
)

var (
	ErrInvalidMessage = errors.New("invalid message")
)

type ProtoCodec struct{}
type JsonCodec struct{}

func (ProtoCodec) Marshal(v interface{}) ([]byte, error) {
	m, ok := v.(proto.Message)
	if !ok {
		return nil, ErrInvalidMessage
	}
	return proto.Marshal(m)
}

func (ProtoCodec) Unmarshal(data []byte, v interface{}) error {
	m, ok := v.(proto.Message)
	if !ok {
		return ErrInvalidMessage
	}
	return proto.Unmarshal(data, m)
}

func (ProtoCodec) Name() string {
	return "proto"
}

func (JsonCodec) Marshal(v interface{}) ([]byte, error) {
	switch val := v.(type) {
	case proto.Message:
		return json.Marshal(v)
	case []byte:
		// 已经是进行过json的marshal之后的byte数组了
		return val, nil
	default:
		return json.Marshal(v)
	}
}

func (JsonCodec) Unmarshal(data []byte, v interface{}) error {
	switch val := v.(type) {
	case proto.Message:
		return json.Unmarshal(data, v)
	case *[]byte:
		// 只需要byte数组
		*val = data
		return nil
	default:
		return json.Unmarshal(data, v)
	}
}

func (JsonCodec) Name() string {
	return "json"
}
