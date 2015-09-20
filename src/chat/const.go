package chat

const (
	MaxRoomCapacity               = 32
	MaxRoomBufferedMessages       = 64
	MinVisitorNameLength          = 1
	MaxVisitorNameLength          = 16
	MaxVisitorBufferedMessages    = 16
	MaxPendingConnections         = 128
	MaxBufferedChangeRoomRequests = 128
	MaxBufferedChangeNameRequests = 64
	MaxMessageLength              = 1024
	LobbyRoomID                   = "Lobby"
	VoidRoomID                    = ""
)
