访问 [https://holiday.dreace.top](https://holiday.dreace.top) 可以获取当天是否为假期，可能的返回：
```json
{
    "date": "2022-12-25",
    "isHoliday": true,
    "note": "周末",
    "type": "假日"
}
```
可以通过 `date` 参数指定要查询的日期，例如 [https://holiday.dreace.top?date=2023-01-02](https://holiday.dreace.top?date=2023-01-02) 将返回：
```json
{
    "date": "2023-01-02",
    "isHoliday": true,
    "note": "元旦",
    "type": "假日"
}
```