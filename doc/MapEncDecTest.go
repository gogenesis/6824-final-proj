// Go provides built-in support for [base64
// encoding/decoding](http://en.wikipedia.org/wiki/Base64).

package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
)
type Message struct {
    Name string
    Body string
    Time int64
}

func main() {
   w := new(bytes.Buffer)
   e := gob.NewEncoder(w)



    // Here's the `string` we'll encode/decode.
    testMap := make(map[string]Message)
    testMap["David"] = Message{"From Taylor", "Test Message", 12345}
    testMap["Taylor"] = Message{"From David", "Test Reply", 54321}
    testMap["Final"] = Message{"Final Data", "Final Msg", 11222}
   
   err := e.Encode(testMap)
   if err != nil {
      fmt.Println("ENCODE FAIL")
   }
   data := w.Bytes()
   r := bytes.NewBuffer(data)
   d := gob.NewDecoder(r)

   var restoreMap map[string]Message

   if d.Decode(&restoreMap) == nil {
      for k,v := range(restoreMap) {
   	fmt.Println(fmt.Sprintf("%s: %s %s %d", k, v.Name, v.Body, v.Time))
      }
   } else {
     fmt.Println("DECODE FAIL")
   }
}

