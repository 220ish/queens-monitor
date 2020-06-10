package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/gocolly/colly"
)

const (
	restockurl = "https://www.queens.cz/kat/2/boty-tenisky-panske/?sort=new"
)

type product struct {
	ID, Name, Price, Thumb, URL string
}

func atc(id, name, url, price, thumb string) {
	webhook1 := "https://monitors.blazingfa.st/queens/?token=Q3CFLMnPMWiszuKedCLD"
	client := &http.Client{}
	data := `{
		"name": "` + name + `",
		"thumb": "` + thumb + `",
		"price": "` + price + `",
		"id": "` + id + `",
		"url": "` + url + `"
	}`
	req, _ := http.NewRequest("POST", webhook1, strings.NewReader(data))

	req.Header.Add("content-type", "application/json")

	_, _ = client.Do(req)

}

// // Products from the Queens sitemap
// type Products struct {
// 	URL []struct {
// 		Loc   string `xml:"loc"`
// 		Image []struct {
// 			Loc   string `xml:"loc"`
// 			Title string `xml:"title"`
// 		} `xml:"image"`
// 	} `xml:"url"`
// }

// func parsexml() {
// 	// xmlFile, err := os.Open("2.changed.xml")
// 	// if err != nil {
// 	// 	fmt.Println(err)
// 	// }

// 	// defer xmlFile.Close()

// 	// byteValue, _ := ioutil.ReadAll(xmlFile)
// 	response, err := http.Get("https://www.queens.cz/sitemap/index/2")
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer response.Body.Close()
// 	contents, err := ioutil.ReadAll(response.Body)

// 	var products Products
// 	xml.Unmarshal(contents, &products)

// 	for i := 0; i < len(products.URL); i++ {
// 		arr := strings.Split(products.URL[i].Loc, "/")
// 		arr = append(arr[:3], arr[4:]...) // remove wear
// 		arr = append(arr[:4], arr[5:]...) // remove number after id
// 		fmt.Println(strings.Join(arr, "/"))
// 	}
// }

// var products []product

func getproducts(url string) []product {
	var products []product
	c := colly.NewCollector()
	c.OnHTML("#categoryItems", func(e *colly.HTMLElement) {
		e.ForEach(".category-item", func(i int, child *colly.HTMLElement) {

			// sk/global
			// fmt.Println(child.ChildAttr("a", "data-name"))
			// cz
			var p product
			p.ID = strings.Split(child.ChildAttr("a", "href"), "/")[4]
			p.Name = child.ChildAttr("a img", "alt")
			p.Thumb = child.ChildAttr("a img.imgr", "data-src")
			p.Price = child.ChildText(".price")
			p.URL = child.ChildAttr("a", "href")

			products = append(products, p)
		})
	})

	c.Visit(url)
	return products

}

func testEq(a, b []string) bool {

	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func helper() {

}

func main() {
	restockDB := redis.NewClient(&redis.Options{
		Addr:     "SG-queens-28130.servers.mongodirector.com:6379",
		Password: "1RUYz0QuhgujaCOKORYmdIa2xrm9Xj7a",
		DB:       1,  
	})
	// soldoutDB := redis.NewClient(&redis.Options{
	// 	Addr:     "localhost:6379",
	// 	Password: "", 
	// 	DB:       1,  
	// })

	for {
		{
			os.Setenv("HTTP_PROXY", "http://hicoria:samosa44@74.231.59.117:2000")
			restocks := getproducts(restockurl)
			// soldouts := getproducts(lasturl)
			fmt.Println("monitoring for new products")
			
			for _, p := range restocks {
				_, err := restockDB.Get(p.ID).Result()
				if err == redis.Nil {
					fmt.Println("new product found:", p.Name)
					atc(p.ID, p.Name, p.URL, p.Price, p.Thumb)
					// err := client.Set(p.ID, strings.Join(p.Sizes, "##"), 0).Err()
					err = restockDB.Set(p.ID, p.Name, 0).Err()
					if err != nil {
						panic(err)
					}
				} else if err != nil {
					panic(err)
				}
			}

		}
		time.Sleep(1 * time.Second)
	}
}
