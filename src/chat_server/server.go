package main

import (
   "net"
   "fmt"
   "log"
   "regexp"
   "strings"
   _ "runtime"
   "time"
   "math/rand"
)

type Server struct {
   Port     int
   Rooms    map[string]*Room
   Lobby    *Room
   Visitors map[string]*Visitor
   ToExit   chan int
   
   ChangeRoomRequests chan *Visitor
   
   RegexpBraces   *regexp.Regexp
}

func (server *Server) CreateMessage (who string, what string) string {
   return fmt.Sprintf ("[%s] %s> %s\n", time.Now ().Local ().Format ("15:04:05"), who, what)
}

func (server *Server) NormalizeName (name string) string {
   return server.RegexpBraces.ReplaceAllString (name, "")
}

func (server *Server) CreateRandomVisitorName () string {
   for {
      var name = fmt.Sprintf ("visitor_%d", 10000 + rand.Intn (9990000))
      if server.Visitors [strings.ToLower (name)] == nil {
         return name
      }
   } 
}

func (server *Server) HandleAccept (listener net.Listener) {
   for {
      var conn, err = listener.Accept ()
      if err != nil {
         log.Printf ("Accept new connection error: %s\n", err.Error ())
      } else {
         var visitor = server.CreateNewVisitor (conn, server.CreateRandomVisitorName ())
         if visitor != nil {
            go visitor.Run ()
         }
      }
   }
}

func (server *Server) HandleChangeRoomRequests () {
   for {
      select {
      case visitor := <- server.ChangeRoomRequests:
         if visitor.CurrentRoom != nil { // &&  visitor.RoomElement != nil
            visitor.CurrentRoom.VisitorLeaveRequests <- visitor
         } else if visitor.NextRoomID == VoidRoomID {
            visitor.EndChangingRoom ()
            log.Printf ("Destroy visitor: %s", visitor.Name)
            server.DestroyVisitor (visitor)
         } else {
            var room = server.Rooms [strings.ToLower (visitor.NextRoomID)]
            if room == nil {
               room = server.CreateNewRoom (visitor.NextRoomID)
               go room.Run ()
            }
            
            room.VisitorEnterRequests <- visitor
         }
      //case room := <- server.RoomCloseRequests:
      }
   }
}

func (server *Server) Run (port int) {
   
   var address = fmt.Sprintf (":%d", port)
   var listener, err = net.Listen ("tcp", address)
   if err != nil {
      log.Fatalf ("Listen error: %s\n", err.Error ())
   }
   
   log.Printf ("Listening ...\n")
   
   rand.Seed (time.Now ().UTC ().UnixNano ())
   
   server.Port = port
   server.Rooms = make (map[string]*Room)
   server.Lobby = server.CreateNewRoom (LobbyRoomID);
   server.Visitors = make (map[string]*Visitor)
   server.ToExit = make (chan int, 1)
   
   server.ChangeRoomRequests = make (chan *Visitor, MaxBufferedChangeRoomRequests)
   
   server.RegexpBraces = regexp.MustCompile ("[{}]")
   
   go server.Lobby.Run ()
   go server.HandleAccept (listener)
   go server.HandleChangeRoomRequests ()
   
   <- server.ToExit
}