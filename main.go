package main

import (
	"embed"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/spf13/pflag"
	"github.com/valyala/fasthttp"
	"sort"
	"strings"
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

func generateICS() string {
	currentYear := time.Now().Year()
	previousYear := currentYear - 1
	
	location, _ := time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(location)
	
	var icsBuilder strings.Builder
	icsBuilder.WriteString("BEGIN:VCALENDAR\r\n")
	icsBuilder.WriteString("VERSION:2.0\r\n")
	icsBuilder.WriteString("PRODID:-//ChinaHolidayAPI//Holiday Calendar//CN\r\n")
	icsBuilder.WriteString("CALSCALE:GREGORIAN\r\n")
	icsBuilder.WriteString("X-WR-CALNAME:中国法定节假日\r\n")
	icsBuilder.WriteString("X-WR-CALDESC:中国法定节假日和调休补班信息\r\n")
	icsBuilder.WriteString("X-WR-TIMEZONE:Asia/Shanghai\r\n")
	
	// Create a sorted list of dates
	var dateKeys []string
	for dateKey := range holidays {
		// Parse the date to check if it's in the target years
		if date, err := time.ParseInLocation("2006-01-02", dateKey, location); err == nil {
			year := date.Year()
			if year == previousYear || year == currentYear {
				dateKeys = append(dateKeys, dateKey)
			}
		}
	}
	sort.Strings(dateKeys)
	
	// Generate events for each holiday
	for _, dateKey := range dateKeys {
		holiday := holidays[dateKey]
		date, _ := time.ParseInLocation("2006-01-02", dateKey, location)
		
		// Create unique UID
		uid := fmt.Sprintf("%s@chinaholidayapi.com", strings.ReplaceAll(dateKey, "-", ""))
		
		// Format the date as YYYYMMDD
		dateFormatted := date.Format("20060102")
		
		// Create the event
		icsBuilder.WriteString("BEGIN:VEVENT\r\n")
		icsBuilder.WriteString(fmt.Sprintf("UID:%s\r\n", uid))
		icsBuilder.WriteString(fmt.Sprintf("DTSTART;VALUE=DATE:%s\r\n", dateFormatted))
		icsBuilder.WriteString(fmt.Sprintf("DTEND;VALUE=DATE:%s\r\n", dateFormatted))
		icsBuilder.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", now.Format("20060102T150405Z")))
		
		// Set summary based on holiday type
		var summary string
		if holiday.Type == "假日" {
			summary = fmt.Sprintf("🎉 %s", holiday.Note)
		} else if holiday.Type == "补班日" {
			summary = fmt.Sprintf("💼 %s", holiday.Note)
		} else {
			summary = holiday.Note
		}
		
		icsBuilder.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", summary))
		icsBuilder.WriteString(fmt.Sprintf("DESCRIPTION:类型: %s\\n说明: %s\r\n", holiday.Type, holiday.Note))
		
		// Add category
		if holiday.Type == "假日" {
			icsBuilder.WriteString("CATEGORIES:假日\r\n")
		} else if holiday.Type == "补班日" {
			icsBuilder.WriteString("CATEGORIES:补班日\r\n")
		}
		
		icsBuilder.WriteString("END:VEVENT\r\n")
	}
	
	icsBuilder.WriteString("END:VCALENDAR\r\n")
	return icsBuilder.String()
}

func icsHandler(ctx *fasthttp.RequestCtx) {
	recordRequest()
	
	icsContent := generateICS()
	
	// Set appropriate headers for ICS file
	ctx.SetContentType("text/calendar; charset=utf-8")
	ctx.Response.Header.Set("Content-Disposition", "attachment; filename=\"china-holidays.ics\"")
	ctx.Response.Header.Set("Cache-Control", "public, max-age=3600")
	
	ctx.SetBodyString(icsContent)
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
		case "/ics":
			icsHandler(ctx)
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
