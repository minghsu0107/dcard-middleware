package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	config *Config
	s      *miniredis.Miniredis
	server *Server
)

func TestMiddleware(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "middleware suite")
}

var _ = BeforeSuite(func() {
	config, _ = NewConfig()
	s = NewMiniRedis()
	server = NewTestServer(config)
})

func NewTestRedisLimiterRepository(s *miniredis.Miniredis, config *Config) *RedisLimiterRepository {
	return &RedisLimiterRepository{
		client: redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		}),
		ctx:        context.Background(),
		expiration: config.RedisConfig.CacheExpiration,
	}
}

func NewMiniRedis() *miniredis.Miniredis {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	return s
}

func NewTestServer(config *Config) *Server {
	testRedisLimiterRepository := NewTestRedisLimiterRepository(s, config)
	engine := NewEngine(config, NewRateLimiterMiddleware(config, testRedisLimiterRepository))
	server := NewServer(config, engine)
	server.RegisterRoutes()
	return server
}

func GetResponse(router *gin.Engine, url string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", url, nil)
	router.ServeHTTP(w, r)
	return w
}

var _ = Describe("middleware", func() {
	var url string
	var maxVisitCount int
	// BeforeEach blocks are run before It blocks
	BeforeEach(func() {
		url = "/hello"
		maxVisitCount = int(config.MaxVisitCount)
	})

	Describe("rate limiting", func() {
		It("should count requests correctly", func() {
			Context("when not exceeding max accepted requests", func() {
				for i := 0; i < maxVisitCount; i++ {
					w := GetResponse(server.Engine, url)
					Expect(w.Code).To(Equal(200))
				}
			})
			Context("when not exceeding max accepted requests", func() {
				for i := maxVisitCount; i < maxVisitCount*2; i++ {
					w := GetResponse(server.Engine, url)
					Expect(w.Code).To(Equal(429))
				}
			})
		})
		It("should maintain expiry", func() {
			s.FastForward(config.RedisConfig.CacheExpiration - 1*time.Second)
			w := GetResponse(server.Engine, url)
			Expect(w.Code).To(Equal(429))
		})
		It("should reset after rate limit expires", func() {
			s.FastForward(time.Second)
			w := GetResponse(server.Engine, url)
			Expect(w.Code).To(Equal(200))
		})
		It("should set X-RateLimit-Remaining and X-Ratelimit-Reset headers", func() {
			w := GetResponse(server.Engine, url)
			var err error
			var curVisitCount int
			var curResetSeconds int
			curVisitCount, err = strconv.Atoi(w.Result().Header.Get("X-RateLimit-Remaining"))
			Expect(err).NotTo(HaveOccurred())
			Expect(curVisitCount >= 1).To(BeTrue())
			Expect(curVisitCount < maxVisitCount).To(BeTrue())

			curResetSeconds, err = strconv.Atoi(w.Result().Header.Get("X-Ratelimit-Reset"))
			Expect(err).NotTo(HaveOccurred())
			Expect(curResetSeconds > 0).To(BeTrue())
			Expect(curResetSeconds <= int(config.RedisConfig.CacheExpiration.Seconds())).To(BeTrue())
		})
	})
})
