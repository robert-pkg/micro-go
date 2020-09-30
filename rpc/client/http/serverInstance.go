package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"time"

	"io/ioutil"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/robert-pkg/micro-go/log"
	"github.com/robert-pkg/micro-go/registry"
	"github.com/robert-pkg/micro-go/rpc/metadata"
)

// 服务实例
type serverInstance struct {
	service *registry.Service
}

func (instance *serverInstance) GetAddr() string {
	return instance.service.Nodes[0].Address
}

func (instance *serverInstance) Call(ctx context.Context, method string, url string, reqBody []byte) (respBody []byte, err error) {

	var client *http.Client

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		//DisableKeepAlives: true,
	}
	client = &http.Client{
		Transport: tr,
		Timeout:   time.Second * 10,
	}

	reqest, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Error("err", "err", err)
		return nil, err
	}

	if true {
		if md, ok := metadata.FromContext(ctx); ok {
			for k, v := range md {
				reqest.Header.Set(k, v)
			}
		}
	}

	tracer := opentracing.GlobalTracer()
	if tracer != nil {

		var parentCtx opentracing.SpanContext
		if pc := ctx.Value("ParentSpanContext"); pc != nil {
			if realPC, ok := pc.(opentracing.SpanContext); ok {
				parentCtx = realPC
			}
		}

		span := tracer.StartSpan(
			method,
			opentracing.ChildOf(parentCtx), // can be nil
			opentracing.Tag{Key: string(ext.Component), Value: "HTTP"},
			ext.SpanKindRPCClient,
		)
		defer span.Finish()

		// 往 http header 中注入 trace 信息
		carrier := opentracing.HTTPHeadersCarrier(reqest.Header)
		err = tracer.Inject(span.Context(), opentracing.HTTPHeaders, carrier)
		if err != nil {
			log.Error("err", "err", err)
			return nil, err
		}
	}

	response, err := client.Do(reqest)
	if err != nil {
		log.Error("err", "err", err)
		return nil, err
	}
	defer response.Body.Close()

	respBody, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Error("err", "err", err)
		return nil, err
	}

	if response.StatusCode != 200 {
		err = errors.Errorf("error: %s", string(respBody))
		log.Error("err", "response.StatusCode", response.StatusCode, "err", err)
		return nil, err
	}

	return
}
