package html_test

import (
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestGoQuery(t *testing.T) {

	t.Run("TestGoQuery", func(t *testing.T) {

		assert := assert.New(t)

		file, err := os.Open("./test_files/goquery.html")

		assert.Nil(err)

		doc, err := goquery.NewDocumentFromReader(file)

		assert.Nil(err)

		{
			// tag id
			el := doc.Find("p#about-desc-p")

			assert.NotNil(el)
			assert.Equal(1, len(el.Nodes))
		}
		{
			// tag class
			el := doc.Find(".main h2.header")

			assert.NotNil(el)
			assert.Equal(3, len(el.Nodes))

		}
		{
			el := doc.Find("h1,h2")

			assert.NotNil(el)
			assert.Equal(4, len(el.Nodes))

		}

		{
			// this is not working as expected
			el := doc.Find(".main>.list li:first")

			assert.NotNil(el)
			assert.Equal(0, len(el.Nodes))

		}

		{
			el := doc.Find(".main .list>li:nth-child(1)")

			assert.NotNil(el)
			assert.Equal(1, len(el.Nodes))
			assert.Equal("Service 1", el.Text())

		}

		{
			el := doc.Find(".main .list>li:first-child")

			assert.NotNil(el)
			assert.Equal(1, len(el.Nodes))
			assert.Equal("Service 1", el.Text())

		}

		{
			el := doc.Find(".main p:first-child")

			assert.NotNil(el)
			assert.Equal(1, len(el.Nodes))
			assert.Equal("This is 1st paragraph", el.Text())

		}

		{
			el := doc.Find(`form>input[type="email"]`)

			assert.NotNil(el)
			assert.Equal(1, len(el.Nodes))
			assert.Equal("", el.AttrOr("required", "false"))
		}

		{
			el := doc.Find(`form>textarea[name="message"]`)

			assert.NotNil(el)
			assert.Equal(1, len(el.Nodes))
			assert.Equal("false", el.AttrOr("required", "false"))
		}

	})

	t.Run("TestGoQuery2", func(t *testing.T) {
		assert := assert.New(t)

		root := `
		<!DOCTYPE html>

		<html>
		  <head>
			<meta charset="UTF-8" />
			<title>This is title 1</title>
		  </head>
		  <body>
			<div class="container">
			  <h1>This is an H1 heading</h1>
			  <h2>This is an H2 heading</h2>
			  <p>This is a paragraph 1</p>
			  <p>This is a paragraph 1</p>
			  <p>Total paragraph 3</p>
			</div>
		  </body>
		</html>
		
		<style>
		  .container {
			background-color: #ffffff;
			text-align: center;
			padding: 16px;
			border-radius: 8px;
		  }
		
		  h1 {
			color: #ff6d5a;
			font-size: 24px;
			font-weight: bold;
			padding: 8px;
			color:red;
		  }
		
		  h2 {
			color: #909399;
			font-size: 18px;
			font-weight: bold;
			padding: 8px;
		  }
		</style>
		
		<script>
		  console.log("Hello World!", data1);
		</script>		
		`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(root))

		assert.Nil(err)

		{
			el := doc.Find(`p:first-of-type`)

			assert.NotNil(el)
			assert.Equal(1, len(el.Nodes))
			html, err := el.Html()
			assert.Nil(err)
			assert.Equal("This is a paragraph 1", html)

		}

	})
}
