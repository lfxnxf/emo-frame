package utils

import (
	"fmt"
	"time"
)

const (
	SecondsOfDay = 24 * 60 * 60
)

var weekDayStringMap = map[string]string{
	"Monday":    "周一",
	"Tuesday":   "周二",
	"Wednesday": "周三",
	"Thursday":  "周四",
	"Friday":    "周五",
	"Saturday":  "周六",
	"Sunday":    "周日",
}

var weekDayIntMap = map[string]int64{
	"Monday":    1,
	"Tuesday":   2,
	"Wednesday": 3,
	"Thursday":  4,
	"Friday":    5,
	"Saturday":  6,
	"Sunday":    7,
}

// DiffNatureDays 计算两个时间戳之间的自然天数之差
func DiffNatureDays(t1, t2 int64) int {
	if t1 == t2 {
		return -1
	}
	if t1 > t2 {
		t1, t2 = t2, t1
	}

	diffDays := 0
	secDiff := t2 - t1
	if secDiff > SecondsOfDay {
		tmpDays := int(secDiff / SecondsOfDay)
		t1 += int64(tmpDays) * SecondsOfDay
		diffDays += tmpDays
	}

	st := time.Unix(t1, 0)
	et := time.Unix(t2, 0)
	dateFormatTpl := "20060102"
	if st.Format(dateFormatTpl) != et.Format(dateFormatTpl) {
		diffDays += 1
	}

	return diffDays
}

func String2Unix(stringTime string) int64 {
	loc, _ := time.LoadLocation("Local")

	timeFormatTpl := TimeFormatYYYYMMDDHHmmSS
	if len(timeFormatTpl) != len(stringTime) {
		timeFormatTpl = timeFormatTpl[0:len(stringTime)]
	}

	theTime, _ := time.ParseInLocation(timeFormatTpl, stringTime, loc)

	return theTime.Unix()
}

func DayString2Unix(stringTime string) int64 {
	loc, _ := time.LoadLocation("Local")

	theTime, _ := time.ParseInLocation("2006-01-02", stringTime, loc)

	return theTime.Unix()
}

func NowTimeString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func GetPoorDayTime(day int) (int64, string) {
	// 时区
	timeZone := time.FixedZone("CST", 8*3600) // 东八区

	// n天
	nowTime := time.Now().In(timeZone)
	poorTime := nowTime.AddDate(0, 0, day)

	// 时间转换格式
	poorTimeS := poorTime.Unix()                                      // 秒时间戳
	poorDate := time.Unix(poorTimeS, 0).Format("2006-01-02 15:04:05") // 固定格式的日期时间戳

	return poorTimeS, poorDate
}

// 获取当天凌晨时间戳
func GetTodayZeroTimeUnix() int64 {
	timeStr := time.Now().Format("2006-01-02")
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr+" 00:00:00", time.Local)
	return t.Unix()
}

// 获取第二天凌晨时间戳
func GetTomorrowZeroTimeUnix() int64 {
	timeStr := time.Now().Format("2006-01-02")
	t, _ := time.ParseInLocation("2006-01-02 15:04:05", timeStr+" 23:59:59", time.Local)
	return t.Unix() + 1
}

func GetNowMonth() string {
	//获取当前月份
	return fmt.Sprintf("%d-%s", time.Now().Year(), time.Now().Format("01"))
}

func GetLastMonth() string {
	return fmt.Sprintf("%d-%s", time.Now().Year(), time.Now().AddDate(0, -1, 0).Format("01"))
}

func GetNextMonth() string {
	return fmt.Sprintf("%d-%s", time.Now().Year(), time.Now().AddDate(0, 1, 0).Format("01"))
}

func GetLastMonthLastDayUnix() int64 {
	year, month, _ := time.Now().Date()
	thisMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	end := thisMonth.AddDate(0, 0, -1).Unix() + 86399
	return end
}

// Unix2String 时间戳转换为时间字符串
func Unix2String(unix int64) string {
	tm := time.Unix(unix, 0)
	return tm.Format(TimeFormatYYYYMMDDHHmmSS)
}

func Unix2StringFmt(unix int64, format string) string {
	if len(format) <= 0 {
		format = TimeFormatYYYYMMDDHHmmSS
	}
	tm := time.Unix(unix, 0)
	return tm.Format(format)
}

func String2Time(stringTime string) (time.Time, error) {
	loc, _ := time.LoadLocation("Local")
	timeFormatTpl := TimeFormatYYYYMMDDHHmmSS
	if len(timeFormatTpl) != len(stringTime) {
		timeFormatTpl = timeFormatTpl[0:len(stringTime)]
	}
	return time.ParseInLocation(timeFormatTpl, stringTime, loc)
}

// 2022-06-07==>2022/06/07
func StringTimeMinus2Slash(stringTime string) (string, error) {
	theTime, err := String2Time(stringTime)
	if err != nil {
		return "", err
	}
	return theTime.Format(TimeFormatYYYYMMDDSlash), nil
}

func GetBetweenDates(startDate, endDate string) []string {
	if startDate == endDate {
		return []string{startDate}
	}
	var d []string
	timeFormatTpl := TimeFormatYYYYMMDDHHmmSS
	if len(timeFormatTpl) != len(startDate) {
		timeFormatTpl = timeFormatTpl[0:len(startDate)]
	}
	date, err := time.Parse(timeFormatTpl, startDate)
	if err != nil {
		// 时间解析，异常
		return d
	}
	date2, err := time.Parse(timeFormatTpl, endDate)
	if err != nil {
		// 时间解析，异常
		return d
	}
	if date2.Before(date) {
		// 如果结束时间小于开始时间，异常
		return d
	}
	// 输出日期格式固定
	date2Str := date2.Format(TimeFormatYYYYMMDD)
	d = append(d, date.Format(TimeFormatYYYYMMDD))

	num := 0
	for {
		// 最大让取1000天的
		if num >= 1000 {
			return d
		}
		date = date.AddDate(0, 0, 1)
		dateStr := date.Format(TimeFormatYYYYMMDD)
		d = append(d, dateStr)
		if dateStr == date2Str {
			break
		}
		num++
	}
	return d
}

func FormatHHMMSS(seconds int64) string {
	//前置补0
	h, m, s := ResolveTime(seconds)
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func ResolveTime(seconds int64) (hour, minute, second int64) {
	var day = seconds / (24 * 3600)
	hour = (seconds - day*3600*24) / 3600
	minute = (seconds - day*24*3600 - hour*3600) / 60
	second = seconds - day*24*3600 - hour*3600 - minute*60
	return
}

// 获取周几
func GetWeekDayString(t time.Time) string {
	return weekDayStringMap[t.Weekday().String()]
}

func GetWeekDayInt(t time.Time) int64 {
	return weekDayIntMap[t.Weekday().String()]
}
