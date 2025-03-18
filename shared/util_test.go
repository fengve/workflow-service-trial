package shared_test

// Command to run this test only
// go test -v shared/util_test.go

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/sugerio/workflow-service-trial/shared"
)

type UtilTestSuite struct {
	suite.Suite
}

type TestObject struct {
	Field1       string        `json:"field1"`
	Field2       int           `json:"field2"`
	NestedStruct *NestedObject `json:"nestedStruct"`
}

type NestedObject struct {
	SubField1 string
}

func Test_UtilTestSuite(t *testing.T) {
	suite.Run(t, new(UtilTestSuite))
}

func (s *UtilTestSuite) TestUtil() {
	s.T().Run("GetAwsSqsArnFromUrl", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		result, err := shared.GetAwsSqsArnFromUrl("https://sqs.us-west-2.amazonaws.com/821785902361/aws-mp-subscription-queue-test-0")
		assert.Nil(err)
		assert.Equal(result, "arn:aws:sqs:us-west-2:821785902361:aws-mp-subscription-queue-test-0")
	})

	s.T().Run("ConvertInterfaceToArray", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		// Valid case
		array, err := shared.ConvertInterfaceToArray[string]([]interface{}{"key1", "key2"})
		assert.Nil(err)
		assert.Equal([]string{"key1", "key2"}, array)

		// Invalid case
		_, err = shared.ConvertInterfaceToArray[string]([]interface{}{1, 2})
		assert.Error(err)
	})

	s.T().Run("EqualObjects", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		a := []TestObject{}
		b := []TestObject{}
		assert.True(shared.EqualObjects(a, b))

		b = []TestObject{{
			Field1: "test-key",
			Field2: 7,
		}}
		assert.False(shared.EqualObjects(a, b))
	})

	s.T().Run("ParseDuration_ISO_8601", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		months, days, err := shared.ParseDuration_ISO_8601("P1Y3M5DT6H7M8S")
		assert.Nil(err)
		assert.Equal(months, 15.0)
		assert.Equal(days, 5.0)

		months, days, err = shared.ParseDuration_ISO_8601("P2Y2.5M")
		assert.Nil(err)
		assert.Equal(months, 26.5)
		assert.Equal(days, 0.0)

		months, days, err = shared.ParseDuration_ISO_8601("PT4H")
		assert.Nil(err)
		assert.Equal(months, 0.0)
		assert.Equal(days, 0.0)

		months, days, err = shared.ParseDuration_ISO_8601("P60D")
		assert.Nil(err)
		assert.Equal(months, 2.0)
		assert.Equal(days, 0.0)
	})

	s.T().Run("TruncateStr", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		str := "abcdefghijk"
		assert.Equal(shared.TruncateStr(str, 3), "abc")
		assert.Equal(shared.TruncateStr(str, 11), "abcdefghijk")
		assert.Equal(shared.TruncateStr(str, 20), "abcdefghijk")
	})

	s.T().Run("UpdateNonEmptyFields", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		type TestStruct struct {
			Name   string
			Source string
			Target string
			Count  int
		}

		t1 := TestStruct{
			Name:   "test1",
			Source: "source1",
			Target: "target1",
			Count:  1,
		}

		t2 := TestStruct{
			Count: 2,
		}

		shared.UpdateNonEmptyFields(&t1, t2)

		assert.Equal(t1.Name, "test1")
		assert.Equal(t1.Source, "source1")
		assert.Equal(t1.Target, "target1")
		assert.Equal(t1.Count, 2)
	})

	s.T().Run("ParseToFloat", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		var value interface{}

		// A real float64
		value = 32.0
		result, err := shared.ParseToFloat(value)
		assert.NoError(err)
		assert.Equal(32.0, result)

		// An integer
		value = 32
		result, err = shared.ParseToFloat(value)
		assert.NoError(err)
		assert.Equal(32.0, result)

		// A string
		value = "32"
		result, err = shared.ParseToFloat(value)
		assert.NoError(err)
		assert.Equal(32.0, result)

		// Empty string
		value = ""
		result, err = shared.ParseToFloat(value)
		assert.Error(err)

		// Boolean
		value = false
		result, err = shared.ParseToFloat(value)
		assert.Error(err)
	})

	s.T().Run("GetNameWithNextVersion", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		assert.Equal(shared.GetNameWithNextVersion("test"), "test_v2")
		assert.Equal(shared.GetNameWithNextVersion("test_v2"), "test_v3")
		assert.Equal(shared.GetNameWithNextVersion("test_v3"), "test_v4")
		assert.Equal(shared.GetNameWithNextVersion("test_v100"), "test_v101")
		assert.Equal(shared.GetNameWithNextVersion("test_vvi"), "test_vvi_v2")
		assert.Equal(shared.GetNameWithNextVersion("test_v35v"), "test_v35v_v2")
	})

	s.T().Run("SetField SetSimpleField", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)

		testObj := &TestObject{}
		err := shared.SetField(testObj, "Field1", "new value")
		assert.NoError(err)
		assert.Equal("new value", testObj.Field1)
	})

	s.T().Run("SetField SetIntegerField", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)

		testObj := &TestObject{}
		err := shared.SetField(testObj, "Field2", 123)
		assert.NoError(err)
		assert.Equal(123, testObj.Field2)
	})

	s.T().Run("SetField - Type mismatch error", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)

		testObj := &TestObject{}
		err := shared.SetField(testObj, "Field1", 123)
		assert.Error(err) // expects an error because of the type mismatch
	})

	s.T().Run("SetField - Non-existent field error", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)

		testObj := &TestObject{}
		err := shared.SetField(testObj, "noSuchField", "value")
		assert.Error(err) // expects an error because the field doesn't exist
	})

	s.T().Run("SetField - Set nested struct field", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)

		testObj := &TestObject{NestedStruct: &NestedObject{}}
		err := shared.SetField(testObj, "NestedStruct.SubField1", "sub value")
		assert.NoError(err)
		assert.Equal("sub value", testObj.NestedStruct.SubField1)
	})

	s.T().Run("BatchProcess", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		// An example worker function that turns each element in the input list into uppercase.
		// Each time the function gets called, we update the count.
		functionCalled := 0
		f := func(inputs []string) []string {
			functionCalled += 1
			results := make([]string, 0, len(inputs))
			for _, x := range inputs {
				results = append(results, strings.ToUpper(x))
			}
			return results
		}

		inputs := []string{"a", "b", "c", "d"}
		expectedOutput := []string{"A", "B", "C", "D"}
		outputsCh := shared.BatchProcess[string, string](inputs, 2, f)
		outputs := shared.ChanToSlice(outputsCh)
		assert.Equal(2, functionCalled)
		assert.Equal(len(inputs), len(outputs))
		assert.Equal(expectedOutput, outputs)
	})

	s.T().Run("BatchProcess with Remainder", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		// An example worker function that turns each element in the input list into uppercase.
		// Each time the function gets called, we update the count.
		functionCalled := 0
		f := func(inputs []string) []string {
			functionCalled += 1
			results := make([]string, 0, len(inputs))
			for _, x := range inputs {
				results = append(results, strings.ToUpper(x))
			}
			return results
		}

		inputs := []string{"a", "b", "c", "d", "e"}
		expectedOutput := []string{"A", "B", "C", "D", "E"}
		outputsCh := shared.BatchProcess[string, string](inputs, 2, f)
		outputs := shared.ChanToSlice(outputsCh)
		assert.Equal(3, functionCalled)
		assert.Equal(len(inputs), len(outputs))
		assert.Equal(expectedOutput, outputs)
	})

	s.T().Run("BatchProcess with Empty Input", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		// An example worker function that turns each element in the input list into uppercase.
		// Each time the function gets called, we update the count.
		functionCalled := 0
		f := func(inputs []string) []string {
			functionCalled += 1
			results := make([]string, 0, len(inputs))
			for _, x := range inputs {
				results = append(results, strings.ToUpper(x))
			}
			return results
		}

		inputs := []string{}
		expectedOutput := []string{}
		outputsCh := shared.BatchProcess[string, string](inputs, 2, f)
		outputs := shared.ChanToSlice(outputsCh)
		assert.Equal(0, functionCalled)
		assert.Equal(len(inputs), len(outputs))
		assert.Equal(expectedOutput, outputs)
	})

	s.T().Run("BatchProcess with BatchSize Same with InputLength", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		// An example worker function that turns each element in the input list into uppercase.
		// Each time the function gets called, we update the count.
		functionCalled := 0
		f := func(inputs []string) []string {
			functionCalled += 1
			results := make([]string, 0, len(inputs))
			for _, x := range inputs {
				results = append(results, strings.ToUpper(x))
			}
			return results
		}

		inputs := []string{"a", "b", "c", "d", "e"}
		expectedOutput := []string{"A", "B", "C", "D", "E"}
		outputsCh := shared.BatchProcess[string, string](inputs, 5, f)
		outputs := shared.ChanToSlice(outputsCh)
		assert.Equal(1, functionCalled)
		assert.Equal(len(inputs), len(outputs))
		assert.Equal(expectedOutput, outputs)
	})

	s.T().Run("ConvertInterface", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		type TestInput struct {
			CurrencyCode string `json:"currencyCode,omitempty"`
			Nanos        int64  `json:"nanos,omitempty"`
			Units        int64  `json:"units,omitempty"`
		}
		type TestOutput struct {
			CurrencyCode string `json:"currencyCode,omitempty"`
			Nanos        int64  `json:"nanos,omitempty"`
			Units        int64  `json:"units,omitempty"`
		}
		testInput := TestInput{
			CurrencyCode: "USD",
			Nanos:        100,
			Units:        200,
		}

		testOutput, err := shared.ConvertInterface[TestOutput](testInput)
		assert.Nil(err)
		assert.Equal(testInput.CurrencyCode, testOutput.CurrencyCode)
		assert.Equal(testInput.Nanos, testOutput.Nanos)
		assert.Equal(testInput.Units, testOutput.Units)
	})

	s.T().Run("RemoveNullOrEmptyFields", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		emptyStr := ""
		validStr := "valid"
		m := map[string]interface{}{
			"field1": "value1", // valid
			"field2": nil,
			"field3": "",
			"field4": 0,         // valid
			"field5": &emptyStr, // pointer to empty string
			"field6": validStr,  // valid
		}
		shared.RemoveNullOrEmptyFields(m)
		assert.Equal(3, len(m))
		assert.Equal("value1", m["field1"])
		assert.Equal(0, m["field4"])
		assert.Equal(validStr, m["field6"])
	})

	s.T().Run("TimeEqualInDate Hour Minute Second", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		time1 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
		time2 := time.Date(2021, 1, 1, 3, 0, 0, 0, time.UTC)
		assert.True(shared.TimeEqualInDate(time1, time2))

		assert.False(shared.TimeEqualInHour(time1, time2))
		time1 = time.Date(2021, 1, 1, 3, 6, 0, 0, time.UTC)
		assert.True(shared.TimeEqualInHour(time1, time2))

		assert.False(shared.TimeEqualInMinute(time1, time2))
		time2 = time.Date(2021, 1, 1, 3, 6, 13, 0, time.UTC)
		assert.True(shared.TimeEqualInMinute(time1, time2))
		time1 = time.Date(2021, 1, 1, 2, 6, 0, 0, time.UTC)
		assert.False(shared.TimeEqualInMinute(time1, time2))
		time1 = time.Date(2021, 1, 2, 3, 6, 0, 0, time.UTC)
		assert.False(shared.TimeEqualInMinute(time1, time2))

		assert.False(shared.TimeEqualInSecond(time1, time2))
		time1 = time.Date(2021, 1, 1, 3, 6, 13, 123, time.UTC)
		assert.True(shared.TimeEqualInSecond(time1, time2))
	})

	s.T().Run("GroundTimeMonth_UTC", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)

		time1 := time.Date(2021, 1, 15, 0, 0, 0, 0, time.UTC)
		time2 := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
		assert.Equal(time2, shared.GroundTimeMonth_UTC(time1))
	})

	s.T().Run("IsEpochBeginning", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)

		time1 := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
		assert.True(shared.IsEpochBeginning(time1))

		time2 := time.Date(1980, 1, 1, 0, 0, 0, 1, time.UTC)
		assert.False(shared.IsEpochBeginning(time2))
	})

	s.T().Run("GetDurationInMonth GetDurationInDay", func(t *testing.T) {
		assert := require.New(s.T())

		startTime := time.Now()
		assert.Equal(shared.GetDurationInMonth(startTime, startTime.AddDate(0, 1, 0)), int64(1))
		assert.Equal(shared.GetDurationInMonth(startTime, startTime.AddDate(0, 1, 24)), int64(2))
		assert.Equal(shared.GetDurationInMonth(startTime, startTime.AddDate(0, 1, 35)), int64(2))
		assert.Equal(shared.GetDurationInMonth(startTime, startTime.AddDate(0, 1, 60)), int64(3))

		startTime = time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
		endTime := time.Date(2024, 7, 31, 0, 0, 0, 0, time.UTC)
		assert.Equal(shared.GetDurationInDay(startTime, endTime), int64(153))
	})

	s.T().Run("MergeMaps", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		map1 := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}
		map2 := map[string]interface{}{
			"key2": "value2-updated",
			"key3": "value3",
		}
		map3 := map[string]interface{}{
			"key3": "value3-updated",
			"key4": "value4",
		}
		merged := shared.MergeMaps(map1, map2, map3)
		assert.Equal(4, len(merged))
		assert.Equal("value1", merged["key1"])
		assert.Equal("value2-updated", merged["key2"])
		assert.Equal("value3-updated", merged["key3"])
		assert.Equal("value4", merged["key4"])
	})

	s.T().Run("Merge Structs", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		type TestWorkloadOfferInfo struct {
			CommitAmount       float64  `json:"commitAmount"`
			Currency           string   `json:"currency"`
			BuyerAwsAccountIds []string `json:"buyerAwsAccountIds,omitempty"`
		}

		type TestWorkloadOffer struct {
			Name         string                `json:"name"`
			CreationTime time.Time             `json:"creationTime"`
			ExpireTime   *time.Time            `json:"expireTime,omitempty"`
			Info         TestWorkloadOfferInfo `json:"info"`
		}

		creationTime := time.Now()
		expireTime := time.Now().Add(48 * time.Hour)
		offer1 := TestWorkloadOffer{
			Name:         "offer1",
			CreationTime: creationTime,
			ExpireTime:   &expireTime,
			Info: TestWorkloadOfferInfo{
				CommitAmount: 648,
				Currency:     "USD",
			},
		}
		offer2 := TestWorkloadOffer{
			Info: TestWorkloadOfferInfo{
				CommitAmount:       100,
				BuyerAwsAccountIds: []string{"123456789012"},
			},
		}

		shared.MergeStruct(&offer1, &offer2)
		assert.Equal("offer1", offer1.Name)
		assert.Equal("USD", offer1.Info.Currency)
		assert.Equal(creationTime, offer1.CreationTime)
		assert.Equal(expireTime, *offer1.ExpireTime)
		assert.Equal(float64(100), offer1.Info.CommitAmount)
		assert.Equal([]string{"123456789012"}, offer1.Info.BuyerAwsAccountIds)
	})

	s.T().Run("GetHtmlContentWithTemplate", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		type TestCommitDimension struct {
			Key         string  `json:"key"`
			Description string  `json:"description"`
			Rate        float64 `json:"rate,omitempty"`
			Quantity    int     `json:"quantity"`
		}

		type TestWorkloadOffer struct {
			Name         string     `json:"name"`
			CreationTime time.Time  `json:"creationTime"`
			ExpireTime   *time.Time `json:"expireTime,omitempty"`
			EndTime      *time.Time `json:"endTime,omitempty"`
		}

		type TestNotificationEvent struct {
			Info interface{} `json:"info"`
		}

		template := ` {{if gt (len .Commits) 0}}
									<p>
										<p style="font-size: 16px; background-color: #f8f8f8; margin-top: 10px; margin-bottom: 0; padding: 5px">Commits</p>
										<table style="font-size: 12px; color: #2e2e2e" border="0" cellpadding="5">
											<tr>
												<th>Key</th>
												<th>Description</th>
												<th>Amount</th>
												<th>Quantity</th>
											</tr>
											{{range .Commits}}
												<tr>
													<td>{{ .Key }}</td>
													<td>{{ .Description }}</td>
													{{ if and .Rate (gt .Rate 0.0) }}
														<td>${{ .Rate }}</td>
													{{ else }}
														<td></td>
													{{ end }}
													<td>{{ .Quantity }}</td>
												</tr>
											{{end}}
										</table>
									</p>
									{{end}}
									
									{{if .Event.Info.EndTime}}
									<p style="font-size: 12px; margin: 0; padding-bottom: 10px; word-break: break-word">
										<strong>End Date</strong>
										<br />
										{{ .Event.Info.EndTime.Format "2006-01-02" }}
									</p>
									{{end}}`
		templateFields := map[string]interface{}{
			"Commits": []TestCommitDimension{
				{
					Key:         "test-key",
					Description: "test-description",
					Rate:        10.0,
					Quantity:    1,
				},
				{
					Key:         "test-key-2",
					Description: "test-description-2",
					Quantity:    2,
				},
			},
			"Event": TestNotificationEvent{
				Info: TestWorkloadOffer{
					EndTime: aws.Time(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)),
				},
			},
		}

		_, err := shared.GetHtmlContentWithTemplate(template, templateFields)
		assert.Nil(err)
	})

	s.T().Run("IsUniqueStringArray", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		// Valid case
		isValid := shared.IsUniqueStringArray([]string{"key1", "key2"})
		assert.True(isValid)

		// Invalid case
		isValid = shared.IsUniqueStringArray([]string{"key1", "key1"})
		assert.False(isValid)
	})

	s.T().Run("GenerateRandomSecret", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		secret1 := shared.GenerateRandomSecret(32)
		secret2 := shared.GenerateRandomSecret(32)
		assert.NotEqual(secret1, secret2)
		decoded1, err := base64.URLEncoding.DecodeString(secret1)
		assert.Nil(err)
		assert.Equal(32, len([]byte(decoded1)))
		decoded2, err := base64.URLEncoding.DecodeString(secret2)
		assert.Nil(err)
		assert.Equal(32, len([]byte(decoded2)))

		secret3 := shared.GenerateRandomSecret(16)
		decoded3, err := base64.URLEncoding.DecodeString(secret3)
		assert.Nil(err)
		assert.Equal(16, len([]byte(decoded3)))
	})

	s.T().Run("SetURLQueryParams", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		rawURL := "https://example.com"
		params := map[string]string{
			"key1": "value1",
			"key2": "value2",
		}
		result := shared.SetURLQueryParams(rawURL, params)
		assert.Equal("https://example.com?key1=value1&key2=value2", result)

		rawURL = "https://example.com/test/"
		params = map[string]string{
			"key1": "value1",
			"key2": "value2",
		}
		result = shared.SetURLQueryParams(rawURL, params)
		assert.Equal("https://example.com/test/?key1=value1&key2=value2", result)

		rawURL = "https://example.com?key1=value1"
		params = map[string]string{
			"key2": "value2",
			"key3": "value3",
		}
		result = shared.SetURLQueryParams(rawURL, params)
		assert.Equal("https://example.com?key1=value1&key2=value2&key3=value3", result)
	})

	s.T().Run("DeduplicateStrArray", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		strArrayWithDuplicates := []string{"key1", "key2", "key1", "key3", "key2"}
		result := shared.DeduplicateStrArray(strArrayWithDuplicates)
		assert.Equal([]string{"key1", "key2", "key3"}, result)

		strArrayWithoutDuplicates := []string{"key1", "key2", "key3"}
		result = shared.DeduplicateStrArray(strArrayWithoutDuplicates)
		assert.Equal([]string{"key1", "key2", "key3"}, result)
	})

	s.T().Run("Md5HashStruct", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())

		strArr := []string{"key1", "key2", "key3"}
		hash, err := shared.Md5HashStruct(strArr)
		assert.Nil(err)
		assert.Equal("57946932dbcc72e0a54fb2f62e35609b", hash)

		type TestStruct struct {
			Key         string  `json:"key"`
			Description string  `json:"description"`
			Rate        float64 `json:"rate,omitempty"`
			Quantity    int     `json:"quantity"`
		}
		testStruct := TestStruct{
			Key:         "test-key",
			Description: "test-description",
			Rate:        10.0,
			Quantity:    1,
		}
		hash, err = shared.Md5HashStruct(testStruct)
		assert.Nil(err)
		assert.Equal("7483a8fbd53c40286fb11bc5b07733f0", hash)
	})

	s.T().Run("Deserialize JSON String from null", func(t *testing.T) {
		assert := require.New(t)

		JSONStr := `
{
  "username": "John Doe",
  "address": null
}
`
		type User struct {
			Username string `json:"username"`
			Address  string `json:"address,omitempty"`
		}

		type UserWithoutEmpty struct {
			Username string `json:"username"`
			Address  string `json:"address"`
		}

		var user User
		err := json.Unmarshal([]byte(JSONStr), &user)
		assert.NoError(err)

		assert.Equal("John Doe", user.Username)
		assert.Equal("", user.Address)

		var userWithoutEmpty UserWithoutEmpty
		err = json.Unmarshal([]byte(JSONStr), &userWithoutEmpty)
		assert.NoError(err)

		assert.Equal("John Doe", userWithoutEmpty.Username)
		assert.Equal("", userWithoutEmpty.Address)
	})

	s.T().Run("DaysInMonth", func(t *testing.T) {
		t.Parallel()
		assert := require.New(t)

		// February 2021
		assert.Equal(28, shared.DaysInMonth(2021, 2))
		// February 2020
		assert.Equal(29, shared.DaysInMonth(2020, 2))
		// January 2021
		assert.Equal(31, shared.DaysInMonth(2021, 1))
		// December 2021
		assert.Equal(31, shared.DaysInMonth(2021, 12))
		// August 2024
		assert.Equal(31, shared.DaysInMonth(2024, 8))
	})
}
