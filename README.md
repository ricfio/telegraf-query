# Query Input Plugin for Telegraf

The telegraf-query is an external telegraf input plugin to collects metrics from a sql query.  
You can use this telegraf input plugin as base for development of other external plugins.  

**NOTES**  
This plugin was implemented for professional and educational purposes to learn: telegraf, telegraf-plugin and go language.  

You can find an official telegraf input plug-in (included as a bundle in the standard telegraf installation) that supports the same features but in a much more mature and comprehensive way:  
- [SQL Input Plugin](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/sql)

## Usage

- Install telegraf  

- Build the plugin:
  ```sh
  make all
  ```

- Install the build to your system:
  ```sh
  mkdir -p /etc/telegraf/plugins/inputs/execd/query
  cp dist/* /etc/telegraf/plugins/inputs/execd/query
  chmod +x /etc/telegraf/plugins/inputs/execd/query/query
  ```

- Edit plugin configuration as needed:  
  `vi /etc/telegraf/plugins/inputs/execd/query/plugin.conf`

- Create the mysql user for the telegraf plugin (for example):
  ```sql
  CREATE USER IF NOT EXISTS telegraf IDENTIFIED BY 'GQ9TpUKDYLkkDM1eguAl'; GRANT SELECT, PROCESS, SHOW DATABASES, SUPER, REPLICATION CLIENT, SHOW VIEW, CREATE USER ON *.* TO telegraf; FLUSH PRIVILEGES;
  ```

- Test plugin execution as standalone:  
  `/etc/telegraf/plugins/inputs/execd/query/query --config /etc/telegraf/plugins/inputs/execd/query/plugin.conf`

- Add to `/etc/telegraf/telegraf.conf` or into file in `/etc/telegraf/telegraf.d/execd-query.conf`
  ```conf
  ## 
  ## Input plugins: Query
  ## 
  [[inputs.execd]]
    command = ["/etc/telegraf/plugins/inputs/execd/query/query", "--config", "/etc/telegraf/plugins/inputs/execd/query/plugin.conf"]
    signal = "none"
  ```

- Restart or reload Telegraf

## Development

```bash
go mod init github.com/ricfio/telegraf-query
go mod tidy
go mod edit -replace github.com/ricfio/telegraf-query=./
```

`make`  
```console
usage: make

TARGETS:
  all          Build all
  clean        Clean all
  install      Install plugin
  run          Run plugin
  test         Run test
```

### [Contributing an External Plugin](https://github.com/influxdata/telegraf/blob/master/CONTRIBUTING.md)

- [How to build and set up your external plugins to run with execd](https://github.com/influxdata/telegraf/blob/master/docs/EXTERNAL_PLUGINS.md#external-plugin-guidelines)
- [Input Plugins](https://github.com/influxdata/telegraf/blob/master/docs/INPUTS.md)
- [Execd Input Plugin](https://github.com/influxdata/telegraf/tree/master/plugins/inputs/execd)
- [How to integrate your plugin with the Execd Go Shim](https://github.com/influxdata/telegraf/blob/master/plugins/common/shim)

### External Input Plugin Examples

- [rand](https://github.com/ssoroka/rand)
- [OpenVPN Plugin](https://github.com/danielnelson//telegraf-execd-openvpn)
- [Amazon CloudWatch Alarms Input](https://github.com/vipinvkmenon/awsalarms)
