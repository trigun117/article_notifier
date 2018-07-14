package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"regexp"

	"golang.org/x/net/html"
)

// Articles struct
type Articles struct {
	NewArticle, CurrentArticle string
	Status                     bool
}

// Article contains fresh article
var Article Articles

var reg = regexp.MustCompile(`^` + link + `\d\d\d\d/\d\d/\d\d/.`)

func (a *Articles) getCurrentArticle(link string) error {
	response, err := http.Get(link)
	if err != nil {
		return err
	}
	defer func() {
		defer response.Body.Close()
		ioutil.ReadAll(response.Body)
		io.Copy(ioutil.Discard, response.Body)
	}()
	z := html.NewTokenizer(response.Body)
	for {
		switch token := z.Next(); token {
		case html.ErrorToken:
			return nil
		case html.StartTagToken:
			t := z.Token()
			isAnchor := t.Data == "a"
			if isAnchor {
				for _, v := range t.Attr {
					if v.Key == "href" {
						if reg.MatchString(v.Val) {
							a.CurrentArticle = v.Val
							return nil
						}
					}
				}
			}
		}
	}

}

func (a *Articles) compare() {
	if a.CurrentArticle != a.NewArticle {
		a.NewArticle = a.CurrentArticle
		a.Status = true
	} else {
		a.Status = false
	}
}
