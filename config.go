
package main

import (
  "fmt"
  "github.com/fatih/color"
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "os"
  "path/filepath"
) // import

type azHealthcheckConfig struct {
  BrowserAgent          string `yaml:"browserAgent"`
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
} // type azHealthcheckConfig_httpCheck


var config azHealthcheckConfig


//
// Config File Load
// Reads in YAML config file
//
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

  printConfigVal("Reading YAML config file", configFilename)
  yamlData, err := ioutil.ReadFile(configFilename)
  errorCheck(err)

  fmt.Println(color.GreenString("Parsing YAML"))
  if err := yaml.Unmarshal([]byte(yamlData), &config); err != nil {
    fmt.Println(err.Error())
  } // if

  fmt.Println(color.GreenString("Options"))
  printConfigVal("browserAgent",          config.BrowserAgent)
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
