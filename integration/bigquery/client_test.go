package bigquery_test

// Command to run this test only
// go test -v shared/bigquery/init_test.go shared/bigquery/client_test.go

import (
	"context"
	"fmt"
	"testing"

	_ "embed"

	"cloud.google.com/go/bigquery"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	sharedBigquery "github.com/sugerio/workflow-service-trial/integration/bigquery"
	"github.com/sugerio/workflow-service-trial/shared/structs"
	"google.golang.org/api/iterator"
)

type ClientTestSuite struct {
	suite.Suite
}

type Item struct {
	Name  string
	Size  float64
	Count int
}

type DataSaver struct {
	data map[string]interface{}
}

func Test_ClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (s *ClientTestSuite) Test() {
	s.T().Run("Test Bigquery Rest API", func(t *testing.T) {
		t.Parallel()
		assert := require.New(s.T())
		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create Bigquery Integration for test
		integration, err := sharedBigquery.CreateBigqueryIntegration_Testing(
			organization.ID, rdsDbQueries, awsSdkClients)
		assert.Nil(err)

		// List Projects
		result, err := sharedBigquery.ListProjects(integration)
		assert.Nil(err)
		assert.NotEmpty(result.Projects)
	})

	s.T().Run("Test Bigquery Client Get Datasets Tables", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		assert := require.New(s.T())

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create Bigquery Integration for test
		integration, err := sharedBigquery.CreateBigqueryIntegration_Testing(
			organization.ID, rdsDbQueries, awsSdkClients)
		assert.Nil(err)

		projectId := "suger-stag"
		client, err := sharedBigquery.NewBigqueryClient(integration, projectId)
		assert.Nil(err)
		assert.NotNil(client)

		// Datasets
		datasetIterator := client.Datasets(ctx)
		for {
			dataset, err := datasetIterator.Next()
			if err != nil {
				break
			}
			assert.NotNil(dataset)
		}

		// Tables
		datasetId := "suger_stag_bigquery_test"
		tableIterator := client.Dataset(datasetId).Tables(ctx)
		for {
			table, err := tableIterator.Next()
			if err != nil {
				break
			}
			assert.NotNil(table)
		}
	})

	s.T().Run("Test Bigquery Client Insert", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		assert := require.New(s.T())

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create Bigquery Integration for test
		integration, err := sharedBigquery.CreateBigqueryIntegration_Testing(
			organization.ID, rdsDbQueries, awsSdkClients)
		assert.Nil(err)

		projectId := "suger-stag"
		client, err := sharedBigquery.NewBigqueryClient(integration, projectId)
		assert.Nil(err)
		assert.NotNil(client)

		datasetId := "suger_stag_bigquery_test"
		tableId := "simple-test-table"

		// Insert Data
		ins := client.Dataset(datasetId).Table(tableId).Inserter()
		ins.SkipInvalidRows = true
		ins.IgnoreUnknownValues = true

		// Insert use Struct, Schema is inferred from the score type.
		items1 := &Item{
			Name:  "n0",
			Size:  11.2,
			Count: 2,
		}
		err = ins.Put(ctx, items1)
		assert.Nil(err, "insert data error", items1)
		// Insert use Struct, Schema is inferred from the score type.
		items2 := []*Item{
			{Name: "n4", Size: 32.2, Count: 7},
			{Name: "n5", Size: 4.2, Count: 2},
			{Name: "n6", Size: 101.2, Count: 1},
		}
		err = ins.Put(ctx, items2)
		assert.Nil(err, "insert data error", items2)
		// Insert use StructSaver
		// Assume schema holds the table's schema.
		schema := []*bigquery.FieldSchema{
			{Name: "Name", Repeated: false, Required: true, Type: "STRING"},
			{Name: "Size", Repeated: false, Required: true, Type: "FLOAT64"},
			{Name: "Count", Repeated: false, Required: true, Type: "INTEGER"},
		}
		savers := []*bigquery.StructSaver{
			{Struct: Item{Name: "n7", Size: 12.1, Count: 1}, Schema: schema, InsertID: "id1"},
			{Struct: Item{Name: "n8", Size: 31.1, Count: 1}, Schema: schema, InsertID: "id2"},
			{Struct: Item{Name: "n9", Size: 7.1, Count: 1}, Schema: schema, InsertID: "id3"},
		}
		err = ins.Put(ctx, savers)
		assert.Nil(err, "insert data error", savers)
		// Insert use ValueSaver
		var vss []*bigquery.ValuesSaver
		for i, name := range []string{"n10", "n11", "n12"} {
			// Assume schema holds the table's schema.
			vss = append(vss, &bigquery.ValuesSaver{
				Schema:   schema,
				InsertID: name,
				Row:      []bigquery.Value{name, 2.2 * float64(i), int64(i)},
			})
		}
		err = ins.Put(ctx, vss)
		assert.Nil(err, "insert data error", vss)

		// Insert use wrap of map[string]interface{}
		datas := []map[string]interface{}{
			{"Name": "n13", "Size": 13.1, "Count": 13},
		}

		bqDataSavers := make([]bigquery.ValueSaver, len(datas))
		for idx, mapVal := range datas {
			bqDataSavers[idx] = &DataSaver{data: mapVal}
		}

		err = ins.Put(ctx, bqDataSavers)
		assert.Nil(err, "insert data error", bqDataSavers)
	})

	s.T().Run("Test Bigquery Client Query", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		assert := require.New(s.T())

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create Bigquery Integration for test
		integration, err := sharedBigquery.CreateBigqueryIntegration_Testing(
			organization.ID, rdsDbQueries, awsSdkClients)
		assert.Nil(err)

		projectId := "suger-stag"
		client, err := sharedBigquery.NewBigqueryClient(integration, projectId)
		assert.Nil(err)
		assert.NotNil(client)

		// Query Public Dataset
		q := client.Query(
			"SELECT * FROM `bigquery-public-data.usa_names.usa_1910_2013` " +
				"WHERE state = \"TX\" " +
				"LIMIT 10")
		q.Location = "US"
		job, err := q.Run(ctx)
		assert.Nil(err)
		status, err := job.Wait(ctx)
		assert.Nil(err)
		err = status.Err()
		assert.Nil(err)
		queryIt, err := job.Read(ctx)
		assert.Nil(err)

		var result []bigquery.Value
		for {
			var row []bigquery.Value
			err := queryIt.Next(&row)
			if err == iterator.Done {
				break
			}
			if err != nil {
				fmt.Println(err)
			}
			result = append(result, row)
		}
		assert.Equal(10, len(result))

		datasetId := "suger_stag_bigquery_test"
		tableId := "simple-test-table"

		// Query Suger Test Table
		q1 := client.Query(fmt.Sprintf("SELECT * FROM `%s.%s.%s` WHERE Name = \"test001\" LIMIT 20", projectId, datasetId, tableId))
		queryIt, err = q1.Read(ctx)
		assert.Nil(err)
		var result1 []bigquery.Value
		for {
			var row []bigquery.Value
			err := queryIt.Next(&row)
			if err == iterator.Done {
				break
			}
			if err != nil {
				fmt.Println(err)
			}
			result1 = append(result1, row)
		}
		assert.NotEmpty(result1)
		// shared.PrintJson("suger test table query", result1)

		// Table Schema
		md, err := client.Dataset(datasetId).Table(tableId).Metadata(ctx)
		assert.Nil(err)
		assert.NotEmpty(md.Schema)
	})

	s.T().Run("Test Bigquery Client Create Table", func(t *testing.T) {
		t.Skip("Skip this test as it will create a new table in Bigquery.")
		ctx := context.Background()
		assert := require.New(s.T())

		// Create Organization for test
		organization := structs.CreateOrganization_Testing(rdsDbQueries, sid, "")

		// Create Bigquery Integration for test
		integration, err := sharedBigquery.CreateBigqueryIntegration_Testing(
			organization.ID, rdsDbQueries, awsSdkClients)
		assert.Nil(err)

		projectId := "suger-stag"
		client, err := sharedBigquery.NewBigqueryClient(integration, projectId)
		assert.Nil(err)
		assert.NotNil(client)

		datasetId := "suger_stag_bigquery_test"

		// Create Table
		schema, err := bigquery.InferSchema(Item{})
		assert.Nil(err)
		table := client.Dataset(datasetId).Table("simple-test-table")
		err = table.Create(ctx,
			&bigquery.TableMetadata{
				Name:   "Simple Test Table",
				Schema: schema,
				// ExpirationTime: time.Now().Add(365 * 24 * time.Hour),
			})
		assert.Nil(err)
	})
}

func (dsv *DataSaver) Save() (map[string]bigquery.Value, string, error) {
	row := make(map[string]bigquery.Value, len(dsv.data))
	for key, val := range dsv.data {
		row[key] = val
	}
	// fmt.Println("dsv::Save", "inserting", row)
	return row, "", nil
}
