package inspection

import (
	"fmt"
	"strconv"
)

func (c *clusterInspection) diagnoseFault() (*inspectionResult, error) {
	// TiKV server down down
	return nil,nil
}

func (c *clusterInspection) diagnoseTiKVServerDown() (*inspectionResult, error) {
	// TiKV server down
	prepareSQL := "set @@tidb_metric_query_step=30;set @@tidb_metric_query_range_duration=30;"
	condition := fmt.Sprintf("where time >= '%s' and time < '%s' ", c.startTime, c.endTime)
	sql := fmt.Sprintf("select  max(value)- min(value) from metrics_schema.pd_cluster_status %s and type='store_disconnected_count';",condition)
	rows, err := querySQL(c.db, prepareSQL + sql)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 || len(rows[0]) == 0 {
		return nil, nil
	}
	count, err := strconv.Atoi(rows[0][0])
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, nil
	}
	sql = fmt.Sprintf(`select t1.instance,t2.min_time from
(select instance from metrics_schema.up %[1]s and job='tikv' group by instance having max(value)-min(value)>0) as t1 join
(select instance,min(time) as min_time from metrics_schema.up %[1]s and job='tikv' and value=0 group by instance) as t2 on t1.instance=t2.instance;`,condition)
	rows, err = querySQL(c.db, prepareSQL + sql)
	if err != nil {
		return nil, err
	}
	detail := fmt.Sprintf("There is %v tikv server disconnect with pd", len(rows))
	for _,row := range rows {
		if len(row)<2 {
			continue
		}
		info := fmt.Sprintf(",\ntikv %s disconnect with prometheus around time '%s'", row[0],row[1])
		detail += info
	}
	fmt.Println(detail)
	fmt.Println()
	return nil,err
}

func (c *clusterInspection) diagnoseServerDown() (*inspectionResult, error) {
	condition := fmt.Sprintf("where time >= '%s' and time < '%s' ", c.startTime, c.endTime)
	prepareSQL := "set @@tidb_metric_query_step=30;set @@tidb_metric_query_range_duration=30;"
	sql := fmt.Sprintf(`select t1.job,t1.instance, t2.min_time from
(select instance,job from metrics_schema.up %[1]s group by instance,job having max(value)-min(value)>0) as t1 join
(select instance,min(time) as min_time from metrics_schema.up %[1]s and value=0 group by instance,job) as t2 on t1.instance=t2.instance;`,condition)
	rows, err := querySQL(c.db, prepareSQL + sql)
	if err != nil {
		return nil, err
	}
	detail := ""
	for i,row := range rows {
		if len(row)<3 {
			continue
		}
		if i > 0 {
			detail += ",\n"
		}
		info := fmt.Sprintf("%s %s disconnect with prometheus around time '%s'", row[0],row[1],row[2])
		detail += info
	}
	fmt.Println(detail)
	fmt.Println()
	return nil,err
}
