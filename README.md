# haproxy-mysql-gr-healthcheck

The healthcheck script for haproxy to monitor MySQL Group Replication members.

```
Require Haproxy version >= 1.6, MySQL version >= 8.0.17
This tested Haproxy 1.8, MySQL 8.0.25
```

Per our test the compiled binary will produce twice less CPU load created by haproxy on doing external checks
rather than doing the same via bash script and mysql cli.
Also you don't need to add mysql cli to haproxy docker container if you are using it.

## Setup

haproxy.cfg:
```
global
    # nbproc Deprecated and removed in HAProxy version 2.5. Per nbproc can increase the 2000 connect sessions.
    nbproc 16
    # nbthread Recommended use in HAProxy version 2.8
    #nbthread 16
    user haproxy
    group haproxy
    stats socket /var/run/haproxy.sock mode 666 level admin
    maxconn 100000
    max-spread-checks 1s
    spread-checks 5
    external-check


defaults
    mode tcp
    timeout connect 30s
    timeout client 3600s
    timeout server 3600s


frontend mysql-gr-front_write
    bind *:5000
    mode tcp
    option contstats
    option dontlognull
    option clitcpka
    default_backend healthcheck_primary

backend healthcheck_primary
    mode tcp
    balance leastconn
    option external-check
    #Sample: external-check path "mysql_user:mysql_password:mysql_checkport"
    external-check path "haproxy:haproxy:13306"
    external-check command /opt/haproxy-mysql/haproxy-mysql-gr-healthcheck
    default-server inter 5s rise 1 fall 3 on-marked-down shutdown-sessions
    #Sample: server mysql1_srv mysql_ip:mysql_port check inter 5s fastinter 500ms rise 1 fall 3
    server mysql1_srv 192.168.1.100:3306 check inter 5s fastinter 500ms rise 1 fall 3
    server mysql2_srv 192.168.1.101:3306 check inter 5s fastinter 500ms rise 1 fall 3
    server mysql3_srv 192.168.1.102:3306 check inter 5s fastinter 500ms rise 1 fall 3


frontend mysql-gr-front_read
    bind *:5001
    mode tcp
    option contstats
    option dontlognull
    option clitcpka
    default_backend healthcheck_secondary

backend healthcheck_secondary
    mode tcp
    balance roundrobin
    option external-check
    #Sample: external-check path "mysql_user:mysql_password:mysql_checkport"
    external-check path "haproxy:haproxy:13306"
    external-check command /opt/haproxy-mysql/haproxy-mysql-gr-healthcheck
    #Sample: server mysql1_srv mysql_ip:mysql_port check inter 5s fastinter 500ms rise 1 fall 3
    server mysql1_srv 192.168.1.100:3306 check inter 5s fastinter 500ms rise 1 fall 3
    server mysql2_srv 192.168.1.101:3306 check inter 5s fastinter 500ms rise 1 fall 3
    server mysql3_srv 192.168.1.102:3306 check inter 5s fastinter 500ms rise 1 fall 3
```

Replace mysql_ip mysql_port mysql_user mysql_password mysql_checkport in haproxy.cfg.

Backends running haproxy-mysql-gr-healthcheck should be given a name with the suffix of either
_primary or _secondary corresponding to the actual role of a Group Replication member.


haproxy.service:
```
haproxy1.8 in centos7:
rh-haproxy18-3.1-2.el7.x86_64
rh-haproxy18-haproxy-1.8.24-3.el7.x86_64
rh-haproxy18-runtime-3.1-2.el7.x86_64


cat /usr/lib/systemd/system/rh-haproxy18-haproxy.service
[Unit]
Description=HAProxy Load Balancer
After=network.target

[Service]
Environment="CONFIG=/etc/opt/rh/rh-haproxy18/haproxy/haproxy.cfg" "PIDFILE=/run/rh-haproxy18-haproxy.pid"
EnvironmentFile=/etc/sysconfig/rh-haproxy18-haproxy
ExecStartPre=/opt/rh/rh-haproxy18/root/usr/sbin/haproxy -f $CONFIG -c -q $OPTIONS
ExecStart=/opt/rh/rh-haproxy18/root/usr/sbin/haproxy -Ws -f $CONFIG -p $PIDFILE $OPTIONS
ExecReload=/opt/rh/rh-haproxy18/root/usr/sbin/haproxy -f $CONFIG -c -q $OPTIONS
ExecReload=/bin/kill -USR2 $MAINPID
KillMode=mixed
Type=notify
LimitNOFILE=1024000
LimitSTACK=infinity
LimitMEMLOCK=infinity
LimitCORE=infinity

[Install]
WantedBy=multi-user.target
```

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
```
mysql -h 127.0.0.1 -P 5000 -u dba -p
Enter password:
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 79344761
Server version: 8.0.25-17 GreatSQL, Release 17, Revision 4733775f703

Copyright (c) 2000, 2023, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SELECT * FROM sys.gr_member_routing_candidate_status;
+------------------+-----------+---------------------+----------------------+-------------+--------------+
| viable_candidate | read_only | transactions_behind | transactions_to_cert | member_role | member_state |
+------------------+-----------+---------------------+----------------------+-------------+--------------+
| YES              | NO        |                   0 |                    0 | PRIMARY     | ONLINE       |
+------------------+-----------+---------------------+----------------------+-------------+--------------+
1 row in set (0.01 sec)


mysql -h 127.0.0.1 -P 5001 -u dba -p
Enter password:
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 174175
Server version: 8.0.25-17 GreatSQL, Release 17, Revision 4733775f703

Copyright (c) 2000, 2023, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SELECT * FROM sys.gr_member_routing_candidate_status;
+------------------+-----------+---------------------+----------------------+-------------+--------------+
| viable_candidate | read_only | transactions_behind | transactions_to_cert | member_role | member_state |
+------------------+-----------+---------------------+----------------------+-------------+--------------+
| YES              | YES       |                   2 |                    0 | SECONDARY   | ONLINE       |
+------------------+-----------+---------------------+----------------------+-------------+--------------+
1 row in set (0.01 sec)

mysql>
```


Build:
```
export GO111MODULE=on
go mod tidy
go build
copy haproxy-mysql-gr-healthcheck to /opt/haproxy-mysql/
```


Manage:
```
yum install socat

echo "help" | socat stdio /var/run/haproxy.sock
echo "show info" | socat stdio /var/run/haproxy.sock
echo "show stat" | socat stdio /var/run/haproxy.sock


#ready/drain/maint
echo "set server healthcheck_secondary/mysql2_srv state maint" | socat stdio /var/run/haproxy.sock
echo "set server healthcheck_secondary/mysql3_srv state maint" | socat stdio /var/run/haproxy.sock

echo "set server healthcheck_secondary/mysql2_srv state ready" | socat stdio /var/run/haproxy.sock
echo "set server healthcheck_secondary/mysql3_srv state ready" | socat stdio /var/run/haproxy.sock
```
