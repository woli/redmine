redmine
=======

A go package that can be used to access the Redmine API (v2.2.3)

$ go get github.com/woli/redmine

#### Usage example:

	package main

	import (
		"fmt"
		"github.com/woli/redmine"
		"net/http"
	)

	func main() {
		auth := &redmine.ApiKeyAuth{"<apikey>"}
		s, _ := redmine.New("<baseurl>", auth, http.DefaultClient)

		feed, err := s.Issues.List().Do()
		if err != nil {
			panic(err)
		} else {
			fmt.Printf("TotalCount:%v Limit:%v Offset:%v\n", feed.TotalCount, feed.Limit, feed.Offset)
			for _, issue := range feed.Issues {
				fmt.Printf("%+v\n", issue)
			}
		}
	}
