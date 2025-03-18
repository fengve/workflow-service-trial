package html_test

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"

	htmlnode "github.com/sugerio/workflow-service-trial/service/workflow_service/nodes/html"
)

func TestGetValue(t *testing.T) {
	assert := require.New(t)

	t.Run("TestGetValue empty", func(t *testing.T) {

		el := &goquery.Selection{}
		valueData := &htmlnode.ParameterExtractionValueOptions{}
		options := htmlnode.ParameterOptions{}

		res, err := htmlnode.GetValue(el, valueData, &options)
		assert.Equal("", res)
		assert.Nil(err)

	})

	t.Run("TestGetValue No return value defined", func(t *testing.T) {

		node := strings.NewReader(`
			<label for="cars">Choose a car:</label>
				<select name="cars" id="cars">
				<optgroup label="Swedish Cars">
					<option value="volvo">Volvo</option>
					<option value="saab">Saab</option>
				</optgroup>
				<optgroup label="German Cars">
					<option value="mercedes">Mercedes</option>
					<option value="audi">Audi</option>
				</optgroup>
				</select>
			`,
		)

		doc, err := goquery.NewDocumentFromReader(node)
		assert.Nil(err)

		selectEl := doc.Find("select")

		assert.NotNil(selectEl)
		assert.Len(selectEl.Nodes, 1)

		valueData := &htmlnode.ParameterExtractionValueOptions{}
		options := htmlnode.ParameterOptions{}

		res, err := htmlnode.GetValue(selectEl, valueData, &options)
		assert.Equal("", res)
		assert.Error(err, "No return value defined")

	})

	t.Run("TestGetValue extract value", func(t *testing.T) {
		{
			node := strings.NewReader(`
			<label for="cars">Choose a car:</label>
				<select name="cars" id="cars">
				<optgroup label="Swedish Cars">
					<option value="volvo">Volvo</option>
					<option value="saab">Saab</option>
				</optgroup>
				<optgroup label="German Cars">
					<option value="mercedes">Mercedes</option>
					<option value="audi">Audi</option>
				</optgroup>
			</select>
			`,
			)

			doc, err := goquery.NewDocumentFromReader(node)
			assert.Nil(err)

			selectEl := doc.Find("select")

			assert.NotNil(selectEl)
			assert.Len(selectEl.Nodes, 1)

			valueData := &htmlnode.ParameterExtractionValueOptions{
				ReturnValue: "value",
			}
			options := htmlnode.ParameterOptions{}

			res, err := htmlnode.GetValue(selectEl, valueData, &options)
			assert.Equal("", res)
			assert.Nil(err)
		}

		{
			// select
			node := strings.NewReader(`
			<label for="cars">Choose a car:</label>
			<select name="cars" id="cars">
				<optgroup label="Swedish Cars">
					<option value="volvo">Volvo</option>
					<option value="saab">Saab</option>
				</optgroup>
				<optgroup label="German Cars">
					<option value="mercedes" selected>Mercedes</option>
					<option value="audi">Audi</option>
				</optgroup>
			</select>
		`,
			)

			doc, err := goquery.NewDocumentFromReader(node)
			assert.Nil(err)

			selectEl := doc.Find("select")

			assert.NotNil(selectEl)
			assert.Len(selectEl.Nodes, 1)

			valueData := &htmlnode.ParameterExtractionValueOptions{
				ReturnValue: "value",
			}
			options := htmlnode.ParameterOptions{}

			res, err := htmlnode.GetValue(selectEl, valueData, &options)
			assert.Equal("mercedes", res)
			assert.Nil(err)
		}

		{
			// textarea
			node := strings.NewReader(`
			<textarea name="message">Hello, World!</textarea>	 
			`,
			)

			doc, err := goquery.NewDocumentFromReader(node)
			assert.Nil(err)

			selectEl := doc.Find("textarea")

			assert.NotNil(selectEl)
			assert.Len(selectEl.Nodes, 1)

			valueData := &htmlnode.ParameterExtractionValueOptions{
				ReturnValue: "value",
			}
			options := htmlnode.ParameterOptions{}

			res, err := htmlnode.GetValue(selectEl, valueData, &options)
			assert.Equal("Hello, World!", res)
			assert.Nil(err)
		}

		{
			// input
			node := strings.NewReader(`
				<input type="text" name="fname" value="Hello, World!">
 			`,
			)

			doc, err := goquery.NewDocumentFromReader(node)
			assert.Nil(err)

			selectEl := doc.Find("input")

			assert.NotNil(selectEl)
			assert.Len(selectEl.Nodes, 1)

			valueData := &htmlnode.ParameterExtractionValueOptions{
				ReturnValue: "value",
			}
			options := htmlnode.ParameterOptions{}

			res, err := htmlnode.GetValue(selectEl, valueData, &options)
			assert.Equal("Hello, World!", res)
			assert.Nil(err)
		}

	})

	t.Run("TestGetValue extract attribute", func(t *testing.T) {

		{
			// img
			node := strings.NewReader(`
			<img src="https://www.suger.io/s/uuxxid" alt="banner">
 			`,
			)

			doc, err := goquery.NewDocumentFromReader(node)
			assert.Nil(err)

			selectEl := doc.Find("img")

			assert.NotNil(selectEl)
			assert.Len(selectEl.Nodes, 1)

			valueData := &htmlnode.ParameterExtractionValueOptions{
				ReturnValue: "attribute",
				Attribute:   "src",
			}

			options := htmlnode.ParameterOptions{}

			res, err := htmlnode.GetValue(selectEl, valueData, &options)
			assert.Equal(`https://www.suger.io/s/uuxxid`, res)
			assert.Nil(err)
		}
	})

	t.Run("TestGetValue trimValues", func(t *testing.T) {

		{
			// textarea value with spaces
			node := strings.NewReader(`
			<textarea name="message"> 
				 Hello, World! 	
			</textarea>
 			`,
			)

			doc, err := goquery.NewDocumentFromReader(node)
			assert.Nil(err)

			selectEl := doc.Find("textarea")

			assert.NotNil(selectEl)
			assert.Len(selectEl.Nodes, 1)

			valueData := &htmlnode.ParameterExtractionValueOptions{
				ReturnValue: "value",
			}
			options := htmlnode.ParameterOptions{
				TrimValues: true,
			}

			res, err := htmlnode.GetValue(selectEl, valueData, &options)
			assert.Equal(`Hello, World!`, res)
			assert.Nil(err)

			options = htmlnode.ParameterOptions{}

			res, err = htmlnode.GetValue(selectEl, valueData, &options)
			assert.Equal(`Hello, World!`, res)
			assert.Nil(err)
		}
	})
}
