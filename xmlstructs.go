package main

import (
	"encoding/xml"
	"time"
)

type envelopeGet struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    bodyGet
}

type bodyGet struct {
	XMLName   xml.Name `xml:"Body"`
	WorkOrder workorder
}

type workorder struct {
	XMLName          xml.Name `xml:"workorder"`
	WowoUserStatus   string
	WowoFeedbackNote string
	WowoEndDate      time.Time
}

type envelopeCreate struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    bodyCreate
}

type bodyCreate struct {
	XMLName                  xml.Name `xml:"Body"`
	WorkOrderCreateSimpleKey workordercreatesimpleKey
}

type workordercreatesimpleKey struct {
	XMLName  xml.Name `xml:"workordercreatesimpleKey"`
	WowoCode int      `xml:"WowoCode"`
}

type envelopeFind struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    bodyFind
}

type bodyFind struct {
	XMLName                        xml.Name `xml:"Body"`
	RetworkorderfindfindParameters retworkorderfindfindParameters
}

type retworkorderfindfindParameters struct {
	XMLName           xml.Name `xml:"retworkorderfindfindParameters"`
	WorkorderfindList workorderfindList
}

type workorderfindList struct {
	XMLName       xml.Name        `xml:"workorderfindList"`
	Workorderfind []workorderfind `xml:"workorderfind"`
}

type workorderfind struct {
	XMLName          xml.Name `xml:"workorderfind"`
	WowoCode         string
	WowoUserStatus   string
	WowoFeedbackNote string
	WowoEndDate      string
	WowoString12     string
	WowoNumber12     string
}
