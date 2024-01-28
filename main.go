package main

import (
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/valyala/fasthttp"
	"sort"
	"sync"
	"time"
)

//go:embed holidays/*.json
var holidayFiles embed.FS

type Holiday struct {
	Type string `json:"type"`
	Note string `json:"note"`
}

var holidays = make(map[string]Holiday)

type APIStats struct {
	Daily   sync.Map // YYYY-MM-DD
	Monthly sync.Map // YYYY-MM
	Years   []string
}

type APIStatsResponse struct {
	RequestCount struct {
		Daily   int `json:"daily"`
		Monthly int `json:"monthly"`
	} `json:"requestCount"`
	Years []string `json:"years"`
}

var stats = APIStats{}

func recordRequest() {
	today := time.Now().Format("2006-01-02")
	month := time.Now().Format("2006-01")

	// 清理过期的每日数据
	stats.Daily.Range(func(key, value interface{}) bool {
		if key.(string) != today {
			stats.Daily.Delete(key)
		}
		return true
	})

	// 清理过期的每月数据
	stats.Monthly.Range(func(key, value interface{}) bool {
		if key.(string) != month {
			stats.Monthly.Delete(key)
		}
		return true
	})

	// 更新或初始化今天的计数
	increment(&stats.Daily, today)

	// 更新或初始化本月的计数
	increment(&stats.Monthly, month)
}

func increment(m *sync.Map, key string) {
	val, _ := m.LoadOrStore(key, 0)
	if count, ok := val.(int); ok {
		m.Store(key, count+1)
	}
}

func isHoliday(date time.Time) (bool, string, string) {
	dateString := date.Format("2006-01-02")
	if holiday, ok := holidays[dateString]; ok {
		if holiday.Type == "假日" {
			return true, "假日", holiday.Note
		} else {
			return false, "工作日", "补班工作日"
		}
	} else {
		if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
			return true, "假日", "周末"
		} else {
			return false, "工作日", "普通工作日"
		}
	}
}

func holidayHandler(ctx *fasthttp.RequestCtx) {
	recordRequest()
	code := fasthttp.StatusOK
	message := map[string]any{}
	defer func() {
		ctx.SetStatusCode(code)
		body, _ := json.Marshal(message)
		ctx.SetBody(body)
		ctx.SetContentTypeBytes([]byte("application/json"))
	}()

	date := time.Now()
	dateBytes := ctx.QueryArgs().Peek("date")
	if len(dateBytes) > 1 {
		location, _ := time.LoadLocation("Asia/Shanghai")
		var err error
		date, err = time.ParseInLocation("2006-01-02", string(dateBytes), location)
		if err != nil {
			code = fasthttp.StatusBadRequest
			message["message"] = fmt.Sprintf("无法解析时间：%s", err)
			return
		}
	}
	message["date"] = date.Format("2006-01-02")

	holiday, dateType, note := isHoliday(date)
	message["isHoliday"] = holiday
	message["type"] = dateType
	message["note"] = note
}

func statsHandler(ctx *fasthttp.RequestCtx) {
	response := APIStatsResponse{}

	// 每日统计
	stats.Daily.Range(func(key, value interface{}) bool {
		response.RequestCount.Daily = value.(int)
		return true
	})

	// 每月统计
	stats.Monthly.Range(func(key, value interface{}) bool {
		response.RequestCount.Monthly = value.(int)
		return true
	})
	response.Years = stats.Years

	// 将统计信息编码为JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		ctx.Error("Error generating JSON", fasthttp.StatusInternalServerError)
		return
	}

	// 设置响应类型为JSON
	ctx.SetContentType("application/json")
	fmt.Fprintf(ctx, "%s", jsonResponse)
}

func main() {
	var host string
	var port int
	pflag.IntVarP(&port, "port", "p", 80, "指定监听端口")
	pflag.StringVarP(&host, "host", "h", "127.0.0.1", "指定监听地址")

	pflag.Parse()

	dirEntries, err := holidayFiles.ReadDir("holidays")
	if err != nil {
		panic(err)
	}

	years := make([]string, 0)
	// Iterate through the embedded files
	for _, entry := range dirEntries {
		if !entry.IsDir() && entry.Type().IsRegular() {
			yearName := entry.Name()[:4]
			years = append(years, yearName)
			filePath := "holidays/" + entry.Name()
			data, err := holidayFiles.ReadFile(filePath)
			if err != nil {
				panic(err)
			}

			var itemList map[string]Holiday
			err = json.Unmarshal(data, &itemList)
			if err != nil {
				panic(err)
			}

			for k, v := range itemList {
				holidays[k] = v
			}
		}
	}
	sort.Strings(years)
	stats.Years = years

	output.Printf("listen on http://%s:%d", host, port)

	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/stats":
			statsHandler(ctx)
			break
		default:
			holidayHandler(ctx)
		}
	}

	err = fasthttp.ListenAndServe(fmt.Sprintf("%s:%d", host, port), ShortColored(m))
	if err != nil {
		output.Fatalf("%v", err)
	}
}
