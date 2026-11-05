package main

import (
	"encoding/json"
	"testing"
	"time"
)

// Test2027OfficialHolidaySchedule verifies the published State Council schedule.
// Source: https://society.people.com.cn/n1/2026/1105/c1008-40828482.html
func Test2027OfficialHolidaySchedule(t *testing.T) {
	data, err := holidayFiles.ReadFile("holidays/2027.json")
	if err != nil {
		t.Fatal(err)
	}
	var actual map[string]Holiday
	if err := json.Unmarshal(data, &actual); err != nil {
		t.Fatal(err)
	}

	expected := make(map[string]Holiday)
	for _, period := range []struct{ start, end, name string }{
		{"2027-01-01", "2027-01-03", "元旦"},
		{"2027-02-05", "2027-02-13", "春节"},
		{"2027-04-03", "2027-04-05", "清明节"},
		{"2027-05-01", "2027-05-05", "劳动节"},
		{"2027-06-09", "2027-06-09", "端午节"},
		{"2027-09-15", "2027-09-15", "中秋节"},
		{"2027-10-01", "2027-10-07", "国庆节"},
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
		"2027-01-31": "春节",
		"2027-02-14": "春节",
		"2027-05-08": "劳动节",
		"2027-09-26": "国庆节",
		"2027-10-09": "国庆节",
	} {
		expected[date] = Holiday{Type: "补班日", Note: name + "补班"}
	}

	if len(expected) != 34 {
		t.Fatalf("expected schedule contains %d entries, want 34", len(expected))
	}
	if len(actual) != len(expected) {
		t.Errorf("schedule has %d entries, want %d", len(actual), len(expected))
	}
	for date, want := range expected {
		if got, ok := actual[date]; !ok {
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

func Test2027HolidayClassification(t *testing.T) {
	data, err := holidayFiles.ReadFile("holidays/2027.json")
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
		date string
		holiday bool
		kind, note string
	}{
		{"2027-01-01", true, "假日", "元旦"},
		{"2027-01-31", false, "工作日", "补班工作日"},
		{"2027-02-14", false, "工作日", "补班工作日"},
		{"2027-06-09", true, "假日", "端午节"},
		{"2027-09-15", true, "假日", "中秋节"},
		{"2027-09-26", false, "工作日", "补班工作日"},
		{"2027-10-09", false, "工作日", "补班工作日"},
		{"2027-06-12", true, "假日", "周末"},
		{"2027-06-10", false, "工作日", "普通工作日"},
	} {
		day, err := time.Parse("2006-01-02", tc.date)
		if err != nil {
			t.Fatal(err)
		}
		gotHoliday, gotKind, gotNote := isHoliday(day)
		if gotHoliday != tc.holiday || gotKind != tc.kind || gotNote != tc.note {
			t.Errorf("%s: got (%t, %q, %q), want (%t, %q, %q)",
				tc.date, gotHoliday, gotKind, gotNote, tc.holiday, tc.kind, tc.note)
		}
	}
}
