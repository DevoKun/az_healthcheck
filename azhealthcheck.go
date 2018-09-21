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
//    StatusBadRequest                   = 400 // RFC 7231, 6.5.1
//    StatusUnauthorized                 = 401 // RFC 7235, 3.1
//    StatusPaymentRequired              = 402 // RFC 7231, 6.5.2
//    StatusForbidden                    = 403 // RFC 7231, 6.5.3
//    StatusNotFound                     = 404 // RFC 7231, 6.5.4
//    StatusMethodNotAllowed             = 405 // RFC 7231, 6.5.5
//    StatusNotAcceptable                = 406 // RFC 7231, 6.5.6
//    StatusProxyAuthRequired            = 407 // RFC 7235, 3.2
//    StatusRequestTimeout               = 408 // RFC 7231, 6.5.7
//    StatusConflict                     = 409 // RFC 7231, 6.5.8
//    StatusGone                         = 410 // RFC 7231, 6.5.9
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
  "crypto/tls"
  //"crypto/x509"
  "fmt"
  "github.com/fatih/color"
  "gopkg.in/yaml.v2"
  "io"
  "io/ioutil"
  "net"
  "net/http"
  "os"
  "path/filepath"
  "strconv"
  "strings"
  "time"
  //"reflect"
  //"github.com/davecgh/go-spew/spew"
  //"syscall"
) // import

//import "os/exec"
//import "bytes"
  

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
  Name                 string  `yaml:"name"`
  Url                  string  `yaml:"url"`
  Headers   map[string]string  `yaml:"headers,omitempty"`
  ClientCertFilename   string  `yaml:"clientcertfilename"`
  ClientKeyFilename    string  `yaml:"clientkeyfilename"`
  //ClientCACertFilename string  `yaml:"clientcacertfilename"`
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
  fmt.Println(color.YellowString(time.Now().Format("2006-01-02 15:04:05 +0000 UTC")), color.GreenString("Starting up asynchronous AZ Health Monitor...\n") )
  for {
    time.Sleep(10 * time.Second)
    httpHealthCheck()
  }
} // func httpHealthMonitor



func httpHealthCheck() {
  
  azHealthcheckErrorCountLocal       := 0
  var azHealthcheckResponseMessages  [100]string
  azHealthcheckResponseMessagesCount := 1
  var tlsConfig    *tls.Config
  var netTransport *http.Transport

  fmt.Println(color.YellowString(time.Now().Format("2006-01-02 15:04:05 +0000 UTC")), color.CyanString("Checking Health of AZ Hosts") )

  for k, v := range config.Hosts {
    fmt.Println( color.WhiteString("  * ") + color.YellowString(k) + color.WhiteString(": ") + color.CyanString(v.Url) )

    useClientCerts := false
    if ( //((v.ClientCACertFilename + "x") != "x") && 
         ((v.ClientCertFilename   + "x") != "x") && 
         ((v.ClientKeyFilename    + "x") != "x") ) {
      useClientCerts = true
      //fmt.Println(color.WhiteString("    ** ") +
      //            color.MagentaString("Client CA Cert Filename ") + 
      //            color.WhiteString(": ") + 
      //            color.RedString(v.ClientCACertFilename) )

      fmt.Println(color.WhiteString("    ** ") +
                  color.MagentaString("Client Cert Filename ") + 
                  color.WhiteString("...: ") + 
                  color.RedString(v.ClientCertFilename) )
      fmt.Println(color.WhiteString("    ** ") +
                  color.MagentaString("Client Key Filename ") + 
                  color.WhiteString("....: ") + 
                  color.RedString(v.ClientKeyFilename) )

      clientKeyPair, err := tls.LoadX509KeyPair(v.ClientCertFilename, v.ClientKeyFilename)
      if err != nil {
        fmt.Println(color.YellowString("    !! ") + 
                    color.RedString("Unable to load Client Key Pair") )
        fmt.Println(err)
      } // if clientKeyPair

      //clientCACert, err := ioutil.ReadFile(v.ClientCACertFilename)
      //if err != nil {
      //  fmt.Println(color.YellowString("    !! ") + 
      //              color.RedString("Unable to load Client CA Cert") )
      //  fmt.Println(err)
      //} // if clientCACert

    //caCertPool := x509.NewCertPool()
    //caCertPool.AppendCertsFromPEM(clientCACert)

    tlsConfig = &tls.Config{
      Certificates: []tls.Certificate{clientKeyPair},
      InsecureSkipVerify: true, // do not fail if CN does not match the url
      //RootCAs:      caCertPool,
    }
    tlsConfig.BuildNameToCertificate()
    //fmt.Println(reflect.TypeOf(tlsConfig))
    } // if



    req, err := http.NewRequest("GET", v.Url, nil)

/*
cmd := exec.Command("lsof") //, "|", "grep", "azheal", "|", "grep", "ESTABLISHED")
var out bytes.Buffer
cmd.Stdout = &out
cmd.Run()
fmt.Println(out.String())
*/

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

    if useClientCerts {
    //fmt.Println(useClientCerts)
      fmt.Println(color.WhiteString("    ** ") + 
                  color.GreenString("Connecting using Client Certs for ") + 
                  color.WhiteString("MutualSSL"))
      netTransport = &http.Transport{
        Dial: (&net.Dialer{
          Timeout:   10 * time.Second,
          KeepAlive: 1 * time.Second,
        }).Dial,
        TLSClientConfig:       tlsConfig, // used for the client ssl cert auth
        TLSHandshakeTimeout:   10 * time.Second,
        ResponseHeaderTimeout: 10 * time.Second,
        ExpectContinueTimeout:  1 * time.Second,
      } // netTransport
      //fmt.Println(reflect.TypeOf(netTransport))

    } else { // if useClientCerts true
      netTransport = &http.Transport{
        Dial: (&net.Dialer{
          Timeout:   10 * time.Second,
          KeepAlive: 1 * time.Second,
        }).Dial,
        TLSHandshakeTimeout:   10 * time.Second,
        ResponseHeaderTimeout: 10 * time.Second,
        ExpectContinueTimeout:  1 * time.Second,
      } // netTransport
      //fmt.Println(reflect.TypeOf(netTransport))
    } // if useClientCerts false

    http.DefaultClient.Timeout = 10 * time.Second
    client := http.Client{
      Timeout:   time.Second * 10,
      Transport: netTransport,
    } // http.Client

    resp, err := client.Do(req)
    if (resp != nil) {
      defer resp.Body.Close() // close connection if it is non-nil
    }


    if err != nil {

      azHealthcheckErrorCountLocal = azHealthcheckErrorCountLocal + 1
      if (strings.Contains(err.Error(), "connection refused")) {
        azHealthcheckResponseMessages[azHealthcheckResponseMessagesCount] = k + " (ECONNREFUSED) Connection Refused: Server is offline or not responding"
      } else {
        azHealthcheckResponseMessages[azHealthcheckResponseMessagesCount] = k + " " + err.Error()
      } // if
      azHealthcheckResponseMessagesCount = azHealthcheckResponseMessagesCount + 1

    } else { // if err != nil
      
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
          azHealthcheckResponseMessages[azHealthcheckResponseMessagesCount] = k + " Unable to get response body from: [" + v.Url + "]"
          azHealthcheckResponseMessagesCount = azHealthcheckResponseMessagesCount + 1
        } else {
          azHealthcheckResponseMessages[azHealthcheckResponseMessagesCount] = k + " successful query to: [" + v.Url + "] (" + strconv.Itoa(resp.StatusCode) + ")"
          azHealthcheckResponseMessagesCount = azHealthcheckResponseMessagesCount + 1
        } // if err

      } // if resp.StatusCode != 200

    } // if err

  if (resp != nil) {
    resp.Body.Close()
  } // if

  } // for

  fmt.Println("")

  t := time.Now()
  now := t.Format("2006-01-02 15:04:05 +0000 UTC")

  azHealthcheckHostStatusesMessage := ""
  for _, rmv := range azHealthcheckResponseMessages {
    if ((rmv + "x") != "x") {
      azHealthcheckHostStatusesMessage = azHealthcheckHostStatusesMessage + rmv + "; "
    }
  } // for

  if (azHealthcheckErrorCount > 0) {
    azHealthcheckStatusMessage = "{\"statusCode\":\"503\",\"statusText\":\"unhealthy\",\"hostStatuses\":\""+azHealthcheckHostStatusesMessage+"\",\"time\":\""+now+"\"}"
  } else {
    azHealthcheckStatusMessage = "{\"statusCode\":\"200\",\"statusText\":\"healthy\",\"hostStatuses\":\""+azHealthcheckHostStatusesMessage+"\",\"time\":\""+now+"\"}"
  } // if else
  azHealthcheckErrorCount = azHealthcheckErrorCountLocal
  
} // func httpHealthCheck



func httpHealthAnswer(w http.ResponseWriter, r *http.Request) {
  
  fmt.Println(color.YellowString(time.Now().Format("2006-01-02 15:04:05 +0000 UTC")), color.CyanString("Answering HTTP Request") )

  /*
  Set the Headers BEFORE calling WriteHeader
  */
  w.Header().Set("X-XSS-Protection",          "1; mode=block")
  w.Header().Set("X-Content-Type-Options",    "nosniff")
  w.Header().Set("Content-Security-Policy",   "default-src 'self';")
  w.Header().Set("X-Frame-Options",           "SAMEORIGIN")
  w.Header().Set("X-Robots-Tag",              "noindex, noarchive, nosnippet")
  w.Header().Set("Strict-Transport-Security", "max-age=631138519")
  /*
  WriteHeader sends an HTTP response header with status code.
  If WriteHeader is not called explicitly,
  the first call to Write will trigger an implicit WriteHeader(http.StatusOK).
  Thus explicit calls to WriteHeader are mainly used to send error codes.
  */
  if r.URL.Path != "/" {
    w.WriteHeader(http.StatusForbidden)
    io.WriteString(w, "\n")
    return
  } else if (azHealthcheckErrorCount > 0) {
    w.WriteHeader(http.StatusServiceUnavailable)
  } // if
  io.WriteString(w, azHealthcheckStatusMessage + "\n")

} // func httpHealthAnswer


func printConfigVal(k string, v string) {
  fmt.Println(color.WhiteString("  * ") + 
              color.YellowString(k + "") + 
              color.WhiteString(": [") + 
              color.CyanString(v + "") + 
              color.WhiteString("]"))

} // func

func azHealthcheckConfigLoad() {

  fmt.Println(color.GreenString("Looking for YAML config file"))
  configFilename := ""
  absConfigFilename,_ := filepath.Abs("./azhealthcheck.yaml");
  if _,err := os.Stat("/etc/azhealthcheck.yaml"); err == nil {
    configFilename = "/etc/azhealthcheck.yaml"
    printConfigVal("Found config file at", configFilename)
  } else if _,err = os.Stat(absConfigFilename); err == nil {
    configFilename = absConfigFilename
    printConfigVal("Found config file at", configFilename)
  } else {
    fmt.Println(color.YellowString("  !! ") +
                color.RedString("Unable to locate ") + 
                color.YellowString("azhealthcheck.yaml") + 
                color.RedString(" config file") +
                color.YellowString(" !!"))
    os.Exit(1)
  } // if fileExists

  //printConfigVal("Reading YAML config file", configFilename)
  fmt.Println(color.GreenString("Reading YAML config file"))
  yamlData, err := ioutil.ReadFile(configFilename)
  //fmt.Printf("File contents: %s", yamlData)
  errorCheck(err)

  fmt.Println(color.GreenString("Parsing YAML"))
  if err := yaml.Unmarshal([]byte(yamlData), &config); err != nil {
    fmt.Println(err.Error())
  } // if

  fmt.Println(color.GreenString("Options"))
  printConfigVal("browserAgent",          config.BrowserAgent)
  printConfigVal("check_mk_service_name", config.Check_mk_service_name)
  printConfigVal("checkInterval",         config.CheckInterval)
  printConfigVal("port",                  config.Port)

  fmt.Println("")
  fmt.Println(color.GreenString("Host Checks"))
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

    //fmt.Println( color.WhiteString("    ** ") + 
    //             color.YellowString("ClientCACertFilename") + 
    //             color.WhiteString(": [") + 
    //             color.CyanString(v.ClientCACertFilename) + 
    //             color.WhiteString("]") )
    fmt.Println( color.WhiteString("    ** ") + 
                 color.YellowString("ClientCertFilename") + 
                 color.WhiteString("..: [") + 
                 color.CyanString(v.ClientCertFilename) + 
                 color.WhiteString("]") )
    fmt.Println( color.WhiteString("    ** ") + 
                 color.YellowString("ClientKeyFilename") + 
                 color.WhiteString("...: [") + 
                 color.CyanString(v.ClientKeyFilename) + 
                 color.WhiteString("]") )

  } // for

  fmt.Println("")

} // func azHealthcheckConfigLoad


func errorCheck(e error) {
  if e != nil {
    panic(e)
  } // if
} // func errorCheck



func main() {

  http.DefaultClient.Timeout = 10 * time.Second

  azHealthcheckConfigLoad()

  go httpHealthMonitor()

  fmt.Println(color.YellowString(time.Now().Format("2006-01-02 15:04:05 +0000 UTC")), color.GreenString("Starting up http listener...") )
  http.HandleFunc("/", httpHealthAnswer)
  http.ListenAndServe(":3000", nil)

} // func main





