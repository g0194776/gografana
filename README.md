# gografana
---

# 总体设计思路

目前市面上针对Grafana v5.x Golang版本的API并不多，而且都不好用。我们的需求是要通过client动态创建dashboard和panels。在第一个版本中`测试通过`，我基本上实现了这个思路，可以支持如下功能:

- 列举所有存在的Dashboard
- 根据FolderId来获取所有Dashboard列表
- 增加对数据源的管理 `本次更新新增`
- 根据UID获取指定Dashboard的详细信息
- 查看指定Dashboard是否存在
- 根据UID删除指定Dashboard

考虑到Grafana多版本间的API参数变化，这次代码的设计在理论上是可以支持多个Grafana版本的，主要设计点在于获取Grafana的Client是通过version来获取的，如下code:
```golang
client, err := gografana.GetClientByVersion("5.x", "http://x.x.x.x:3000", "xxxxxxxxxxxxxxxxx")
if err != nil {
  panic(err)
}
```

这看起来很妙，不是吗？通过传递远程Grafana服务期端的版本，就可以从内部生成出对应的client实例来，从而也就解决了多版本参数不兼容的问题。如下代码，是经过测试的，用于通过client来生成动态Dashboard以及Panels:

```golang
client, err := gografana.GetClientByVersion("5.x", "http://x.x.x.x:3000", "xxxxxxxxxxxxxxxxx")
if err != nil {
  panic(err)
}
title := fmt.Sprintf("DY_%s", time.Now())
existed, _, err := client.IsBoardExists(title)
if err != nil {
  panic(err)
}
if existed {
  fmt.Printf("Dashboard: %s has been existed.\n", title)
  os.Exit(0)
}
fmt.Printf("Start creating new dashboard: %s\n", title)
board, err := client.NewDashboard(&gografana.Board{
  Title:    title,
  Editable: true,
  Rows: []*gografana.Row{
    {Panels: []gografana.Panel_5_0{
      {Datasource: "Kubernetes Prod Cluster",
        DashLength:      10,
        Pointradius:     5,
        Linewidth:       1,
        SeriesOverrides: []interface{}{},
        Type:            "graph",
        Title:           "Traefik CPU Usage",
        Targets: []struct {
          Expr           string `json:"expr"`
          Format         string `json:"format"`
          Instant        bool   `json:"instant"`
          IntervalFactor int    `json:"intervalFactor"`
          LegendFormat   string `json:"legendFormat"`
          RefID          string `json:"refId"`
        }{
          {
            Expr:         "avg(sum(irate(container_cpu_usage_seconds_total{pod_name=~\"^traefik-ingress.*\"}[1h])) by (pod_name)*100) by (pod_name)",
            Format:       "time_series",
            LegendFormat: "{{pod_name}}",
            Instant:      false,
          },
        }},
    }},
  },
}, 0, false)
if err != nil {
  panic(err)
}
fmt.Printf("%#v\n", board)
fmt.Println("--- RETRIEVE IT AGAIN ---")
b, err := client.GetDashboardDetails(board.UID)
if err != nil {
  panic(err)
}
fmt.Printf("%#v\n", b)
fmt.Println("--- RETRIEVE IT END ---")
ok, err := client.DeleteDashboard(board.UID)
if err != nil {
  panic(err)
}
fmt.Printf("Dashboard deletion result: %t\n", ok)
```
目前先实现到这个级别，如果大家有别的需求也请随时给我提ISSUE。但是需要特别提出的一点是，目前我还没有保证这里面用到的数据结构是否跟官方的字段一个不落下的保持一致，也请各位使用时自己注意。
