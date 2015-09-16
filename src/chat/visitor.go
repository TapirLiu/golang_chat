package chat

import (
   "net"
   "fmt"
   "log"
   "strings"
   "bufio"
   "container/list"
)

type Visitor struct {
   Server         *Server
   
   Connection     net.Conn
   Input          *bufio.Reader
   Output         *bufio.Writer
   OutputMessages chan string
   
   Name           string
   
   RoomElement    *list.Element // in CurrentRoom.Visitors
   CurrentRoom    *Room
   NextRoomID     string
   RoomChanged    chan int
   
   ToClose        chan int
   Closed         chan int
}

func (server *Server) createNewVisitor (c net.Conn, name string) *Visitor {
   var visitor = &Visitor {
      Server: server,
      
      Connection: c,
      Input: bufio.NewReader (c),
      Output: bufio.NewWriter (c),
      OutputMessages: make (chan string, MaxVisitorBufferedMessages),
      
      Name: name,
      
      RoomElement: nil,
      CurrentRoom: nil,
      NextRoomID: VoidRoomID,
      RoomChanged: make (chan int),
      
      ToClose: make (chan int, 2), // 2 for read and write
      Closed: make (chan int),
   }
   
   server.Visitors [strings.ToLower (name)] = visitor
   
   visitor.beginChangingRoom (LobbyRoomID)
   
   log.Printf ("New visitor: %s", visitor.Name)
   
   return visitor
}

func (server *Server) destroyVisitor (visitor *Visitor) {
   if visitor.CurrentRoom != nil {
      log.Printf ("EnterVisDestroyVisitoritor: visitor.CurrentRoom != nil");
   }
    
   delete (server.Visitors, visitor.Name)
   
   visitor.Connection.Close ()
}

func (visitor *Visitor) beginChangingRoom (newRoomID string) {
   visitor.RoomChanged = make (chan int) // to block Visitor.Read before room is changed.
   visitor.NextRoomID = newRoomID
   visitor.Server.ChangeRoomRequests <- visitor
}

func (visitor *Visitor) endChangingRoom () {
   visitor.NextRoomID = VoidRoomID
   close (visitor.RoomChanged)
}

func (visitor *Visitor) run () {
   go visitor.read ()
   go visitor.write ()
   
   <- visitor.ToClose
   close (visitor.Closed)
   
   // let server close visitor
   visitor.beginChangingRoom (VoidRoomID)
}

func (visitor *Visitor) read () {
   var server = visitor.Server
   
   for {
      <- visitor.RoomChanged // wait server change room for vistor, when server has done it, this channel will be closed.
      
      select {
      case <- visitor.Closed:
         goto EXIT
      default:
      }
      
      var line, err = visitor.Input.ReadString ('\n')
      if err != nil {
         visitor.ToClose <- 1
         goto EXIT
      }
      
      if len (line) > MaxMessageLength {
         line = line [:MaxMessageLength]
      }
      
      if strings.HasPrefix (line, "/") {
         if strings.HasPrefix (line, "/exit") {
            visitor.ToClose <- 1
            goto EXIT
         } else if strings.HasPrefix (line, "/room") {
            line = strings.TrimPrefix (line, "/room")
            line = strings.TrimSpace (line)
            if len (line) == 0 { // show current room name
               if visitor.CurrentRoom == nil {
                  visitor.OutputMessages <- server.CreateMessage ("Server", "{lobby}")
               } else {
                  visitor.OutputMessages <- server.CreateMessage ("Server", fmt.Sprintf ("{%d}", visitor.CurrentRoom.ID))
               }
            } else { // change room, 
               line = server.NormalizeName (line)
               
               visitor.beginChangingRoom (line)
            }
            
            continue;
         //} else if strings.HasPrefix (line, "/name") {
         }      
      }
      
      if visitor.CurrentRoom != nil {
         if visitor.CurrentRoom == server.Lobby {
            line = server.CreateMessage ("Server", "you are current in lobby, please input /room room_name to enter a room")
         } else {
            line = server.CreateMessage (visitor.Name, line)
         }
         
         visitor.CurrentRoom.Messages <- line
      }
   }
   
EXIT:
}

func (visitor *Visitor) write () {
   for {
      select {
      case <- visitor.Closed:
         goto EXIT
      case message := <- visitor.OutputMessages:
         num, err := visitor.Output.WriteString (message)
         _ = num
         if err != nil {
            visitor.ToClose <- 1
            goto EXIT
         }
         
         for {
            select {
            case message = <- visitor.OutputMessages:
               num, err = visitor.Output.WriteString (message)
               _ = num
               if err != nil {
                  visitor.ToClose <- 1
                  goto EXIT
               }
            default:
               goto FLUSH
            }
         }
         
FLUSH:
            
         if visitor.Output.Flush () != nil {
            visitor.ToClose <- 1
            goto EXIT
         }
      }
   }
   
EXIT:
}
