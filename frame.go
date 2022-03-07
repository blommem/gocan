package gocan

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
)

type CANFrameType struct {
	Type      int
	Responses int
}

func (c *CANFrameType) SetResponseCount(no int) {
	c.Responses = no
}

func (c *CANFrameType) GetResponseCount() int {
	return c.Responses
}

var (
	Incoming         = CANFrameType{Type: 0, Responses: 0}
	Outgoing         = CANFrameType{Type: 1, Responses: 0}
	ResponseRequired = CANFrameType{Type: 2, Responses: 1}
)

type CANFrame interface {
	// Return frame identifier
	Identifier() uint32
	// Return frame data length
	Length() int
	// Return data of frame
	Data() []byte
	// Return type of frame
	Type() CANFrameType
	// Return fancy string version of frame
	String() string
	// Set response timeour
	SetTimeout(time.Duration)
	// Return response timeout
	Timeout() time.Duration
}

type Frame struct {
	identifier uint32
	data       []byte
	frameType  CANFrameType
	timeout    time.Duration
}

func NewFrame(identifier uint32, data []byte, frameType CANFrameType) *Frame {
	return &Frame{
		identifier: identifier,

		data:      data,
		frameType: frameType,
	}
}

func (f *Frame) Identifier() uint32 {
	return f.identifier
}

func (f *Frame) Length() int {
	return len(f.data)
}

func (f *Frame) Data() []byte {
	return f.data
}

func (f *Frame) Type() CANFrameType {
	return f.frameType
}

func (f *Frame) SetTimeout(t time.Duration) {
	f.timeout = t
}

func (f *Frame) Timeout() time.Duration {
	return f.timeout
}

var (
	yellow    = color.New(color.FgHiBlue).SprintfFunc()
	red       = color.New(color.FgRed).SprintfFunc()
	green     = color.New(color.FgGreen).SprintfFunc()
	printable = regexp.MustCompile("[^A-Za-z0-9.,!?]+")
)

func (f *Frame) String() string {
	var out strings.Builder

	switch f.frameType.Type {
	case 0:
		out.WriteString("<i> || ")
	case 1:
		out.WriteString("<o> || ")
	case 2:
		out.WriteString("<r> || ")

	}

	out.WriteString(green("0x%03X", f.identifier) + " || ")

	var hexView strings.Builder

	for i, b := range f.data {
		hexView.WriteString(fmt.Sprintf("%02X", b))
		if i != len(f.data)-1 {
			hexView.WriteString(" ")
		}
	}

	out.WriteString(fmt.Sprintf("%-23s", hexView.String()))

	out.WriteString(" || ")

	var binView strings.Builder
	for i, b := range f.data {
		binView.WriteString(fmt.Sprintf("%08b", b))
		if i != len(f.data)-1 {
			binView.WriteString(" ")
		}
	}

	out.WriteString(red(fmt.Sprintf("%-72s", binView.String())))

	out.WriteString(" || ")
	out.WriteString(yellow("%8s", printable.ReplaceAllString(string(f.data), ".")))
	return out.String()
}
