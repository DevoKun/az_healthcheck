// http response code constant names comes from: https://godoc.org/net/http#ResponseWriter
//    StatusOK                            = 200 // RFC 7231, 6.3.1
//    StatusCreated                       = 201 // RFC 7231, 6.3.2
//    StatusAccepted                      = 202 // RFC 7231, 6.3.3
//    StatusNonAuthoritativeInfo          = 203 // RFC 7231, 6.3.4
//    StatusNoContent                     = 204 // RFC 7231, 6.3.5
//    StatusResetContent                  = 205 // RFC 7231, 6.3.6
//    StatusPartialContent                = 206 // RFC 7233, 4.1
//    StatusMultiStatus                   = 207 // RFC 4918, 11.1
//    StatusAlreadyReported               = 208 // RFC 5842, 7.1
//    StatusIMUsed                        = 226 // RFC 3229, 10.4.1
//
//    StatusInternalServerError           = 500 // RFC 7231, 6.6.1
//    StatusNotImplemented                = 501 // RFC 7231, 6.6.2
//    StatusBadGateway                    = 502 // RFC 7231, 6.6.3
//    StatusServiceUnavailable            = 503 // RFC 7231, 6.6.4
//    StatusGatewayTimeout                = 504 // RFC 7231, 6.6.5
//    StatusHTTPVersionNotSupported       = 505 // RFC 7231, 6.6.6
//    StatusVariantAlsoNegotiates         = 506 // RFC 2295, 8.1
//    StatusInsufficientStorage           = 507 // RFC 4918, 11.5
//    StatusLoopDetected                  = 508 // RFC 5842, 7.2
//    StatusNotExtended                   = 510 // RFC 2774, 7
//    StatusNetworkAuthenticationRequired = 511 // RFC 6585, 6

//
// Package
//

package main

//
// Import
//

import (
  "io"
  "net"
  "net/http"
  "fmt"
  "time"
  "gopkg.in/yaml.v2"
  "github.com/fatih/color"
  //"github.com/davecgh/go-spew/spew"
  "io/ioutil"
  "path/filepath"
  "strconv"
  "strings"
  //"os"
  //"syscall"
) // import



  

//
// Custom Types
//

type azHealthcheckConfig struct {
  BrowserAgent          string `yaml:"browserAgent"`
  Check_mk_service_name string `yaml:"check_mk_service_name"`
  CheckInterval         string `yaml:"checkInterval"`
  Port                  string `yaml:"port"`
  Hosts                 map[string]azHealthcheckConfigHost `yaml:"hosts"`
} // type azHealthcheckConfig

type azHealthcheckConfigHost struct {
  Name     string            `yaml:"name"`
  Url      string            `yaml:"url"`
  Headers  map[string]string `yaml:"headers,omitempty"`
} // type azHealthcheckConfig_httpCheck



//
// Global Variables
//

var config azHealthcheckConfig
var azHealthcheckErrorCount    = 0
var azHealthcheckStatusMessage = ""

//
// Functions
//

func keepLines(s string, n int) string {
  result := strings.Join(strings.Split(s, "\n")[:n], "\n")
  return strings.Replace(result, "\r", "", -1)
}


//
// HTTP Health Monitor
// Makes call out to each server in the AZ
// and sets global variable which is used by the http listener
//

func httpHealthMonitor() {
  fmt.Println(color.YellowString(time.Now().String()), color.GreenString("Starting up asynchronous AZ Health Monitor...\n") )
  for {
    time.Sleep(3 * time.Second)
    httpHealthCheck()
  }
} // func httpHealthMonitor



func httpHealthCheck() {
  
  //fmt.Println("113 - azHealthcheckErrorCount: ", azHealthcheckErrorCount)
  azHealthcheckErrorCountLocal       := 0
  //fmt.Println("115 - azHealthcheckErrorCount: ", azHealthcheckErrorCount)
  var azHealthcheckResponseMessages  [100]string
  azHealthcheckResponseMessagesCount := 1

  fmt.Println(color.YellowString(time.Now().String()), color.CyanString("Checking Health of AZ Hosts") )

  for k, v := range config.Hosts {
    fmt.Println( color.WhiteString("  * ") + color.YellowString(k) + color.WhiteString(": ") + color.CyanString(v.Url) )

    req, err := http.NewRequest("GET", v.Url, nil)
    if err != nil {
      errorCheck(err)
    } // if err

    for hk,hv := range v.Headers {
      fmt.Println(color.WhiteString("    ** -h ") + 
                  color.MagentaString(hk) + 
                  color.WhiteString(": ") + 
                  color.RedString(hv) )
      req.Header.Set(hk, hv)
    } // for


    netTransport := &http.Transport{
      Dial: (&net.Dialer{
        Timeout:   10 * time.Second,
        KeepAlive: 10 * time.Second,
      }).Dial,
      TLSHandshakeTimeout:   10 * time.Second,
      ResponseHeaderTimeout: 10 * time.Second,
      ExpectContinueTimeout:  1 * time.Second,
    }

    client := http.Client{
      Timeout:   time.Second * 10,
      Transport: netTransport,
    } // http.Client

    resp, err := client.Do(req)
    //resp, err := http.DefaultClient.Do(req)
    if err != nil {

      azHealthcheckErrorCountLocal = azHealthcheckErrorCountLocal + 1
      if (strings.Contains(err.Error(), "connection refused")) {
        azHealthcheckResponseMessages[azHealthcheckResponseMessagesCount] = "(ECONNREFUSED) Connection Refused: Server is offline or not responding"
      } else {
        azHealthcheckResponseMessages[azHealthcheckResponseMessagesCount] = err.Error()
      } // if
      azHealthcheckResponseMessagesCount = azHealthcheckResponseMessagesCount + 1

    } else { // if err != nil
      defer resp.Body.Close()

      if resp.StatusCode != 200 { // OK
        // non http 200 response code

        azHealthcheckErrorCountLocal = azHealthcheckErrorCountLocal + 1
        azHealthcheckResponseMessages[azHealthcheckResponseMessagesCount] = strconv.Itoa(resp.StatusCode) + " ERROR from: [" + v.Url + "]"
        azHealthcheckResponseMessagesCount = azHealthcheckResponseMessagesCount + 1

      } else {

        _, err2 := ioutil.ReadAll(resp.Body)
        if err2 != nil {
          fmt.Println( color.YellowString(" !!!! ") + color.RedString("Unable to get response body") + color.YellowString(" !!!! ") )
          azHealthcheckErrorCountLocal = azHealthcheckErrorCountLocal + 1
          azHealthcheckResponseMessages[azHealthcheckResponseMessagesCount] = "Unable to get response body from: [" + v.Url + "]"
          azHealthcheckResponseMessagesCount = azHealthcheckResponseMessagesCount + 1
        } else {
          azHealthcheckResponseMessages[azHealthcheckResponseMessagesCount] = "successful query to: [" + v.Url + "] (" + strconv.Itoa(resp.StatusCode) + ")"
          azHealthcheckResponseMessagesCount = azHealthcheckResponseMessagesCount + 1
        } // if err

      } // if resp.StatusCode != 200

    } // if err


  } // for

  fmt.Println("")

  t := time.Now()
  //fmt.Println(t.String())
  //fmt.Println(t.Format("2006-01-02 15:04:05 +0000 UTC"))
  now := t.Format("2006-01-02 15:04:05 +0000 UTC")

  //fmt.Println("203 - azHealthcheckErrorCount: ", azHealthcheckErrorCount)
  if (azHealthcheckErrorCount > 0) {
    azHealthcheckStatusMessage = "{\"statusCode\":\"503\",\"statusText\":\"unhealthy\",\"time\":\""+now+"\"}"
  } else {
    azHealthcheckStatusMessage = "{\"statusCode\":\"200\",\"statusText\":\"healthy\",\"time\":\""+now+"\"}"
  } // if else
  azHealthcheckErrorCount = azHealthcheckErrorCountLocal



} // func httpHealthCheck



func httpHealthAnswer(w http.ResponseWriter, r *http.Request) {
  
  fmt.Println(color.YellowString(time.Now().String()), color.CyanString("Answering HTTP Request") )

  //for k, v := range config.Hosts {
  //  fmt.Println( color.WhiteString("  * ") + color.YellowString(k) + color.WhiteString(": ") + color.CyanString(v.Url) )

  //fmt.Println("")

  /*
  WriteHeader sends an HTTP response header with status code.
  If WriteHeader is not called explicitly,
  the first call to Write will trigger an implicit WriteHeader(http.StatusOK).
  Thus explicit calls to WriteHeader are mainly used to send error codes.
  */
  if (azHealthcheckErrorCount > 0) {
    w.WriteHeader(http.StatusServiceUnavailable)
  } // if
  io.WriteString(w, azHealthcheckStatusMessage + "\n")
  //for _, azHealthcheckResponseMessage := range azHealthcheckResponseMessages {    
  //  if ( "x" + azHealthcheckResponseMessage != "x" ) {
  //    io.WriteString(w, azHealthcheckResponseMessage + "\n")
  //  } // if
  //} // for

} // func httpHealthAnswer


func printConfigVal(k string, v string) {
  fmt.Println(color.WhiteString("  * ") + 
              color.YellowString(k + "") + 
              color.WhiteString(": [") + 
              color.CyanString(v + "") + 
              color.WhiteString("]"))

} // func

func azHealthcheckConfigLoad() {

  configFilename, _ := filepath.Abs("./az_healthcheck.yaml")

  yamlData, err := ioutil.ReadFile(configFilename)
  errorCheck(err)

  if err := yaml.Unmarshal([]byte(yamlData), &config); err != nil {
    fmt.Println(err.Error())
  } // if

  fmt.Println(color.YellowString("Options"))
  printConfigVal("browserAgent",          config.BrowserAgent)
  printConfigVal("check_mk_service_name", config.Check_mk_service_name)
  printConfigVal("checkInterval",         config.CheckInterval)
  printConfigVal("port",                  config.Port)

  fmt.Println("")
  fmt.Println(color.YellowString("Host Checks"))
  for k, v := range config.Hosts {
    fmt.Println( color.WhiteString("  * ") + color.YellowString(k) )

    fmt.Println( color.WhiteString("    ** ") + 
                 color.YellowString("Name") + 
                 color.WhiteString(": [") + 
                 color.CyanString(v.Name) + 
                 color.WhiteString("]") )

    fmt.Println( color.WhiteString("    ** ") + 
                 color.YellowString("URL") + 
                 color.WhiteString(": [") + 
                 color.CyanString(v.Url) + 
                 color.WhiteString("]") )

    fmt.Println( color.WhiteString("    ** ") + 
                 color.YellowString("Headers") )
    for hk,hv := range v.Headers {
      fmt.Println(color.WhiteString("       *** ") + 
                  color.MagentaString(hk) + 
                  color.WhiteString(": ") + 
                  color.RedString(hv) )
    } // for

  } // for

  fmt.Println("")

} // func azHealthcheckConfigLoad


func errorCheck(e error) {
  if e != nil {
    panic(e)
  } // if
} // func errorCheck



func main() {

  azHealthcheckConfigLoad()

  go httpHealthMonitor()

  fmt.Println(color.YellowString(time.Now().String()), color.GreenString("Starting up http listener...") )
  http.HandleFunc("/", httpHealthAnswer)
  http.ListenAndServe(":3000", nil)

} // func main





