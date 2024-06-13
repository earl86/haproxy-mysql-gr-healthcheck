# haproxy-mysql-gr-healthcheck

The healthcheck script for haproxy to monitor MySQL Group Replication members.
haproxy version >=1.6

Per our test the compiled binary will produce twice less CPU load created by haproxy on doing external checks
rather than doing the same via bash script and mysql cli.
Also you don't need to add mysql cli to haproxy docker container if you are using it.

## Setup

haproxy.cfg:
```
global
    max-spread-checks 1s
    spread-checks 5
    external-check

frontend mysql-gr-front_write
    bind *:5000
    mode tcp
    default_backend healthcheck_primary

backend healthcheck_primary
    mode tcp
    balance leastconn
    option external-check
    #Sample: external-check path "mysql_user:mysql_password:mysql_checkport"
    external-check path "haproxy:haproxy:13306"
    external-check command /opt/haproxy-mysql/haproxy-mysql-gr-healthcheck
    default-server inter 3s fall 3 rise 2 on-marked-down shutdown-sessions
    #Sample: server mysql1_srv mysql_ip:mysql_port check inter 5s fastinter 500ms rise 1 fall 2
    server mysql1_srv 192.168.1.100:3306 check inter 5s fastinter 500ms rise 1 fall 2
    server mysql2_srv 192.168.1.101:3306 check inter 5s fastinter 500ms rise 1 fall 2
    server mysql3_srv 192.168.1.102:3306 check inter 5s fastinter 500ms rise 1 fall 2


frontend mysql-gr-front_read
    bind *:5001
    mode tcp
    default_backend healthcheck_secondary

backend healthcheck_secondary
    mode tcp
    balance roundrobin
    option external-check
    #Sample: external-check path "mysql_user:mysql_password:mysql_checkport"
    external-check path "haproxy:haproxy:13306"
    external-check command /opt/haproxy-mysql/haproxy-mysql-gr-healthcheck
    #Sample: server mysql1_srv mysql_ip:mysql_port check inter 5s fastinter 500ms rise 1 fall 2
    server mysql1_srv 192.168.1.100:3306 check inter 5s fastinter 500ms rise 1 fall 2
    server mysql2_srv 192.168.1.101:3306 check inter 5s fastinter 500ms rise 1 fall 2
    server mysql3_srv 192.168.1.102:3306 check inter 5s fastinter 500ms rise 1 fall 2
```

Replace mysql_ip mysql_port mysql_user mysql_password mysql_checkport in haproxy.cfg.

Backends running haproxy-mysql-gr-healthcheck should be given a name with the suffix of either
_primary or _secondary corresponding to the actual role of a Group Replication member.

MySQL user grants:
```
mysql> show grants for haproxy;
+-----------------------------------------------------------------------------+
| Grants for haproxy@%                                                        |
+-----------------------------------------------------------------------------+
| GRANT USAGE ON *.* TO `haproxy`@`%`                                         |
| GRANT SELECT ON `sys`.`gr_member_routing_candidate_status` TO `haproxy`@`%` |
+-----------------------------------------------------------------------------+
2 rows in set (0.00 sec)

Attention: If mysql_checkport is admin_port the haproxy user need SERVICE_CONNECTION_ADMIN privilege.

```

Additional SQL schema of `sys.gr_member_routing_candidate_status` to exec gr_member_routing_candidate_status.sql on the MySQL GR primary node.


Build:
```
export GO111MODULE=on
go mod tidy
go build
copy haproxy-mysql-gr-healthcheck to /opt/haproxy-mysql/
```
