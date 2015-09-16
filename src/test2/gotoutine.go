package main

import (
   "fmt"
   "runtime"
   "time"
)

func main () {
   var exit = make (chan int, 1)

   var n int64 = int64 (runtime.GOMAXPROCS (0))
   for i := 0; int64 (i) < n; i ++ {
      fmt.Printf ("Start WithoutSleep #%d\n", n)
      go WithoutSleep (n, n, exit)
   }
   
   fmt.Printf ("Start WithSleep\n")
   go WithSleep ()
   
   for {
      <- exit
      if n --; n == 0 {
         break
      }
   }
   
   fmt.Printf ("Exit\n")
}

func WithSleep () {
   time.Sleep (1)
   
   fmt.Printf ("Exit WithSleep (in fact, will never go here if runtime.Gosched is not called).\n")
}

func WithoutSleep (a, b int64, exit chan<-int) int64 {
   var r int64
   for {
      r++
      if (r == 0x7FFFFFFF) {
         runtime.Gosched ()
         fmt.Printf ("Gosched\n")
      }
      if (r == 0xFFFFFFFF) {
         fmt.Printf ("break\n")
         break
      }
   }
   
   fmt.Printf ("Exit WithoutSleep\n")
   
   exit <- 1
   
   return r
}



