
package main

import (
  "crypto/tls"
  "fmt"
  "github.com/fatih/color"
  "io"
  "io/ioutil"
  "net"
  "net/http"
  "strconv"
  "strings"
  "time"
) // import

var azHealthcheckErrorCount    = 0
var azHealthcheckStatusMessage = ""


//
// HTTP Health Monitor
// Makes call out to each target server
// and sets global variable which is used by the http listener
//

func httpHealthMonitor() {
  fmt.Println(color.YellowString(time.Now().Format("2006-01-02 15:04:05 +0000 UTC")), color.GreenString("Starting up asynchronous AZ Health Monitor...\n") )
  for {
    time.Sleep(10 * time.Second)
    httpHealthCheck()
  }
} // func httpHealthMonitor



//
// HTTP Health Check
// Performs the HTTP-based client request to the target server.
//
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
    if (((v.ClientCertFilename   + "x") != "x") && 
        ((v.ClientKeyFilename    + "x") != "x") ) {
      useClientCerts = true

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

    tlsConfig = &tls.Config{
      Certificates: []tls.Certificate{clientKeyPair},
      InsecureSkipVerify: true, // do not fail if CN does not match the url
    }
    tlsConfig.BuildNameToCertificate()
    } // if

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

    if useClientCerts {
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

    } else {
      netTransport = &http.Transport{
        Dial: (&net.Dialer{
          Timeout:   10 * time.Second,
          KeepAlive: 1 * time.Second,
        }).Dial,
        TLSHandshakeTimeout:   10 * time.Second,
        ResponseHeaderTimeout: 10 * time.Second,
        ExpectContinueTimeout:  1 * time.Second,
      } // netTransport
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

  t   := time.Now()
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


//
// HTTP Health Answer
// Answer HTTP client request
// Return server status as an HTTP Response Code
// and the contents of azHealthcheckStatusMessage, which should be a json string.
//
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
