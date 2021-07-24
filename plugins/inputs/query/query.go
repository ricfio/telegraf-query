package query

import (
	"bytes"
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

const (
	pluginName = "query"
)

type PluginData struct {
	Server      string   `toml:"server"`
	Measurement string   `toml:"measurement"`
	Database    string   `toml:"database"`
	Query       string   `toml:"query"`
	Tags        []string `toml:"tags"`

	Log telegraf.Logger `toml:"-"`

	initialized bool
	dbExists    bool
}

var (
	defaultServer      string = "tcp(127.0.0.1:3306)/"
	defaultMeasurement string = "query_plugin"
	defaultDatabase    string = "mysql"
)

const sampleConfig = `
  ## specify mysql server connection via a url matching:
  ##  [username[:password]@][protocol[(address)]]/[?tls=[true|false|skip-verify|custom]]
  ##  see https://github.com/go-sql-driver/mysql#dsn-data-source-name
  ##  e.g.
  ##    server = ["user:passwd@tcp(127.0.0.1:3306)/?tls=false"]
  ##    server = ["user@tcp(127.0.0.1:3306)/?tls=false"]
  #
  ## If no servers are specified, then localhost is used as the host.
  # server = "tcp(127.0.0.1:3306)/"

  ## Measurement
  # measurement = "query_plugin"

  ## Metric Database (this database must exists to enable metrics collection)
  # database = "mysql"

  ## Metric Query (this query and its field aliases are used to collect the metrics)
  query = "SELECT 10 AS field_integer, 20.30 AS field_decimal, 'helloworld' AS field_string, 'tagValue1' AS tag_1, 'tagValue2' AS tag_2 FROM DUAL"

  ## Metric Tags (these query fields will be treated as tags)
  tags = ["tag_1", "tag_2"]
`

func getFunctionName() string {
	pc, _, _, _ := runtime.Caller(1)
	nameFull := runtime.FuncForPC(pc).Name()
	nameEnd := filepath.Ext(nameFull)
	name := strings.TrimPrefix(nameEnd, ".")
	return name
}

func getDataSourceName(dsn string) (string, error) {
	conf, err := mysql.ParseDSN(dsn)
	if err != nil {
		return "", err
	}

	if conf.Timeout == 0 {
		conf.Timeout = time.Second * 5
	}

	return conf.FormatDSN(), nil
}

func openDatabase(dsn string) (*sql.DB, error) {
	dataSourceName, err := getDataSourceName(dsn)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func existsDatabase(db *sql.DB, dbName string) (bool, error) {
	query := "SHOW DATABASES LIKE '" + dbName + "'"
	rows, err := db.Query(query)
	if err != nil {
		return false, fmt.Errorf("%v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}

	return count == 1, nil
}

// parseSqlValue can be used to convert values such as "ON","OFF","Yes","No" to 0,1
func parseSqlValue(value sql.RawBytes) (interface{}, bool) {
	if bytes.EqualFold(value, []byte("YES")) || bytes.Equal(value, []byte("ON")) {
		return 1, true
	}

	if bytes.EqualFold(value, []byte("NO")) || bytes.Equal(value, []byte("OFF")) {
		return 0, true
	}

	if val, err := strconv.ParseInt(string(value), 10, 64); err == nil {
		return val, true
	}
	if val, err := strconv.ParseFloat(string(value), 64); err == nil {
		return val, true
	}

	if len(string(value)) > 0 {
		return string(value), true
	}
	return nil, false
}

func getFieldsFromQuery(db *sql.DB, rows *sql.Rows, columns []string) (map[string]interface{}, error) {
	fields := make(map[string]interface{})
	values := make([]interface{}, len(columns))
	for i := range columns {
		values[i] = &sql.RawBytes{}
	}
	err := rows.Scan(values...)
	if err != nil {
		return nil, err
	}
	for i := range columns {
		key := columns[i]
		fields[key], _ = parseSqlValue(*values[i].(*sql.RawBytes))
	}
	return fields, nil
}

func useFieldAsTag(key string, fields map[string]interface{}, tags map[string]string) {
	if value, ok := fields[key]; ok {
		if value != nil {
			tags[key] = value.(string)
		}
		delete(fields, key)
	}
}

func (x *PluginData) SampleConfig() string {
	return sampleConfig
}

func (x *PluginData) Description() string {
	return "Collects metrics from a SQL query"
}

func (x *PluginData) initData(db *sql.DB) error {
	dbName := x.Database
	dbFound, err := existsDatabase(db, dbName)
	if err != nil {
		return err
	}
	x.Log.Debugf("%v - Check if database '%v' exists: %v", getFunctionName(), dbName, dbFound)

	// Inizialization stuffs here
	x.initialized = true
	x.dbExists = dbFound

	return nil
}

func (x *PluginData) gatherFromQuery(db *sql.DB, query string, measurement string, tagNames []string, acc telegraf.Accumulator) error {
	var err error = nil
	tags := map[string]string{}
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("query error: %v - %v", query, err)
	}
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	for rows.Next() {
		fields, err := getFieldsFromQuery(db, rows, columns)
		if err != nil {
			return err
		}
		if len(fields) > 0 {
			for i := range tagNames {
				tagName := tagNames[i]
				useFieldAsTag(tagName, fields, tags)
			}
			acc.AddFields(measurement, fields, tags)
		}
	}
	return nil
}

func (x *PluginData) Gather(acc telegraf.Accumulator) error {
	var err error = nil

	db, err := openDatabase(x.Server)
	if err != nil {
		return err
	}
	defer db.Close()

	if !x.initialized {
		err = x.initData(db)
		if err != nil {
			return err
		}
	}

	if x.Query != "" {
		err = x.gatherFromQuery(db, x.Query, x.Measurement, x.Tags, acc)
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	inputs.Add(pluginName, func() telegraf.Input {
		return &PluginData{
			Server:      defaultServer,
			Measurement: defaultMeasurement,
			Database:    defaultDatabase,
		}
	})
}
