package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var logAddress = flag.String("f", "", "慢sq日志文件所在的位置 例子：/var/log/mysql4306-slow.log")
var startTime = flag.String("startTime", "", "开始时间 格式：yyyy-mm-dd hh:mm:ss")
var endTime = flag.String("endTime", "", "结束时间 格式：yyyy-mm-dd hh:mm:ss")

func hasDuplicate(items []string, target string) bool {

	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

type SlowSqlInfoSliceDecrement []SlowSqlInfo

func (s SlowSqlInfoSliceDecrement) Len() int { return len(s) }

func (s SlowSqlInfoSliceDecrement) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s SlowSqlInfoSliceDecrement) Less(i, j int) bool { return s[i].Time95 > s[j].Time95 }

func main() {
	//pt_cmd := "./cmd/pt-query-digest  /Users/macbookpro/Documents/go-code/mysql_slow/mysql_slow_agent/data/mysql4306-slow.log --output json  --progress time,1 --charset=utf8mb4 --since='2023-07-24 00:00:00' --until='2023-07-24 23:59:59' >mysql_slow.json"
	flag.Parse()

	if *logAddress == "" {
		panic("慢日志文件不能为空")
	}
	if *startTime == "" || *endTime == "" {
		panic("查询时间不能为空")
	}
	ptCmd := fmt.Sprintf("./cmd/pt-query-digest  %s --output json  --noversion-check  --progress time,1 --charset=utf8mb4 --since='%s' --until='%s' >mysql_slow.json", *logAddress, *startTime, *endTime)
	println(ptCmd)
	cmd := exec.Command("/bin/bash", "-c", ptCmd)
	_, err := cmd.CombinedOutput()
	if err != nil {
		println(err.Error())
	}

	fileName := fmt.Sprintf("mysql-%s.html", time.Now().Format("2006-01-02"))
	newFile, _ := os.Create(fileName)
	file, err := os.Open("mysql_slow.json")
	defer file.Close()
	var report Report
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&report)
	var slowSqlInfos []SlowSqlInfo
	allSqlInfo := report.Classes
	for _, sqlInfo := range allSqlInfo {
		var allTables []string
		var slowSqlInfo SlowSqlInfo
		for _, slowTable := range sqlInfo.Tables {
			s := strings.Split(slowTable.Create, ".")
			//getTableName := "`table_name`\\G"
			re := regexp.MustCompile("`([^`]+)`")
			match := re.FindStringSubmatch(s[1])
			tableName := match[1]
			flag := hasDuplicate(allTables, tableName)
			if !flag {
				allTables = append(allTables, tableName)
			}
		}
		slowSqlInfo.RowsSum = sqlInfo.Metrics.RowsExamined.Sum
		slowSqlInfo.RowsMax = sqlInfo.Metrics.RowsExamined.Max
		slowSqlInfo.LengthSum = sqlInfo.Metrics.QueryLength.Sum
		slowSqlInfo.LengthMax = sqlInfo.Metrics.QueryLength.Max
		slowSqlInfo.TimeMax = sqlInfo.Metrics.QueryTime.Max
		slowSqlInfo.TimeMin = sqlInfo.Metrics.QueryTime.Min
		pct95, _ := strconv.ParseFloat(sqlInfo.Metrics.QueryTime.Pct95, 64)
		t := int64(pct95)
		slowSqlInfo.Time95 = t
		slowSqlInfo.TimeMedian = sqlInfo.Metrics.QueryTime.Median
		slowSqlInfo.RowSendMax = sqlInfo.Metrics.RowsSent.Max
		slowSqlInfo.QueryDb = sqlInfo.Metrics.Db.Value
		slowSqlInfo.QueryCount = sqlInfo.QueryCount
		slowSqlInfo.Sql = sqlInfo.Example.Query
		slowSqlInfo.QueryTables = allTables
		slowSqlInfo.Id = sqlInfo.Checksum
		slowSqlInfos = append(slowSqlInfos, slowSqlInfo)
	}
	sort.Sort(SlowSqlInfoSliceDecrement(slowSqlInfos))

	if err != nil {
	}
	tmpl, err := template.ParseFiles("./template/template.html")
	if err != nil {
		fmt.Errorf(err.Error())
		panic(err)

	}
	tmpl.Execute(newFile, slowSqlInfos)
}

type Report struct {
	Global struct {
		UniqueQueryCount int `json:"unique_query_count"`
		Files            []struct {
			Name string `json:"name"`
			Size int    `json:"size"`
		} `json:"files"`
		QueryCount int `json:"query_count"`
		Metrics    struct {
			QueryLength struct {
				Sum    string `json:"sum"`
				Stddev string `json:"stddev"`
				Min    string `json:"min"`
				Avg    string `json:"avg"`
				Median string `json:"median"`
				Max    string `json:"max"`
				Pct95  string `json:"pct_95"`
			} `json:"Query_length"`
			LockTime struct {
				Pct95  string `json:"pct_95"`
				Max    string `json:"max"`
				Median string `json:"median"`
				Avg    string `json:"avg"`
				Min    string `json:"min"`
				Stddev string `json:"stddev"`
				Sum    string `json:"sum"`
			} `json:"Lock_time"`
			RowsExamined struct {
				Avg    string `json:"avg"`
				Min    string `json:"min"`
				Median string `json:"median"`
				Max    string `json:"max"`
				Pct95  string `json:"pct_95"`
				Sum    string `json:"sum"`
				Stddev string `json:"stddev"`
			} `json:"Rows_examined"`
			RowsSent struct {
				Max    string `json:"max"`
				Pct95  string `json:"pct_95"`
				Avg    string `json:"avg"`
				Min    string `json:"min"`
				Median string `json:"median"`
				Sum    string `json:"sum"`
				Stddev string `json:"stddev"`
			} `json:"Rows_sent"`
			QueryTime struct {
				Median string `json:"median"`
				Min    string `json:"min"`
				Avg    string `json:"avg"`
				Pct95  string `json:"pct_95"`
				Max    string `json:"max"`
				Stddev string `json:"stddev"`
				Sum    string `json:"sum"`
			} `json:"Query_time"`
		} `json:"metrics"`
	} `json:"global"`
	Classes []struct {
		Distillate string `json:"distillate"`
		Example    struct {
			QueryTime string `json:"Query_time"`
			Query     string `json:"query"`
			Ts        string `json:"ts"`
			AsSelect  string `json:"as_select,omitempty"`
		} `json:"example"`
		Histograms struct {
			QueryTime []int `json:"Query_time"`
		} `json:"histograms"`
		Fingerprint string `json:"fingerprint"`
		Metrics     struct {
			LockTime struct {
				Pct    string `json:"pct"`
				Stddev string `json:"stddev"`
				Sum    string `json:"sum"`
				Pct95  string `json:"pct_95"`
				Max    string `json:"max"`
				Median string `json:"median"`
				Avg    string `json:"avg"`
				Min    string `json:"min"`
			} `json:"Lock_time"`
			QueryLength struct {
				Pct    string `json:"pct"`
				Stddev string `json:"stddev"`
				Sum    string `json:"sum"`
				Pct95  string `json:"pct_95"`
				Max    string `json:"max"`
				Median string `json:"median"`
				Avg    string `json:"avg"`
				Min    string `json:"min"`
			} `json:"Query_length"`
			RowsSent struct {
				Max    string `json:"max"`
				Pct95  string `json:"pct_95"`
				Avg    string `json:"avg"`
				Min    string `json:"min"`
				Median string `json:"median"`
				Pct    string `json:"pct"`
				Sum    string `json:"sum"`
				Stddev string `json:"stddev"`
			} `json:"Rows_sent"`
			User struct {
				Value string `json:"value"`
			} `json:"user"`
			Db struct {
				Value string `json:"value"`
			} `json:"db,omitempty"`
			RowsExamined struct {
				Median string `json:"median"`
				Min    string `json:"min"`
				Avg    string `json:"avg"`
				Pct95  string `json:"pct_95"`
				Max    string `json:"max"`
				Stddev string `json:"stddev"`
				Sum    string `json:"sum"`
				Pct    string `json:"pct"`
			} `json:"Rows_examined"`
			Host struct {
				Value string `json:"value"`
			} `json:"host"`
			QueryTime struct {
				Avg    string `json:"avg"`
				Min    string `json:"min"`
				Median string `json:"median"`
				Max    string `json:"max"`
				Pct95  string `json:"pct_95"`
				Sum    string `json:"sum"`
				Stddev string `json:"stddev"`
				Pct    string `json:"pct"`
			} `json:"Query_time"`
		} `json:"metrics"`
		TsMin      string `json:"ts_min"`
		Attribute  string `json:"attribute"`
		TsMax      string `json:"ts_max"`
		Checksum   string `json:"checksum"`
		QueryCount int    `json:"query_count"`
		Tables     []struct {
			Status string `json:"status"`
			Create string `json:"create"`
		} `json:"tables,omitempty"`
	} `json:"classes"`
}

type SlowSqlInfo struct {
	Id          string
	RowsSum     string
	RowsMax     string
	LengthSum   string
	LengthMax   string
	TimeMax     string
	TimeMin     string
	Time95      int64
	TimeMedian  string
	RowSendMax  string
	QueryDb     string
	QueryCount  int
	QueryTables []string
	Sql         string
}
