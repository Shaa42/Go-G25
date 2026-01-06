package protocol




import (
	"encoding/binary"
	"io"
)

func EncodeMessage(w io.Writer, msg Message) error {
	length := uint32(1 + len(msg.Payload)) 

	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return err
	}

	if err := binary.Write(w, binary.BigEndian, msg.Type); err != nil {
		return err
	}

	if _, err := w.Write(msg.Payload); err != nil {
		return err
	}

	return nil
}
