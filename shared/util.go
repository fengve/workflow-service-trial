package shared

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-kit/log"
	"github.com/gocarina/gocsv"
	rdsDbLib "github.com/sugerio/workflow-service-trial/rds-db/lib"
	"golang.org/x/exp/slices"
)

const (
	ENV_LOCAL_DEV   = "local-dev"
	ENV_LOCAL_TEST  = "local-test"
	ENV_GITHUB_TEST = "github-test"
	ENV_PROD        = "prod"

	// AWS auth method using local SSO profile. for local testing purpose.
	AWS_AUTH_METHOD_SSO = "SSO"
	// AWS auth method using environment variables. for github testing purpose.
	AWS_AUTH_METHOD_ENV = "ENV"
	// AWS auth method using IRSA.
	AWS_AUTH_METHOD_IRSA = "IRSA"
)

// Function to get environment variable, if it is not available, then return the fallback value as default.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// Generate organization resource in the format of "organization/organizationId".
func GenerateOrgResource(orgId string) string {
	return fmt.Sprintf("organization/%s", orgId)
}

// Check whether the string array contains the value, return true if yes.
func StringArrayContains(strArr []string, value string) bool {
	for _, str := range strArr {
		if strings.EqualFold(str, value) {
			return true
		}
	}
	return false
}

// Get the domain of the input email address. For example: test@example.com -> example.com.
// Return empty string if failed.
func GetEmailDomain(email string) string {
	components := strings.Split(email, "@")
	if len(components) == 2 {
		return components[1]
	}
	return ""
}

// Exit server if any errors in the initialization process of service.
func Check(logger log.Logger, msg string, err error) {
	if err != nil {
		logger.Log("msg", msg, "error", err)
		os.Exit(1)
	}
}

// Function to check whether the user (by email) exists. If yes, return the IdentityUser.
func CheckIfUserExistsByEmail(
	email string, queries *rdsDbLib.Queries) (bool, *rdsDbLib.IdentityUser, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	user, err := queries.GetUserByEmail(ctx, email)
	if err != nil {
		if err == sql.ErrNoRows {
			// there were no rows, but otherwise no error occurred.
			return false, nil, nil
		} else {
			return false, nil, err
		}
	}

	return true, &user, nil
}

// Return a future time so it won't expire. 100 years late.
func GetHundredYearsLaterTime() time.Time {
	return time.Now().AddDate(100, 0, 0)
}

// Get the AWS SQS queue Arn from the SQS queue url
func GetAwsSqsArnFromUrl(url string) (string, error) {
	temp := strings.ReplaceAll(url, "https://", "")
	temp = strings.ReplaceAll(temp, "/", ".")
	items := strings.Split(temp, ".")
	if len(items) != 6 {
		return "", fmt.Errorf("the AWS SQS queue url %s is invalid", url)
	}

	return fmt.Sprintf("arn:aws:sqs:%s:%s:%s", items[1], items[4], items[5]), nil
}

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}

// CopyCommonFields copies src fields into dest fields. A src field is copied
// to a dest field if they have the same field name & field type.
// Dest and src must be pointers to structs.
func CopyCommonFields(src, dest interface{}) {
	srcType := reflect.TypeOf(src).Elem()
	destType := reflect.TypeOf(dest).Elem()
	destFieldsMap := map[string]int{}

	for i := 0; i < destType.NumField(); i++ {
		destFieldsMap[destType.Field(i).Name] = i
	}

	for i := 0; i < srcType.NumField(); i++ {
		if j, ok := destFieldsMap[srcType.Field(i).Name]; ok {
			if srcType.Field(i).Type == destType.Field(j).Type {
				reflect.ValueOf(dest).Elem().Field(j).Set(
					reflect.ValueOf(src).Elem().Field(i),
				)
			}
		}
	}
}

// Overwite fields in dest with non-empty fields in src.
func UpdateNonEmptyFields(dest, src interface{}) {
	destVal := reflect.ValueOf(dest).Elem()
	srcVal := reflect.ValueOf(src)

	// Iterate over the fields of the source struct
	for i := 0; i < srcVal.NumField(); i++ {
		srcField := srcVal.Field(i)
		srcFieldName := srcVal.Type().Field(i).Name

		if !srcField.IsZero() {
			// Find the corresponding field in the destination struct
			destField := destVal.FieldByName(srcFieldName)
			// Check if the field exists in the destination struct and is settable
			if destField.IsValid() && destField.CanSet() {
				destField.Set(srcField)
			}
		}
	}
}

func LogErrorWithTrace(logger log.Logger, err error) {
	// Get the RunTime code file, line & function.
	pc := make([]uintptr, 10)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	logger.Log(
		"location", fmt.Sprintf("%s:%d", frame.File, frame.Line), "msg", fmt.Sprintf("Failed to %s", frame.Function),
		"error", err,
	)
}

func MinInt64(a int64, b int64) int64 {
	if a >= b {
		return b
	} else {
		return a
	}
}

func MaxInt64(a int64, b int64) int64 {
	if a >= b {
		return a
	} else {
		return b
	}
}

// Check if the error is pq.Error of Code=23505 duplicate key error.
func IsDuplicateKeyError(err error) bool {
	return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}

// Compare two objects by converting them to json string and compare the json string.
func EqualObjects(a interface{}, b interface{}) bool {
	aJson, _ := json.Marshal(a)
	bJson, _ := json.Marshal(b)
	return string(aJson) == string(bJson)
}

// Compare two float64 values with the given precision.
func FloatsAlmostEqual(a, b float64) bool {
	return math.Abs(a-b) <= 0.0000001
}

// Parse the ISO 8601 duration syntax (such as P3M, P2DT1H), output the number of months & days.
// The time duration units under Day, such as Hour, Minute or Second, are ignored.
func ParseDuration_ISO_8601(duration string) (float64, float64, error) {
	durationRegex := regexp.MustCompile(`P((\d+(?:\.\d+)?)Y)?((\d+(?:\.\d+)?)M)?((\d+(?:\.\d+)?)D)?(T((\d+(?:\.\d+)?)H)?((\d+(?:\.\d+)?)M)?((\d+(?:\.\d+)?)S)?)?`)
	matches := durationRegex.FindStringSubmatch(duration)
	if matches == nil {
		return 0.0, 0.0, fmt.Errorf("the input string %s is of incorrect format of ISO 8601 duration", duration)
	}

	numMonths := 0.0
	if matches[2] != "" {
		years, err := strconv.ParseFloat(matches[2], 64)
		if err != nil {
			return 0.0, 0.0, err
		}
		numMonths = numMonths + years*12
	}

	if matches[4] != "" {
		months, err := strconv.ParseFloat(matches[4], 64)
		if err != nil {
			return 0.0, 0.0, err
		}
		numMonths = numMonths + months
	}

	numDays := 0.0
	if matches[6] != "" {
		days, err := strconv.ParseFloat(matches[6], 64)
		if err != nil {
			return 0.0, 0.0, err
		}
		numDays = days
	}

	numMonths = numMonths + math.Floor(numDays/30.0)
	numDays = numDays - math.Floor(numDays/30.0)*30.0

	return numMonths, numDays, nil
}

func Md5HashStruct(data interface{}) (string, error) {
	dataJson, err := json.Marshal(data)
	return fmt.Sprintf("%x", md5.Sum(dataJson)), err
}

// Return the date in format of UTC RFC3339 without hour, minute, second or millisecond, such as "2022-11-19T00:00:00.000Z"
func FormatTime_Ground_UTC_RFC3339(t time.Time) string {
	return time.Date(t.UTC().Year(), t.UTC().Month(), t.UTC().Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
}

// Return the date in UTC with zero values in hour, minute, second or millisecond.
func GroundTime_UTC(t time.Time) time.Time {
	return time.Date(t.UTC().Year(), t.UTC().Month(), t.UTC().Day(), 0, 0, 0, 0, time.UTC)
}

// Return the first day in UTC with zero values in hour, minute, second or millisecond.
// Beaware that the day is 1-indexed, not 0-indexed.
func GroundTimeMonth_UTC(t time.Time) time.Time {
	return time.Date(t.UTC().Year(), t.UTC().Month(), 1, 0, 0, 0, 0, time.UTC)
}

// Return the date in UTC with zero values in minute, second or millisecond.
func GroundTimeHour_UTC(t time.Time) time.Time {
	return time.Date(t.UTC().Year(), t.UTC().Month(), t.UTC().Day(), t.UTC().Hour(), 0, 0, 0, time.UTC)
}

// Return whether the two input time is in the same date, without considering hour, minute, second or millisecond.
func TimeEqualInDate(t1 time.Time, t2 time.Time) bool {
	t1DateStr := GroundTime_UTC(t1).Format("2006-01-02")
	t2DateStr := GroundTime_UTC(t2).Format("2006-01-02")
	return t1DateStr == t2DateStr
}

// TimeEqualInHour returns whether the two input time is the same time on hour,
// without considering minute, second or millisecond.
func TimeEqualInHour(t1 time.Time, t2 time.Time) bool {
	t1HourStr := t1.UTC().Format("2006-01-02T15")
	t2HourStr := t2.UTC().Format("2006-01-02T15")
	return t1HourStr == t2HourStr
}

// TimeEqualInMinute returns whether the two input time is the same time on minute,
// without considering second or millisecond.
func TimeEqualInMinute(t1 time.Time, t2 time.Time) bool {
	t1Minute := t1.UTC().Truncate(time.Minute)
	t2Minute := t2.UTC().Truncate(time.Minute)
	return t1Minute.Equal(t2Minute)
}

// TimeEqualInSecond returns whether the two input time is the same in epoch unix second.
func TimeEqualInSecond(t1 time.Time, t2 time.Time) bool {
	return t1.UTC().Unix() == t2.UTC().Unix()
}

// IsEpochBeginning returns whether the given time is the beginning of the epoch.
// The beginning of the epoch is defined as the time at 00:00:00 UTC on January 1, 1970.
func IsEpochBeginning(t time.Time) bool {
	return t.UTC().Unix() == 0
}

func IsLastDayOfMonth(t time.Time) bool {
	nextDay := t.AddDate(0, 0, 1)
	return t.Month() != nextDay.Month()
}

// GetCurrentTime is used to mock current time by gomonkey.
func GetCurrentTime() time.Time {
	return time.Now()
}

// Return the number of days of the given year and month.
func DaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func TruncateStr(str string, length int) string {
	if len(str) > length {
		return str[0:length]
	}
	return str
}

// Marshal the given object to json string.
func ToJsonStr(obj interface{}) string {
	objJson, _ := json.Marshal(obj)
	return string(objJson)
}

// Print the given object in json format.
func PrintJson(objName string, obj interface{}) {
	objJson := ToJsonStr(obj)
	fmt.Println(objName, objJson)
	fmt.Println("")
}

// Save the given object in json format to the given file path.
func SaveJson(filePath string, obj interface{}) error {
	objJson := ToJsonStr(obj)
	return os.WriteFile(filePath, []byte(objJson), 0644)
}

func ReadCsvFromPath[T any](path string) ([]T, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results []T
	if err = gocsv.Unmarshal(file, &results); err != nil {
		return nil, err
	}

	return results, nil

}

// Generics function to read csv from url and convert records to []struct.
func ReadCsvFromUrl[T any](url string) ([]T, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var results []T
	if err = gocsv.UnmarshalBytes(body, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// Generics function to read content from the given S3 object and convert it to struct.
func GetStructFromS3Object[T any](s3Client *s3.Client, bucketName string, objectKey string) (*T, error) {
	getObjectInput := s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}
	getObjectOutput, err := s3Client.GetObject(context.Background(), &getObjectInput)
	if err != nil {
		return nil, err
	}
	defer getObjectOutput.Body.Close()

	body, err := io.ReadAll(getObjectOutput.Body)
	if err != nil {
		return nil, err
	}

	var result T
	if err = json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Calculate the sum of the float64 array.
func SumArrayFloat64(arr []float64) float64 {
	sum := 0.0
	for _, v := range arr {
		sum += v
	}
	return sum
}

// Parse a value to float32, if the value is either float or int.
func ParseAsNumber(value interface{}) (float32, bool) {
	if value == nil {
		return 0, false
	}
	if number, ok := value.(float32); ok {
		return number, true
	}
	if number, ok := value.(float64); ok {
		return float32(number), true
	}
	if number, ok := value.(int); ok {
		return float32(number), true
	}
	return 0, false
}

// Generic function to filter the array by the given filter function.
func FilterArray[T any](arr []T, filterFunc func(T) bool) []T {
	var result []T
	for _, v := range arr {
		if filterFunc(v) {
			result = append(result, v)
		}
	}
	return result
}

// Generic function to remove duplicate items in the array.
func RemoveDuplicates[T comparable](arr []T) []T {
	var result []T
	for _, v := range arr {
		if !slices.Contains(result, v) {
			result = append(result, v)
		}
	}
	return result
}

// GetFieldFromObject returns the field value from the given object as the specified type T.
// If the field doesn't exist or the field value is not of the specified type, return error.
func GetFieldFromObject[T any](object map[string]interface{}, field string) (*T, error) {
	if value, ok := object[field]; ok {
		if value, ok := value.(T); ok {
			return &value, nil
		}
		return nil, fmt.Errorf("field %s is not of given type", field)
	}
	return nil, fmt.Errorf("field %s not found", field)
}

// GetFieldFromObjectOrNil returns the field value from the given object as the specified type T.
// If the field doesn't exist or the field value is not of the specified type, return nil.
func GetFieldFromObjectOrNil[T any](object map[string]interface{}, field string) *T {
	if object == nil {
		return nil
	}
	if value, ok := object[field]; ok {
		if value, ok := value.(T); ok {
			return &value
		}
		return nil
	}
	return nil
}

// GetTimeFieldFromObjectOrNil gets the value of the given field and converts that into a `time.Time` instance.
// If format is left blank, the default value is time.RFC3339.
// If the field doesn't exist, or we failed to parse it with RFC3339, return nil.
func GetTimeFieldFromObjectOrNil(object map[string]interface{}, field string, format string) *time.Time {
	if format == "" {
		format = time.RFC3339
	}
	s := GetFieldFromObjectOrNil[string](object, field)
	if s == nil {
		return nil
	}
	dt, err := time.Parse(format, *s)
	if err != nil {
		return nil
	}
	return &dt
}

func ParseToFloat(value interface{}) (float64, error) {
	v, ok := value.(float64)
	if ok {
		return v, nil
	}

	// Try casting to an integer
	vInt, ok := value.(int)
	if ok {
		return float64(vInt), nil
	}

	// Try casting to a string
	vStr, ok := value.(string)
	if ok {
		return strconv.ParseFloat(strings.TrimSpace(vStr), 64)
	}

	return 0, fmt.Errorf("cannot parse value to float64: %v", value)
}

// ChanToSlice consumes up a channel and stores the elements in a slice.
func ChanToSlice[T any](ch <-chan T) []T {
	result := make([]T, 0)
	for v := range ch {
		result = append(result, v)
	}
	return result
}

// ResultChanToSlice consumes up a channel of Result[T] and stores the elements in a slice.
// The consumption stops at the first error encountered.
func ResultChanToSlice[T any](ch <-chan Result[T]) ([]T, error) {
	result := make([]T, 0)
	for v := range ch {
		if v.Err != nil {
			return result, v.Err
		}
		result = append(result, *v.Value)
	}
	return result, nil
}

func NullOrEmptyStr(v interface{}) bool {
	return v == nil || v == ""
}

// Convert the given interface array to the given type array.
func ConvertInterfaceArray[T any](interfaceArray []interface{}) ([]T, error) {
	result := make([]T, 0)
	for _, item := range interfaceArray {
		if item, ok := item.(T); ok {
			result = append(result, item)
		} else {
			return nil, fmt.Errorf("failed to convert interface array to %T", item)
		}
	}
	return result, nil
}

// Convert the given interface to the given type array.
func ConvertInterfaceToArray[T any](input interface{}) ([]T, error) {
	if input == nil {
		return nil, errors.New("input interface is nil")
	}
	if v, ok := input.([]T); ok {
		return v, nil
	}
	if v, ok := input.([]interface{}); ok {
		return ConvertInterfaceArray[T](v)
	} else if reflect.TypeOf(input).Kind() == reflect.Slice || reflect.TypeOf(input).Kind() == reflect.Array {
		// json marshal & unmarshal to convert the interface to array
		inputJSON, err := json.Marshal(input)
		if err != nil {
			return nil, err
		}
		var result []T
		err = json.Unmarshal(inputJSON, &result)
		if err != nil {
			return nil, err
		}
		return result, nil
	}
	return nil, fmt.Errorf("failed to convert interface to array")
}

// Get the HTML content with HTML template and template fields.
func GetHtmlContentWithTemplate(htmlTemplate string, templateFields interface{}) (string, error) {
	tmpl, err := template.New("template").Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	err = tmpl.Execute(&result, templateFields)
	if err != nil {
		return "", err
	}

	return result.String(), nil
}

// Get the name with the next version suffix (e.g., "_v2")
func GetNameWithNextVersion(name string) string {
	// Check if the name already has a version suffix (e.g., "_v2")
	re := regexp.MustCompile(`^(.+)_v(\d+)$`)
	matches := re.FindStringSubmatch(name)

	// If a version is found, increment it
	if len(matches) == 3 {
		// Extract the base name and the version number
		baseName := matches[1]
		versionNum, err := strconv.Atoi(matches[2])
		if err != nil {
			// If the version number is not a number, append "_v2"
			return name + "_v2"
		}
		// Return the incremented version
		return fmt.Sprintf("%s_v%d", baseName, versionNum+1)
	}

	// If no version is found, append "_v2"
	return name + "_v2"
}

// SetField sets the field with the given field path (using "." to leverage the path) to the given value.
func SetField(object interface{}, fieldPath string, value interface{}) error {
	// Check if it's a nested field.
	fields := strings.Split(fieldPath, ".")
	rv := reflect.ValueOf(object)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf(
			"object must be a non-nil pointer. object: %v, fieldPath: %s, value: %v", object, fieldPath, value,
		)
	}
	rv = rv.Elem()
	for i, fieldName := range fields {
		// Get the field by name
		field := rv.FieldByName(fieldName)

		// Check if the field is found
		if !field.IsValid() {
			return fmt.Errorf("no such field: %s in obj", fieldName)
		}
		// Check if the field can be set
		if !field.CanSet() {
			return fmt.Errorf("cannot set field %s", fieldName)
		}
		if (i + 1) < len(fields) { // Not yet the last field in the path
			// Check if the nested struct is nil, and if so, initialize it
			if field.Kind() == reflect.Ptr && field.IsNil() {
				// Make sure we are dealing with a struct
				if field.Type().Elem().Kind() != reflect.Struct {
					return fmt.Errorf("%s is not a pointer to struct", fieldName)
				}
				// Create a new struct and set the pointer to it
				field.Set(reflect.New(field.Type().Elem()))
			}
			rv = field.Elem()
		} else { // Last field in the path
			if field.Kind() == reflect.Bool {
				// If it's a boolean field, we need to handle a few possible cases for the value
				switch v := value.(type) {
				case bool:
					// If it's already a boolean, we're good to set it directly
					field.Set(reflect.ValueOf(v))
				case string:
					// If it's a string, we try to interpret it as a boolean
					if v == "true" {
						field.Set(reflect.ValueOf(true))
					} else if v == "false" {
						field.Set(reflect.ValueOf(false))
					} else {
						return fmt.Errorf("invalid value for boolean field: %s", v)
					}
				default:
					// If it's another type, we have an issue
					return fmt.Errorf("provided value type %T didn't match obj field type bool", value)
				}
			} else {
				// Set the field with the provided value
				val := reflect.ValueOf(value)
				if field.Type() != val.Type() {
					return fmt.Errorf("provided value type %s didn't match obj field type %s", val.Type(), field.Type())
				}
				field.Set(val)
			}
		}
	}
	return nil
}

type Pair[T, U any] struct {
	First  T
	Second U
}

// BatchChannel Converts a channel of T instances into []T with each slice having at most batchSize elements.
func BatchChannel[T any](ctx context.Context, ch <-chan T, batchSize int) <-chan []T {
	resultCh := make(chan []T)
	go func() {
		defer close(resultCh)
		var batch []T
		for item := range ch {
			select {
			case <-ctx.Done():
				// If the context is done, stop producing batches and return
				return
			default:
				batch = append(batch, item)
				if len(batch) == batchSize {
					resultCh <- batch
					batch = nil
				}
			}
		}
		if len(batch) > 0 {
			resultCh <- batch
		}
	}()
	return resultCh
}

// BatchProcess batches the inputs into smaller task groups, and calls the worker function to process each.
// The worker function should take a task group (same type with inputs), and return a list of results.
// Eventually this function emits each result to the caller.
func BatchProcess[T any, R any](inputs []T, batchSize int, worker func([]T) []R) <-chan R {
	ch := make(chan R)
	if batchSize == 0 || len(inputs) == 0 {
		close(ch)
		return ch
	}

	numBatches := (len(inputs) + batchSize - 1) / batchSize
	go func() {
		for i := 0; i < numBatches; i++ {
			start := i * batchSize
			end := start + batchSize
			if end > len(inputs) {
				end = len(inputs)
			}
			batch := inputs[start:end]
			results := worker(batch)
			for _, result := range results {
				ch <- result
			}
		}
		close(ch)
	}()
	return ch
}

// Convert the given interface to the given type via json marshal & unmarshal.
func ConvertInterface[T any](input interface{}) (*T, error) {
	if input == nil {
		return nil, errors.New("input interface is nil")
	}
	inputJson, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	var result T
	if err = json.Unmarshal(inputJson, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Deduplicate the given string array. return the deduplicated array.
// The order of the items in the result array is the same as the input array.
// The input array is not modified.
func DeduplicateStrArray(arr []string) []string {
	result := make([]string, 0)
	for _, v := range arr {
		if !StringArrayContains(result, v) {
			result = append(result, v)
		}
	}
	return result
}

// DeepCopy makes a deep copy of the given input.
func DeepCopy[T any](input T) T {
	inputJson, _ := json.Marshal(input)
	var result T
	_ = json.Unmarshal(inputJson, &result)
	return result
}

// Result is a wrapper of an actual value and an error.
// Useful when returning a channel of items where any of them items can be an error.
// E.g., calling Salesforce API to obtain a list of records, where any of the records can potentially fail.
type Result[T any] struct {
	Value *T
	Err   error
}

func NewResult[T any](value *T) Result[T] {
	return Result[T]{Value: value, Err: nil}
}

func NewResultE[T any](err error) Result[T] {
	return Result[T]{Value: nil, Err: err}
}

// UnwrapResultChannel unwraps a channel of Result[T] to a channel of T.
// Elements from the source channel are unwrapped one by one until an error is encountered.
func UnwrapResultChannel[T any](ch <-chan Result[T]) <-chan T {
	resultCh := make(chan T)
	go func() {
		defer close(resultCh)
		for result := range ch {
			if result.Err != nil {
				break
			}
			resultCh <- *result.Value
		}
	}()
	return resultCh
}

// RemoveNullOrEmptyFields removes the fields with null or empty string values from the given map.
// A field is considered empty when:
// - it's nil
// - it's a string and its length is 0
// - it's a non-nil pointer but points to an empty string
// The operation happens in-place (i.e., the input map is modified).
// NOTE(shiman) deleting key from map while iterating over it is SAFE. See https://stackoverflow.com/a/23231539/5628717
func RemoveNullOrEmptyFields(m map[string]any) {
	for k, v := range m {
		if v == nil {
			delete(m, k)
			continue
		}
		if s, ok := v.(string); ok && s == "" {
			delete(m, k)
			continue
		}
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Ptr {
			if val.IsNil() {
				delete(m, k)
				continue
			}
			elem := val.Elem()
			if elem.Kind() == reflect.String && elem.String() == "" {
				delete(m, k)
				continue
			}
		}
	}
}

// GetDurationInMonth returns the duration in months between startTime and endTime.
func GetDurationInMonth(startTime time.Time, endTime time.Time) int64 {
	diffInYear := endTime.Year() - startTime.Year()
	diffInMonth := endTime.Month() - startTime.Month()
	diffInDay := endTime.Day() - startTime.Day()
	return int64(math.Round(float64(diffInDay)/30.0 + float64(diffInMonth) + float64(diffInYear)*12))
}

// GetDurationInMonth returns the duration (float) in months between startTime and endTime.
func GetDurationInMonth_Float(startTime time.Time, endTime time.Time) float64 {
	diffInYear := endTime.Year() - startTime.Year()
	diffInMonth := endTime.Month() - startTime.Month()
	diffInDay := endTime.Day() - startTime.Day()
	return float64(diffInDay)/30.0 + float64(diffInMonth) + float64(diffInYear)*12
}

// GetDurationInDay returns the duration in days between startTime and endTime.
func GetDurationInDay(startTime time.Time, endTime time.Time) int64 {
	diff := endTime.Sub(startTime)
	return int64(diff.Hours() / 24)
}

// Merge multiple maps into one. If there are duplicate keys, the value from the last map will be used.
func MergeMaps(maps ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// MergeStruct merges 2 structs recursively by copying non-zero values from src to dst.
func MergeStruct(dst, src interface{}) {
	dstVal := reflect.ValueOf(dst).Elem()
	srcVal := reflect.ValueOf(src).Elem()

	for i := 0; i < dstVal.NumField(); i++ {
		dstField := dstVal.Field(i)
		srcField := srcVal.Field(i)

		// Check if the source field is set or non-zero and both fields are settable
		if srcField.IsValid() && srcField.CanSet() {
			if isZeroValue(srcField) {
				continue // Skip zero value fields in source
			}

			// If the field type is struct and not a primitive type, merge recursively
			if srcField.Kind() == reflect.Struct {
				MergeStruct(dstField.Addr().Interface(), srcField.Addr().Interface())
			} else {
				dstField.Set(srcField) // Set the source value to the destination
			}
		}
	}
}

// isZeroValue is a helper function to check if a reflect.Value is zero
func isZeroValue(v reflect.Value) bool { // Safeguard to avoid panicking on unexported fields
	if !v.CanInterface() {
		return false
	}

	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !isZeroValue(v.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !isZeroValue(v.Field(i)) {
				return false
			}
		}
		return true
	default:
		// Default case for all other data types that support direct comparison
		zero := reflect.Zero(v.Type()).Interface()
		return reflect.DeepEqual(v.Interface(), zero)
	}
}

// MergeStringArrayUnique merges two string arrays and ensure that the final array contains unique elements
func MergeStringArrayUnique(arr1, arr2 []string) []string {
	uniqueMap := map[string]bool{}

	for _, elem := range arr1 {
		uniqueMap[elem] = true
	}
	for _, elem := range arr2 {
		uniqueMap[elem] = true
	}

	merged := make([]string, 0, len(uniqueMap))
	for key := range uniqueMap {
		merged = append(merged, key)
	}

	return merged
}

// Check if the string array is unique.
// Return true if all elements are unique. Otherwise, return false.
func IsUniqueStringArray(arr []string) bool {
	uniqueMap := map[string]bool{}
	for _, elem := range arr {
		if _, ok := uniqueMap[elem]; ok {
			return false
		}
		uniqueMap[elem] = true
	}
	return true
}

// Check if the current environment is local development.
func IsLocalDevEnv() bool {
	return os.Getenv("ENV") == ENV_LOCAL_DEV
}

// IsTestEnv returns true if the current environment is for testing.
func IsTestEnv() bool {
	return os.Getenv("ENV") == ENV_LOCAL_TEST || os.Getenv("ENV") == ENV_GITHUB_TEST
}

// IsLocalTestEnv returns true if the current environment is local test.
func IsLocalTestEnv() bool {
	return os.Getenv("ENV") == ENV_LOCAL_TEST
}

// IsProdEnv returns true if the current environment is for production.
func IsProductionEnv() bool {
	return os.Getenv("ENV") == ENV_PROD
}

// FormatFloat formats the float to string with precision 2.
// It is used in invoice display.
func FormatFloatByPrecision2(f float64) string {
	// Format the float with 'f' format, precision 2
	s := strconv.FormatFloat(f, 'f', 2, 64)
	return s
}

// Rounds like 12.3416 -> 12.35
func RoundCeil(val float64, precision int) float64 {
	return math.Ceil(val*(math.Pow10(precision))) / math.Pow10(precision)
}

// Rounds like 12.3496 -> 12.34
func RoundFloor(val float64, precision int) float64 {
	return math.Floor(val*(math.Pow10(precision))) / math.Pow10(precision)
}

// Rounds to nearest like 12.3456 -> 12.35
func Round(val float64, precision int) float64 {
	return math.Round(val*(math.Pow10(precision))) / math.Pow10(precision)
}

// Generate random string of given byte length using base64.URLEncoding.
// base64.URLEncoding is better than hex.EncodeToString in the token secret scenario.
func GenerateRandomSecret(byteSize int) string {
	randomBytes := make([]byte, byteSize)
	rand.Read(randomBytes)
	return base64.URLEncoding.EncodeToString(randomBytes)
}

// SetURLQueryParams sets the query parameters to the given URL.
// The query parameters are passed as a map of key-value pairs.
// The function returns the updated URL with the query parameters.
// The original input URL is not modified.
// The query parameters in the original input URL are kept or overrided in the output URL.
func SetURLQueryParams(rawURL string, queryParams map[string]string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	for key, value := range queryParams {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// GetCurrencySymbol returns the currency symbol for the given currency code.
// If the currency code is not found, return the currency code itself.
func GetCurrencySymbol(currency string) string {
	switch currency {
	case "CAD":
		return "CA$"
	case "CNY":
		return "CN¥"
	case "EUR":
		return "€"
	case "GBP":
		return "£"
	case "JPY":
		return "JP¥"
	case "USD":
		return "$"
	default:
		return currency
	}
}
