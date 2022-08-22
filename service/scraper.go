
package service 

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
  "golang.org/x/exp/slices"
)

type Article struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	Source string `json:"source"`
}

type Source struct {
	URL    string
	Source string
}

var sourceMap = map[string]Source {
  "times": {URL: "https://www.thetimes.co.uk/environment/climate-change", Source: "The Times"},
  "guardian": {URL: "http://www.theguardian.com/environment/climate-crisis", Source: "The Guardian"},
  "telegraph": {URL: "http://www.telegraph.co.uk/climate-change", Source: "The Telegraph"},
}

func getDoc(url string) *goquery.Document {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("Status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

  return doc
}

func newsScrape(url string, source string, ch chan<- Article, wg *sync.WaitGroup) {
	defer wg.Done()
  doc := getDoc(url)
	doc.Find("a:contains('climate')").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Text())
		url, exists := s.Attr("href")
		if exists == true {
			ch <- Article{title, url, source}
		}
	})
}

func GetNewsBySource(c *gin.Context) {
  source := c.Param("source")
  validSources := []string {
    "times",
    "guardian",
    "telegraph",
  }
  
  if !slices.Contains(validSources, source){
    c.AbortWithStatus(http.StatusBadRequest)
  }

  doc := getDoc(sourceMap[source].URL)

  var articles []Article
	doc.Find("a:contains('climate')").Each(func(i int, s *goquery.Selection) {
		title := strings.TrimSpace(s.Text())
		url, exists := s.Attr("href")
		if exists == true {
		  articles = append(articles, Article{title, url, sourceMap[source].Source})
		}
	})

	c.JSON(http.StatusOK, articles)
}

func GetNews(c *gin.Context) {
	var wg sync.WaitGroup
	ch := make(chan Article)
  done := make(chan bool)

  for _, source := range sourceMap {
	  wg.Add(1)
	  go newsScrape(source.URL, source.Source, ch, &wg)
  }

	var articles []Article
	go func() {
		for r := range ch {
			articles = append(articles, r)
		}
    done <- true
	}()

	wg.Wait()
	close(ch)
  <- done
	c.JSON(http.StatusOK, articles)
}

