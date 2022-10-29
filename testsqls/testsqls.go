package testsqls

import (
	"errors"
	"fmt"
	"github.com/qaqcatz/impomysql/connector"
	"io/ioutil"
	"log"
	"path"
	"runtime"
)

// sudo docker run -itd --name test -p 13306:3306 -e MYSQL_ROOT_PASSWORD=123456 mysql
const (
	host = "127.0.0.1"
	port = 13306
	username = "root"
	password = "123456"
	dbname = "TEST"
)

// InitDBTEST:
//   DROP DATABASE IF EXISTS TEST
//   CREATE DATABASE TEST
func InitDBTEST() error {
	conn, err := connector.NewConnector(host, port, username, password, "", "")
	if err != nil {
		return err
	}
	result := conn.ExecSQL("DROP DATABASE IF EXISTS " + dbname)
	if result.Err != nil {
		return err
	}
	result = conn.ExecSQL("CREATE DATABASE " + dbname)
	if result.Err != nil {
		return err
	}
	return nil
}

// InitDBTEST:
//   CREATE DATABASE IF NOT EXISTS TEST
func EnsureDBTEST() error {
	conn, err := connector.NewConnector(host, port, username, password, "", "")
	if err != nil {
		return err
	}
	result := conn.ExecSQL("CREATE DATABASE IF NOT EXISTS TEST " + dbname)
	if result.Err != nil {
		return err
	}
	return nil
}

func GetConnector() (*connector.Connector, error) {
	conn, err := connector.NewConnector(host, port, username, password, dbname, "")
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// SQLExec: Execute the sql, print the result into standard output stream.
func SQLExec(sql string) error {
	conn, err := GetConnector()
	if err != nil {
		return err
	}
	fmt.Println("Exec SQL:", sql)
	result := conn.ExecSQL(sql)
	if result.Err != nil {
		return result.Err
	}
	fmt.Println("Exec result:", result.ToString())
	return nil
}

// SQLExecS: see connector.Connector .SQLExecS
func SQLExecS(sql string) error {
	conn, err := GetConnector()
	if err != nil {
		return err
	}
	fmt.Println("Exec SQL:", sql)
	result := conn.ExecSQLS(sql)
	if result.Err != nil {
		return result.Err
	}
	fmt.Println("Exec result:", result.ToString())
	return nil
}


// table benchmark:

// InitTableCOMPANY:
//   DROP TABLE IF EXISTS COMPANY
//   CREATE TABLE COMPANY (ID INT, NAME TEXT, AGE INT, CITY TEXT)
//   INSERT INTO COMPANY VALUES
//   (1, 'A', 18, 'a'), (2, 'B', 19, 'b'), (3, 'C', 20, 'c'),
//   (4, 'A', 19, 'c'), (5, 'A', 19, 'c'), (6, 'B', 18, 'b')
func InitTableCOMPANY() error {
	conn, err := GetConnector()
	if err != nil {
		return err
	}
	result := conn.ExecSQL("DROP TABLE IF EXISTS COMPANY")
	if result.Err != nil {
		return result.Err
	}
	result = conn.ExecSQL("CREATE TABLE COMPANY (ID INT, NAME TEXT, AGE INT, CITY TEXT)")
	if result.Err != nil {
		return result.Err
	}
	result = conn.ExecSQL("INSERT INTO COMPANY VALUES (1, 'A', 18, 'a'), (2, 'B', 19, 'b'), " +
		"(3, 'C', 20, 'c'), (4, 'A', 19, 'c'), (5, 'A', 19, 'c'), (6, 'B', 18, 'b')")
	if result.Err != nil {
		return result.Err
	}
	return nil
}

// sql benchmark:
const (
	SQLAGG = "SELECT S, G, CITY FROM ( " +
		"   SELECT SUM(ID+1) AS S, GROUP_CONCAT(NAME ORDER BY NAME DESC) AS G, CITY " +
		"   FROM COMPANY " +
		"   GROUP BY CITY " +
		"   HAVING COUNT(DISTINCT AGE) >= 1 " +
		") AS T " +
		"WHERE T.S > 0;"
	SQLWindow = "SELECT " +
		"   ID AS id, CITY, AGE, " +
		"   SUM(AGE) OVER w " +
		"   AS sum_age, " +
		"   AVG(AGE) OVER (PARTITION BY CITY ORDER BY ID ROWS BETWEEN 1 PRECEDING AND 1 FOLLOWING) " +
		"   AS avg_age, " +
		"   ROW_NUMBER() OVER (PARTITION BY CITY ORDER BY ID) " +
		"   AS rn " +
		"   FROM COMPANY " +
		"   WINDOW w AS (PARTITION BY CITY ORDER BY ID ROWS UNBOUNDED PRECEDING)"
	SQLSelectValue = "SELECT 1"
	SQLSelectValue2 = "SELECT 1.0001"
	SQLSelectValue3 = "SELECT 'a'"
	SQLSubQuery = "SELECT * FROM COMPANY WHERE ID = (SELECT ID FROM COMPANY WHERE ID = 1)"
	SQLSubQuery2 = "SELECT * FROM COMPANY WHERE ID = ANY (SELECT ID FROM COMPANY WHERE ID > 1)"
	SQLSubQuery3 = "SELECT * FROM COMPANY WHERE ID NOT IN (SELECT ID FROM COMPANY WHERE ID IN (1, 2))"
	SQLSubQuery4 = "SELECT * FROM COMPANY WHERE ID > ALL (SELECT ID FROM COMPANY WHERE ID <= 1)"
	SQLSubQuery5 = "SELECT * FROM COMPANY WHERE NOT EXISTS (SELECT ID FROM COMPANY WHERE ID <= 1)"
	// In MySQL, JOIN, CROSS JOIN, and INNER JOIN are syntactic equivalents (they can replace each other).
	SQLJOIN = "SELECT * FROM COMPANY JOIN (SELECT * FROM COMPANY WHERE ID = 1) AS T1 ON COMPANY.ID > T1.ID"
	SQLJOIN2 = "SELECT * FROM COMPANY LEFT OUTER JOIN (SELECT * FROM COMPANY WHERE ID = 2) AS T1 ON COMPANY.ID > T1.ID"
	SQLJOIN3 = "SELECT * FROM COMPANY RIGHT OUTER JOIN (SELECT * FROM COMPANY WHERE ID = 2) AS T1 ON COMPANY.ID > T1.ID"
	SQLJOIN4 = "SELECT * FROM COMPANY STRAIGHT_JOIN (SELECT * FROM COMPANY WHERE ID = 2) AS T1 ON COMPANY.ID > T1.ID"
	SQLJOIN5 = "SELECT * FROM COMPANY NATURAL JOIN (SELECT * FROM COMPANY WHERE ID = 2) AS T1"
	SQLJOIN6 = "SELECT * FROM COMPANY NATURAL LEFT JOIN (SELECT * FROM COMPANY WHERE ID = 2) AS T1"
	SQLLIMIT = "SELECT * FROM COMPANY LIMIT 2147483647,1"
	SQLLIMIT2 = "SELECT * FROM COMPANY LIMIT 1"
	SQLUNION = "SELECT * FROM COMPANY UNION ALL SELECT * FROM (SELECT * FROM COMPANY UNION SELECT * FROM COMPANY) AS T1"
	SQLUNION2 = "(SELECT * FROM COMPANY UNION ALL SELECT * FROM (SELECT * FROM COMPANY UNION SELECT * FROM COMPANY) AS T1) " +
		"UNION ALL " +
		"SELECT * FROM (SELECT * FROM COMPANY UNION SELECT * FROM COMPANY) AS T1"
	SQLWITH = "WITH XX AS (SELECT * FROM COMPANY) SELECT * FROM XX"
	SQLWITH2 = "WITH RECURSIVE fibonacci (n, fib_n, next_fib_n) AS " +
		"(SELECT 1, 0, 1 UNION ALL SELECT n + 1, next_fib_n, fib_n + next_fib_n " +
		"FROM fibonacci WHERE n < 10 ) SELECT * FROM fibonacci"
	SQLIN = "SELECT 1 IN (1, 2, 3)"
	SQLIN2 = "SELECT * FROM COMPANY WHERE ID IN (1, 2, 3)"
	SQLWHERE = "SELECT * FROM COMPANY WHERE TRUE"
	SQLHAVING = "SELECT * FROM COMPANY HAVING TRUE"
	SQLLIKE = "SELECT * FROM COMPANY WHERE 'abc' NOT LIKE 'A_%' ESCAPE '_'"
	SQLRegExp = "SELECT * FROM COMPANY WHERE 'abc' NOT REGEXP '^A[B]*C$'"
	SQLBetween = "SELECT * FROM COMPANY WHERE ID BETWEEN 1 AND 3"
	SQLBetween2 = "SELECT * FROM COMPANY WHERE ID BETWEEN '1' AND '3'"
	SQLBetween3 = "SELECT * FROM COMPANY WHERE NAME BETWEEN 0 AND 'A'"
)

// sql file benchmark:

// getPackagePath: get the package actual path, then you can read files under the path.
func getPackagePath() (string, error) {
	if _, file, _, ok := runtime.Caller(0); !ok {
		return "", errors.New("PackagePath: runtime.Caller(0) error ")
	} else {
		return path.Join(file, "../"), nil
	}
}

const (
	SQLFileQuote = "quote"
	SQLFileTest = "test"
	SQLFileAgg = "agg"
	SQLFileWindow = "window"
)

// ReadSQLFile: read the sql file under testsqls with the help of runtime.Caller().
//
// The third return value is the absolute filepath,
// you can use it to get the actual location of the file
func ReadSQLFile(sqlFileName string) ([]byte, error, string) {
	sqlFileName += ".sql"
	packagePath, err := getPackagePath()
	if err != nil {
		return nil, errors.New("ReadSQLFile: getPackagePath() error "), ""
	}
	sqlFilePath := path.Join(packagePath, sqlFileName)
	data, err := ioutil.ReadFile(sqlFilePath)
	if err != nil {
		return nil, errors.New("ReadSQLFile: read " + sqlFilePath + " error: " + err.Error()), ""
	}
	return data, nil, sqlFilePath
}

// test .zz, .yy

const (
	zzTest = "test.zz.lua"
	yyTest = "test.yy"
)

func GetTestZZPath() string {
	packagePath, err := getPackagePath()
	if err != nil {
		log.Fatal(err)
	}
	return path.Join(packagePath, zzTest)
}

func GetTestYYPath() string {
	packagePath, err := getPackagePath()
	if err != nil {
		log.Fatal(err)
	}
	return path.Join(packagePath, yyTest)
}
