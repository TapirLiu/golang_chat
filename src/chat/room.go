package chat

import (
   "log"
   "fmt"
   "strings"
   "container/list"
)

type Room struct {
   Server   *Server
   
   ID       string
   Name     string
   
   Visitors *list.List   // go list is not fun
   Messages chan string
   
   VisitorLeaveRequests chan *Visitor
   VisitorEnterRequests chan *Visitor
   
   //Closed   chan int // todo: when number visitors is 0, close room
}

func (server *Server) createNewRoom (id string) *Room {
   var room = &Room {
      Server: server,
      
      ID: id,
      
      Visitors: list.New (),
      Messages: make (chan string, MaxRoomBufferedMessages),
      
      VisitorLeaveRequests: make (chan *Visitor, MaxRoomCapacity),
      VisitorEnterRequests: make (chan *Visitor, MaxRoomCapacity),
   }
   
   if id == LobbyRoomID {
      room.Name = id
   } else {
      room.Name = fmt.Sprintf ("Room#%s", id)
   }
   
   server.Rooms [strings.ToLower (id)] = room
   
   log.Printf ("New room: %s", id)
   
   return room
}

func (room *Room) enterVisitor (visitor *Visitor) {
   if visitor.RoomElement != nil || visitor.CurrentRoom != nil {
      log.Printf ("EnterVisitor: visitor has already entered a room")
   }
   
   visitor.CurrentRoom = room
   visitor.RoomElement = room.Visitors.PushBack (visitor)
}

func (room *Room) leaveVisitor (visitor *Visitor) {
   if visitor.RoomElement == nil || visitor.CurrentRoom == nil {
      log.Printf ("LeaveVisitor: visitor has not entered any room yet")
      return
   }
   if visitor.CurrentRoom != room {
      log.Printf ("LeaveVisitor: visitor.CurrentRoom != room")
      return
   }
   
   if visitor != room.Visitors.Remove (visitor.RoomElement) {
      log.Printf ("LeaveVisitor: visitor != element.value")
      return
   }
   
   visitor.RoomElement = nil 
   visitor.CurrentRoom = nil
}

func (room *Room) run () {
   var server = room.Server
   var visitor *Visitor
   var message string
   var ok bool
   
   for {
      select {
      case visitor = <- room.VisitorLeaveRequests:
         //if (visitor.CurrentRoom == room) {
            room.leaveVisitor (visitor)
            visitor.OutputMessages <- server.CreateMessage (room.Name, "<= you leaved this room.")
            //room.Messages <- server.CreateMessage (room.Name, fmt.Sprintf ("<= visitor#%s leaved this room.", visitor.Name))
         //}
         server.ChangeRoomRequests <- visitor
      case visitor = <- room.VisitorEnterRequests:
         //if (visitor.CurrentRoom == room) {
            if room.Visitors.Len () >= MaxRoomCapacity {
               visitor.OutputMessages <- server.CreateMessage (room.Name, "Sorry, I am full. :(")
            } else {
               room.enterVisitor (visitor)
               visitor.OutputMessages <- server.CreateMessage (room.Name, "<= you entered this room.")
               //room.Messages <- server.CreateMessage (room.Name, fmt.Sprintf ("<= visitor#%s entered this room.", visitor.Name))
            }
            visitor.endChangingRoom ()
         //}
      case message = <- room.Messages:
         for e := room.Visitors.Front(); e != nil; e = e.Next() {
            visitor, ok = e.Value. (*Visitor)
            if ok {
               visitor.OutputMessages <- message
            }
         }
      }
   }
   
}
