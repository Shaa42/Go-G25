package protocol



type MessageType uint8

const (
        MsgRequestAudio  MessageType = 1
        MsgResponseAudio MessageType = 2
        MsgError         MessageType = 3
)

type Message struct {
        Type    MessageType
        Payload []byte
}
:
