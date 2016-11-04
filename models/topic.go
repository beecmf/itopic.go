package models

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"time"

	"encoding/json"
	"github.com/russross/blackfriday"
)

//Topic struct
type Topic struct {
	TopicID  string
	Title    string
	Time     time.Time
	Tag      []*TopicTag
	Content  string
	IsPublic bool //true for public，false for protected
}

//MonthList Show The Topic Group By Month
type MonthList struct {
	Month  string
	Topics []*Topic
}

//InitTopicList Load All The Topic On Start
func InitTopicList() error {
	Topics = Topics[:0]
	return filepath.Walk(topicMarkdownFolder, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		t, err := GetTopicByPath(path)
		if err != nil {
			return err
		}
		SetTopicToTag(t)
		SetTopicToMonth(t)
		//按时间倒序排列
		for i := range Topics {
			if t.Time.After(Topics[i].Time) {
				Topics = append(Topics, nil)
				copy(Topics[i+1:], Topics[i:])
				Topics[i] = t
				return nil
			}
		}
		Topics = append(Topics, t)
		return nil
	})
}

//GetTopicByPath Read The Topic By Path
func GetTopicByPath(path string) (*Topic, error) {
	fp, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	t := &Topic{
		Title:    strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)),
		IsPublic: true,
	}
	var tHeadStr string
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		s := scanner.Text()
		tHeadStr += s
		if len(s) == 0 {
			break
		}
	}
	tHeadStr = strings.Trim(tHeadStr, "```")
	type tHeadJSON struct {
		URL      string
		Time     string
		Tag      string
		IsPublic string `json:"public"`
	}
	var thj tHeadJSON
	if err := json.Unmarshal([]byte(tHeadStr), &thj); err != nil {
		return nil, err
	}
	t.TopicID = thj.URL
	t.Time, err = time.Parse("2006/01/02 15:04", thj.Time)
	if err != nil {
		return nil, err
	}
	if strings.Compare(thj.IsPublic, "no") == 0 {
		t.IsPublic = false
	}
	tagArray := strings.Split(thj.Tag, ",")
	for _, tagName := range tagArray {
		for kc := range TopicsGroupByTag {
			if strings.Compare(tagName, TopicsGroupByTag[kc].TagID) == 0 {
				t.Tag = append(t.Tag, TopicsGroupByTag[kc])
				break
			}
		}
	}
	var content bytes.Buffer
	for scanner.Scan() {
		content.Write(scanner.Bytes())
		content.WriteString("\n")
	}
	t.Content = string(blackfriday.MarkdownCommon(content.Bytes()))
	return t, nil
}

//SetTopicToTag set topic to tag struct
func SetTopicToTag(t *Topic) {
	if t.IsPublic == false {
		return
	}
	for k := range TopicsGroupByTag {
		for i := range t.Tag {
			if TopicsGroupByTag[k].TagID != t.Tag[i].TagID {
				continue
			}
			for j := range TopicsGroupByTag[k].Topics {
				if t.Time.After(TopicsGroupByTag[k].Topics[j].Time) {
					TopicsGroupByTag[k].Topics = append(TopicsGroupByTag[k].Topics, nil)
					copy(TopicsGroupByTag[k].Topics[j+1:], TopicsGroupByTag[k].Topics[j:])
					TopicsGroupByTag[k].Topics[j] = t
					return
				}
			}
			TopicsGroupByTag[k].Topics = append(TopicsGroupByTag[k].Topics, t)
		}
	}
}

//SetTopicToMonth set topic to month struct
func SetTopicToMonth(t *Topic) {
	if t.IsPublic == false {
		return
	}
	month := t.Time.Format("2006-01")
	ml := &MonthList{}
	for _, m := range TopicsGroupByMonth {
		if m.Month == month {
			ml = m
		}
	}
	if ml.Month == "" {
		ml.Month = month
		isFind := false
		for i := range TopicsGroupByMonth {
			if strings.Compare(ml.Month, TopicsGroupByMonth[i].Month) > 0 {
				TopicsGroupByMonth = append(TopicsGroupByMonth, nil)
				copy(TopicsGroupByMonth[i+1:], TopicsGroupByMonth[i:])
				TopicsGroupByMonth[i] = ml
				isFind = true
				break
			}
		}
		if isFind == false {
			TopicsGroupByMonth = append(TopicsGroupByMonth, ml)
		}
	}
	for i := range ml.Topics {
		if t.Time.After(ml.Topics[i].Time) {
			ml.Topics = append(ml.Topics, nil)
			copy(ml.Topics[i+1:], ml.Topics[i:])
			ml.Topics[i] = t
			return
		}
	}
	ml.Topics = append(ml.Topics, t)
}
