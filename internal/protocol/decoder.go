package protocol

import (
	"encoding/binary"
	"io"
)


func DecodeMessage(r io.Reader) (Message, error) {
	var length uint32

	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return Message{}, err
	}

	var msgType MessageType
	if err := binary.Read(r, binary.BigEndian, &msgType); err != nil {
		return Message{}, err
	}

	payloadLen := int(length) - 1
	payload := make([]byte, payloadLen)

	if _, err := io.ReadFull(r, payload); err != nil {
		return Message{}, err
	}

	return Message{
		Type:    msgType,
		Payload: payload,
	}, nil
}
