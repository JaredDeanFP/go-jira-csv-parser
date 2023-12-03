package main

import (
	"encoding/csv"
	"fmt"
	"html/template"
	"log"
	"os"
	"strings"
	"time"
)

type Record map[string]interface{}

type Watcher struct {
	Name string
	Id   string
}
type JiraRecord struct {
	Summary         string
	IssueKey        string `json:"Issue key"` // Use struct tags to handle the space in the field name
	IssueId         string `json:"Issue id"`
	ParentId        string `json:"Parent id"`
	Parent          string
	IssueType       string `json:"Issue Type"`
	Status          string
	ProjectLead     string `json:"Project lead"`
	ProjectLeadId   string `json:"Project lead id"`
	Priority        string
	Resolution      string
	Assignee        string
	AssigneeId      string `json:"Assignee Id"`
	Reporter        string
	ReporterId      string `json:"Reporter Id"`
	Creator         string
	CreatorId       string `json:"Creator Id"`
	Created         int64
	Updated         int64
	LastViewed      int64 `json:"Last Viewed"`
	Resolved        int64
	Description     string
	Watchers        []Watcher
	EpicLinkSummary string `json:"Epic Link Summary"`
	Comment         []string
}

func main() {
	file, err := os.Open("./Jira.csv")

	if err != nil {
		log.Fatal("Failed to open Jira.csv")
		return
	}

	// Close file when program finishes
	defer file.Close()

	csvReader := csv.NewReader(file)
	csvReader.FieldsPerRecord = -1

	headers, err := csvReader.Read()

	if err != nil {
		log.Fatal("Failed to read csv.")
		return
	}

	fmt.Println(headers)

	var data []JiraRecord
	var watchers []Watcher
	for {
		row, err := csvReader.Read()
		if err != nil {
			break
		}

		watcherNames := []string{}
		watcherIds := []string{}

		ticket := JiraRecord{}
		for i, val := range row {
			switch headers[i] {
			case "Summary":
				ticket.Summary = val
				break
			case "Issue key":
				ticket.IssueKey = val
				break
			case "Issue id":
				ticket.IssueId = val
				break
			case "Parent id":
				ticket.ParentId = val
				break
			case "Parent":
				ticket.Parent = val
				break
			case "Issue Type":
				ticket.IssueType = val
				break
			case "Status":
				ticket.Status = val
				break
			case "Project lead":
				ticket.ProjectLead = val
				break
			case "Project lead id":
				ticket.ProjectLeadId = val
				break
			case "Priority":
				ticket.Priority = val
				break
			case "Resolution":
				ticket.Resolution = val
				break
			case "Assignee":
				ticket.Assignee = toName(val)
				break
			case "Assignee Id":
				ticket.AssigneeId = val
				break
			case "Reporter":
				ticket.Reporter = toName(val)
				break
			case "Reporter Id":
				ticket.ReporterId = val
				break
			case "Creator":
				ticket.Creator = toName(val)
				break
			case "Creator Id":
				ticket.CreatorId = val
				break
			case "Created":
				ticket.Created = toUnixTime(val)
				break
			case "Updated":
				ticket.Updated = toUnixTime(val)
				break
			case "Last Viewed":
				ticket.LastViewed = toUnixTime(val)
				break
			case "Resolved":
				ticket.Resolved = toUnixTime(val)
				break
			case "Description":
				ticket.Description = val
				break
			case "Watchers":
				watcherNames = append(watcherNames, val)
				break
			case "Watchers Id":
				watcherIds = append(watcherIds, val)
				break
			case "Epic Link Summary":
				ticket.EpicLinkSummary = val
				break
			case "Comment":
				ticket.Comment = append(ticket.Comment, val)
				break

			default:
				break

			}

		}

		for i := 0; i < len(watcherIds); i++ {

			if watcherNames[i] == "" || watcherIds[i] == "" {
				continue
			}

			watchers = append(watchers, Watcher{
				Id:   watcherIds[i],
				Name: watcherNames[i],
			})
		}

		data = append(data, ticket)
	}

	//var caption []string

	type TemplateEdit struct {
		UnixTime int64
		Person   string
		Action   string
		Route    string
		Color    string
	}
	tmpl, err := template.New("TicketCreate").Parse("{{.UnixTime}}|{{.Person}}|{{.Action}}|{{.Route}}|{{.Color}}\n")

	type CaptionEdit struct {
		UnixTime int64
		Message  string
	}
	captionTmpl, err := template.New("TicketCreate").Parse("{{.UnixTime}}|{{.Message}}\n")

	jiraLog, err := os.Create("./jira1.txt")
	captionLog, err := os.Create("./caption.txt")

	defer jiraLog.Close()
	defer captionLog.Close()

	for _, ticket := range data {
		// Ticket Created
		if ticket.Created == -1 {
			continue
		}
		tmpl.Execute(jiraLog, TemplateEdit{
			UnixTime: ticket.Created,
			Person:   ticket.Creator,
			Action:   "C",
			Route:    toRoute(data, ticket),
			Color:    toHex(ticket.IssueType),
		})

		if ticket.IssueType == "Epic" {
			captionTmpl.Execute(captionLog, CaptionEdit{
				UnixTime: ticket.Created,
				Message:  ticket.Summary,
			})
		}

		// Ticket Modified
		if ticket.Updated == -1 || ticket.Assignee == "" {
			continue
		}
		tmpl.Execute(jiraLog, TemplateEdit{
			UnixTime: ticket.Updated,
			Person:   ticket.Assignee,
			Action:   "M",
			Route:    toRoute(data, ticket),
			Color:    toHex(ticket.IssueType),
		})

		// Ticket Completed
		if ticket.Resolved == -1 || ticket.Reporter == "" {
			continue
		}
		tmpl.Execute(jiraLog, TemplateEdit{
			UnixTime: ticket.Resolved,
			Person:   ticket.Reporter,
			Action:   "D",
			Route:    toRoute(data, ticket),
			Color:    toHex(ticket.IssueType),
		})

	}
}

func toUnixTime(rawTime string) int64 {
	var layout string

	if strings.Contains(rawTime, "AM") || strings.Contains(rawTime, "PM") {
		layout = "02/Jan/06 3:04 PM"
	} else {
		layout = "1/2/06 15:04"
	}
	if rawTime == "" {
		return -1
	}
	Time, err := time.Parse(layout, rawTime)

	if err != nil {
		log.Fatal("Failed to parse time " + rawTime)
		return -1
	}

	return Time.Unix()
}

func toHex(in string) string {
	switch in {
	case "Task":
		return "ADD8E6"
	case "Epic":
		return "CBC3E3"
	case "Story":
		return "90EE90"
	case "Bug":
		return "FFCCCB"
	default:
		return "FFFFFF"
	}
}

func toRoute(data []JiraRecord, ticket JiraRecord) string {
	route := "JIRA/"
	parent := ""
	for _, potentialParent := range data {
		if ticket.Parent == "" && ticket.ParentId == "" {
			break
		}
		if potentialParent.IssueId == ticket.Parent || potentialParent.IssueId == ticket.ParentId {
			parent = potentialParent.IssueKey
			break
		}
	}
	if parent != "" {
		parent += "/"
	}

	return route + parent + ticket.IssueKey
}

func toName(dirtyName string) string {
	return strings.ReplaceAll(dirtyName, ".", " ")
}
