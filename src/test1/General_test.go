package test

import "fmt"
import "testing"
import "strconv"

func fi () int {
   return 2;
}

func TestSwitch (t *testing.T) {
   switch x := fi (); x {
   case 0, 1:
      t.Fail ()
   case 2:
      //t.Fail ();
   }
}

func TestSwitch2 (t *testing.T) {
   switch x := fi (); {
   case false:
      fallthrough
   case x < 1:
      t.Fail ()
   case x > 1:
   }
}

func TestSwitch3 (t *testing.T) {
   switch {
   case false:
      fallthrough
   case 2 < 1:
      t.Fail ()
   case 2 > 1:
   }
}

func BenchmarkStringConcatation (b *testing.B) {
   for i := 0; i < b.N; i ++ {
      _ = fmt.Sprintf ("%s%d", "heelo ", strconv.Itoa (i))
   }
}

func BenchmarkStringConcatation2 (b *testing.B) {
   b.RunParallel ( func (pb *testing.PB) {
      i := 0
      for pb.Next () {
         _ = fmt.Sprintf ("%s%d", "hello ", strconv.Itoa (i))
         i ++
      }
   })
}