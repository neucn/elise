package elise

import (
	"errors"
	"fmt"
	"github.com/neucn/neugo"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	defaultCourseTableUrl       = "http://219.216.96.4/eams/courseTableForStd.action"
	defaultCourseTableActionUrl = "http://219.216.96.4/eams/courseTableForStd!courseTable.action"
	defaultCurrentWeekUrl       = "http://219.216.96.4/eams/homeExt.action"
)

type session struct {
	courseTableUrl,
	courseTableActionUrl,
	currentWeekUrl string
	client *http.Client
}

type GenerateFunc func(courses []*Course, startDay time.Time, output string) (generated string, err error)

func (s *session) Generate(generateFunc GenerateFunc, output string) (path string, err error) {
	var body string
	if body, err = s.getCourseTablePage(); err != nil {
		return
	}
	var week int
	if week, err = s.getCurrentWeek(); err != nil {
		return
	}

	startDay := getSemesterStartDay(week)
	fmt.Println("当前为第", week, "教学周")
	fmt.Println("计算得到本学期开始于", startDay.Format("2006-01-02"))
	fmt.Println("官方校历 http://www.neu.edu.cn/xl/list.htm")

	fmt.Println("\n======开始生成课程表======")
	courses := parseCourses(body)
	path, err = generateFunc(courses, startDay, output)
	fmt.Println("\n======课程表生成成功======")

	return
}

func (s *session) getCourseTablePage() (content string, err error) {
	var resp *http.Response

	// 发送
	if resp, err = s.client.Get(s.courseTableUrl); err != nil {
		return
	}

	// 读取
	if content, err = readBody(resp); err != nil {
		return
	}
	// 检查
	if !strings.Contains(content, "bg.form.addInput(form,\"ids\",\"") {
		return "", errors.New("获取必要参数ids失败")
	}
	content = content[strings.Index(content, "bg.form.addInput(form,\"ids\",\"")+29 : strings.Index(content, "bg.form.addInput(form,\"ids\",\"")+50]
	ids := content[:strings.Index(content, "\");")]
	semesterId := resp.Cookies()[0].Value

	// 第二次请求
	requestBody := fmt.Sprintf(
		"ignoreHead=1&showPrintAndExport=1&setting.kind=std&startWeek=&semester.id=%s&ids=%s",
		semesterId, ids)

	req, _ := http.NewRequest(http.MethodPost, s.courseTableActionUrl, strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/85.0.4183.102 Safari/537.36")

	// 发送
	if resp, err = s.client.Do(req); err != nil {
		return
	}

	// 读取
	if content, err = readBody(resp); err != nil {
		return
	}
	// 检查
	if !strings.Contains(content, "课表格式说明") {
		return "", errors.New("获取课表失败")
	}

	return
}

func (s *session) getCurrentWeek() (week int, err error) {
	var resp *http.Response
	var content string
	// 发送
	if resp, err = s.client.Get(s.currentWeekUrl); err != nil {
		return
	}

	// 读取
	if content, err = readBody(resp); err != nil {
		return
	}

	// 检查
	if !strings.Contains(content, "教学周") {
		return 0, errors.New("获取当前教学周失败")
	}
	content = content[strings.Index(content, "id=\"teach-week\">") : strings.Index(content, "教学周")+10]

	reg := regexp.MustCompile(`学期\s*<font size="\d+px">(\d+)</font>\s*教学周`)
	res := reg.FindStringSubmatch(content)
	if len(res) < 2 {
		return 0, errors.New("无法获取当前教学周")
	}

	return strconv.Atoi(res[1])
}

type Generator interface {
	// 传入一个解析函数与目标输出路径，返回一个最终的绝对路径与错误
	Generate(generateFunc GenerateFunc, output string) (path string, err error)
}

var _ Generator = &session{}

func New(username, password string, webVPN bool) (Generator, error) {
	s := new(session)
	var platform neugo.Platform
	if webVPN {
		platform = neugo.WebVPN
		s.currentWeekUrl = neugo.EncryptWebVPNUrl(defaultCurrentWeekUrl)
		s.courseTableUrl = neugo.EncryptWebVPNUrl(defaultCourseTableUrl)
		s.courseTableActionUrl = neugo.EncryptWebVPNUrl(defaultCourseTableActionUrl)
	} else {
		platform = neugo.CAS
		s.currentWeekUrl = defaultCurrentWeekUrl
		s.courseTableUrl = defaultCourseTableUrl
		s.courseTableActionUrl = defaultCourseTableActionUrl
	}

	client := neugo.NewSession()
	if err := neugo.Use(client).WithAuth(username, password).On(platform).Login(); err != nil {
		return nil, err
	}
	s.client = client
	return s, nil
}

func readBody(resp *http.Response) (string, error) {
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	_ = resp.Body.Close()
	return string(content), nil
}

func getSemesterStartDay(week int) time.Time {
	now := time.Now()
	location := time.FixedZone("UTC+8", 8*60*60)
	daySum := int(now.Weekday()) + week*7 - 7
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location).
		AddDate(0, 0, -daySum)
}

// 课程信息
type Course struct {
	ID          string
	Name        string
	RoomID      string
	RoomName    string
	Weeks       string
	CourseTimes []CourseTime
}

// 课程具体时间，周几第几节
type CourseTime struct {
	DayOfTheWeek int
	TimeOfTheDay int
}

func parseCourses(body string) []*Course {
	var courses []*Course
	reg1 := regexp.MustCompile(`TaskActivity\(actTeacherId.join\(','\),actTeacherName.join\(','\),"(.*)","(.*)\(.*\)","(.*)","(.*)","(.*)",null,null,assistantName,"",""\);((?:\s*index =\d+\*unitCount\+\d+;\s*.*\s)+)`)
	reg2 := regexp.MustCompile(`\s*index =(\d+)\*unitCount\+(\d+);\s*`)
	coursesStr := reg1.FindAllStringSubmatch(body, -1)
	for _, courseStr := range coursesStr {
		course := &Course{}
		course.ID = courseStr[1]
		course.Name = courseStr[2]
		course.RoomID = courseStr[3]
		course.RoomName = courseStr[4]
		course.Weeks = courseStr[5]
		for _, indexStr := range strings.Split(courseStr[6], "table0.activities[index][table0.activities[index].length]=activity;") {
			if !strings.Contains(indexStr, "unitCount") {
				continue
			}
			var courseTime CourseTime
			courseTime.DayOfTheWeek, _ = strconv.Atoi(reg2.FindStringSubmatch(indexStr)[1])
			courseTime.TimeOfTheDay, _ = strconv.Atoi(reg2.FindStringSubmatch(indexStr)[2])
			course.CourseTimes = append(course.CourseTimes, courseTime)
		}
		courses = append(courses, course)
	}
	return courses
}
