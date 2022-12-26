import json
import logging
import re
from datetime import datetime, timedelta
from typing import List

"""
从官网复制带有标题的完整通知，如 http://www.gov.cn/fuwu/2022-12/08/content_5730853.htm
国务院办公厅关于2023年部分节假日安排的通知
各省、自治区、直辖市人民政府，国务院各部委、各直属机构：

经国务院批准，现将2023年元旦、春节、清明节、劳动节、端午节、中秋节和国庆节放假调休日期的具体安排通知如下。

一、元旦：2022年12月31日至2023年1月2日放假调休，共3天。

二、春节：1月21日至27日放假调休，共7天。1月28日（星期六）、1月29日（星期日）上班。

三、清明节：4月5日放假，共1天。

四、劳动节：4月29日至5月3日放假调休，共5天。4月23日（星期日）、5月6日（星期六）上班。

五、端午节：6月22日至24日放假调休，共3天。6月25日（星期日）上班。

六、中秋节、国庆节：9月29日至10月6日放假调休，共8天。10月7日（星期六）、10月8日（星期日）上班。

节假日期间，各地区、各部门要妥善安排好值班和安全、保卫、疫情防控等工作，遇有重大突发事件，要按规定及时报告并妥善处置，确保人民群众祥和平安度过节日假期。

国务院办公厅
2022年12月8日
"""


def get_input() -> str:
    text = ""

    while True:
        try:
            t = input()
        except EOFError:
            break
        text += t + "\n"
    return text


def get_dates_between(start_date, end_date) -> List[str]:
    # 将开始日期和结束日期转换为 datetime 对象
    start_date = datetime.strptime(start_date, '%Y-%m-%d')
    end_date = datetime.strptime(end_date, '%Y-%m-%d')

    # 创建一个空列表，用于存储日期
    dates = []

    # 当前日期
    current_date = start_date

    # 遍历所有日期
    while current_date <= end_date:
        # 将日期转换为字符串，并添加到列表中
        dates.append(current_date.strftime('%Y-%m-%d'))
        current_date += timedelta(days=1)

    # 返回日期列表
    return dates


def main():
    text = get_input()
    matches = re.search(r"(\d{4}).*节假日安排的通知", text)
    year = 0
    if not matches:
        logging.error("无法从通知提取放假年份")
    else:
        year = int(matches.group(1))

    holiday_text_list = []
    results = {}
    for line in text.split("\n"):
        if re.search(r"共\d{1,2}天", line):
            holiday_name_re = re.search(r"、(.*)：", line)
            holiday_name = holiday_name_re.group(1)

            t = line.split("放假")
            holiday_text = t[0]
            make_up_text = t[1] if len(t) > 1 else ""

            if holiday_text.find("至") == -1:
                # 放假一天
                holiday_re = re.search(r"(\d+)月(\d+)日", holiday_text)
                results[f"{year}-{int(holiday_re.group(1)):02d}-{int(holiday_re.group(2)):02d}"] = {
                    "type": "假日",
                    "note": holiday_name
                }
            else:
                # 放假多天
                holiday_re = re.search(r"(\d+)月(\d+)日至.*?(\d+)日", holiday_text)
                start_year = end_year = year
                start_month = end_month = int(holiday_re.group(1))
                start_day = int(holiday_re.group(2))
                end_day = int(holiday_re.group(3))
                if start_day > end_day:
                    # 跨月份
                    end_month = (end_month + 1) % 12
                if start_month > end_month:
                    # 跨年份
                    start_year -= 1
                for date in get_dates_between(f"{start_year}-{start_month}-{start_day}",
                                              f"{end_year}-{end_month}-{end_day}"):
                    results[date] = {
                        "type": "假日",
                        "note": holiday_name
                    }
            if make_up_text.find("上班") != -1:
                for match in re.finditer(r"(\d+)月(\d+)日", make_up_text):
                    results[f"{year}-{int(match.group(1)):02d}-{int(match.group(2)):02d}"] = {
                        "type": "补班日",
                        "note": holiday_name + "补班"
                    }
    with open(f"holiday_{year}.json", "w+", encoding="utf8") as output_file:
        output_file.write(json.dumps(results, ensure_ascii=False, indent=2))


if __name__ == '__main__':
    main()
