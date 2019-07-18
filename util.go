
package main

import (
  "fmt"
  "github.com/fatih/color"
  "strings"
) // import

func keepLines(s string, n int) string {
  result := strings.Join(strings.Split(s, "\n")[:n], "\n")
  return strings.Replace(result, "\r", "", -1)
} // func keepLines

func errorCheck(e error) {
  if e != nil {
    panic(e)
  } // if
} // func errorCheck

func printConfigVal(k string, v string) {
  fmt.Println(color.WhiteString("  * ") + 
              color.YellowString(k + "") + 
              color.WhiteString(": [") + 
              color.CyanString(v + "") + 
              color.WhiteString("]"))

} // func printConfigVal
