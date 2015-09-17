package chat

import (
   "net"
   _ "io"
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
   
   ReadClosed     chan int
   WriteClosed    chan int
   Closed         chan int
}

func (server *Server) createNewVisitor (c net.Conn, name string) *Visitor {
   var visitor = &Visitor {
      Server: server,
      
      Connection: c,
      Input: bufio.NewReader (c), // todo:io.LimitedReader 
      Output: bufio.NewWriter (c),
      OutputMessages: make (chan string, MaxVisitorBufferedMessages),
      
      Name: name,
      
      RoomElement: nil,
      CurrentRoom: nil,
      NextRoomID: VoidRoomID,
      RoomChanged: make (chan int),
      
      ReadClosed: make (chan int),
      WriteClosed: make (chan int),
      Closed: make (chan int),
   }
   
   server.Visitors [strings.ToLower (name)] = visitor
   
   visitor.beginChangingRoom (LobbyRoomID)
   
   log.Printf ("New visitor: %s", visitor.Name)
   
   return visitor
}

func (server *Server) destroyVisitor (visitor *Visitor) {
   if visitor.CurrentRoom != nil {
      log.Printf ("destroyVisitor: visitor.CurrentRoom != nil")
   }
    
   delete (server.Visitors, visitor.Name)
   
   visitor.CloseConnection ()
}

func (visitor *Visitor) CloseConnection () error {
    //defer func() {
    //    if err := recover(); err != nil {
    //        log.Println ("CloseConnection paniced", err)
    //    }
    //}()
    // above is to avoid panic on reclose a connection. It may be not essential.
   
   return visitor.Connection.Close ()
}

func (visitor *Visitor) beginChangingRoom (newRoomID string) {
   visitor.RoomChanged = make (chan int) // to block Visitor. Read before room is changed.
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
   
   <- visitor.WriteClosed
   <- visitor.ReadClosed
   close (visitor.Closed)
   
   // let server close visitor
   visitor.beginChangingRoom (VoidRoomID)
}

func (visitor *Visitor) read () {
   var server = visitor.Server
   
   for {
      select {
      //case <- visitor.ReadClosed:
      //   goto EXIT
      case <- visitor.WriteClosed:
         goto EXIT
      case <- visitor.Closed:
         goto EXIT
      default:
      }
      
      <- visitor.RoomChanged // wait server change room for vistor, when server has done it, this channel will be closed.
      
      var line, err = visitor.Input.ReadString ('\n') // todo: use io.LimitedReader insstead 
      if err != nil {
         goto EXIT
      }
      
      rn := []rune (line)
      if len (rn) > MaxMessageLength {
         rn = rn [:MaxMessageLength -1]
         line = fmt.Sprintf ("%s\n", string(rn))
      }
      
      if strings.HasPrefix (line, "/") {
         if strings.HasPrefix (line, "/exit") {
            goto EXIT
         } else if strings.HasPrefix (line, "/room") {
            line = strings.TrimPrefix (line, "/room")
            line = strings.TrimSpace (line)
            if len (line) == 0 { // show current room name
               if visitor.CurrentRoom == nil {
                  visitor.OutputMessages <- server.CreateMessage ("Server", "you are in lobby now")
               } else {
                  visitor.OutputMessages <- server.CreateMessage ("Server", fmt.Sprintf ("your are in %s now}", visitor.CurrentRoom.Name))
               }
            } else { // change room, 
               line = server.NormalizeName (line)
               
               visitor.beginChangingRoom (line)
            }
            
            continue;
         } else if strings.HasPrefix (line, "/name") {
            line = strings.TrimPrefix (line, "/name")
            line = strings.TrimSpace (line)
            if len (line) != 0 {
               rn := []rune (line)
               if len (rn) > MaxVisitorNameLength {
                  line = string (rn [:MaxVisitorNameLength])
               }
               line = server.NormalizeName (line)
               
               visitor.Name = line
               
               visitor.OutputMessages <- server.CreateMessage ("Server", fmt.Sprintf ("you changed your name to %s", visitor.Name)) 
            }
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
   
   close (visitor.ReadClosed)
}

func (visitor *Visitor) write () {
   for {
      select {
      case <- visitor.ReadClosed:
         goto EXIT
      //case <- visitor.WriteClosed:
      //   goto EXIT
      case <- visitor.Closed:
         goto EXIT
      case message := <- visitor.OutputMessages:
         num, err := visitor.Output.WriteString (message)
         _ = num
         if err != nil {
            goto EXIT
         }
         
         for {
            select {
            case message = <- visitor.OutputMessages:
               num, err = visitor.Output.WriteString (message)
               _ = num
               if err != nil {
                  goto EXIT
               }
            default:
               goto FLUSH
            }
         }
         
FLUSH:
            
         if visitor.Output.Flush () != nil {
            goto EXIT
         }
      }
   }
   
EXIT:
   
   visitor.CloseConnection () // to avoud read blocking
   
   close (visitor.WriteClosed)
}
