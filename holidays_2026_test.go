package main

import (
	"encoding/json"
	"testing"
	"time"
)

// Test2026OfficialHolidaySchedule checks the schedule in 国办发明电〔2025〕7号.
// Source: http://www.scio.gov.cn/zdgz/jj/202511/t20251110_938367_m.html
func Test2026OfficialHolidaySchedule(t *testing.T) {
	data, err := holidayFiles.ReadFile("holidays/2026.json")
	if err != nil {
		t.Fatal(err)
	}

	var actual map[string]Holiday
	if err := json.Unmarshal(data, &actual); err != nil {
		t.Fatal(err)
	}

	expected := make(map[string]Holiday)
	for _, period := range []struct {
		start, end, name string
	}{
		{"2026-01-01", "2026-01-03", "元旦"},
		{"2026-02-15", "2026-02-23", "春节"},
		{"2026-04-04", "2026-04-06", "清明节"},
		{"2026-05-01", "2026-05-05", "劳动节"},
		{"2026-06-19", "2026-06-21", "端午节"},
		{"2026-09-25", "2026-09-27", "中秋节"},
		{"2026-10-01", "2026-10-07", "国庆节"},
	} {
		start, err := time.Parse("2006-01-02", period.start)
		if err != nil {
			t.Fatal(err)
		}
		end, err := time.Parse("2006-01-02", period.end)
		if err != nil {
			t.Fatal(err)
		}
		for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
			expected[day.Format("2006-01-02")] = Holiday{Type: "假日", Note: period.name}
		}
	}

	for date, name := range map[string]string{
		"2026-01-04": "元旦",
		"2026-02-14": "春节",
		"2026-02-28": "春节",
		"2026-05-09": "劳动节",
		"2026-09-20": "国庆节",
		"2026-10-10": "国庆节",
	} {
		expected[date] = Holiday{Type: "补班日", Note: name + "补班"}
	}

	if len(expected) != 39 {
		t.Fatalf("test schedule has %d dates, want 39", len(expected))
	}
	if len(actual) != len(expected) {
		t.Errorf("2026 schedule has %d dates, want %d", len(actual), len(expected))
	}
	for date, want := range expected {
		got, ok := actual[date]
		if !ok {
			t.Errorf("missing date %s", date)
		} else if got != want {
			t.Errorf("%s: got %+v, want %+v", date, got, want)
		}
	}
	for date := range actual {
		if _, ok := expected[date]; !ok {
			t.Errorf("unexpected date %s", date)
		}
	}
}

func Test2026HolidayClassification(t *testing.T) {
	data, err := holidayFiles.ReadFile("holidays/2026.json")
	if err != nil {
		t.Fatal(err)
	}
	var schedule map[string]Holiday
	if err := json.Unmarshal(data, &schedule); err != nil {
		t.Fatal(err)
	}

	original := holidays
	holidays = schedule
	t.Cleanup(func() { holidays = original })

	for _, tc := range []struct {
		date      string
		isHoliday bool
		kind      string
		note      string
	}{
		{"2026-01-04", false, "工作日", "补班工作日"},
		{"2026-02-15", true, "假日", "春节"},
		{"2026-09-20", false, "工作日", "补班工作日"},
		{"2026-09-25", true, "假日", "中秋节"},
		{"2026-09-19", true, "假日", "周末"},
		{"2026-09-18", false, "工作日", "普通工作日"},
	} {
		day, err := time.Parse("2006-01-02", tc.date)
		if err != nil {
			t.Fatal(err)
		}
		isHoliday, kind, note := isHoliday(day)
		if isHoliday != tc.isHoliday || kind != tc.kind || note != tc.note {
			t.Errorf("%s: got (%t, %q, %q), want (%t, %q, %q)",
				tc.date, isHoliday, kind, note, tc.isHoliday, tc.kind, tc.note)
		}
	}
}
