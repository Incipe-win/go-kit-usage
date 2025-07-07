package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/log"
)

// 1 业务逻辑抽象为接口
type AddService interface {
	Sum(ctx context.Context, a, b int) (int, error)
	Concat(ctx context.Context, a, b string) (string, error)
}

// 实现接口
type addService struct {
}

var (
	ErrOverflow    = errors.New("overflow")
	ErrEmptyString = errors.New("empty string")
	ErrInternal    = errors.New("internal error")
)

// Sum 返回两个数的和
func (s *addService) Sum(ctx context.Context, a, b int) (int, error) {
	ret := a + b
	if ret > math.MaxInt || ret < math.MinInt {
		return 0, ErrOverflow
	}
	return ret, nil
}

// Concat 返回两个字符串的拼接
func (s *addService) Concat(ctx context.Context, a, b string) (string, error) {
	if a == "" || b == "" {
		return "", ErrEmptyString
	}
	return a + b, nil
}

func NewService() AddService {
	return &addService{}
}

type logMiddleware struct {
	logger log.Logger
	next   AddService
}

func NewLogMiddleware(logger log.Logger, srv AddService) AddService {
	return &logMiddleware{
		logger: logger,
		next:   srv,
	}
}

func (s *logMiddleware) Sum(ctx context.Context, a, b int) (res int, err error) {
	defer func(start time.Time) {
		s.logger.Log(
			"method", "sum",
			"a", a,
			"b", b,
			"res", res,
			"err", err,
			"cast", time.Since(start),
		)
	}(time.Now())
	res, err = s.next.Sum(ctx, a, b)
	return
}

func (s *logMiddleware) Concat(ctx context.Context, a, b string) (res string, err error) {
	defer func(start time.Time) {
		s.logger.Log(
			"method", "concat",
			"a", a,
			"b", b,
			"res", res,
			"err", err,
			"cast", time.Since(start),
		)
	}(time.Now())
	res, err = s.next.Concat(ctx, a, b)
	return
}

type instrumentingMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	countResult    metrics.Histogram
	next           AddService
}

func (im *instrumentingMiddleware) Sum(ctx context.Context, a, b int) (res int, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "sum", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
		im.countResult.Observe(float64(res))
	}(time.Now())
	res, err = im.next.Sum(ctx, a, b)
	return
}

func (im *instrumentingMiddleware) Concat(ctx context.Context, a, b string) (res string, err error) {
	defer func(begin time.Time) {
		lvs := []string{"method", "concat", "error", fmt.Sprint(err != nil)}
		im.requestCount.With(lvs...).Add(1)
		im.requestLatency.With(lvs...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	res, err = im.next.Concat(ctx, a, b)
	return
}

type withTrimMiddleware struct {
	next        AddService
	trimService endpoint.Endpoint
}

func NewServiceWithTrim(trimEndpoint endpoint.Endpoint, srv AddService) AddService {
	return &withTrimMiddleware{
		trimService: trimEndpoint,
		next:        srv,
	}
}

// 为 withTrimMiddleware 实现 AddService 接口
func (tm *withTrimMiddleware) Sum(ctx context.Context, a, b int) (res int, err error) {
	return tm.next.Sum(ctx, a, b)
}

func (tm *withTrimMiddleware) Concat(ctx context.Context, a, b string) (res string, err error) {
	// 需要新的逻辑处理
	// 外部调用我们的Concat方法时
	// 1. 发起RPC调用 trim_service 对数据进行处理 （调用其他服务/依赖其他的服务）
	respA, err := tm.trimService(ctx, TrimRequest{S: a})
	if err != nil {
		return "", err
	}
	respB, err := tm.trimService(ctx, TrimRequest{S: b})
	if err != nil {
		return "", err
	}
	trimA, trimB := respA.(TrimResponse), respB.(TrimResponse)

	// 2. 拿到处理后的数据再拼接
	return tm.next.Concat(ctx, trimA.S, trimB.S)
}
