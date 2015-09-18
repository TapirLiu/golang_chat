package chat

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"regexp"
	_ "runtime"
	"strings"
	"time"
)

type Server struct {
	Rooms    map[string]*Room
	Lobby    *Room
	Visitors map[string]*Visitor
	ToExit   chan int

	PendingConnections chan net.Conn
	ChangeRoomRequests chan *Visitor
	ChangeNameRequests chan *Visitor

	RegexpBraces *regexp.Regexp
}

func (server *Server) CreateMessage(who string, what string) string {
	return fmt.Sprintf("[%s] %s> %s\n", time.Now().Local().Format("15:04:05"), who, what)
}

func (server *Server) NormalizeName(name string) string {
	return server.RegexpBraces.ReplaceAllString(name, "")
}

func (server *Server) CreateRandomVisitorName() string {
	for {
		var name = fmt.Sprintf("visitor_%d", 10000+rand.Intn(9990000))
		if server.Visitors[strings.ToLower(name)] == nil {
			return name
		}
	}
}

func (server *Server) run() {

	rand.Seed(time.Now().UTC().UnixNano())

	server.Rooms = make(map[string]*Room)
	server.Lobby = server.createNewRoom(LobbyRoomID)
	server.Visitors = make(map[string]*Visitor)
	server.ToExit = make(chan int, 1)

	server.PendingConnections = make(chan net.Conn, MaxPendingConnections)
	server.ChangeRoomRequests = make(chan *Visitor, MaxBufferedChangeRoomRequests)
	server.ChangeNameRequests = make(chan *Visitor, MaxBufferedChangeNameRequests)

	server.RegexpBraces = regexp.MustCompile("[{}]")

	go server.Lobby.run()

	go func(server *Server) {
		for {
			select {
			case conn := <-server.PendingConnections:
				var visitor = server.createNewVisitor(conn, server.CreateRandomVisitorName())
				if visitor != nil {
					visitor.OutputMessages <- server.CreateMessage("Server", fmt.Sprintf("your name: %s. You can input /name new_name to change your name.", visitor.Name))
					go visitor.run()
				}
			case visitor := <-server.ChangeNameRequests:
				visitor.changeName()
			case visitor := <-server.ChangeRoomRequests:
				if visitor.CurrentRoom != nil { // &&  visitor.RoomElement != nil
					visitor.CurrentRoom.VisitorLeaveRequests <- visitor
				} else if visitor.NextRoomID == VoidRoomID {
					visitor.endChangingRoom()
					log.Printf("Destroy visitor: %s", visitor.Name)
					server.destroyVisitor(visitor)
				} else {
					var room = server.Rooms[strings.ToLower(visitor.NextRoomID)]
					if room == nil {
						room = server.createNewRoom(visitor.NextRoomID)
						go room.run()
					}

					room.VisitorEnterRequests <- visitor
				}
				//case room := <- server.RoomCloseRequests:
			}
		}
	}(server)

	<-server.ToExit
}

func (server *Server) OnNewConnection(conn net.Conn) {
	server.PendingConnections <- conn
}

func CreateChatServer() *Server {
	var server = new(Server)
	go server.run()

	return server
}
