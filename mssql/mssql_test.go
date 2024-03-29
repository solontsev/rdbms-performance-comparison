package mssql

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"regexp"
	"testing"
	"time"

	"github.com/solontsev/rdbms-performance-comparison/config"

	"github.com/docker/go-connections/nat"
	. "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/microsoft/go-mssqldb"
)

var db *sql.DB // Database connection pool.

const defaultDbName = "master"
const testDbName = "test"
const port = "1433/tcp"
const user = "SA"
const password = "myStrong(!)Password"

var dockerImages = []string{
	"mcr.microsoft.com/mssql/server:2019-CU20-ubuntu-20.04",
	"mcr.microsoft.com/mssql/server:2022-CU4-ubuntu-20.04",
}

var env = map[string]string{
	"ACCEPT_EULA":       "Y",
	"MSSQL_USER":        user,
	"MSSQL_SA_PASSWORD": password,
	"MSSQL_PID":         "Developer",
}

var dbURL = func(host string, port nat.Port) string {
	return fmt.Sprintf("sqlserver://%s:%s@%s:%s?database=%s", user, password, host, port.Port(), defaultDbName)
}

func StreamToString(stream io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.String()
}

// read all *.sql files from testdata folder
func ReadInitSqlFiles() []ContainerFile {
	var files []ContainerFile
	fileInfos, err := os.ReadDir("./testdata")
	if err != nil {
		log.Fatal(err)
	}

	for _, fileInfo := range fileInfos {
		match, _ := regexp.MatchString(".*init.*\\.sql", fileInfo.Name())

		if fileInfo.IsDir() || !match {
			continue
		}

		files = append(files, ContainerFile{
			HostFilePath:      "./testdata/" + fileInfo.Name(),
			ContainerFilePath: "/tmp/" + fileInfo.Name(),
			FileMode:          700,
		})
	}
	return files
}

func startContainer(ctx context.Context, dockerImage string, t *testing.T) (Container, string, error) {
	containerFiles := ReadInitSqlFiles()

	req := ContainerRequest{
		Image:         dockerImage,
		ImagePlatform: "linux/amd64",
		ExposedPorts:  []string{port},
		Env:           env,
		WaitingFor:    wait.ForSQL(nat.Port(port), "sqlserver", dbURL).WithStartupTimeout(config.ContainerStartupTimeout),
		//WaitingFor: wait.ForSQL(nat.Port(port), "sqlserver", dbURL).WithStartupTimeout(config.ContainerStartupTimeout).WithQuery("SELECT 1"), // custom query
		Files: containerFiles,
	}
	container, err := GenericContainer(ctx, GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		container.Terminate(ctx)
	})

	var reader io.Reader
	result, reader, err := container.Exec(ctx, []string{
		"/opt/mssql-tools/bin/sqlcmd", "-S", "localhost", "-U", user, "-P", password, "-d", "master", "-i", "/tmp/init_db.sql",
	})
	if err != nil {
		t.Fatal(err)
	}
	log.Printf("Init script(/tmp/init_db.sql) result = %d, output:\n%s\n", result, StreamToString(reader))

	mappedPort, err := container.MappedPort(ctx, nat.Port(port))
	if err != nil {
		t.Fatal(err)
	}

	connectionString := fmt.Sprintf("server=%s;user id=%s;password=%s;port=%d;database=%s;",
		"127.0.0.1", user, password, mappedPort.Int(), testDbName)

	return container, connectionString, err
}

func Ping(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
}

// FilterById the database for the information requested and prints the results.
// If the query fails exit the program with an error.
func FilterById(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var result int32
	err := db.QueryRowContext(ctx, "select count(*) from dbo.test_table as p where status_id = @status_id;", sql.Named("status_id", 1)).Scan(&result)
	if err != nil {
		log.Fatal("unable to execute search query", err)
	}
	//log.Println("result = ", result)
}

func FilterByName(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var result int32
	err := db.QueryRowContext(ctx, "select count(*) from dbo.test_table as p where status = @status;", sql.Named("status", "active")).Scan(&result)
	if err != nil {
		log.Fatal("unable to execute search query", err)
	}
	//log.Println("result = ", result)
}

func TestContainerWithWaitForSQL(t *testing.T) {

	ctx, stop := context.WithCancel(context.Background())
	defer stop()

	appSignal := make(chan os.Signal, 3)
	signal.Notify(appSignal, os.Interrupt)

	go func() {
		<-appSignal
		stop()
	}()

	data := []struct {
		name       string
		initScript string
		f          func(context.Context)
	}{
		{"q0", "q0_init.sql", FilterById},
		{"q1", "q1_init.sql", FilterById},
		{"q2", "q2_init.sql", FilterByName},
		{"q3", "q3_init.sql", FilterByName},
		{"q4", "q4_init.sql", FilterByName},
	}

	result := make(map[string]string, len(data))

	for _, dockerImage := range dockerImages {
		container, dbConnectionString, err := startContainer(ctx, dockerImage, t)
		if err != nil {
			t.Fatal(err)
		}

		db, err = sql.Open("sqlserver", dbConnectionString)
		if err != nil {
			log.Fatal("Error creating connection: ", err.Error())
		}

		for _, d := range data {
			initScriptContainerPath := fmt.Sprintf("/tmp/%s", d.initScript)
			execResult, reader, err := container.Exec(ctx, []string{
				"/opt/mssql-tools/bin/sqlcmd", "-S", "localhost", "-U", user, "-P", password, "-d", "master", "-i", initScriptContainerPath,
			})
			if err != nil {
				t.Fatal(err)
			}
			log.Printf("Init script(%s) result = %d, output:\n%s\n", initScriptContainerPath, execResult, StreamToString(reader))

			t.Run(d.name, func(t *testing.T) {
				log.Printf("Starting test %s on image %s...", d.name, dockerImage)

				db.SetConnMaxLifetime(0)
				db.SetMaxIdleConns(3)
				db.SetMaxOpenConns(3)

				Ping(ctx)

				for i := 0; i < config.WarmUpExecutions; i++ {
					d.f(ctx)
				}

				start := time.Now()

				for i := 0; i < config.TestExecutions; i++ {
					d.f(ctx)
				}

				elapsed := time.Since(start)

				key := fmt.Sprintf("%s - %s", dockerImage, d.name)
				result[key] = fmt.Sprintf("%s", elapsed/time.Duration(config.TestExecutions))
			})
		}

		db.Close()
	}

	prettyResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Println("error:", err)
	}

	log.Println(string(prettyResult))
}
