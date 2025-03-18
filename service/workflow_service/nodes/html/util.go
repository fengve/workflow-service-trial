package html

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sugerio/workflow-service-trial/service/workflow_service/core"
)

// // The extraction functions
func extractFunctions(selectedElement *goquery.Selection, valueData *ParameterExtractionValueOptions) (string, error) {

	returnValue := valueData.ReturnValue
	if returnValue == "" {
		return "", fmt.Errorf("No return value defined")
	}

	switch returnValue {

	case "attribute":
		attribute := valueData.Attribute
		if attribute == "" {
			return "", fmt.Errorf("No attribute defined")
		}
		val, ok := selectedElement.Attr(attribute)
		if !ok {
			return "", fmt.Errorf("Attribute %s not found", attribute)
		}

		return val, nil

	case "html":

		return selectedElement.Html()

	case "text":

		return selectedElement.Text(), nil

	case "value":

		if selectedElement.Is("textarea") {
			return selectedElement.Text(), nil
		}

		var val string
		if selectedElement.Is("select") {
			selectedElement.Find("option").EachWithBreak(
				func(i int, option *goquery.Selection) bool {
					selected, exists := option.Attr("selected")
					if exists && selected != "false" {
						val = option.AttrOr("value", option.Text())
						return false // break
					}
					return true
				})
			return val, nil

		}

		// if el.Is("input") {
		// }
		val, _ = selectedElement.Attr("value")
		return val, nil

	}

	return "", fmt.Errorf("Unknown return value %s", returnValue)
}

/**
 * Simple helper function which applies options
 */
func GetValue(selectedElement *goquery.Selection, valueData *ParameterExtractionValueOptions, options *ParameterOptions) (string, error) {
	if selectedElement == nil || valueData == nil {
		return "", fmt.Errorf("GetValue: Invalid arguments")
	}

	if selectedElement.Length() == 0 {
		return "", nil // no elements found
	}

	value, err := extractFunctions(selectedElement, valueData)

	if err != nil {
		return "", err
	}

	// FIXME: default value issue
	// trimValues, ok := options.TrimValues
	trimValues := true // default value

	if yes, err := core.ConvertToBool(trimValues); yes && err == nil {
		value = strings.TrimSpace(value)
	}

	if yes, err := core.ConvertToBool(options.CleanUpText); yes && err == nil {
		value = strings.ReplaceAll(value, "\r\n", "")
		value = strings.ReplaceAll(value, "\n", "")
		value = strings.ReplaceAll(value, "\r", "")
		re := regexp.MustCompile(`\s+`)         // \s+ matches one or more whitespace characters.
		value = re.ReplaceAllString(value, " ") // Replace all whitespace with a single space
	}

	return value, nil
}
