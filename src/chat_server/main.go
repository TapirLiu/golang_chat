package main

import (
   "fmt"
   "log"
   "net"
   
   "chat"
)

func run (port int) {
   
   var address = fmt.Sprintf (":%d", port)
   var listener, err = net.Listen ("tcp", address)
   if err != nil {
      log.Fatalf ("Listen error: %s\n", err.Error ())
   }
   
   log.Printf ("Listening ...\n")
   
   var chat_server = chat.CreateChatServer ()
   
   for {
      var conn, err = listener.Accept ()
      if err != nil {
         log.Printf ("Accept new connection error: %s\n", err.Error ())
      } else {
         chat_server.OnNewConnection (conn)
      }
   }
}

func main () {
   run (9981)
}