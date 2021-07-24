package query

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	dataSourceName string = defaultServer
)

func makePluginData() *PluginData {
	return &PluginData{
		Server:      dataSourceName,
		Measurement: defaultMeasurement,
		Database:    defaultDatabase,
		Query:       "SELECT 10 AS field_integer, 20.30 AS field_decimal, 'helloworld' AS field_string, 'tagValue1' AS tag_1, 'tagValue2' AS tag_2 FROM DUAL",
		Tags:        []string{"tag_1", "tag_2"},
		Log:         testutil.Logger{},
	}
}

func TestConnectivity(t *testing.T) {
	var flagtests = []struct {
		key string
		in  string
		out bool
	}{
		{"fail", "fakeuser:password@tcp(127.0.0.1:3306)/", false},
		{"succ", dataSourceName, true},
	}
	for _, tt := range flagtests {
		t.Run(tt.in, func(t *testing.T) {
			db, err := openDatabase(tt.in)
			require.NoError(t, err)
			defer db.Close()

			exists, err := existsDatabase(db, "mysql")
			if tt.out {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			assert.Equal(t, tt.out, exists)
		})
	}
}

func TestInitializing(t *testing.T) {
	var flagtests = []struct {
		key string
		in  string
		out bool
	}{
		{"succ", "mysql", true},
		{"fail", "fakedatabase", false},
	}
	for _, tt := range flagtests {
		t.Run(tt.in, func(t *testing.T) {
			x := makePluginData()
			x.Database = tt.in

			assert.False(t, x.initialized)
			assert.False(t, x.dbExists)

			var acc testutil.Accumulator
			x.Gather(&acc)

			assert.True(t, x.initialized)
			assert.Equal(t, tt.out, x.dbExists)
		})
	}
}

func TestMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	x := makePluginData()
	x.Query = "SELECT 10 AS field_integer, 20.30 AS field_decimal, 'helloworld' AS field_string, 'tagValue1' AS tag_1, 'tagValue2' AS tag_2 FROM DUAL"
	x.Tags = []string{"tag_1", "tag_2"}

	var acc testutil.Accumulator
	err := x.Gather(&acc)
	require.NoError(t, err)

	assert.True(t, acc.HasMeasurement("query_plugin"))
	assert.True(t, acc.HasTag("query_plugin", "tag_1"))
	assert.True(t, acc.HasTag("query_plugin", "tag_2"))
	assert.True(t, acc.HasField("query_plugin", "field_integer"))
	assert.True(t, acc.HasField("query_plugin", "field_decimal"))
	assert.True(t, acc.HasField("query_plugin", "field_string"))
}
