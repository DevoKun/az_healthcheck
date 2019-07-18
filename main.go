
package main

import (
  "fmt"
  "github.com/fatih/color"
  "net/http"
  "time"
) // import


func main() {

  http.DefaultClient.Timeout = 10 * time.Second

  azHealthcheckConfigLoad()

  go httpHealthMonitor()

  fmt.Println(color.YellowString(time.Now().Format("2006-01-02 15:04:05 +0000 UTC")) + 
              color.GreenString(" Starting up http listener on port ") + 
              color.WhiteString(config.Port) + 
              color.GreenString(" ...") )
  http.HandleFunc("/", httpHealthAnswer)
  http.ListenAndServe(":" + config.Port, nil)

} // func main





