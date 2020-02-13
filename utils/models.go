package utils

import (
	"fmt"
)

// ParseTask struct for encoding/decoding data shared between dummy producer and parser
type ParseTask struct {
	FileName string
	Data     []string
}

// CDR record
type CDR struct {
	recordType            int
	recordID              int
	startTimestamp        string
	callingPartyNumber    string
	calledPartyNumber     string
	redirectingNumber     string
	callIDNumber          int
	supplementaryServices string
	cause                 int
	callingPartyCategory  int
	callDuration          int
	callStatus            int
	connectedNumber       string
	imsiCalling           string
	imeiCalling           string
	imsiCalled            string
	imeiCalled            string
	msisdnCalling         string
	msisdnCalled          string
	mscNumber             string
	vlrNumber             int
	locationLac           int
	locationCell          int
	forwardingReason      int
	roamingNumber         string
	ssCode                string
	ussd                  string
	operatorID            int
	dateAndTime           string
	callDirection         int
	seizureTime           string
	answerTime            string
	releaseTime           string
}

// NewCdr create new instance of Cdr with default fields
func NewCdr() CDR {
	c := CDR{}
	c.recordType = -1
	c.recordID = -1
	c.startTimestamp = "-"
	c.callingPartyNumber = "-"
	c.calledPartyNumber = "-"
	c.redirectingNumber = "-"
	c.callIDNumber = -1
	c.supplementaryServices = "-"
	c.cause = -1
	c.callingPartyCategory = -1
	c.callDuration = -1
	c.callStatus = -1
	c.connectedNumber = "-"
	c.imsiCalling = "-"
	c.imeiCalling = "-"
	c.imsiCalled = "-"
	c.imeiCalled = "-"
	c.msisdnCalling = "-"
	c.msisdnCalled = "-"
	c.mscNumber = "-"
	c.vlrNumber = -1
	c.locationLac = -1
	c.locationCell = -1
	c.forwardingReason = -1
	c.roamingNumber = "-"
	c.ssCode = "-"
	c.ussd = "-"
	c.operatorID = -1
	c.dateAndTime = "-"
	c.callDirection = -1
	c.seizureTime = "-"
	c.answerTime = "-"
	c.releaseTime = "-"
	return c
}

func (c CDR) csvRow() string {
	return fmt.Sprintf("%d,%d,%s,%s,%s,%s,%d,%s,%d,%d,%d,%d,%s,%s,%s,%s,%s,%s,%s,%s,%d,%d,%d,%d,%s,%s,%s,%d,%s,%d,%s,%s,%s", c.recordType, c.recordID,
		c.startTimestamp, c.callingPartyNumber, c.calledPartyNumber, c.redirectingNumber, c.callIDNumber, c.supplementaryServices, c.cause, c.callingPartyCategory,
		c.callDuration, c.callStatus, c.connectedNumber, c.imsiCalling, c.imeiCalling, c.imsiCalled, c.imeiCalled, c.msisdnCalling, c.msisdnCalled,
		c.mscNumber, c.vlrNumber, c.locationLac, c.locationCell, c.forwardingReason, c.roamingNumber, c.ssCode, c.ussd, c.operatorID, c.dateAndTime, c.callDirection, c.seizureTime, c.answerTime, c.releaseTime)
}
