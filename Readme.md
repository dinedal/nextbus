# nextbus.go

## Summary

Nextbus Public API implementation in Go. Not complete. Works for me. No warrenty expressed, given, or implied.

### Why?

I wanted to know when my next MUNI tram would arrive.

## Usage

```go
package main

import (
    "fmt"
    "github.com/dinedal/nextbus"
)

func main() {
    nb := nextbus.DefaultClient
    alist, aerr := nb.GetAgencyList()
    if aerr != nil {
        fmt.Println(aerr)
    }
    fmt.Println(alist)

    rlist, rerr := nb.GetRouteList("sf-muni")
    if rerr != nil {
        fmt.Println(rerr)
    }
    fmt.Println(rlist)

    rconfig, rcerr := nb.GetRouteConfig("sf-muni", nextbus.RouteConfigTag("N"))
    if rcerr != nil {
        fmt.Println(rcerr)
    }
    fmt.Println(rconfig)

    // Get predictions for the N train at stop with tag 5205.
    predictions, perr := nb.GetPredictions("sf-muni", "N", "5205")
    if perr != nil {
        fmt.Println(perr)
    }
    fmt.Println(predictions)

    // Get predictions for all routes at stop with id 14961.
    stopPreds, sperr := nb.GetStopPredictions("sf-muni", "14961")
    if perr != nil {
        fmt.Println(sperr)
    }
    fmt.Println(stopPreds)
}
```

## License
MIT
