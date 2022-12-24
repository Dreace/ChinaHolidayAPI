package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/valyala/fasthttp"
	"log"
	"time"
)

//go:embed holiday.json
var holidayData []byte

type Holiday struct {
	Type string `json:"type"`
	Note string `json:"note"`
}

var holidays map[string]Holiday

func isHoliday(date time.Time) (bool, string, string) {
	dateString := date.Format("2006-01-02")
	if holiday, ok := holidays[dateString]; ok {
		if holiday.Type == "假日" {
			return true, "假日", holiday.Note
		} else {
			return false, "补班日", "补班工作日"
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
	code := fasthttp.StatusOK
	message := map[string]any{}
	defer func() {
		ctx.SetStatusCode(code)
		body, _ := json.Marshal(message)
		ctx.SetBody(body)
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

func main() {
	var host string
	var port int
	pflag.IntVarP(&port, "port", "p", 80, "指定监听端口")
	pflag.StringVarP(&host, "host", "h", "127.0.0.1", "指定监听地址")

	pflag.Parse()

	err := json.Unmarshal(holidayData, &holidays)
	if err != nil {
		log.Fatalf("%v", err)
		return
	}
	log.Printf("listen on http://%s:%d", host, port)
	err = fasthttp.ListenAndServe(fmt.Sprintf("%s:%d", host, port), ShortColored(holidayHandler))
	if err != nil {
		log.Fatalf("%v", err)
	}
}
