package schedule_trigger

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

const (
	// Category is the category of ManualTriggerNode.
	Category = structs.CategoryTrigger

	// Name is the name of ManualTriggerNode.
	Name = "n8n-nodes-base.scheduleTrigger"
)

const (
	FieldSeconds        = "seconds"
	FieldMinutes        = "minutes"
	FieldHours          = "hours"
	FieldDays           = "days"
	FieldWeeks          = "weeks"
	FieldMonths         = "months"
	FieldCronExpression = "cronExpression"
)

var (
	//go:embed node.json
	rawJson []byte
)

type (
	ScheduleTrigger struct {
		spec *structs.WorkflowNodeSpec
	}

	ScheduleParams struct {
		Rule scheduleParamsRule `json:"rule"`
	}

	scheduleParamsRule struct {
		Interval []scheduleParamsRuleInterval `json:"interval"`
	}

	scheduleParamsRuleInterval struct {
		Field               string `json:"field"`
		TriggerAtDayOfMonth int    `json:"triggerAtDayOfMonth"`
		TriggerAtHour       int    `json:"triggerAtHour"`
		TriggerAtMinute     int    `json:"triggerAtMinute"`

		TriggerAtDay []int `json:"triggerAtDay"` // 0-6, 0 is Sunday

		Expression string `json:"expression"`

		DaysInterval    int `json:"daysInterval"`
		MonthsInterval  int `json:"monthsInterval"`
		HoursInterval   int `json:"hoursInterval"`
		MinutesInterval int `json:"minutesInterval"`
		SecondsInterval int `json:"secondsInterval"`
		WeeksInterval   int `json:"weeksInterval"`
	}

	scheduleTriggerSpecOption struct {
		Name   string
		Values []scheduleTriggerSpecOptionValue
	}

	scheduleTriggerSpecOptionValue struct {
		Default interface{}
		Name    string
	}
)

func init() {
	trigger := &ScheduleTrigger{
		spec: &structs.WorkflowNodeSpec{},
	}
	trigger.spec.JsonConfig = rawJson
	trigger.spec.GenerateSpec()

	core.Register(trigger)
}

func (trigger *ScheduleTrigger) Category() structs.NodeObjectCategory {
	return Category
}

func (trigger *ScheduleTrigger) Name() string {
	return Name
}

func (trigger *ScheduleTrigger) DefaultSpec() interface{} {
	return trigger.spec
}

func (trigger *ScheduleTrigger) Execute(ctx context.Context, _ *structs.NodeExecuteInput) *structs.NodeExecutionResult {
	ts := time.Now()
	data := structs.NodeData{
		{
			"timestamp":     ts.Unix(),
			"Readable date": ts.Format(time.DateTime),
			"Readable time": ts.Format(time.TimeOnly),
			"Day of week":   ts.Weekday().String(),
			"Year":          ts.Year(),
			"Month":         ts.Month().String(),
			"Day of month":  ts.Day(),
			"Hour":          ts.Hour(),
			"Minute":        ts.Minute(),
			"Second":        ts.Second(),
			"Timezone":      ts.Location().String(),
		},
	}
	return core.GenerateSuccessResponse(data, []structs.NodeData{})
}

func (trigger *ScheduleTrigger) Trigger(ctx context.Context, input *structs.WorkflowNode) string {
	// 1. parse input.Params.Parameters to ScheduleParams
	// 2. use the cron expression, otherwise implement a cron.Schedule algorithm for intervals.
	// 3. generate cron job

	// there must be a rule.interval field, so we can safely assume it exists
	params := input.Parameters["rule"].(map[string]interface{})["interval"].([]interface{})
	intervals := trigger.SetDefaultValues(params)

	if len(intervals) == 0 {
		return ""
	}

	// We only need the first interval to generate the cron expression.
	// All other intervals are ignored.
	return trigger.GenerateCronExpression(intervals[0])
}

func (trigger *ScheduleTrigger) GenerateCronExpression(interval scheduleParamsRuleInterval) string {
	switch interval.Field {
	case FieldCronExpression:
		return interval.Expression
	case FieldSeconds:
		return fmt.Sprintf("@every %ds", interval.SecondsInterval)
	case FieldMinutes:
		return fmt.Sprintf("@every %dm", interval.MinutesInterval)
	case FieldHours:
		if interval.HoursInterval == 1 {
			return fmt.Sprintf("%d %d * * *", interval.TriggerAtMinute, interval.TriggerAtHour)
		}
		return fmt.Sprintf("%d */%d * * *", interval.TriggerAtMinute, interval.HoursInterval)
	case FieldDays:
		if interval.DaysInterval == 1 {
			return fmt.Sprintf("%d %d * * *", interval.TriggerAtMinute, interval.TriggerAtHour)
		}
		return fmt.Sprintf("%d %d */%d * *", interval.TriggerAtMinute, interval.TriggerAtHour, interval.DaysInterval)
	case FieldWeeks:
		var days string
		if len(interval.TriggerAtDay) == 0 {
			days = "*"
		} else {
			daysList := make([]string, len(interval.TriggerAtDay))
			for i := range interval.TriggerAtDay {
				daysList[i] = fmt.Sprintf("%d", interval.TriggerAtDay[i])
			}
			days = strings.Join(daysList, ",")
		}
		return fmt.Sprintf("%d %d * * %s", interval.TriggerAtMinute, interval.TriggerAtHour, days)
	case FieldMonths:
		if interval.MonthsInterval == 1 {
			return fmt.Sprintf("%d %d %d * *", interval.TriggerAtMinute, interval.TriggerAtHour, interval.TriggerAtDayOfMonth)
		}
		return fmt.Sprintf("%d %d %d */%d *", interval.TriggerAtMinute, interval.TriggerAtHour, interval.TriggerAtDayOfMonth, interval.MonthsInterval)
	}
	return ""
}

func (trigger *ScheduleTrigger) SetDefaultValues(paramsIntervals []interface{}) []scheduleParamsRuleInterval {
	options := trigger.getDefaultOptions()
	defaultValue := make(map[string]interface{}, len(options.Values))
	for i := range options.Values {
		value := options.Values[i]
		defaultValue[value.Name] = value.Default
	}

	defaultInterval := scheduleParamsRuleInterval{}
	trigger.unmarshal(defaultValue, &defaultInterval)

	var res []scheduleParamsRuleInterval
	if len(paramsIntervals) == 0 {
		res = []scheduleParamsRuleInterval{
			defaultInterval,
		}
		return res
	}

	for i := range paramsIntervals {
		v := paramsIntervals[i].(map[string]interface{})
		for j := range defaultValue {
			if _, ok := v[j]; !ok {
				v[j] = defaultValue[j]
			}
		}
		var interval scheduleParamsRuleInterval
		err := trigger.unmarshal(v, &interval)
		if err != nil {
			core.Errorf("failed to unmarshal interval %d: %v", i, err)
		}
		res = append(res, interval)
	}

	return res
}

func (trigger *ScheduleTrigger) unmarshal(data map[string]interface{}, target interface{}) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, &target)
}

func (trigger *ScheduleTrigger) getDefaultOptions() *scheduleTriggerSpecOption {
	options := scheduleTriggerSpecOption{}
	for i := range trigger.spec.NodeSpec.Properties {
		property := trigger.spec.NodeSpec.Properties[i]
		if trigger.spec.NodeSpec.Properties[i].Name == "rule" {
			// safe to ignore error here
			optionsRaw, _ := json.Marshal(property.Options[0])
			_ = json.Unmarshal(optionsRaw, &options)
			break
		}
	}
	return &options
}
