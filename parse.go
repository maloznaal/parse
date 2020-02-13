package main

import (
	"bufio"
	"compress/flate"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"offline_parser/utils"
	"os"
	"path/filepath"
	"time"

	_ "github.com/lib/pq"
	"github.com/mholt/archiver/v3"
)

// initialize byte slice instead of valz string ?
var valz []string
var cdrs []utils.CDR
var zipNames []string

// postgres cred-s for iTest?
// time spent on each batch insert?

var m = map[string]bool{}

// const paths
const (
	tmpDirName   = "/app/tmp"
	tmpDirPath   = "/app/tmp/"
	cleanZipPath = "/app/czips/"
	dirtyZipPath = "/app/zips/"
	gpfdistAddr  = "gpfdist:8888"
	itestGpWait  = 20
)

func init() {
	utils.Gen(len(utils.S), 3, "", &m)
}

func connectDb() *sql.DB {
	// wait untill gp db ready to accept connections for iTest
	time.Sleep(itestGpWait * time.Second)
	conStr := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"master", 5432, "gpadmin", "greenplum", "gpadmin")
	fmt.Println("con url!", conStr)
	db, err := sql.Open("postgres", conStr)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		utils.HandleError(err, "No healty connection to db")
		panic(err)
	}

	fmt.Println("Successfully connected!")
	return db
}

func createStatement() string {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	batch := r.Intn((1 << 31) - 1)

	fName := "*.gz"
	rez := fmt.Sprintf("CREATE EXTERNAL TABLE ext_batch%d ( LIKE cdr_temporary ) ", batch) +
		fmt.Sprintf("LOCATION ('gpfdist://%s/%s')", gpfdistAddr, fName) +
		"FORMAT 'CSV' (HEADER FORCE NOT NULL start_timestamp, called_party_number)" +
		"LOG ERRORS SEGMENT REJECT LIMIT 10 ROWS;" +
		fmt.Sprintf("INSERT INTO cdr_temporary SELECT * FROM ext_batch%d;", batch) +
		fmt.Sprintf("DROP EXTERNAL TABLE IF EXISTS ext_batch%d;", batch)
	return rez
}

func main() {

	// get db connection
	db := connectDb()
	defer db.Close()

	stopChan := make(chan bool)
	z := archiver.NewTarGz()
	z.CompressionLevel = flate.DefaultCompression
	z.SingleThreaded = false

	err := z.Create(os.Stdout)

	if err != nil {
		utils.HandleError(err, fmt.Sprintf("Err creating archiver with stdout writer"))
	}
	defer z.Close()

	files, err := ioutil.ReadDir(dirtyZipPath)
	if err != nil {
		utils.HandleError(err, fmt.Sprintf("Error reading dir %s check permission granted", dirtyZipPath))
		panic(err)
	}

	for _, f := range files {
		if !f.IsDir() {
			zipNames = append(zipNames, f.Name())
		}
	}

	// making transaction
	rez := createStatement()
	fmt.Println("final query for batch - ", rez)

	txn, err := db.Begin()
	if err != nil {
		utils.HandleError(err, "database is closed while starting transaction")
		panic(err)
	}
	fmt.Println(txn)

	/*
		("CREATE EXTERNAL TABLE ext_batch%1$s ( LIKE cdr_temporary ) " +
				"LOCATION ('gpfdist://%2$s/%3$s') " +
				"FORMAT 'CSV' (HEADER FORCE NOT NULL start_timestamp, called_party_number)" +
				"LOG ERRORS SEGMENT REJECT LIMIT 10 ROWS; " +
				"" +
				"INSERT INTO cdr_temporary SELECT * FROM ext_batch%1$s; " +
				"DROP EXTERNAL TABLE IF EXISTS ext_batch%1$s;", suffix, feederDirectory, filename);
	*/
	// table should be passed as env argument

	scanner := bufio.NewScanner(nil)
	for _, name := range zipNames {
		// walkFn for each zip
		err := z.Walk(dirtyZipPath+name, func(f archiver.File) error {
			if f.IsDir() {
				return nil
			}
			scanner = bufio.NewScanner(f)
			for scanner.Scan() {
				valz = append(valz, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				msg := "Error scanning file with name %s and size %d"
				utils.HandleError(err, fmt.Sprintf(msg, f.Name(), f.Size()))
				return errors.New(msg)
			}
			utils.ParseJob(&valz, &m, &cdrs)
			utils.WriteJob(f.Name(), tmpDirName, &cdrs)
			return nil
		})

		if err != nil {
			utils.HandleError(err, fmt.Sprintf("Err occured while walking file %s", err))
		}

		files, err := filepath.Glob(fmt.Sprintf("%s*", tmpDirPath))

		if err != nil {
			utils.HandleError(err, fmt.Sprintf("Err while reading contents of %s dir", tmpDirPath))
		}

		// produce clean zip at /czips dir
		z.Archive(files, fmt.Sprintf("%s%s", cleanZipPath, name))

		// flush /tmp dir
		if ok := utils.RemoveContents(tmpDirPath); ok != nil {
			utils.HandleError(ok, fmt.Sprintf("Error removing contents at path %s", tmpDirPath))
		}

		/*
					CREATE EXTERNAL TABLE ext_batch{batch_num} ( LIKE cdr )
			            LOCATION ('gpfdist://{gpfdist_addr}/{gz_filename}')
			            FORMAT 'CSV' (HEADER FORCE NOT NULL start_timestamp, called_party_number)
			            LOG ERRORS SEGMENT REJECT LIMIT 10 ROWS;

			            INSERT INTO {table_name} SELECT * FROM ext_batch{batch_num};

			            DROP EXTERNAL TABLE IF EXISTS ext_batch{batch_num};
		*/

		//stmt, err := txn.Prepare(pq.CopyIn("users", "name", "age"))

		// if err != nil {
		// 	log.Fatal(err)
		// }
		// _, err = stmt.Exec(user.Name, int64(user.Age))
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// _, err = stmt.Exec()
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// err = stmt.Close()
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// err = txn.Commit()
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// // flush /czips dir, but wait gpfdist complete!
		// if ok := utils.RemoveContents(cleanZipPath); ok != nil {
		// 	utils.HandleError(ok, fmt.Sprintf("Error removing contents at path %s", cleanZipPath))
		// }

		log.Println("flushed contents of a pure_csv dir!, ready for next job")
	}

	// blocking
	<-stopChan
}
