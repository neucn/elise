// modified from https://github.com/whoisnian/getMyCourses/blob/master/generate/generate.go
// original author @whoisnian
package ics

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/neucn/elise"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// 作息时间表，浑南上课时间
var ClassStartTimeHunnan = []string{
	"083000",
	"093000",
	"104000",
	"114000",
	"140000",
	"150000",
	"161000",
	"171000",
	"183000",
	"193000",
	"203000",
	"213000",
}

// 作息时间表，浑南下课时间
var classEndTimeHunnan = []string{
	"092000",
	"102000",
	"113000",
	"123000",
	"145000",
	"155000",
	"170000",
	"180000",
	"192000",
	"202000",
	"212000",
	"222000",
}

// 作息时间表，南湖上课时间
var ClassStartTimeNanhu = []string{
	"080000",
	"090000",
	"101000",
	"111000",
	"140000",
	"150000",
	"161000",
	"171000",
	"183000",
	"193000",
	"203000",
	"213000",
}

// 作息时间表，南湖下课时间
var classEndTimeNanhu = []string{
	"085000",
	"095000",
	"110000",
	"120000",
	"145000",
	"155000",
	"170000",
	"180000",
	"192000",
	"202000",
	"212000",
	"222000",
}

// ics文件用到的星期几简称
var dayOfWeek = []string{"MO", "TU", "WE", "TH", "FR", "SA", "SU"}

// 导出ICS格式的课程表
func Generate(courses []*elise.Course, startDay time.Time, output string) (string, error) {
	// 生成ics文件头
	var icsData string
	icsData = `BEGIN:VCALENDAR
PRODID:-//nian//getMyCourses 20190522//EN
VERSION:2.0
CALSCALE:GREGORIAN
METHOD:PUBLISH
X-WR-CALNAME:myCourses
X-WR-TIMEZONE:Asia/Shanghai
BEGIN:VTIMEZONE
TZID:Asia/Shanghai
X-LIC-LOCATION:Asia/Shanghai
BEGIN:STANDARD
TZOFFSETFROM:+0800
TZOFFSETTO:+0800
TZNAME:CST
DTSTART:19700101T000000
END:STANDARD
END:VTIMEZONE` + "\n"

	num := 0
	for _, course := range courses {
		var weekDay, st, en int
		weekDay = course.CourseTimes[0].DayOfTheWeek
		st = 12
		en = -1
		// 课程上下课时间
		for _, courseTime := range course.CourseTimes {
			if st > courseTime.TimeOfTheDay {
				st = courseTime.TimeOfTheDay
			}
			if en < courseTime.TimeOfTheDay {
				en = courseTime.TimeOfTheDay
			}
		}

		// debug信息
		num++
		fmt.Printf("\n#%d %s\n", num, course.Name)
		fmt.Println("周" + strconv.Itoa(weekDay+1) + " 第" + strconv.Itoa(st+1) + "-" + strconv.Itoa(en+1) + "节")

		// 统计要上课的周
		var periods []string
		startWeek := []int{}
		byday := dayOfWeek[weekDay]
		for i := 0; i < 53; i++ {
			if course.Weeks[i] != '1' {
				continue
			}
			if i+1 >= 53 {
				startWeek = append(startWeek, i)
				periods = append(periods, "RRULE:FREQ=WEEKLY;WKST=SU;COUNT=1;INTERVAL=1;BYDAY="+byday)
				// debug信息
				fmt.Println("第" + strconv.Itoa(i) + "周")
				continue
			}
			if course.Weeks[i+1] == '1' {
				// 连续周合并
				var j int
				for j = i + 1; j < 53; j++ {
					if course.Weeks[j] != '1' {
						break
					}
				}
				startWeek = append(startWeek, i)
				periods = append(periods, "RRULE:FREQ=WEEKLY;WKST=SU;COUNT="+strconv.Itoa(j-i)+";INTERVAL=1;BYDAY="+byday)
				// debug信息
				fmt.Println("第" + strconv.Itoa(i) + "-" + strconv.Itoa(j-1) + "周")
				i = j - 1
			} else {
				// 单双周合并
				var j int
				for j = i + 1; j+1 < 53; j += 2 {
					if course.Weeks[j] == '1' || course.Weeks[j+1] == '0' {
						break
					}
				}
				startWeek = append(startWeek, i)
				periods = append(periods, "RRULE:FREQ=WEEKLY;WKST=SU;COUNT="+strconv.Itoa((j+1-i)/2)+";INTERVAL=2;BYDAY="+byday)
				// debug信息
				if i%2 == 0 {
					fmt.Printf("双")
				} else {
					fmt.Printf("单")
				}
				fmt.Println(strconv.Itoa(i) + "-" + strconv.Itoa(j-1) + "周")
				i = j - 1
			}
		}

		// 生成ics文件中的EVENT
		for i := 0; i < len(periods); i++ {
			var eventData string
			eventData = `BEGIN:VEVENT` + "\n"
			startDate := startDay.AddDate(0, 0, (startWeek[i]-1)*7+weekDay+1)

			if strings.Contains(course.RoomName, "浑南") {
				eventData = eventData + `DTSTART;TZID=Asia/Shanghai:` + startDate.Format("20060102T") + ClassStartTimeHunnan[st] + "\n"
				eventData = eventData + `DTEND;TZID=Asia/Shanghai:` + startDate.Format("20060102T") + classEndTimeHunnan[en] + "\n"
			} else {
				eventData = eventData + `DTSTART;TZID=Asia/Shanghai:` + startDate.Format("20060102T") + ClassStartTimeNanhu[st] + "\n"
				eventData = eventData + `DTEND;TZID=Asia/Shanghai:` + startDate.Format("20060102T") + classEndTimeNanhu[en] + "\n"
			}
			eventData = eventData + periods[i] + "\n"
			eventData = eventData + `DTSTAMP:` + time.Now().Format("20060102T150405Z") + "\n"
			eventData = eventData + `UID:` + uuid.New().String() + "\n"
			eventData = eventData + `CREATED:` + time.Now().Format("20060102T150405Z") + "\n"
			eventData = eventData + `DESCRIPTION:` + "\n"
			eventData = eventData + `LAST-MODIFIED:` + time.Now().Format("20060102T150405Z") + "\n"
			eventData = eventData + `LOCATION:` + course.RoomName + "\n"
			eventData = eventData + `SEQUENCE:0
STATUS:CONFIRMED` + "\n"
			eventData = eventData + `SUMMARY:` + course.Name + "\n"

			eventData = eventData + `TRANSP:OPAQUE
END:VEVENT` + "\n"
			icsData = icsData + eventData
		}
	}
	icsData = icsData + `END:VCALENDAR`

	return export(output, icsData)
}

func export(output, content string) (string, error) {
	if len(output) == 0 {
		output = "./courses.ics"
	}
	if err := ioutil.WriteFile(output, []byte(content), 0644); err != nil {
		return "", err
	}
	return filepath.Abs(output)
}
