# url-shortener

In this project I am going to follow a tutorial for a url shortener in **Go** and I'll be using **Redis** as a store mechanism for super fast data retrieval in the implementation.

### 1. Project setup
lets setup the project and install all the dependencies that will be needed.
- initialize the go project.

```
go mod init github.com/eDyrr/url-shortener
```

- create `main.go` file and add the code below for checking the setup.

```
package main

import "fmt"

func main() {
	fmt.Printf("hello go url shortener")
}
```

then run `go run main.go`.

- installing project dependencies.
```
go get github.com/go-redis/redis/v8
```

```
go get github.com/gin-gonic/gin
```
- installing redis locally

### 2. Igniting the web server

now we can launch the web server, and return some data from the its API endpoint.

here's the updated `main.go` file where we create a server which returns a message with some data at the root endpoint ("/")

```
package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "hey Go URL shortener",
		})
	})

	err := r.Run(":9808")
	if err != nil {
		panic(fmt.Sprintf("Failed to start the web server - Error %v", err))
	}
}
```

run the following to get the content of `message`

```
curl -X GET http://localhost:9808/
```

here's the expected output

```
{"message":"hey Go URL shortener"}
```

---

in this segment I am going to focus on building the storage layer of our application, so mainly we're going to:

1. setup the store service.
2. storage API design and implementation.
3. unit & integration testing.

### 1. Store service Setup

first we create our `store` folder, then we create 2 go file `store.service.go` and `store.service_test.go`

- we will start by setting up our wrappers around Redis, the wrappers will be used as interface for persisting and retrieving our application data mapping.

here's the `store.service.go` file:

```
package store

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// define the struct wrapper around raw Redis client
type StorageService struct {
	redisClient *redis.Client
}

// top level declarations for the storeService and Redis context
var (
	storeService = &StorageService{}
	ctx          = context.Background()
)

// note that in a real world usage, the cache duration shouldnt have
// an expiration time, an LRU policy should be set where the
// values that are retrieved less often are purged automatically from
// the cache and stored back in RDBMS whenever the cache is full

const CacheDuration = 6 * time.Hour
```

- after defining wrapper structs we can finally be able to initialize the store service, in this case our Redis client.

```
// initializing the store service and return a store pointer
func InitializeStore() *StorageService {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Error init Redis %v:", err))
	}

	fmt.Printf("\nRedis started successfully: pong message = {%s}", pong)
	storeService.redisClient = redisClient
	return storeService
}
```

### 2. storage API design and implementation

```
// we want to be able to save the mapping between the original url
// and the generated url
func SaveUrlMapping(shortUrl string, originalUrl string, userId string) {

}

// we should be able to retrieve the initial long url once the short is provided
// this is when users will be calling the short link in the url, so what we need 
// to do is to retrieve the long url and think about redirect.
func RetrieveInitialUrl(shortUrl string) string {

}
```

- the next step is to implement our storage API.

```
func SaveUrlMapping(shortUrl string, originalUrl string, userId string) {
	err := storeService.redisClient.Set(ctx, shortUrl, originalUrl, cacheDurations).Err()
	if err != nil {
		// handle err
	}
}

func RetrieveUrl(shortUrl string) string {
	res, err := storageService.redisClient.Get(ctx, shortUrl).Result()
	if err != nil {
		// handle err
	}
	return res
}
```

### 3. Unit and Integration testing

to preserve the best practices and avoiding unintentional regressions in the future, we are going to have to think about unit and integration tests for our storage layer implementation, now lets install the testing tools:

```
go get github.com/stretchr/testify
```

- first we setup the testing env

```
package store

var testStoreService = &StorageService{}

func init() {
	testStoreService = InitializeStore()
}
```

- now we unit test the store init.

```
func TestStoreInit(t *testing.T) {
	assert.True(t, testStoreService != nil)
}
```

- finally we will test for the storage APIs

```
func TestInsertionAndRetrieval(t *testing.T) {
	originalURL := "https://www.guru3d.com/news-story/spotted-ryzen-threadripper-pro-3995wx-processor-with-8-channel-ddr4,2.html"
	shortURL := "e0dba740-fc4b-4977-872c-d360239e6b1a"
	userUUId := "Jsz4k57oAX"

	// persist data mapping
	SaveUrlMapping(shortURL, originalURL, userUUId)

	retrievedURL := RetrieveInitialUrl(shortURL)

	assert.Equal(t, retrievedURL, originalURL)
}
```

run the tests and they should pass.

---

Now we are going to work on the algorithm we will be using to hash and process the initial input or the long url into a smaller and shorter mapping that corresponds to it.

when doing the choice for the algorithm we do have a number of objectives to keep in mind:
- the final input should be shorter: maximum 8 characters.
- should be easily human readable, avoid confusing characters mix up, character that often similar in most fonts.
- the entropy should be fairly large to avoid collision in short link generation.

### 1. Generator algorithm

during this implementation we are going to use two main schemes:
a hash function and a binary to text encoding algorithm.

first, we create 2 files, `shorturl_generator.go` and `shorturl_generator_test.go`, and put them under a folder called `shortener`.

### 2. Shortener implementation

#### 2.1. SHA256

we will be using SHA256 to hash the initial inputs.
we will be using Golang's built-in implementation:

```
// shorturl_generator.go
package shortener

import "crypto/sha256"

func sha2560f(input string) []byte {
	algorithm := sha256.New()
	algorithm.Write([]byte(input))
	return algorithm.Sum(nil)
}
```

#### 2.2. BASE58

this binary to text will be used to provide the final output of the process.

first, we install the base58 dependency library:

```
go get github.com/itchyny/bas58-go/base58
```

now we add the implementation code:

```
package shortener

import (
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/itchyny/base58-go"
)

func sha2560f(input string) []byte {
	algorithm := sha256.New()
	algorithm.Write([]byte(input))
	return algorithm.Sum(nil)
}

func base58Encoded(bytes []byte) string {
	encoding := base58.BitcoinEncoding
	encoded, err := encoding.Encode(bytes)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return string(encoded)
}
```

#### 2.3. Final algorithm

the final algorithm will be super straightforward now as we have our 2 main building blocks already setup, it will go as follow:

- hashing `initialUrl + userId` url with **sha256**. here `userId` is added to prevent providing similar shortened urls to separate users in case they want to shorten exact same link, its a design decision, so some implementations do this differently.

- derive a big integer number from the hash bytes generated during the hashing.

- finally apply **base58** on the derived big integer value and pick the first 8 characters.

```
func GeneratedLink(initialLink string, userId string) string {
	urlHashBytes := sha256(initialLink + userId)
	generatedNumber := new(big.Int).SetBytes(urlHashBytes).Uint64()
	finalString := base58Encoded([]byte(fmt.Sprintf("%d", generatedNumber)))
	return finalString[:8]
}
```

#### 3. Shortener unit tests

now we write tests for the algorithm implementation:

```
package shortener

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const UserId = "e0dba740-fc4b-4977-872c-d360239e6b1a"

func TestShortLinkGenerator(t *testing.T) {
	initialLink1 := "https://www.guru3d.com/news-story/spotted-ryzen-threadripper-pro-3995wx-processor-with-8-channel-ddr4,2.html"
	shortLink1 := GeneratedShortLink(initialLink1, UserId)

	initialLink2 := "https://www.eddywm.com/lets-build-a-url-shortener-in-go-with-redis-part-2-storage-layer/"
	shortLink2 := GeneratedShortLink(initialLink2, UserId)

	initialLink3 := "https://spectrum.ieee.org/automaton/robotics/home-robots/hello-robots-stretch-mobile-manipulator"
	shortLink3 := GeneratedShortLink(initialLink3, UserId)

	assert.Equal(t, shortLink1, "jTa4L57P")
	assert.Equal(t, shortLink2, "d66yfx7N")
	assert.Equal(t, shortLink3, "dhZTayYQ")
}
```

---

now we'll begin to code the APIs, to be specific 2 main endpoints to our API service:

- one endpoint that will be used to generate a short url and return it, when the initial long url is provided. `/create-short-url`

- the other one will be used to provide the actual redirection from the shortend version to the original longer URL `/:short-url`

### 1. Handlers and endpoints

#### 1.1. setup and definitions

lets create our `handler` package and define our handers functions in there.

create a folder called `handler` with a `handler.go` file.

after that's done, lets define our handlers stubs.

```
package handler

import "github.com/gin-gonic/gin"

func CreateShortUrl(c *gin.Context) {

}

func HandleShortUrlRedirects(c *gin.Context) {
	
}
```

after that's done, we go to the `main.go` file and add our **endpoints**.

```
package main

import (
	"fmt"
	"net/http"

	"github.com/eDyrr/url-shortener/handler"
	"github.com/eDyrr/url-shortener/store"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "hey Go URL shortener",
		})
	})

	r.POST("/create-short-url", func(c *gin.Context) {
		handler.CreateShortUrl(c)
	})

	r.GET("/:short-url", func(c *gin.Context) {
		handler.HandleShortUrlRedirects(c)
	})

	store.InitializeStore()
	err := r.Run(":9808")
	if err != nil {
		panic(fmt.Sprintf("Failed to start the web server - Error %v", err))
	}
}
```

#### 1.2. Implementations

now its time to write the implementation code.

**STEP 1**: implement the `CreateShortUrl()` handler function:

- recieve creation request body, parse it then extract the initial long url and the userId.

- call `shortener.GenerateShortLink()` that we implemented and generate our shortened hash.

- finally store the mapping of our output `hash/shortUrl` with the initial long url, here, we will be using the `store.SaveUrlMapping()` we implemented.

```
package handler

import (
	"net/http"

	"github.com/eDyrr/url-shortener/shortener"
	"github.com/eDyrr/url-shortener/store"
	"github.com/gin-gonic/gin"
)

type UrlCreationRequest struct {
	LongUrl string `json:"long_url" binding:"required"`
	UserId  string `json:"user_id" binding:"required"`
}

func CreateShortUrl(c *gin.Context) {
	var creationRequest UrlCreationRequest
	if err := c.ShouldBindJSON(&creationRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shortUrl := shortener.GeneratedShortLink(creationRequest.LongUrl, creationRequest.UserId)
	store.SaveUrlMapping(shortUrl, creationRequest.LongUrl, creationRequest.UserId)

	host := "http://localhost:9808/"
	c.JSON(200, gin.H{
		"message":   "short url created successfully",
		"short_url": host + shortUrl,
	})
}
```

**STEP 2**: the second and last step will be about implementing the redirection handler, `handleShortUrlRedirect()`, it will consist of:

- getting the short url from the path parameter `/:shortUrl`

- call the store to retrieve the initial url that corresponds to the short one provided in the path.

- and finally apply the http redirection function.

```
func HandleShortUrlRedirects(c *gin.Context) {
	shortUrl := c.Param("shortUrl")
	initialUrl := store.RetrieveInitialUrl(shortUrl)
	c.Redirect(302, initialUrl)
}
```

### 2. Testing

after finishing the handlers, now lets test them.

- **Step 1**: run/start the project (`main.go` file is the entry point).

```
{"message":"hey Go URL shortener"}
```

- **Step 2**: request url shortening action.

we can post the request body below to the specified endpoing.

```
curl --request POST \
--data '{
    "long_url": "https://www.guru3d.com/news-story/spotted-ryzen-threadripper-pro-3995wx-processor-with-8-channel-ddr4,2.html",
    "user_id" : "e0dba740-fc4b-4977-872c-d360239e6b10"
}' \
  http://localhost:9808/create-short-url
```

here's the output:

```
{
    "message": "short url created successfully",
    "short_url": "http://localhost:9808/9Zatkhpi"
}
```

**Step 3**: testing the redirection
