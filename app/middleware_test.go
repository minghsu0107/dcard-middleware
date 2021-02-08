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
)

func NewTestRedisLimiterRepository(s *miniredis.Miniredis, config *Config) *RedisLimiterRepository {
	return &RedisLimiterRepository{
		client: redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		}),
		ctx:        context.Background(),
		expiration: config.RedisConfig.CacheExpiration,
	}
}

func NewTestRecorder(router *gin.Engine) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w, r)
	return w
}

func Test_RateLimiter(t *testing.T) {
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	config, _ := NewConfig()
	testRedisLimiterRepository := NewTestRedisLimiterRepository(s, config)
	engine := NewEngine(config, NewRateLimiterMiddleware(config, testRedisLimiterRepository))
	server := NewServer(config, engine)
	server.RegisterRoutes()

	getResponse := func(router *gin.Engine) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/", nil)
		router.ServeHTTP(w, r)
		return w
	}

	maxVisitCount := int(config.MaxVisitCount)
	for i := 0; i < maxVisitCount*2; i++ {
		w := getResponse(server.Engine)
		if i < maxVisitCount && w.Code != 200 {
			t.Fatal("error before reach maximum visit count", i, w.Code, w.Result().Header)
		} else if i >= maxVisitCount && w.Code != 429 {
			t.Fatal("not sending 429 code", i, w.Code, w.Result().Header)
		}
	}

	// test if ttl works as expected
	s.FastForward(testRedisLimiterRepository.expiration - 1*time.Second)
	w := getResponse(server.Engine)
	if w.Code != 429 {
		t.Fatal("not sending 429 code before expiration")
	}
	s.FastForward(time.Second)
	w = getResponse(server.Engine)
	curVisitCount, err := strconv.Atoi(w.Result().Header.Get("X-RateLimit-Remaining"))
	if err != nil {
		t.Fatal("X-RateLimit-Remaining is not a number")
	}
	curResetSeconds, err := strconv.Atoi(w.Result().Header.Get("X-Ratelimit-Reset"))
	if err != nil {
		t.Fatal("X-Ratelimit-Reset is not a number")
	}
	if (curVisitCount != maxVisitCount-1) || (curResetSeconds != int(config.RedisConfig.CacheExpiration.Seconds())) {
		t.Fatal("TTL is not working")
	}
}
