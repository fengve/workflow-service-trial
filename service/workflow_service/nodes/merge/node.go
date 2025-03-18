package mergenode

import (
	"context"
	_ "embed"
	"fmt"
	"math"
	"reflect"
	"strings"

	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
	"github.com/sugerio/workflow-service-trial/shared/structs"
)

const (
	Category = structs.CategoryExecutor
	Name     = "n8n-nodes-base.merge"
)

var (
	//go:embed node.json
	rawJson []byte
)

type (
	MergeExecutor struct {
		spec *structs.WorkflowNodeSpec
	}

	//parameters.options
	ParameterOptions struct {
		ClashHandling      ClashHandlingOption `json:"clashHandling,omitempty"`
		FuzzyCompare       bool                `json:"fuzzyCompare,omitempty"`
		IncludeUnpaired    bool                `json:"includeUnpaired,omitempty"`
		DisableDotNotation bool                `json:"disableDotNotation,omitempty"`
		MultipleMatches    string              `json:"multipleMatches,omitempty"`
	}

	ClashHandlingOption struct {
		Values ClashHandlingValue `json:"values,omitempty"`
	}

	ClashHandlingValue struct {
		ResolveClash  string `json:"resolveClash,omitempty"`
		MergeMode     string `json:"mergeMode,omitempty"`
		OverrideEmpty bool   `json:"overrideEmpty,omitempty"`
	}

	// parameters.mergeByFields
	MergeByFields struct {
		Values []PairToMatch `json:"values,omitempty"`
	}

	PairToMatch struct {
		Field1 string `json:"field1,omitempty"`
		Field2 string `json:"field2,omitempty"`
	}

	// for code logic not completely form parameters
	MatchFieldsOptions struct {
		JoinMode           string `json:"joinMode,omitempty"`
		OutputDataFrom     string `json:"outputDataFrom,omitempty"`
		MultipleMatches    string `json:"multipleMatches,omitempty"`
		DisableDotNotation bool   `json:"disableDotNotation,omitempty"`
		FuzzyCompare       bool   `json:"fuzzyCompare,omitempty"`
	}

	MergeMethod func(dest interface{}, source ...interface{}) interface{}

	MatchResult struct {
		EntryMatches []EntryMatches   `json:"entryMatches,omitempty"`
		Matched2     structs.NodeData `json:"matched2,omitempty"`
		Unmatched1   structs.NodeData `json:"unmatched1,omitempty"`
		Unmatched2   structs.NodeData `json:"unmatched2,omitempty"`
	}

	EntryMatches struct {
		Entry   structs.NodeSingleData
		Matches structs.NodeData
	}

	CompareFunction func(entry1, entry2 interface{}) bool
)

func init() {
	me := &MergeExecutor{
		spec: &structs.WorkflowNodeSpec{},
	}
	me.spec.JsonConfig = rawJson
	me.spec.GenerateSpec()

	core.Register(me)
}

func (me *MergeExecutor) Category() structs.NodeObjectCategory {
	return Category
}

func (me *MergeExecutor) Name() string {
	return Name
}

func (me *MergeExecutor) DefaultSpec() interface{} {
	return me.spec
}

func (me *MergeExecutor) Execute(ctx context.Context, input *structs.NodeExecuteInput) *structs.NodeExecutionResult {
	items1 := core.GetInputDataByIndex(input.Data, 0)
	items2 := core.GetInputDataByIndex(input.Data, 1)
	result := structs.NodeData{}

	mode, err := core.GetNodeParameterAsBasicType(Name, "mode", "append", input, 0)
	if err != nil {
		return core.GenerateFailedResponse(Name, err)
	}
	if mode == "append" {
		result = append(result, items1...)
		result = append(result, items2...)
		return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{result})
	}

	if mode == "chooseBranch" {
		output, err := core.GetNodeParameterAsBasicType(Name, "output", "input1", input, 0)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		if output == "input1" {
			result = append(result, items1...)
		} else if output == "input2" {
			result = append(result, items2...)
		} else if output == "empty" {
			result = append(result, structs.NodeSingleData{
				"json": map[string]interface{}{},
			})
		}
		return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{result})
	}

	if mode == "combine" {
		// Parse options
		options, err := core.GetNodeParameterAsType(Name, "options",
			ParameterOptions{}, input, 0,
		)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}

		combinationMode, err := core.GetNodeParameterAsBasicType(Name, "combinationMode", "mergeByFields", input, 0)
		if err != nil {
			return core.GenerateFailedResponse(Name, err)
		}
		if combinationMode == "multiplex" {
			if len(items1) == 0 || len(items2) == 0 {
				return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{result})
			}
			clashHandling := options.ClashHandling.Values
			if clashHandling.ResolveClash == "preferInput1" {
				items1, items2 = items2, items1
			}
			if clashHandling.ResolveClash == "addSuffix" {
				items1 = addSuffixToEntriesKeys(items1, "1")
				items2 = addSuffixToEntriesKeys(items2, "2")
			}

			mergeIntoSingleObject := selectMergeMethod(clashHandling)
			for _, item1 := range items1 {
				for _, item2 := range items2 {
					result = append(result, structs.NodeSingleData{
						"json": mergeIntoSingleObject(item1["json"], item2["json"]),
					})
				}
			}
			return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{result})
		} else if combinationMode == "mergeByPosition" {
			if len(items1) == 0 || len(items2) == 0 {
				result = append(result, items1...)
				result = append(result, items2...)
				return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{result})
			}

			clashHandling := options.ClashHandling.Values
			if clashHandling.ResolveClash == "preferInput1" {
				items1, items2 = items2, items1
			}
			if clashHandling.ResolveClash == "addSuffix" {
				items1 = addSuffixToEntriesKeys(items1, "1")
				items2 = addSuffixToEntriesKeys(items2, "2")
			}

			var maxCount int
			if options.IncludeUnpaired {
				maxCount = int(math.Max(float64(len(items1)), float64(len(items2))))
			} else {
				maxCount = int(math.Min(float64(len(items1)), float64(len(items2))))
			}

			mergeMethod := selectMergeMethod(clashHandling)
			for i := 0; i < maxCount; i++ {
				if i >= len(items1) {
					result = append(result, items2[i])
					continue
				}
				if i >= len(items2) {
					result = append(result, items1[i])
					continue
				}
				result = append(result, structs.NodeSingleData{
					"json": mergeMethod(items1[i]["json"], items2[i]["json"]),
				})
			}
			return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{result})
		} else if combinationMode == "mergeByFields" {
			clashHandling := options.ClashHandling.Values
			mergeByFields, err := core.GetNodeParameterAsType(Name, "mergeByFields",
				MergeByFields{}, input, 0,
			)
			if err != nil {
				return core.GenerateFailedResponse(Name, err)
			}
			matchFields := mergeByFields.Values
			checkErr := checkMatchFieldsInput(matchFields)
			if checkErr != nil {
				return core.GenerateFailedResponse(Name, checkErr)
			}

			joinMode, err := core.GetNodeParameterAsBasicType(Name, "joinMode",
				"keepMatches", input, 0)
			if err != nil {
				return core.GenerateFailedResponse(Name, err)
			}

			outputDataFrom, err := core.GetNodeParameterAsBasicType(Name, "outputDataFrom",
				"both", input, 0)
			if err != nil {
				return core.GenerateFailedResponse(Name, err)
			}

			matchFieldsOptions := MatchFieldsOptions{
				JoinMode:           joinMode,
				OutputDataFrom:     outputDataFrom,
				MultipleMatches:    options.MultipleMatches,
				DisableDotNotation: options.DisableDotNotation,
				FuzzyCompare:       options.DisableDotNotation,
			}

			if len(items1) == 0 || len(items2) == 0 {
				if joinMode == "keepMatches" {
					return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{result})
				} else if joinMode == "enrichInput1" && len(items1) == 0 {
					return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{result})
				} else if joinMode == "enrichInput2" && len(items2) == 0 {
					return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{result})
				} else {
					result = append(result, items1...)
					result = append(result, items2...)
					return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{result})
				}
			}

			matches := findMatches(items1, items2, matchFields, matchFieldsOptions)

			if joinMode == "keepMatches" || joinMode == "keepEverything" {
				output := structs.NodeData{}
				if outputDataFrom == "input1" {
					for _, entryMatch := range matches.EntryMatches {
						output = append(output, entryMatch.Entry)
					}
				} else if outputDataFrom == "input2" {
					output = matches.Matched2
				} else if outputDataFrom == "both" {
					output = mergeMatched(matches.EntryMatches, clashHandling, "")
				}

				if joinMode == "keepEverything" {
					unmatched1 := matches.Unmatched1
					unmatched2 := matches.Unmatched2
					if clashHandling.ResolveClash == "addSuffix" {
						unmatched1 = addSuffixToEntriesKeys(unmatched1, "1")
						unmatched2 = addSuffixToEntriesKeys(unmatched2, "2")
					}
					output = append(output, unmatched1...)
					output = append(output, unmatched2...)
				}
				result = append(result, output...)
			} else if joinMode == "keepNonMatches" {
				if outputDataFrom == "input1" {
					result = matches.Unmatched1
				} else if outputDataFrom == "input2" {
					result = matches.Unmatched2
				} else if outputDataFrom == "both" {
					output := structs.NodeData{}
					unmatched1 := addSourceField(matches.Unmatched1, "input1")
					unmatched2 := addSourceField(matches.Unmatched2, "input2")
					output = append(output, unmatched1...)
					output = append(output, unmatched2...)
					result = output
				}
				return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{result})
			} else if joinMode == "enrichInput1" || joinMode == "enrichInput2" {
				mergedEntries := mergeMatched(matches.EntryMatches, clashHandling, joinMode)
				if joinMode == "enrichInput1" {
					if clashHandling.ResolveClash == "addSuffix" {
						result = append(result, mergedEntries...)
						result = append(result, addSuffixToEntriesKeys(matches.Unmatched1, "1")...)
					} else {
						result = append(result, mergedEntries...)
						result = append(result, matches.Unmatched1...)
					}
				} else {
					if clashHandling.ResolveClash == "addSuffix" {
						result = append(result, mergedEntries...)
						result = append(result, addSuffixToEntriesKeys(matches.Unmatched2, "2")...)
					} else {
						result = append(result, mergedEntries...)
						result = append(result, matches.Unmatched2...)
					}
				}
			}
		}
	}
	return core.GenerateSuccessResponse(structs.NodeData{}, []structs.NodeData{result})
}

func mergeMatched(matched []EntryMatches, clashHandling ClashHandlingValue, joinMode string) structs.NodeData {
	result := structs.NodeData{}
	resolveClash := clashHandling.ResolveClash
	mergeIntoSingleObject := selectMergeMethod(clashHandling)
	for _, match := range matched {
		entry := match.Entry
		matches := match.Matches

		var mergedJson interface{}
		if resolveClash == "addSuffix" {
			entry = addSuffixToEntryKey(entry, "1")
			matches = addSuffixToEntriesKeys(matches, "2")
			matchesJsonObjets := []interface{}{}
			for _, match := range matches {
				matchesJsonObjets = append(matchesJsonObjets, match["json"])
			}
			mergedJson = mergeIntoSingleObject(entry["json"],
				matchesJsonObjets...)
		} else {
			if resolveClash == "" {
				if joinMode != "enrichInput2" {
					resolveClash = "preferInput2"
				} else {
					resolveClash = "preferInput1"
				}
			}
			if resolveClash == "preferInput1" {
				firstMatch := matches[0]
				restMatches := matches[1:]
				restMatchesJsonObjets := []interface{}{}
				for _, match := range restMatches {
					restMatchesJsonObjets = append(restMatchesJsonObjets, match["json"])
				}
				restMatchesJsonObjets = append(restMatchesJsonObjets, entry["json"])
				mergedJson = mergeIntoSingleObject(firstMatch["json"], restMatchesJsonObjets...)
			} else if resolveClash == "preferInput2" {
				matchesJsonObjets := []interface{}{}
				for _, match := range matches {
					matchesJsonObjets = append(matchesJsonObjets, match["json"])
				}
				mergedJson = mergeIntoSingleObject(entry["json"], matchesJsonObjets...)
			}
		}
		result = append(result, map[string]interface{}{
			"json": mergedJson,
		})
	}
	return result
}

func addSuffixToEntryKey(entry structs.NodeSingleData, suffix string) structs.NodeSingleData {
	// If entry is not a map(aka object), it can't be added a suffix.
	if oldJson, ok := entry["json"].(map[string]interface{}); ok {
		newJson := map[string]interface{}{}
		for key, val := range oldJson {
			newJson[fmt.Sprintf("%s_%s", suffix, key)] = val
		}
		newItem := entry
		newItem["json"] = newJson
		return newItem
	} else {
		return entry
	}
}

func addSuffixToEntriesKeys(data structs.NodeData, suffix string) structs.NodeData {
	result := structs.NodeData{}
	for _, item := range data {
		newItem := addSuffixToEntryKey(item, suffix)
		result = append(result, newItem)
	}
	return result
}

func selectMergeMethod(clashHandling ClashHandlingValue) MergeMethod {
	// TODO support OverrideEmpty
	mergeMode := clashHandling.MergeMode
	if clashHandling.OverrideEmpty {

	} else {
		if mergeMode == "deepMerge" {
			return merge
		} else if mergeMode == "shallowMerge" {
			return shallowMerge
		}
	}
	return merge
}

func shallowMerge(object interface{}, otherArgs ...interface{}) interface{} {
	result := object
	for _, arg := range otherArgs {
		result = shallowMergeObjects(result, arg)
	}
	return result
}

func shallowMergeObjects(dest, src interface{}) interface{} {
	destMap, ok1 := dest.(map[string]interface{})
	srcMap, ok2 := src.(map[string]interface{})

	if !ok1 || !ok2 {
		return src
	}

	result := make(map[string]interface{})
	for key, value := range destMap {
		result[key] = value
	}

	for key, srcValue := range srcMap {
		result[key] = srcValue
	}

	return result

}

func merge(object interface{}, otherArgs ...interface{}) interface{} {
	result := object
	for _, arg := range otherArgs {
		result = mergeObjects(result, arg)
	}
	return result
}

func mergeObjects(dest, src interface{}) interface{} {
	// TODO if both dest and src are array, if we merge the array?
	// the array items are anonymous, they may have different structs.

	// If any of dest and scr is not map(aka object), merge end.
	destMap, ok1 := dest.(map[string]interface{})
	srcMap, ok2 := src.(map[string]interface{})

	if !ok1 || !ok2 {
		return src
	}

	// Merge object recursively
	result := make(map[string]interface{})
	for key, value := range destMap {
		result[key] = value
	}

	for key, srcValue := range srcMap {
		if destValue, ok := destMap[key]; ok && isMap(destValue) && isMap(srcValue) {
			destMap := destValue.(map[string]interface{})
			srcMap := srcValue.(map[string]interface{})
			result[key] = mergeObjects(destMap, srcMap)
		} else {
			result[key] = srcValue
		}
	}

	return result
}

func isMap(value interface{}) bool {
	_, ok := value.(map[string]interface{})
	return ok
}

func checkMatchFieldsInput(values []PairToMatch) error {
	if len(values) == 1 && values[0].Field1 == "" && values[0].Field2 == "" {
		return fmt.Errorf("you need to define at least one pair of fields")
	}
	for i, pair := range values {
		if pair.Field1 == "" || pair.Field2 == "" {
			return fmt.Errorf("you need to define both fields for pair %d", i+1)
		}
	}
	return nil
}

// Find mathced item between items1 and items2 by specified fields.
// When two items in items1 and items2 has the same target field, then they can be merged.
func findMatches(items1 structs.NodeData, items2 structs.NodeData,
	fieldsToMatch []PairToMatch, options MatchFieldsOptions) MatchResult {
	isEntriesEqual := fuzzyCompare(options.FuzzyCompare)
	disableDotNotation := options.DisableDotNotation
	multipleMatches := "all"
	if options.MultipleMatches != "" {
		multipleMatches = options.MultipleMatches
	}

	matchResult := MatchResult{}
	matchedInItems2 := make(map[int]bool)

	for _, entry1 := range items1 {
		// If the item.json is not map(aks object), skip.
		if !isMap(entry1["json"]) {
			continue
		}
		entry1Json := entry1["json"].(map[string]interface{})

		// Filter target filed & value in items1
		lookup := structs.NodeSingleData{}
		lookupValid := true
		for _, matchCase := range fieldsToMatch {
			var valueToCompare interface{}
			if disableDotNotation {
				valueToCompare = entry1Json[matchCase.Field1]
			} else {
				valueToCompare = getValFromObject(entry1Json, matchCase.Field1)
			}
			lookup[matchCase.Field2] = valueToCompare
			if valueToCompare == nil {
				lookupValid = false
			}
		}
		if !lookupValid {
			continue
		}

		// Find matched items in items2
		matchedItems, matchedIndexes := findEntryMatches(items2, lookup, disableDotNotation, isEntriesEqual, multipleMatches == "all")
		for _, matchedIndex := range matchedIndexes {
			matchedInItems2[matchedIndex] = true
		}

		if len(matchedItems) > 0 {
			if options.OutputDataFrom == "both" || options.JoinMode == "enrichInput1" ||
				options.JoinMode == "enrichInput2" {
				for _, matchedItem := range matchedItems {
					matchResult.EntryMatches = append(matchResult.EntryMatches, EntryMatches{
						Entry:   entry1,
						Matches: append(structs.NodeData{}, matchedItem),
					})
				}
			} else {
				matchResult.EntryMatches = append(matchResult.EntryMatches, EntryMatches{
					Entry:   entry1,
					Matches: matchedItems,
				})
			}
		} else {
			matchResult.Unmatched1 = append(matchResult.Unmatched1, entry1)
		}
	}

	for i, entry := range items2 {
		if matchedInItems2[i] {
			matchResult.Matched2 = append(matchResult.Matched2, entry)
		} else {
			matchResult.Unmatched2 = append(matchResult.Unmatched2, entry)
		}
	}

	return matchResult
}

func findEntryMatches(data structs.NodeData, lookup structs.NodeSingleData,
	disableDotNotation bool, isEntriesEqual CompareFunction, findAll bool) (structs.NodeData, []int) {
	result := structs.NodeData{}
	indexes := []int{}
	for i, entry := range data {
		if entry == nil {
			continue
		}
		if !isMap(entry["json"]) {
			continue
		}
		entryJson := entry["json"].(map[string]interface{})

		match := true
		for key, expectedValue := range lookup {
			var entryFieldValue interface{}
			if disableDotNotation {
				entryFieldValue = entryJson[key]
			} else {
				entryFieldValue = getValFromObject(entryJson, key)
			}
			if !isEntriesEqual(expectedValue, entryFieldValue) {
				match = false
				break
			}
		}
		if match {
			result = append(result, entry)
			indexes = append(indexes, i)
			if !findAll {
				return result, indexes
			}
		}
	}
	return result, indexes
}

func getValFromObject(object map[string]interface{}, path string) interface{} {
	keys := strings.Split(path, ".")
	for _, key := range keys {
		value, ok := object[key]
		if !ok {
			return nil
		}
		if nestedObject, ok := value.(map[string]interface{}); ok {
			object = nestedObject
		} else {
			return value
		}
	}

	return nil
}

func fuzzyCompare(useFuzzyCompare bool) CompareFunction {
	// TODO handle different data types fuzzyCompare

	return func(entry1, entry2 interface{}) bool {
		return reflect.DeepEqual(entry1, entry2)
	}
}

func addSourceField(data structs.NodeData, sourceField string) structs.NodeData {
	result := structs.NodeData{}
	for _, item := range data {
		if jsonObject, ok := item["json"].(map[string]interface{}); ok {
			jsonObject["_source"] = sourceField
			newItem := item
			newItem["json"] = jsonObject
			result = append(result, newItem)
		} else {
			result = append(result, item)
		}
	}
	return result
}
