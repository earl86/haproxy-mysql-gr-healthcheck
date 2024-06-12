# haproxy-mysql-gr-healthcheck

The healthcheck script for haproxy to monitor MySQL Group Replication members.

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
    default-server inter 1s rise 1 fall 1 on-marked-down shutdown-sessions

backend healthcheck_primary
    option external-check
    external-check command /opt/haproxy-mysql/haproxy-mysql-gr-healthcheck -x healthcheck_primary -n mysql_ip -p mysql_port -u mysql_user -a mysql_password
    server mysql1_srv mysql_ip:mysql_port check inter 1s fastinter 500ms rise 1 fall 2

backend healthcheck_secondary
    option external-check
    external-check command /opt/haproxy-mysql/haproxy-mysql-gr-healthcheck -x healthcheck_secondary -n mysql_ip -p mysql_port -u mysql_user -a mysql_password
    server mysql1_srv mysql_ip:mysql_port check inter 5s fastinter 500ms rise 1 fall 2
```

Replace mysql_ip mysql_port mysql_user mysql_password in haproxy.cfg.

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
```

Additional SQL schema of `sys.gr_member_routing_candidate_status` to exec gr_member_routing_candidate_status.sql on the MySQL GR primary node.
