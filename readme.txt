Demo: http://www.tapirgames.com/blog/golang-chat

To build:

   export GOPATH=`pwd`

   go get github.com/gorilla/websocket

   go install chat_wsserver

To run:

   bin/chat_wsserver

Then open http://localhost:6636 in a browser to local test.

In chatting, you can change your name by inputting

  /name your_new_name

and change current room by inputting

  /room new_room
