package models

import (
	"fmt"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops/v7/build"
)

// PipelineStatus represents the status of a pipeline
type PipelineStatus string

const (
	StatusInProgress PipelineStatus = "InProgress"
	StatusCompleted  PipelineStatus = "Completed"
	StatusCancelled  PipelineStatus = "Cancelled"
	StatusFailed     PipelineStatus = "Failed"
	StatusSucceeded  PipelineStatus = "Succeeded"
	StatusNone       PipelineStatus = "None"
)

// Pipeline represents a simplified view of a build pipeline
type Pipeline struct {
	ID             int
	Number         string
	Status         PipelineStatus
	Result         string
	Definition     string
	SourceBranch   string
	RequestedBy    string
	StartTime      *time.Time
	FinishTime     *time.Time
	QueueTime      *time.Time
	Build          *build.Build // Keep reference to original
}

// FromBuild converts an Azure DevOps build to our Pipeline model
func FromBuild(b *build.Build) *Pipeline {
	p := &Pipeline{
		Build: b,
	}

	if b.Id != nil {
		p.ID = *b.Id
	}

	if b.BuildNumber != nil {
		p.Number = *b.BuildNumber
	}

	if b.Status != nil {
		p.Status = PipelineStatus(*b.Status)
	}

	if b.Result != nil {
		p.Result = string(*b.Result)
	}

	if b.Definition != nil && b.Definition.Name != nil {
		p.Definition = *b.Definition.Name
	}

	if b.SourceBranch != nil {
		p.SourceBranch = *b.SourceBranch
	}

	if b.RequestedFor != nil && b.RequestedFor.DisplayName != nil {
		p.RequestedBy = *b.RequestedFor.DisplayName
	}

	if b.StartTime != nil {
		t := b.StartTime.Time
		p.StartTime = &t
	}

	if b.FinishTime != nil {
		t := b.FinishTime.Time
		p.FinishTime = &t
	}

	if b.QueueTime != nil {
		t := b.QueueTime.Time
		p.QueueTime = &t
	}

	return p
}

// Duration returns the duration of the pipeline run
func (p *Pipeline) Duration() string {
	if p.StartTime == nil {
		return "Not started"
	}

	endTime := time.Now()
	if p.FinishTime != nil {
		endTime = *p.FinishTime
	}

	duration := endTime.Sub(*p.StartTime)

	// Format duration nicely
	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm %ds", int(duration.Minutes()), int(duration.Seconds())%60)
	}
	return fmt.Sprintf("%dh %dm", int(duration.Hours()), int(duration.Minutes())%60)
}

// IsRunning returns true if the pipeline is currently running
func (p *Pipeline) IsRunning() bool {
	return p.Status == StatusInProgress
}

// StageInfo represents information about a pipeline stage
type StageInfo struct {
	Name       string
	State      string
	Result     string
	StartTime  *time.Time
	FinishTime *time.Time
	Jobs       []JobInfo
}

// JobInfo represents information about a job within a stage
type JobInfo struct {
	Name       string
	State      string
	Result     string
	StartTime  *time.Time
	FinishTime *time.Time
}

// ParseTimeline converts an Azure DevOps timeline to our stage/job models
func ParseTimeline(timeline *build.Timeline) []StageInfo {
	if timeline == nil || timeline.Records == nil {
		return []StageInfo{}
	}

	stages := make([]StageInfo, 0)
	stageMap := make(map[string]*StageInfo)

	// First pass: create stages
	for _, record := range *timeline.Records {
		if record.Type == nil {
			continue
		}

		if *record.Type == "Stage" {
			stage := StageInfo{
				Name:   getRecordName(record),
				State:  getRecordState(record),
				Result: getRecordResult(record),
				Jobs:   make([]JobInfo, 0),
			}

			if record.StartTime != nil {
				t := record.StartTime.Time
				stage.StartTime = &t
			}
			if record.FinishTime != nil {
				t := record.FinishTime.Time
				stage.FinishTime = &t
			}

			if record.Id != nil {
				stageMap[record.Id.String()] = &stage
			}
			stages = append(stages, stage)
		}
	}

	// Second pass: add jobs to stages
	for _, record := range *timeline.Records {
		if record.Type == nil {
			continue
		}

		if *record.Type == "Job" && record.ParentId != nil {
			job := JobInfo{
				Name:   getRecordName(record),
				State:  getRecordState(record),
				Result: getRecordResult(record),
			}

			if record.StartTime != nil {
				t := record.StartTime.Time
				job.StartTime = &t
			}
			if record.FinishTime != nil {
				t := record.FinishTime.Time
				job.FinishTime = &t
			}

			if stage, ok := stageMap[record.ParentId.String()]; ok {
				stage.Jobs = append(stage.Jobs, job)
			}
		}
	}

	return stages
}

func getRecordName(record build.TimelineRecord) string {
	if record.Name != nil {
		return *record.Name
	}
	return "Unknown"
}

func getRecordState(record build.TimelineRecord) string {
	if record.State != nil {
		return string(*record.State)
	}
	return "Unknown"
}

func getRecordResult(record build.TimelineRecord) string {
	if record.Result != nil {
		return string(*record.Result)
	}
	return "None"
}
