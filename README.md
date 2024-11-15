## 介绍
<p align="center">
    <img alt="Dynamic JSON Badge" src="https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fholiday.dreace.top%2Fstats&query=%24.requestCount.daily&label=%E6%9C%8D%E5%8A%A1%E6%AC%A1%E6%95%B0%EF%BC%88%E5%A4%A9%EF%BC%89&cacheSeconds=3600">
    <img alt="Dynamic JSON Badge" src="https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fholiday.dreace.top%2Fstats&query=%24.requestCount.monthly&label=%E6%9C%8D%E5%8A%A1%E6%AC%A1%E6%95%B0%EF%BC%88%E6%9C%88%EF%BC%89&cacheSeconds=3600">
    <img alt="Dynamic JSON Badge" src="https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fholiday.dreace.top%2Fstats&query=%24..years%5B-1%3A%5D&label=%E6%9B%B4%E6%96%B0%E5%88%B0&color=orange&cacheSeconds=3600">
    <img alt="Dynamic JSON Badge" src="https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fholiday.dreace.top&query=%24.date&label=date&color=red&cacheSeconds=3600">
    <img alt="Dynamic JSON Badge" src="https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fholiday.dreace.top&query=%24.isHoliday&label=isHoliday&color=green&cacheSeconds=3600">
    <img alt="Dynamic JSON Badge" src="https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fholiday.dreace.top&query=%24.type&label=type&color=blue&cacheSeconds=3600">
    <img alt="Dynamic JSON Badge" src="https://img.shields.io/badge/dynamic/json?url=https%3A%2F%2Fholiday.dreace.top&query=%24.note&label=note&color=yellow&cacheSeconds=3600">
</p>



提供 HTTP 服务（Go）查询当前或指定日期是否为假日，返回中区分了假日、工作日。可以用于在特定日期类型完成自动化，比如使用 iOS 的快捷指令自动记录基金定投等。

提供脚本（Python）自动从 [中国政府官网](http://www.gov.cn/) 生成调休、补班数据。[生成假期 JSON 文件](#生成假期-json-文件) 提供了其他生成方式可供参考。

可以自行从 [Releases](https://github.com/Dreace/ChinaHolidayAPI/releases) 下载可执行文件并部署，或直接访问 [https://holiday.dreace.top](https://holiday.dreace.top) 查询假日。

目前可以查询 2023 至 2025 年的假期数据。

## 启动参数

|   短参数名   | 长参数名 | 类型     | 必填  | 说明                   | 示例        |
|:--------:|------|--------|-----|----------------------|-----------|
| `--host` | `-h` | string | 否   | 监听地址，默认为 `127.0.0.1` | `0.0.0.0` |
| `--port` | `-p` | int    | 否   | 监听端口，默认为 `80`        | `8081`    |

## HTTP 接口参数

### 请求参数

|  参数名   | 类型     | 必填  | 说明               | 示例           |
|:------:|--------|-----|------------------|--------------|
| `date` | string | 否   | 要查询的日期，如果不填则查询当天 | `2023-01-02` |

### 响应参数

|     参数名     | 类型     | 说明                                                                                                                     | 示例           |
|:-----------:|--------|------------------------------------------------------------------------------------------------------------------------|--------------|
|   `date`    | string | 查询的日期                                                                                                                  | `2023-01-02` |
| `isHoliday` | bool   | 查询的日期是否为假期                                                                                                             | `false`      |
|   `type`    | string | 查询的日期类型，可能为：<br/> `假日` <br/> `工作日`                                                                                     | `假日`         |
|   `note`    | string | 对日期的详细描述，当 `type` 为 `假日` 时可能为：<br/> `周末`<br/>`<假日描述>`（非固定，可能是 `元旦节` 等）<br/>当 `type` 为工作日时，可能为：<br/>`普通工作日`<br/>`补班工作日` | `普通工作日`      |

## 例子

访问 [https://holiday.dreace.top](https://holiday.dreace.top) 可以获取当天是否为假期，可能的返回：

```json
{
  "date": "2022-12-25",
  "isHoliday": true,
  "note": "周末",
  "type": "假日"
}
```

如果需要在基金可交易日进行自动记账，判断 `note` 是否为 `普通工作日` 即可（补班日基金不可交易）。

可以通过 `date` 参数指定要查询的日期，例如 [https://holiday.dreace.top?date=2023-01-02](https://holiday.dreace.top?date=2023-01-02) 将返回：

```json
{
  "date": "2023-01-02",
  "isHoliday": true,
  "note": "元旦",
  "type": "假日"
}
```

## 生成假期 JSON 文件
要生成假期数据文件需要从官网复制带有标题的完整通知，如 [国务院办公厅关于2024年部分节假日安排的通知](https://www.gov.cn/zhengce/content/202310/content_6911527.htm)。

### 使用 Python 脚本生成
复制通知全文粘贴到 `export_holiday.py` 中执行即可。

### 使用大语言模型（LLM）生成
可以使用 GPT-4 等大语言模型输入通知全文和参考格式快速生成假期数据，参考提示词（prompt）如下：
````text
步骤一：请分析下面使用 ``` 包裹的 2024 年放假数据
```
<通知全文>
```
步骤二：分析下面的使用 ``` 包裹的 2023 年已生成数据作为格式参考：
```
<参考数据，可以从直接粘贴往年的 JSON 文件内容>
```
生成要求如下：
1. 日期需要按升序排序
2. 严格按照参考格式生成
````
生成后请务必核对生成的数据是否准确，上面的提示词**仅在 GPT-4、o1-preview 中经过测试**可以生成准确的假期数据文件。

## 支持功能及未来计划

- [x] 脚本自动生成调休和补班 JSON 数据
- [x] 提供 HTTP 服务
- [x] 支持查询历史年份假期
- [ ] 从命令行更新服务和优雅重启

