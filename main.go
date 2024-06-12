package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	// 🐒 patching of "database/sql".
	_ "github.com/go-sql-driver/mysql"
)

const (
	query       = "SELECT * FROM sys.gr_member_routing_candidate_status"
	mysqlParams = "collation=utf8mb4_0900_ai_ci"
)

var (
	// Version version.
	Version = "DEV"

	// We pass credentials in this env var as there is no better way of doing this from haproxy.
	//mysqlCredentials   = os.Getenv("PATH")
	//haproxyBackendName = os.Getenv("HAPROXY_PROXY_NAME")
)

type eventRow struct {
	ViableCandidate    string
	ReadOnly           string
	TransactionsBehind string
	TransactionsToCert string
	MemberRole         string
	MemberState        string
}

func debugMsg(isDebug bool, msg string) {
	if isDebug {
		fmt.Println(msg)
	}
}

func main() {
	var versionFlag, debugFlag bool
	var mysql_ip, mysql_username, mysql_password, haproxyBackendName string
	var mysql_port int
	flag.BoolVar(&versionFlag, "v", false, "show version")
	flag.BoolVar(&debugFlag, "d", false, "enable debug output")
	flag.StringVar(&mysql_ip, "n", "", "mysql host ip")
	flag.IntVar(&mysql_port, "p", 3306, "mysql host port")
	flag.StringVar(&mysql_username, "u", "", "mysql username")
	flag.StringVar(&mysql_password, "a", "", "mysql password")
	flag.StringVar(&haproxyBackendName, "x", "", "haproxyBackendName")

	flag.Parse()
	if versionFlag {
		fmt.Println("Version", Version)
		os.Exit(0)
	}

	if !strings.HasSuffix(haproxyBackendName, "_primary") && !strings.HasSuffix(haproxyBackendName, "_secondary") {
		debugMsg(debugFlag, "Haproxy backend name does not end with either _primary or _secondary.")
		os.Exit(1)
	}

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/?%s", mysql_username, mysql_password, mysql_ip, mysql_port, mysqlParams))
	if err != nil {
		fmt.Println("Error connecting to MySQL", err)
		os.Exit(1)
	}

	//| viable_candidate | read_only | transactions_behind | transactions_to_cert | member_role | member_state |
	rows, err := db.Query(query)
	if err != nil {
		fmt.Println("Error selecting from MySQL table:", err)
		os.Exit(1)
	}

	var row eventRow
	for rows.Next() {
		rows.Scan(&row.ViableCandidate, &row.ReadOnly, &row.TransactionsBehind, &row.TransactionsToCert, &row.MemberRole, &row.MemberState)
		debugMsg(debugFlag, fmt.Sprintf("MySQL query result: %+v\n", row))
		break
	}

	if row.ViableCandidate != "YES" {
		debugMsg(debugFlag, "GR member is not viable candidate.")
		os.Exit(1)
	}

	if strings.HasSuffix(haproxyBackendName, "_primary") && row.ReadOnly == "NO" && row.MemberRole == "PRIMARY" && row.MemberState == "ONLINE" {
		debugMsg(debugFlag, "HEALTHCHECK PRIMARY - OK")
		return
	} else if strings.HasSuffix(haproxyBackendName, "_secondary") && row.ReadOnly == "YES" && row.MemberRole == "SECONDARY" && row.MemberState == "ONLINE" {
		debugMsg(debugFlag, "HEALTHCHECK SECONDARY - OK")
		return
	}

	debugMsg(debugFlag, "HEALTHCHECK - NOT OK")
	os.Exit(1)
}
