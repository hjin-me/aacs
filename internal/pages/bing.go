package pages

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type BingResp struct {
	Images []*PicInfo `json:"images"`
}
type PicInfo struct {
	Url       string `json:"url"`
	Copyright string `json:"copyright"`
	Title     string `json:"title"`
}

var lastBingPic *PicInfo = nil
var mux sync.Mutex
var once sync.Once

func BingPic(ctx context.Context) (*PicInfo, error) {
	once.Do(func() {
		refreshBing()
	})
	select {
	case <-ctx.Done():
		return nil, nil
	default:

	}
	if lastBingPic != nil {
		return lastBingPic, nil
	}
	r, err := getBingPic(ctx)
	if err != nil {
		return nil, err
	}

	if lastBingPic == nil {
		lastBingPic = r
	}
	return lastBingPic, nil
}

func getBingPic(ctx context.Context) (*PicInfo, error) {
	mux.Lock()
	defer mux.Unlock()

	resp, err := otelhttp.Get(ctx, "https://cn.bing.com/HPImageArchive.aspx?format=js&idx=0&n=1&mkt=zh-CN")
	if err != nil {
		return nil, errors.WithMessage(err, "请求 bing 失败")
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithMessage(err, "请求解析壁纸失败")
	}
	data := BingResp{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		return nil, errors.WithMessage(err, "请求解析壁纸失败")
	}
	if len(data.Images) < 1 {
		return nil, errors.New("没有找到壁纸")
	}
	data.Images[0].Url = "https://cn.bing.com" + data.Images[0].Url

	return data.Images[0], nil
}
func refreshBing() {
	r, err := getBingPic(context.Background())
	if err != nil {
		panic(err)
	}
	lastBingPic = r
	go func() {
		for range time.Tick(time.Hour) {
			r, err := getBingPic(context.Background())
			if err != nil {
				continue
			}
			lastBingPic = r
		}
	}()
}
