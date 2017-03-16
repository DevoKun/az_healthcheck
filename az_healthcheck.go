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
  "net/http"
  "fmt"
  "time"
  "gopkg.in/yaml.v2"
  "github.com/fatih/color"
  "github.com/davecgh/go-spew/spew"
  "io/ioutil"
  "path/filepath"
  "strconv"
  "strings"
  //"os"
) // import



  

//
// Custom Types
//

type azHealthcheckConfig struct {
  AllowedFailedChecks int                                      `yaml:"allowed_failed_checks"`
  Options             map[string]azHealthcheckConfig_options   `yaml:"options"`
  HttpChecks          map[string]azHealthcheckConfig_httpCheck `yaml:"httpchecks"`
} // type azHealthcheckConfig

type azHealthcheckConfig_options struct {
  statusFileName string `yaml:"status_file_name"`
} // type azHealthcheckConfig_httpCheck

type azHealthcheckConfig_httpCheck struct {
  Url     string            `yaml:"url"`
  Headers map[string]string `yaml:"headers"`
} // type azHealthcheckConfig_httpCheck



//
// Global Variables
//

var a azHealthcheckConfig

//
// Functions
//

func keepLines(s string, n int) string {
  result := strings.Join(strings.Split(s, "\n")[:n], "\n")
  return strings.Replace(result, "\r", "", -1)
}

func httpHealthAnswer(w http.ResponseWriter, r *http.Request) {
  
  azHealthcheckErrorCount            := 0
  var azHealthcheckResponseMessages  [100]string
  azHealthcheckResponseMessagesCount := 1

  fmt.Println(color.YellowString(time.Now().String()), color.CyanString("Answering HTTP Request") )

  //allowed_failed_checks := a.AllowedFailedChecks
  for k, v := range a.HttpChecks {
    fmt.Println( color.WhiteString("  * ") + color.YellowString(k) + color.WhiteString(": ") + color.CyanString(v.Url) )

    req, err := http.NewRequest("GET", v.Url, nil)
    if err != nil {
      // handle err
    } // if err

    for hk,hv := range v.Headers {
      fmt.Println(color.WhiteString("    ** -h ") + color.MagentaString(hk) + color.WhiteString(": ") + color.RedString(hv) )
      req.Header.Set(hk, hv)
    } // for

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
      // handle err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 { // OK
      // non http 200 response code

      //w.WriteHeader(http.StatusServiceUnavailable)
      //io.WriteString(w, strconv.Itoa(resp.StatusCode) + " ERROR")
      azHealthcheckErrorCount = azHealthcheckErrorCount + 1
      azHealthcheckResponseMessages[azHealthcheckResponseMessagesCount] = strconv.Itoa(resp.StatusCode) + " ERROR from: [" + v.Url + "]"
      azHealthcheckResponseMessagesCount = azHealthcheckResponseMessagesCount + 1
    } else {

      //bodyBytes, err2 := ioutil.ReadAll(resp.Body)
      _, err2 := ioutil.ReadAll(resp.Body)
      //bodyString := string(bodyBytes)
      if err2 != nil {
        fmt.Println( color.YellowString(" !!!! ") + color.RedString("Unable to get response body") + color.YellowString(" !!!! ") )
        //w.WriteHeader(http.StatusServiceUnavailable)
        //io.WriteString(w, "Unable to get response body")
        azHealthcheckErrorCount = azHealthcheckErrorCount + 1
        azHealthcheckResponseMessages[azHealthcheckResponseMessagesCount] = "Unable to get response body from: [" + v.Url + "]"
        azHealthcheckResponseMessagesCount = azHealthcheckResponseMessagesCount + 1
      } else {
        //fmt.Println("[" + bodyString + "]")
        //w.WriteHeader(http.StatusOK)
        //io.WriteString(w, "[" + bodyString + "]")
        azHealthcheckErrorCount = azHealthcheckErrorCount + 1
        //azHealthcheckResponseMessages[azHealthcheckResponseMessagesCount] = "[" + bodyString + "]"
        azHealthcheckResponseMessages[azHealthcheckResponseMessagesCount] = "successful query to: [" + v.Url + "]"
        azHealthcheckResponseMessagesCount = azHealthcheckResponseMessagesCount + 1
      } // if err

    } // if resp.StatusCode != 200

  } // for

  fmt.Println("")

  if (azHealthcheckErrorCount > 0) {
    w.WriteHeader(http.StatusServiceUnavailable)
    status_file_msg := "unhealthy since $now UTC";
  } else {
    w.WriteHeader(http.StatusOK)
    status_file_msg := "healthy since $now UTC";
  } // if else

  for _, azHealthcheckResponseMessage := range azHealthcheckResponseMessages {    
    if ( "x" + azHealthcheckResponseMessage != "x" ) {
      io.WriteString(w, azHealthcheckResponseMessage + "\n")
    } // if
  //for i := 0; i < len(azHealthcheckResponseMessages); i++ {
    //io.WriteString(w, azHealthcheckResponseMessages[i] + "\n")
  } // for
  

  spew.Dump(a.Options)
  /*
  f, err := os.Create(a.Options.statusFileName)
  check(err)
  defer f.Close()
  n3, err := f.WriteString(status_file_msg)
  f.Sync()
  */

} // func hello




func azHealthcheckConfigLoad() {

  configFilename, _ := filepath.Abs("./az_healthcheck.yaml")

  yamlData, err := ioutil.ReadFile(configFilename)
  check(err)

  if err := yaml.Unmarshal([]byte(yamlData), &a); err != nil {
    fmt.Println(err.Error())
  } // if

  //spew.Dump(a)

  fmt.Println(color.YellowString("allowed_failed_checks") + color.WhiteString(": [") + color.CyanString(strconv.Itoa(a.AllowedFailedChecks)) + color.WhiteString("]"))
  fmt.Println(color.YellowString("Options"))
  for k, v := range a.Options {
    fmt.Println(color.WhiteString("  * ") + color.YellowString(k) + color.WhiteString(": [") + color.CyanString(v) + color.WhiteString("]"))
  } // for
  fmt.Println("")
  fmt.Println(color.YellowString("HTTP Checks"))
  for k, v := range a.HttpChecks {
    fmt.Println( color.WhiteString("  * ") + color.YellowString(k) )
    fmt.Println( color.WhiteString("    ** ") + color.YellowString("URL") + color.WhiteString(": [") + color.CyanString(v.Url) + color.WhiteString("]") )
    fmt.Println( color.WhiteString("    ** ") + color.YellowString("Headers") )
    for _,h := range v.Headers {
      fmt.Println(color.WhiteString("      *** ") + color.RedString(h) )
    } // for

  } // for

  fmt.Println("")

} // func yamltest


func check(e error) {
  if e != nil {
    panic(e)
  } // if
} // func check




func main() {

  azHealthcheckConfigLoad()

  fmt.Println(color.YellowString(time.Now().String()), color.GreenString("Starting up http listener...") )
  http.HandleFunc("/", httpHealthAnswer)
  http.ListenAndServe(":8000", nil)
} // func main





