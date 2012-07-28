redmine
=======

A go package that can be used to access the Redmine API

$ go get github.com/woli/redmine

#### Usage example:

    package main

    import (
            "fmt"
            "github.com/woli/redmine"
            "net/url"
    )

    func main() {
            // necessary if using https
            redmine.SetCertAuth([]byte("<cert>"))

            r := redmine.New("<url>")
			r.SetAPIKey("<apiKey>")

			v := &url.Values{}
			v.Add("limit", "10")
			if issues, pag, err := r.GetIssues(v); err != nil {
				fmt.Println(err)
			} else {
				fmt.Printf("TotalCount:%v Limit:%v Offset:%v\n", pag.TotalCount, pag.Limit, pag.Offset)
				for i := range issues {
					fmt.Printf("%+v\n", issues[i])
				}
			}
    }
