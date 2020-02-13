package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// S alphabet
const S string = "ABCDE0123456789"

// Headers of CDR record
const Headers string = "record_type,record_id,start_timestamp,calling_party_number,called_party_number,redirecting_number,call_id_number,supplementary_services,cause,calling_party_category,call_duration,call_status,connected_number,imsi_calling,imei_calling,imsi_called,imei_called,msisdn_calling,msisdn_called,msc_number,vlr_number,location_lac,location_cell,forwarding_reason,roaming_number,ss_code,ussd,operator_id,date_and_time,call_direction,seizure_time,answer_time,release_time"

// TQueueName - test queue name for dummy producer-consumer exchange
const TQueueName string = "test"

var r = []rune(S)

// HandleError handles error
func HandleError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
	}
}

// TimeTrack returns time per function
func TimeTrack(start time.Time, name string) {
	dur := time.Since(start)
	log.Printf("%s took %s", name, dur)
}

// ReadFile reads file from path to slice of strings
func ReadFile(path string, valz *[]string) {
	file, err := os.Open(path)
	if err != nil {
		HandleError(err, fmt.Sprintf("Error opening file at path %s", path))
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		*valz = append(*valz, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		HandleError(err, fmt.Sprintf("Error scanning file at path %s", path))
	}
}

// WriteJob writes to dirname and dealocates memory from reusable var cdrs
func WriteJob(file string, dir string, cdrs *[]CDR) {
	fo, err := os.Create(fmt.Sprintf("%s/%s", dir, file))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := fo.Close(); err != nil {
			HandleError(err, fmt.Sprintf("Error while closing file with name: %s", file))
			panic(err)
		}
	}()
	w := bufio.NewWriter(fo)

	// write headers, then write contents
	w.Write([]byte(Headers + "\n"))
	for _, c := range *cdrs {
		if count, err := w.Write([]byte(c.csvRow() + "\n")); err != nil {
			HandleError(err, fmt.Sprintf("Error short write chunk with expected byte len %d got %d", len(c.csvRow())+(1<<1), count))
			panic(err)
		}
	}

	if err = w.Flush(); err != nil {
		panic(err)
	}

	*cdrs = (*cdrs)[:0]
	*(cdrs) = nil
}

// ParseJob parses slice of strings in place -> ready for writing
func ParseJob(valz *[]string, m *map[string]bool, cdrs *[]CDR) {
	chunkz := make([]string, 0, 20)
	for _, val := range *valz {
		chunkz = strings.Split(val, ",")
		c, r := chunkz[3], chunkz[4]
		// assume contains black prefix
		suf := c[3:]
		if len(c) > 11 && AreDigits(suf) && (*m)[c[:3]] {
			chunkz[3] = suf
		}

		suf = r[3:]
		if len(r) > 11 && AreDigits(suf) && (*m)[r[:3]] {
			chunkz[4] = suf
		}

		cdr := NewCdr()
		cdr.callingPartyNumber = chunkz[3]
		cdr.calledPartyNumber = chunkz[4]
		// imsi calling called?
		cdr.mscNumber = chunkz[13]
		// cdr start timestamp add
		*cdrs = append(*cdrs, cdr)
	}

	*valz = (*valz)[:0]
	*(valz) = nil
}

// RemoveContents - removes content of a dir
func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

// AreDigits - O(n) ASCI check of each rune 2 times faster than regexp.MatchString("^[0-9]*$", suf)
func AreDigits(suf string) bool {
	for _, c := range suf {
		if c < 48 || c > 57 {
			return false
		}
	}
	return true
}

// Gen  - generates all permuts of size k=3 with alphabet size - n, excluding all only digit subsets from result set
func Gen(n int, k int, res string, m *map[string]bool) {
	if k == 0 {
		if !AreDigits(res) {
			(*m)[res] = true
		}
		return
	}
	for i := 0; i < n; i++ {
		prefix := res + string(r[i])
		Gen(n, k-1, prefix, m)
	}
}
